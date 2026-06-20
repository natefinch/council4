package parser

import "testing"

func TestControllerHandDiscardSyntaxIsTypedAndFailsClosed(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		oracle string
		want   bool
	}{
		{oracle: "Discard two cards.", want: true},
		{oracle: "You discard a card.", want: true},
		{oracle: "Target player discards two cards."},
		{oracle: "Each opponent discards two cards."},
		{oracle: "Discard X cards."},
		{oracle: "Discard two cards at random."},
		{oracle: "Discard two land cards."},
	} {
		t.Run(test.oracle, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.oracle, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("Parse(%q) diagnostics = %#v", test.oracle, diagnostics)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 || effects[0].HandDiscard.Present != test.want {
				t.Fatalf("Parse(%q) HandDiscard = %#v, want Present=%v", test.oracle, effects, test.want)
			}
		})
	}
}
