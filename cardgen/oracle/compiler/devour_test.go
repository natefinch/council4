package compiler

import "testing"

// TestCompileDevourKeyword verifies that the compiler maps the parser's Devour
// as-enters replacement through to a typed EffectDevour with its multiplier
// preserved, without inspecting the printed keyword text (CR 702.81).
func TestCompileDevourKeyword(t *testing.T) {
	t.Parallel()
	source := "Devour 2 (As this creature enters, you may sacrifice any number of creatures. It enters with twice that many +1/+1 counters on it.)"
	compilation, diagnostics := compileSource(source, pipelineContext{CardName: "Devourer"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Kind != AbilityReplacement {
		t.Fatalf("kind = %v, want AbilityReplacement", ability.Kind)
	}
	if len(ability.Content.Effects) != 1 {
		t.Fatalf("effects = %#v", ability.Content.Effects)
	}
	effect := ability.Content.Effects[0]
	if effect.Kind != EffectDevour {
		t.Fatalf("effect kind = %v, want EffectDevour", effect.Kind)
	}
	if !effect.EntersDevour {
		t.Fatal("EntersDevour = false, want true")
	}
	if effect.EntersDevourMultiplier != 2 {
		t.Fatalf("EntersDevourMultiplier = %d, want 2", effect.EntersDevourMultiplier)
	}
}
