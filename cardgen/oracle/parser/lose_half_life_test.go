package parser

import (
	"testing"
)

// TestParseLoseHalfLife checks the "loses half their life, rounded up/down"
// wording types to an exact EffectLose whose amount is the half-player-life
// dynamic amount with the recognized rounding, and whose possessive "their" is
// consumed (leaving only the losing-player reference).
func TestParseLoseHalfLife(t *testing.T) {
	t.Parallel()
	cases := []struct {
		text    string
		roundUp bool
	}{
		{"Whenever this creature deals combat damage to a player, that player loses half their life, rounded up.", true},
		{"Whenever this creature deals combat damage to a player, that player loses half their life, rounded down.", false},
	}
	for _, tc := range cases {
		document, diagnostics := Parse(tc.text, Context{})
		if len(diagnostics) != 0 {
			t.Fatalf("%q diagnostics = %#v", tc.text, diagnostics)
		}
		effect := document.Abilities[0].Sentences[0].Effects[0]
		if effect.Kind != EffectLose || !effect.Exact ||
			effect.Amount.DynamicKind != EffectDynamicAmountHalfPlayerLife ||
			effect.Amount.DynamicForm != EffectDynamicAmountFormHalfLife ||
			effect.Amount.RoundUp != tc.roundUp {
			t.Fatalf("%q effect = %#v", tc.text, effect)
		}
	}
}
