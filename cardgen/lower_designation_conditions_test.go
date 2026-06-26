package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerControllerDesignationInterveningConditions verifies that the
// controller-designation intervening-if conditions ("if you're the monarch", "if
// you have the initiative", "if you have the city's blessing") lower onto the
// matching live single-player game-state predicate on the trigger's intervening
// condition.
func TestLowerControllerDesignationInterveningConditions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		oracle string
		field  func(game.Condition) bool
	}{
		{
			name:   "monarch",
			oracle: "At the beginning of your end step, if you're the monarch, you draw a card.",
			field:  func(c game.Condition) bool { return c.ControllerIsMonarch },
		},
		{
			name:   "initiative",
			oracle: "At the beginning of your end step, if you have the initiative, you draw a card.",
			field:  func(c game.Condition) bool { return c.ControllerHasInitiative },
		},
		{
			name:   "city's blessing",
			oracle: "At the beginning of your end step, if you have the city's blessing, you draw a card.",
			field:  func(c game.Condition) bool { return c.ControllerHasCityBlessing },
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: tc.oracle,
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
			}
			trigger := face.TriggeredAbilities[0].Trigger
			if trigger.InterveningIf == "" || !trigger.InterveningCondition.Exists {
				t.Fatalf("trigger = %+v, want intervening condition", trigger)
			}
			cond := trigger.InterveningCondition.Val
			if cond.Negate {
				t.Error("condition Negate = true, want false")
			}
			if !tc.field(cond) {
				t.Errorf("condition = %+v, want designation predicate set", cond)
			}
		})
	}
}

// TestControllerDesignationConditionRejectedOutsideInterveningTrigger verifies
// the designation predicates fail closed when used as a static "as long as"
// condition, which the runtime does not evaluate for these designations.
func TestControllerDesignationConditionRejectedOutsideInterveningTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "As long as you're the monarch, Test Bear gets +1/+1.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	for _, ability := range face.TriggeredAbilities {
		if ability.Trigger.InterveningCondition.Val.ControllerIsMonarch {
			t.Fatalf("static monarch condition unexpectedly lowered: %+v", ability)
		}
	}
}
