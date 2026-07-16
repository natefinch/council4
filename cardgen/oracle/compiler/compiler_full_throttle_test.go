package compiler

import "testing"

func TestCompileFullThrottleTypedMechanics(t *testing.T) {
	t.Parallel()
	source := "After this main phase, there are two additional combat phases.\n" +
		"At the beginning of each combat this turn, untap all creatures that attacked this turn."
	compilation, diagnostics := compileSource(source, pipelineContext{
		CardName:         "Test Sorcery",
		InstantOrSorcery: true,
	})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(compilation.Abilities) != 2 {
		t.Fatalf("ability count = %d, want two spell paragraphs", len(compilation.Abilities))
	}
	extra := compilation.Abilities[0].Content.Effects[0]
	if extra.Kind != EffectAdditionalCombatPhase ||
		!extra.AdditionalCombatPhase ||
		extra.AdditionalCombatPhaseCount != 2 ||
		extra.AdditionalMainPhase {
		t.Fatalf("extra combat effect = %#v", extra)
	}
	delayed := compilation.Abilities[1].Content.Effects[0]
	if delayed.Kind != EffectDelayedTrigger ||
		delayed.DelayedTriggerAbility == nil ||
		delayed.DelayedTriggerOneShot {
		t.Fatalf("delayed effect = %#v", delayed)
	}
}
