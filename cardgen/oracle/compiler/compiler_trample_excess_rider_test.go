package compiler

import "testing"

// excessRedirectEffect returns the compiled deal-damage effect that models an
// "excess damage is dealt to ... instead" redirect (its amount is the excess
// damage dealt this way), or nil when the ability carries none.
func excessRedirectEffect(effects []CompiledEffect) *CompiledEffect {
	for i := range effects {
		if effects[i].Kind == EffectDealDamage &&
			effects[i].Amount.DynamicKind == DynamicAmountExcessDamageDealtThisWay {
			return &effects[i]
		}
	}
	return nil
}

// TestCompileSourceTrampleExcessRiderMarker proves the compiler carries Ram
// Through's RequireSourceTrample marker onto the excess-damage redirect while
// leaving Flame Spill's unconditional redirect unmarked, and that Ram Through's
// stripped rider leaves the ability with no leftover conditions.
func TestCompileSourceTrampleExcessRiderMarker(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		source         string
		requireTrample bool
	}{
		{
			"Ram Through",
			"Target creature you control deals damage equal to its power to target creature you don't control. " +
				"If the creature you control has trample, excess damage is dealt to that creature's controller instead.",
			true,
		},
		{
			"Flame Spill",
			"Flame Spill deals 4 damage to target creature. Excess damage is dealt to that creature's controller instead.",
			false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(test.source, pipelineContext{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			content := compilation.Abilities[0].Content
			redirect := excessRedirectEffect(content.Effects)
			if redirect == nil {
				t.Fatalf("no excess-damage redirect compiled from %q", test.source)
			}
			if redirect.RequireSourceTrample != test.requireTrample {
				t.Fatalf("RequireSourceTrample = %v, want %v", redirect.RequireSourceTrample, test.requireTrample)
			}
			if test.requireTrample && len(content.Conditions) != 0 {
				t.Fatalf("conditions = %#v, want none (stripped)", content.Conditions)
			}
		})
	}
}
