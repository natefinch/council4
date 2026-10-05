package parser

import "testing"

func TestRemovedCounterQuantityOwnership(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		text string
		want []int
	}{
		{"Remove two charge counters from this artifact. You gain that much life.", []int{1}},
		{"Remove all +1/+1 counters from this creature. Draw that many cards.", []int{1}},
		{"Remove a charge counter from this artifact. Untap this artifact. You gain life equal to the number of counters removed this way.", []int{1}},
		{"Remove a charge counter from this artifact. You gain that much life. Remove two charge counters from this artifact. Draw that many cards.", []int{1, 3}},
		{"You gain life equal to the number of counters removed this way. Remove a charge counter from this artifact.", []int{0}},
		{"Remove a charge counter from this artifact. Remove a charge counter from this artifact. You gain life equal to the number of counters removed this way.", []int{0}},
		{"Remove a charge counter from this artifact. Draw a card. You gain that much life.", []int{0}},
	} {
		t.Run(tt.text, func(t *testing.T) {
			t.Parallel()
			doc, diagnostics := Parse(tt.text, Context{})
			if len(diagnostics) != 0 || len(doc.Abilities) != 1 {
				t.Fatalf("parse = %#v, %#v", doc, diagnostics)
			}
			var amounts []EffectAmountSyntax
			for _, sentence := range doc.Abilities[0].Sentences {
				for _, effect := range sentence.Effects {
					if effect.Amount.DynamicKind == EffectDynamicAmountRemovedCounterCount {
						amounts = append(amounts, effect.Amount)
					}
				}
			}
			if len(amounts) != len(tt.want) {
				t.Fatalf("amounts=%#v, want producers %v", amounts, tt.want)
			}
			for i, amount := range amounts {
				if amount.ProducerClauseID != tt.want[i] || amount.Multiplier != 1 {
					t.Fatalf("amount[%d]=%#v, want producer %d", i, amount, tt.want[i])
				}
			}
		})
	}
}
