package compiler

import "testing"

const greatTrainHeistOracle = "Spree (Choose one or more additional costs.)\n" +
	"+ {2}{R} — Untap all creatures you control. If it's your combat phase, there is an additional combat phase after this phase.\n" +
	"+ {2} — Creatures you control get +1/+0 and gain first strike until end of turn.\n" +
	"+ {R} — Choose target opponent. Whenever a creature you control deals combat damage to that player this turn, create a tapped Treasure token."

func TestCompileGreatTrainHeistTypedModes(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(greatTrainHeistOracle, pipelineContext{
		CardName:         "Great Train Heist",
		InstantOrSorcery: true,
	})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	content := compilation.Abilities[0].Content
	if len(content.Modes) != 3 || content.Modes[0].Modal == nil {
		t.Fatalf("compiled modes = %#v", content.Modes)
	}
	if got := content.Modes[0].SpreeCost.ManaValue(); got != 3 {
		t.Fatalf("mode 1 cost = %d, want 3", got)
	}
	var sawCombatCondition bool
	for _, condition := range content.Modes[0].Content.Conditions {
		sawCombatCondition = sawCombatCondition || condition.Predicate == ConditionPredicateControllerCombatPhase
	}
	if !sawCombatCondition {
		t.Fatalf("mode 1 conditions = %#v, want controller combat phase", content.Modes[0].Content.Conditions)
	}
	mode3 := content.Modes[2].Content
	if len(mode3.Targets) != 1 || len(mode3.Effects) != 1 {
		t.Fatalf("mode 3 content = %#v", mode3)
	}
	delayed := mode3.Effects[0]
	if delayed.Kind != EffectDelayedTrigger ||
		!delayed.DelayedTriggerBindEventPlayer ||
		delayed.DelayedTriggerAbility == nil {
		t.Fatalf("mode 3 delayed effect = %#v", delayed)
	}
}
