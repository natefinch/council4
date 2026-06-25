package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// scriptedChoiceAgent answers every ChoiceRequest with a fixed selection.
type scriptedChoiceAgent struct {
	answer []int
}

func (scriptedChoiceAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Action{}
}

func (a scriptedChoiceAgent) ChooseChoice(_ PlayerObservation, _ game.ChoiceRequest) []int {
	return a.answer
}

func creatureWithManaValue(name string, value int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		Types:    []types.Card{types.Creature},
		ManaCost: opt.Val(cost.Mana{cost.O(value)}),
	}}
}

func totalManaValueReanimationInstruction(amount, maxTotal int) *game.Instruction {
	return &game.Instruction{
		Primitive: game.ReturnFromGraveyardChoice(
			game.ControllerReference(),
			game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou},
			game.Fixed(amount),
			zone.Battlefield,
			false,
			opt.Val(maxTotal),
			false,
			"",
		),
	}
}

// TestReturnFromGraveyardTotalManaValueWithinCapReanimates verifies a subset of
// creature cards whose combined mana value stays within the cap is put onto the
// battlefield.
func TestReturnFromGraveyardTotalManaValueWithinCapReanimates(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Source",
		Types: []types.Card{types.Artifact},
	}})
	obj := triggeredObjFor(source)

	two := addCardToGraveyard(g, game.Player1, creatureWithManaValue("Two", 2))
	threeAlso := addCardToGraveyard(g, game.Player1, creatureWithManaValue("Three", 2))

	// Choose both creatures: 2 + 2 = 4, exactly at the cap.
	agents := [game.NumPlayers]PlayerAgent{game.Player1: scriptedChoiceAgent{answer: []int{0, 1}}}
	engine.resolveInstructionWithChoices(g, obj, totalManaValueReanimationInstruction(2, 4), agents, &TurnLog{})

	for _, cardID := range []id.ID{two, threeAlso} {
		if !onBattlefieldByCard(g, cardID) {
			t.Fatalf("card %v within the cap was not put onto the battlefield", cardID)
		}
		if g.Players[game.Player1].Graveyard.Contains(cardID) {
			t.Fatalf("card %v still in graveyard", cardID)
		}
	}
}

// TestReturnFromGraveyardTotalManaValueOverCapRejected verifies an over-cap
// selection is rejected and the engine falls back to the empty default, leaving
// every card in the graveyard.
func TestReturnFromGraveyardTotalManaValueOverCapRejected(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Source",
		Types: []types.Card{types.Artifact},
	}})
	obj := triggeredObjFor(source)

	big := addCardToGraveyard(g, game.Player1, creatureWithManaValue("Big", 3))
	bigger := addCardToGraveyard(g, game.Player1, creatureWithManaValue("Bigger", 3))

	// Choose both creatures: 3 + 3 = 6, over the cap of 4. The engine must
	// reject the selection and fall back to returning nothing.
	agents := [game.NumPlayers]PlayerAgent{game.Player1: scriptedChoiceAgent{answer: []int{0, 1}}}
	engine.resolveInstructionWithChoices(g, obj, totalManaValueReanimationInstruction(2, 4), agents, &TurnLog{})

	for _, cardID := range []id.ID{big, bigger} {
		if onBattlefieldByCard(g, cardID) {
			t.Fatalf("over-cap card %v was put onto the battlefield", cardID)
		}
		if !g.Players[game.Player1].Graveyard.Contains(cardID) {
			t.Fatalf("over-cap card %v left the graveyard", cardID)
		}
	}
}
