package rules

import (
	"math/rand/v2"
	"testing"

	cards "github.com/natefinch/council4/mtg/cards/l"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

// leylineHandEntry pairs a curated card's name with its opening-hand card id, so
// multi-card pregame assertions can report which Leyline failed.
type leylineHandEntry struct {
	name   string
	cardID id.ID
}

// curatedLeylineCycle lists the curated opening-hand Leyline-cycle cards under
// verification, so the pregame-entry checks below run over each one.
var curatedLeylineCycle = []struct {
	name string
	ctor func() *game.CardDef
}{
	{"Leyline Axe", cards.LeylineAxe},
	{"Leyline of Abundance", cards.LeylineOfAbundance},
	{"Leyline of Anticipation", cards.LeylineOfAnticipation},
	{"Leyline of Hope", cards.LeylineOfHope},
	{"Leyline of Lightning", cards.LeylineOfLightning},
	{"Leyline of the Meek", cards.LeylineOfTheMeek},
	{"Leyline of the Void", cards.LeylineOfTheVoid},
	{"Leyline of Vitality", cards.LeylineOfVitality},
}

// TestLeylineCycleCarriesOpeningHandMarker proves every curated definition
// carries the pregame "begin the game on the battlefield" permission (CR 103.6a)
// that #3067's opening-hand subsystem consumes.
func TestLeylineCycleCarriesOpeningHandMarker(t *testing.T) {
	for _, tc := range curatedLeylineCycle {
		t.Run(tc.name, func(t *testing.T) {
			def := tc.ctor()
			if !def.BeginsGameOnBattlefield() {
				t.Fatalf("%s: BeginsGameOnBattlefield() = false, want true", tc.name)
			}
		})
	}
}

// TestLeylineCycleEntersFromOpeningHandWhenAccepted proves each curated card,
// held in an opening hand, moves to the battlefield under its owner via the
// pregame action when the player accepts.
func TestLeylineCycleEntersFromOpeningHandWhenAccepted(t *testing.T) {
	for _, tc := range curatedLeylineCycle {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			cardID := addCardToHand(g, game.Player1, tc.ctor())

			engine.performOpeningHandBattlefieldActions(g, agentsAll(&recordingChoiceAgent{answer: []int{1}}))

			if g.Players[game.Player1].Hand.Contains(cardID) {
				t.Fatalf("%s: stayed in hand after acceptance", tc.name)
			}
			permanent := permanentForCard(g, cardID)
			if permanent == nil {
				t.Fatalf("%s: did not enter the battlefield", tc.name)
			}
			if permanent.Owner != game.Player1 || permanent.Controller != game.Player1 {
				t.Fatalf("%s: owner/controller = %v/%v, want Player1/Player1", tc.name, permanent.Owner, permanent.Controller)
			}
			entered := false
			for _, event := range g.Events {
				if event.Kind == game.EventPermanentEnteredBattlefield && event.CardID == cardID {
					entered = true
				}
			}
			if !entered {
				t.Fatalf("%s: no enters-the-battlefield event emitted", tc.name)
			}
		})
	}
}

// TestLeylineCycleDeclinedFromOpeningHandStaysInHand proves the pregame action is
// optional for each curated card: declining leaves it in hand off the battlefield.
func TestLeylineCycleDeclinedFromOpeningHandStaysInHand(t *testing.T) {
	for _, tc := range curatedLeylineCycle {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			cardID := addCardToHand(g, game.Player1, tc.ctor())

			engine.performOpeningHandBattlefieldActions(g, agentsAll(&recordingChoiceAgent{answer: []int{0}}))

			if !g.Players[game.Player1].Hand.Contains(cardID) {
				t.Fatalf("%s: left hand despite declining", tc.name)
			}
			if len(g.Battlefield) != 0 {
				t.Fatalf("%s: battlefield size = %d, want 0", tc.name, len(g.Battlefield))
			}
		})
	}
}

