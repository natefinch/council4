package compiler

import "testing"

func TestCompileCanAttackAsThoughDefenderEffect(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"{G}: This creature can attack this turn as though it didn't have defender.",
		pipelineContext{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Content.Effects
	if len(effects) != 1 {
		t.Fatalf("effects = %#v, want exactly one", effects)
	}
	effect := effects[0]
	if effect.Kind != EffectCanAttackAsThoughDefender {
		t.Errorf("effect.Kind = %v, want EffectCanAttackAsThoughDefender", effect.Kind)
	}
	if !effect.Exact {
		t.Error("effect.Exact = false, want true")
	}
	if effect.Duration != DurationThisTurn {
		t.Errorf("effect.Duration = %v, want DurationThisTurn", effect.Duration)
	}
}

func TestCompileCanAttackAsThoughDefenderFailsClosed(t *testing.T) {
	t.Parallel()
	// Wordings that deviate from the exact temporary permission must not compile
	// to an exact EffectCanAttackAsThoughDefender resolving effect.
	rejected := []string{
		"{G}: This creature can attack as though it didn't have defender.",
		"{G}: This creature can't attack this turn.",
		"{G}: This creature can attack this turn as though it weren't tapped.",
	}
	for _, source := range rejected {
		compilation, _ := compileSource(source, pipelineContext{})
		if len(compilation.Abilities) == 0 {
			continue
		}
		for _, effect := range compilation.Abilities[0].Content.Effects {
			if effect.Kind == EffectCanAttackAsThoughDefender && effect.Exact {
				t.Errorf("compileSource(%q) produced an exact EffectCanAttackAsThoughDefender, want fail closed", source)
			}
		}
	}
}
