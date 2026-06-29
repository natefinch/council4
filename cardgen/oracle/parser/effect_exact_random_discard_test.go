package parser

import "testing"

func TestNonControllerRandomDiscardSyntaxIsTypedAndFailsClosed(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		oracle string
		want   bool
	}{
		{oracle: "Target player discards two cards at random.", want: true},
		{oracle: "Each player discards a card at random.", want: true},
		{oracle: "Each opponent discards a card at random.", want: true},
		{oracle: "Discard a card at random."},
		{oracle: "Discard two cards at random."},
		{oracle: "Target player discards two cards."},
		{oracle: "Each player discards a card."},
		{oracle: "Target player discards X cards at random.", want: true},
		{oracle: "Each player discards their hand."},
	} {
		t.Run(test.oracle, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.oracle, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("Parse(%q) diagnostics = %#v", test.oracle, diagnostics)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) == 0 || effects[0].RandomDiscard != test.want {
				t.Fatalf("Parse(%q) RandomDiscard = %#v, want %v", test.oracle, effects, test.want)
			}
		})
	}
}
