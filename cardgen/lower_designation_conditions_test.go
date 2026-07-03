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

// TestLowerMonarchInsteadEscalationEffectGate verifies the monarch "instead"
// escalation cycle ("At the beginning of your upkeep, <base>. If you're the
// monarch, <escalated> instead.", the Court cycle) lowers both sub-effects with
// per-effect gates: the base effect runs only when the controller is not the
// monarch (a negated ControllerIsMonarch gate) and the escalated effect runs
// only when they are.
func TestLowerMonarchInsteadEscalationEffectGate(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Test Court",
		Layout:   "normal",
		TypeLine: "Enchantment",
		OracleText: "When this enchantment enters, you become the monarch.\n" +
			"At the beginning of your upkeep, this enchantment deals 2 damage to any target. " +
			"If you're the monarch, it deals 7 damage instead.",
	})
	var upkeep *game.AbilityContent
	for i := range face.TriggeredAbilities {
		content := face.TriggeredAbilities[i].Content
		if len(content.Modes) == 1 && len(content.Modes[0].Sequence) == 2 {
			upkeep = &face.TriggeredAbilities[i].Content
			break
		}
	}
	if upkeep == nil {
		t.Fatalf("no upkeep trigger with a two-instruction sequence: %#v", face.TriggeredAbilities)
	}
	seq := upkeep.Modes[0].Sequence
	base := seq[0].Condition
	escalated := seq[1].Condition
	if !base.Exists || !base.Val.Condition.Exists ||
		!base.Val.Condition.Val.ControllerIsMonarch || !base.Val.Condition.Val.Negate {
		t.Fatalf("base gate = %#v, want negated ControllerIsMonarch", base)
	}
	if !escalated.Exists || !escalated.Val.Condition.Exists ||
		!escalated.Val.Condition.Val.ControllerIsMonarch || escalated.Val.Condition.Val.Negate {
		t.Fatalf("escalated gate = %#v, want ControllerIsMonarch", escalated)
	}
}

// TestLowerMonarchInsteadTokenEscalation verifies the token-creation "instead"
// escalation (Court of Grace: "create a 1/1 Spirit token. If you're the monarch,
// create a 4/4 Angel token instead.") lowers both branches with the correct
// per-effect gates. The trailing " instead" on the escalation clause must not
// keep the create effect from being parser-exact.
func TestLowerMonarchInsteadTokenEscalation(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Test Court of Grace",
		Layout:   "normal",
		TypeLine: "Enchantment",
		OracleText: "When this enchantment enters, you become the monarch.\n" +
			"At the beginning of your upkeep, create a 1/1 white Spirit creature token with flying. " +
			"If you're the monarch, create a 4/4 white Angel creature token with flying instead.",
	})
	var upkeep *game.AbilityContent
	for i := range face.TriggeredAbilities {
		content := face.TriggeredAbilities[i].Content
		if len(content.Modes) == 1 && len(content.Modes[0].Sequence) == 2 {
			upkeep = &face.TriggeredAbilities[i].Content
			break
		}
	}
	if upkeep == nil {
		t.Fatalf("no upkeep trigger with a two-instruction sequence: %#v", face.TriggeredAbilities)
	}
	seq := upkeep.Modes[0].Sequence
	if _, ok := seq[0].Primitive.(game.CreateToken); !ok {
		t.Fatalf("base primitive = %#v, want CreateToken", seq[0].Primitive)
	}
	if _, ok := seq[1].Primitive.(game.CreateToken); !ok {
		t.Fatalf("escalation primitive = %#v, want CreateToken", seq[1].Primitive)
	}
	base := seq[0].Condition
	escalated := seq[1].Condition
	if !base.Exists || !base.Val.Condition.Exists ||
		!base.Val.Condition.Val.ControllerIsMonarch || !base.Val.Condition.Val.Negate {
		t.Fatalf("base gate = %#v, want negated ControllerIsMonarch", base)
	}
	if !escalated.Exists || !escalated.Val.Condition.Exists ||
		!escalated.Val.Condition.Val.ControllerIsMonarch || escalated.Val.Condition.Val.Negate {
		t.Fatalf("escalation gate = %#v, want ControllerIsMonarch", escalated)
	}
}

// TestLowerStandaloneCreateTokenInsteadFailsClosed proves the trailing-"instead"
// create-token exactness only lowers as a gated sequence escalation: a lone
// "Create <token> instead." with no preceding effect to replace fails closed
// rather than creating the token unconditionally. No real card hits this — every
// "create ... instead." is inline-gated into a sequence — but the guard keeps
// the single-effect path correct by construction.
func TestLowerStandaloneCreateTokenInsteadFailsClosed(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Standalone Instead",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Create a 4/4 white Angel creature token with flying instead.",
	})
}
