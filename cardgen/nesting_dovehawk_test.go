package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLowerNestingDovehawkPopulateTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Nesting Dovehawk",
		Layout:     "normal",
		TypeLine:   "Creature — Bird",
		ManaCost:   "{3}{W}",
		Power:      new("2"),
		Toughness:  new("2"),
		OracleText: "Flying\nAt the beginning of combat on your turn, populate. (Create a token that's a copy of a creature token you control.)\nWhenever a creature token you control enters, put a +1/+1 counter on this creature.",
	})
	if len(face.TriggeredAbilities) != 2 {
		t.Fatalf("triggered abilities = %d, want 2", len(face.TriggeredAbilities))
	}
	populate := face.TriggeredAbilities[0]
	if populate.Trigger.Pattern.Event != game.EventBeginningOfStep ||
		populate.Trigger.Pattern.Step != game.StepBeginningOfCombat ||
		populate.Trigger.Pattern.Controller != game.TriggerControllerYou {
		t.Fatalf("populate trigger = %#v", populate.Trigger.Pattern)
	}
	create, ok := populate.Content.Modes[0].Sequence[0].Primitive.(game.CreateToken)
	spec, copyOK := create.Source.TokenCopy()
	if !ok || !copyOK ||
		spec.Source != game.TokenCopySourceChosenControlledCreatureToken {
		t.Fatalf("populate instruction = %#v", populate.Content.Modes[0].Sequence)
	}
}
