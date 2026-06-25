package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func landCard(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  name,
		Types: []types.Card{types.Land},
	}}
}

func anyNumberReturnInstruction() *game.Instruction {
	return &game.Instruction{
		Primitive: game.ReturnFromGraveyardChoice(
			game.ControllerReference(),
			game.Selection{RequiredTypes: []types.Card{types.Land}},
			game.Quantity{},
			zone.Battlefield,
			true,
			opt.V[int]{},
			true,
			"",
		),
	}
}

// TestReturnFromGraveyardAnyNumberPutsChosenSubset verifies the "put any number
// of <type> cards onto the battlefield" form lets the player put any chosen
// subset of the matching graveyard cards onto the battlefield, leaving the rest
// (and non-matching cards) in the graveyard.
func TestReturnFromGraveyardAnyNumberPutsChosenSubset(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Source",
		Types: []types.Card{types.Artifact},
	}})
	obj := triggeredObjFor(source)

	forest := addCardToGraveyard(g, game.Player1, landCard("Forest"))
	island := addCardToGraveyard(g, game.Player1, landCard("Island"))
	bear := addCardToGraveyard(g, game.Player1, creatureWithManaValue("Bear", 2))

	// Two lands are candidates; choose exactly one of them (the first offered).
	agents := [game.NumPlayers]PlayerAgent{game.Player1: scriptedChoiceAgent{answer: []int{0}}}
	engine.resolveInstructionWithChoices(g, obj, anyNumberReturnInstruction(), agents, &TurnLog{})

	onBattlefield := 0
	for _, cardID := range []id.ID{forest, island} {
		if onBattlefieldByCard(g, cardID) {
			onBattlefield++
		}
	}
	if onBattlefield != 1 {
		t.Fatalf("expected exactly one chosen land on the battlefield, got %d", onBattlefield)
	}
	if onBattlefieldByCard(g, bear) {
		t.Fatal("non-matching creature was put onto the battlefield")
	}
	if !g.Players[game.Player1].Graveyard.Contains(bear) {
		t.Fatal("non-matching creature left the graveyard")
	}
}

// TestReturnFromGraveyardAnyNumberAllowsEmptyChoice verifies the any-number form
// permits choosing none, leaving every matching card in the graveyard.
func TestReturnFromGraveyardAnyNumberAllowsEmptyChoice(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Source",
		Types: []types.Card{types.Artifact},
	}})
	obj := triggeredObjFor(source)

	forest := addCardToGraveyard(g, game.Player1, landCard("Forest"))
	island := addCardToGraveyard(g, game.Player1, landCard("Island"))

	agents := [game.NumPlayers]PlayerAgent{game.Player1: scriptedChoiceAgent{answer: []int{}}}
	engine.resolveInstructionWithChoices(g, obj, anyNumberReturnInstruction(), agents, &TurnLog{})

	for _, cardID := range []id.ID{forest, island} {
		if onBattlefieldByCard(g, cardID) {
			t.Fatalf("card %v was put onto the battlefield despite the empty choice", cardID)
		}
		if !g.Players[game.Player1].Graveyard.Contains(cardID) {
			t.Fatalf("card %v left the graveyard despite the empty choice", cardID)
		}
	}
}
