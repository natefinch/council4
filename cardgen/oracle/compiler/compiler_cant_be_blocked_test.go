package compiler

import "testing"

func TestCompileCantBeBlockedThisTurnEffect(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"Target creature can't be blocked this turn.",
		pipelineContext{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Content.Effects
	if len(effects) != 1 {
		t.Fatalf("effects = %#v, want exactly one", effects)
	}
	effect := effects[0]
	if effect.Kind != EffectCantBeBlocked {
		t.Errorf("effect.Kind = %v, want EffectCantBeBlocked", effect.Kind)
	}
	if !effect.Exact {
		t.Error("effect.Exact = false, want true")
	}
	if effect.Duration != DurationThisTurn {
		t.Errorf("effect.Duration = %v, want DurationThisTurn", effect.Duration)
	}
	targets := compilation.Abilities[0].Content.Targets
	if len(targets) != 1 || targets[0].Selector.Kind != SelectorCreature {
		t.Fatalf("targets = %#v, want a single creature target", targets)
	}
}

func TestCompileCantBeBlockedThisTurnFailsClosed(t *testing.T) {
	t.Parallel()
	// Wordings that deviate from the exact temporary restriction must not compile
	// to an exact EffectCantBeBlocked resolving effect.
	rejected := []string{
		"Target creature can't be blocked.",
		"Target creature can't be blocked until end of turn.",
		"Target creature can't be blocked this turn except by Walls.",
		"Up to two target creatures can't be blocked this turn.",
		"Target creature can't block this turn.",
		"Target creature can't attack this turn.",
	}
	for _, source := range rejected {
		compilation, _ := compileSource(source, pipelineContext{InstantOrSorcery: true})
		if len(compilation.Abilities) == 0 {
			continue
		}
		for _, effect := range compilation.Abilities[0].Content.Effects {
			if effect.Kind == EffectCantBeBlocked && effect.Exact {
				t.Errorf("compileSource(%q) produced an exact EffectCantBeBlocked, want fail closed", source)
			}
		}
	}
}
