package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// conditionalLookAtTopBattlefieldSequence is the runtime shape cardgen emits for
// the "look at the top card of your library; if it's a land card, you may put it
// onto the battlefield" building block with no trailing else clause (Into the
// Wilds, Explorer's Scope, Raiders' Karve, Mobile Homestead). A declined or
// non-matching card is left atop the library because Else is zone.None.
func conditionalLookAtTopBattlefieldSequence(cardType types.Card, tapped bool) []game.Instruction {
	looked := game.CardReference{Kind: game.CardReferenceLinked, LinkID: "looked"}
	return []game.Instruction{
		{
			Primitive: game.LookAtLibraryTop{
				Player:        game.ControllerReference(),
				PublishLinked: "looked",
			},
		},
		{
			Primitive: game.ConditionalDestinationPlace{
				Card:     looked,
				FromZone: zone.Library,
				CardCondition: opt.Val(game.CardSelection{
					Card:      looked,
					Selection: game.Selection{RequiredTypesAny: []types.Card{cardType}},
				}),
				EntryTapped: tapped,
			},
		},
	}
}

// TestConditionalLookAtTopBattlefieldAcceptEntersTapped proves that when the
// looked-at card matches the gate and the controller accepts, the card enters the
// battlefield tapped and leaves the library.
func TestConditionalLookAtTopBattlefieldAcceptEntersTapped(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	topID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Secret Forest",
		Types: []types.Card{types.Land},
	}})
	addInstructionSpellToStack(g, conditionalLookAtTopBattlefieldSequence(types.Land, true))
	agent := &lookChoiceAgent{}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = agent

	NewEngine(nil).resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if g.Players[game.Player1].Library.Contains(topID) {
		t.Fatal("accepted land remained in library")
	}
	var found *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == topID {
			found = permanent
		}
	}
	if found == nil {
		t.Fatal("accepted land did not enter the battlefield")
	}
	if !found.Tapped {
		t.Fatal("accepted land should have entered tapped")
	}
}

// TestConditionalLookAtTopBattlefieldNonMatchingStaysOnLibrary proves that when
// the looked-at card does not match the gate type, no put is offered and, because
// there is no else clause (Else = zone.None), the card stays atop the library
// instead of moving anywhere.
func TestConditionalLookAtTopBattlefieldNonMatchingStaysOnLibrary(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	topID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Secret Relic",
		Types: []types.Card{types.Artifact},
	}})
	addInstructionSpellToStack(g, conditionalLookAtTopBattlefieldSequence(types.Land, false))
	agent := &lookChoiceAgent{}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = agent

	NewEngine(nil).resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Library.Contains(topID) {
		t.Fatal("non-matching card should have stayed in the library")
	}
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == topID {
			t.Fatal("non-matching card should not have entered the battlefield")
		}
	}
	if g.Players[game.Player1].Hand.Contains(topID) {
		t.Fatal("non-matching card should not have entered the hand with no else clause")
	}
}

// TestConditionalLookAtTopBattlefieldDeclineStaysOnLibrary proves that when the
// gate holds but the controller declines the optional battlefield put, the card
// is left atop the library because there is no else clause.
func TestConditionalLookAtTopBattlefieldDeclineStaysOnLibrary(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	topID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Secret Forest",
		Types: []types.Card{types.Land},
	}})
	addInstructionSpellToStack(g, conditionalLookAtTopBattlefieldSequence(types.Land, false))
	agent := conditionalDestinationAgent{acceptPut: false}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = agent

	NewEngine(nil).resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Library.Contains(topID) {
		t.Fatal("declined land should have stayed in the library")
	}
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == topID {
			t.Fatal("declined land should not have entered the battlefield")
		}
	}
}