// TestLeylineCycleMultipleFromOneOpeningHand proves the whole curated cycle held
// in a single opening hand all begin the game on the battlefield together, while
// an ineligible card is never offered or moved.
func TestLeylineCycleMultipleFromOneOpeningHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	cardIDs := make([]leylineHandEntry, 0)
	for _, tc := range curatedLeylineCycle {
		cardIDs = append(cardIDs, leylineHandEntry{name: tc.name, cardID: addCardToHand(g, game.Player1, tc.ctor())})
	}
	vanilla := addCardToHand(g, game.Player1, vanillaHandCard("Bear"))

	agent := &recordingChoiceAgent{answer: []int{1}}
	engine.performOpeningHandBattlefieldActions(g, agentsAll(agent))

	for _, entry := range cardIDs {
		if permanentForCard(g, entry.cardID) == nil {
			t.Fatalf("%s: did not enter from a multi-Leyline opening hand", entry.name)
		}
	}
	if !g.Players[game.Player1].Hand.Contains(vanilla) {
		t.Fatal("ineligible card left the hand")
	}
	if len(agent.requests) != len(curatedLeylineCycle) {
		t.Fatalf("agent consulted %d times, want %d (one per eligible Leyline)", len(agent.requests), len(curatedLeylineCycle))
	}
}

// TestLeylineCycleEntersAcrossPlayers proves several different curated Leylines,
// one in each player's opening hand, all begin the game on the battlefield under
// their respective owners.
func TestLeylineCycleEntersAcrossPlayers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	cardByPlayer := map[game.PlayerID]leylineHandEntry{}
	for seat := range game.NumPlayers {
		player := game.PlayerID(seat)
		tc := curatedLeylineCycle[seat%len(curatedLeylineCycle)]
		cardByPlayer[player] = leylineHandEntry{name: tc.name, cardID: addCardToHand(g, player, tc.ctor())}
	}

	engine.performOpeningHandBattlefieldActions(g, agentsAll(&recordingChoiceAgent{answer: []int{1}}))

	for player, entry := range cardByPlayer {
		permanent := permanentForCard(g, entry.cardID)
		if permanent == nil {
			t.Fatalf("player %v %s: did not enter the battlefield", player, entry.name)
		}
		if permanent.Controller != player {
			t.Fatalf("player %v %s: controller = %v, want %v", player, entry.name, permanent.Controller, player)
		}
	}
}

// TestLeylineCycleGoldfishAcceptsWholeOpeningHand runs each curated card end to
// end through a goldfish game: a full opening hand of that Leyline all begins the
// game on the battlefield under the goldfish's control.
func TestLeylineCycleGoldfishAcceptsWholeOpeningHand(t *testing.T) {
	for _, tc := range curatedLeylineCycle {
		t.Run(tc.name, func(t *testing.T) {
			commander := &game.CardDef{CardFace: game.CardFace{
				Name:       "Goldfish Commander",
				Supertypes: []types.Super{types.Legendary},
				Types:      []types.Card{types.Creature},
			}}
			config := game.PlayerConfig{Name: "Goldfish", Commander: commander, Deck: repeatedCard(tc.ctor(), 99)}

			engine := NewEngine(rand.New(rand.NewPCG(7, 11)))
			g := engine.NewGoldfishGame(config)
			result := engine.RunGoldfish(g, goldfishAcceptAgent{}, 1)

			if len(result.OpeningHand) != openingHandSize {
				t.Fatalf("%s: opening hand size = %d, want %d", tc.name, len(result.OpeningHand), openingHandSize)
			}
			entered := 0
			for _, permanent := range g.Battlefield {
				card, ok := g.GetCardInstance(permanent.CardInstanceID)
				if !ok || card.Def.Name != tc.name {
					continue
				}
				entered++
				if permanent.Controller != game.Player1 {
					t.Fatalf("%s: permanent controller = %v, want Player1", tc.name, permanent.Controller)
				}
			}
			if entered != openingHandSize {
				t.Fatalf("%s: %d permanents entered from opening hand, want %d", tc.name, entered, openingHandSize)
			}
		})
	}
}
