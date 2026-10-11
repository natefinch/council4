package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestLocalProductRepeatFramesInheritReadOnlyAndHidePublishers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Players[game.Player1].Life = 20
	addCardToLibrary(g, game.Player1, vanillaCreature("Observed", 2, 3))
	obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1}
	key := game.LinkedKey("local-link")
	condition := opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{
		Object:        opt.Val(game.LinkedObjectReference(string(key))),
		ObjectMatches: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
	})})
	reader := game.Instruction{Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(1)}, Condition: condition}
	readOnly := game.RepeatProcess{Times: game.Fixed(2), Body: game.Mode{Sequence: []game.Instruction{reader}}.Ability()}
	hidden := game.RepeatProcess{Times: game.Fixed(3), Body: game.Mode{Sequence: []game.Instruction{
		{Primitive: game.LookAtLibraryTop{Player: game.ControllerReference(), PublishLinked: key},
			ConditionGate: "absent", LocalProducts: game.LocalProducts{Links: []game.LinkedKey{key}}},
		{Primitive: readOnly},
	}}.Ability()}
	sequence := []game.Instruction{
		{Primitive: game.LookAtLibraryTop{Player: game.ControllerReference(), PublishLinked: key},
			LocalProducts: game.LocalProducts{Links: []game.LinkedKey{key}}},
		{Primitive: readOnly},
		{Primitive: hidden},
		reader,
	}
	NewEngine(nil).resolveInstructionSequence(g, obj, sequence, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if got := g.Players[game.Player1].Life; got != 23 {
		t.Fatalf("life=%d, want 23: nested reader inherited an inner skipped publisher or lost its outer product", got)
	}
	if len(g.LinkedObjects) != 0 || len(obj.LocalLinkedProducts) != 0 {
		t.Fatal("successive or nested repeat invocation leaked local subjects")
	}
}
