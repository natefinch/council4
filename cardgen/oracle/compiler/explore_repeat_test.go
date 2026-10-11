package compiler

import "testing"

func TestCompileExploreRepeatTypedParameters(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		text    string
		value   int
		x       bool
		dynamic DynamicAmountKind
	}{
		{"When this creature enters, it explores two times.", 2, false, DynamicAmountNone},
		{"When this creature enters, it explores X times.", 0, true, DynamicAmountNone},
		{"Target creature explores X times, where X is the number of artifacts you control.", 0, false, DynamicAmountCount},
	} {
		t.Run(tc.text, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(tc.text, pipelineContext{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %+v", diagnostics)
			}
			effects := compilation.Abilities[0].Content.Effects
			if len(effects) != 1 || effects[0].Kind != EffectExplore || !effects[0].Exact {
				t.Fatalf("effects = %+v", effects)
			}
			amount := effects[0].Amount
			if amount.Value != tc.value || amount.VariableX != tc.x || amount.DynamicKind != tc.dynamic {
				t.Fatalf("amount = %+v", amount)
			}
		})
	}
}
