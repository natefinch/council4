package parser

import "testing"

// TestParseWhicheverIsGreaterMaxAmount proves the "<A> or <B>, whichever is
// greater" wording compiles to a EffectDynamicAmountMaxOf combinator carrying
// both operands in source order (Willowdusk, Essence Seer's "the amount of life
// you gained this turn or the amount of life you lost this turn, whichever is
// greater"), and that the same combinator works regardless of the surrounding
// effect. Unrecognized or single-amount wording fails closed.
func TestParseWhicheverIsGreaterMaxAmount(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source       string
		wantOperands []EffectDynamicAmountKind
	}{
		{
			"You gain life equal to the amount of life you gained this turn or the amount of life you lost this turn, whichever is greater.",
			[]EffectDynamicAmountKind{
				EffectDynamicAmountLifeGainedThisTurn,
				EffectDynamicAmountLifeLostThisTurn,
			},
		},
		{
			"You draw cards equal to the amount of life you lost this turn or the amount of life you gained this turn, whichever is greater.",
			[]EffectDynamicAmountKind{
				EffectDynamicAmountLifeLostThisTurn,
				EffectDynamicAmountLifeGainedThisTurn,
			},
		},
		// Fail closed: a single amount is not a max combinator.
		{"You gain life equal to the amount of life you gained this turn.", nil},
		// Fail closed: "whichever is greater" without two recognized amounts.
		{"You gain life equal to the amount of life you spent this turn or the amount of life you saved this turn, whichever is greater.", nil},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{InstantOrSorcery: true})
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v, want one", effects)
			}
			amount := effects[0].Amount
			if test.wantOperands == nil {
				if amount.DynamicKind == EffectDynamicAmountMaxOf {
					t.Fatalf("amount = %#v, want non-max", amount)
				}
				return
			}
			if amount.DynamicKind != EffectDynamicAmountMaxOf {
				t.Fatalf("amount dynamic kind = %v, want EffectDynamicAmountMaxOf", amount.DynamicKind)
			}
			if len(amount.Operands) != len(test.wantOperands) {
				t.Fatalf("operands = %#v, want %d", amount.Operands, len(test.wantOperands))
			}
			for i, want := range test.wantOperands {
				if got := amount.Operands[i].DynamicKind; got != want {
					t.Fatalf("operand[%d] kind = %v, want %v", i, got, want)
				}
			}
		})
	}
}
