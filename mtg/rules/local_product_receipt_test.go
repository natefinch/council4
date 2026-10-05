package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLocalProductConditionalRepeatCapturesCurrentReceiptBeforeExit(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	for range 3 {
		addCardToLibrary(g, game.Player1, vanillaCreature("Observed", 2, 3))
	}
	obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1,
		ResolutionResults: map[string]game.InstructionResolutionResult{"continue": {Succeeded: true, Amount: 99}}}
	body := game.Mode{Sequence: []game.Instruction{{
		Primitive:     game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(1)},
		PublishResult: "continue", LocalProducts: game.LocalProducts{Results: []game.ResultKey{"continue"}},
	}}}.Ability()
	NewEngine(nil).resolveInstructionSequence(g, obj, []game.Instruction{{
		Primitive: game.RepeatProcess{Body: body, ContinueResult: "continue"},
	}}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if g.Players[game.Player1].Hand.Size() != 3 {
		t.Fatal("conditional repeat read a restored outer receipt instead of its current iteration")
	}
	if receipt := obj.ResolutionResults["continue"]; !receipt.Succeeded || receipt.Amount != 99 {
		t.Fatal("conditional repeat destroyed the enclosing receipt")
	}
}
