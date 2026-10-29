package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestMadnessDiscardGoesToExile(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cardID := addCardToHand(g, game.Player1, madnessSorcery(cost.Mana{cost.O(1)}))

	if !discardCardFromHand(g, game.Player1, cardID) {
		t.Fatal("discardCardFromHand() = false, want true")
	}

	if g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("madness card went to graveyard instead of exile")
	}
	if !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("madness card did not go to exile")
	}
	assertEvent(t, g.Events, game.EventCardDiscarded, func(event game.Event) bool {
		return event.CardID == cardID && event.FromZone == zone.Hand && event.ToZone == zone.Exile
	})
}

func TestMadnessTriggerCastsCardFromExile(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, madnessSorcery(cost.Mana{cost.G}))
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)

	discardCardFromHand(g, game.Player1, cardID)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("madness trigger was not put on the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if !forest.Tapped {
		t.Fatal("madness cost did not tap mana source")
	}
	if g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("cast madness card remained in exile")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.Kind != game.StackSpell || obj.SourceID != cardID {
		t.Fatalf("stack top = %+v, want madness spell", obj)
	}
	assertEvent(t, g.Events, game.EventSpellCast, func(event game.Event) bool {
		return event.CardID == cardID && event.FromZone == zone.Exile && event.ToZone == zone.Stack
	})
}

func TestUnpayableMadnessTriggerMovesCardToGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, madnessSorcery(cost.Mana{cost.O(1)}))

	discardCardFromHand(g, game.Player1, cardID)
	engine.putTriggeredAbilitiesOnStack(g)
	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("unpayable madness card remained in exile")
	}
	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("unpayable madness card did not move to graveyard")
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want no madness spell", g.Stack.Size())
	}
}

func TestDeclinedMadnessTriggerMovesCardToGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, madnessSorcery(cost.Mana{cost.G}))
	addBasicLandPermanent(g, game.Player1, types.Forest)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}}}

	discardCardFromHand(g, game.Player1, cardID)
	engine.putTriggeredAbilitiesOnStack(g)
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("declined madness card remained in exile")
	}
	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("declined madness card did not move to graveyard")
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want no madness spell", g.Stack.Size())
	}
}

// TestMadnessCastPaysSpellAdditionalCosts proves the madness cost replaces only
// the mana cost (CR 702.35b, 601.2f): the spell's additional cost is still part
// of the total cost, so the cast pays it alongside the madness mana and cannot be
// made when the additional cost is unpayable.
func TestMadnessCastPaysSpellAdditionalCosts(t *testing.T) {
	for _, tc := range []struct {
		name     string
		life     int
		cast     bool
		wantLife int
	}{
		{"payable", 20, true, 17},
		{"unpayable", 2, false, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			def := madnessSorcery(cost.Mana{cost.G})
			def.AdditionalCosts = []cost.Additional{{Kind: cost.AdditionalPayLife, Amount: 3, Text: "Pay 3 life"}}
			cardID := addCardToHand(g, game.Player1, def)
			forest := addBasicLandPermanent(g, game.Player1, types.Forest)
			player := g.Players[game.Player1]
			player.Life = tc.life

			discardCardFromHand(g, game.Player1, cardID)
			if !engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("madness trigger was not put on the stack")
			}
			engine.resolveTopOfStack(g, &TurnLog{})

			obj, ok := g.Stack.Peek()
			if cast := ok && obj.Kind == game.StackSpell && obj.SourceID == cardID; cast != tc.cast {
				t.Fatalf("madness cast = %v, want %v", cast, tc.cast)
			}
			if player.Life != tc.wantLife || forest.Tapped != tc.cast {
				t.Fatalf("life = %d, forest tapped = %v; want %d, %v", player.Life, forest.Tapped, tc.wantLife, tc.cast)
			}
			if !tc.cast && (!player.Graveyard.Contains(cardID) || player.Exile.Contains(cardID)) {
				t.Fatal("uncast madness card did not move from exile to the graveyard")
			}
		})
	}
}

func madnessSorcery(manaCost cost.Mana) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Madness Sorcery",
		Types: []types.Card{types.Sorcery},
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.MadnessKeyword{Cost: manaCost}},
		}}},
	}
}
