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

func TestDiscardEntireHandSyntaxIsTypedAndFailsClosed(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		oracle string
		want   bool
	}{
		{oracle: "Discard your hand.", want: true},
		{oracle: "You discard your hand.", want: true},
		{oracle: "Each player discards their hand.", want: true},
		{oracle: "Each opponent discards their hand.", want: true},
		{oracle: "Target player discards their hand.", want: true},
		{oracle: "Discard two cards."},
		{oracle: "Each player discards two cards."},
		{oracle: "Each player may discard their hand."},
		{oracle: "Target creature's controller discards their hand."},
	} {
		t.Run(test.oracle, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.oracle, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("Parse(%q) diagnostics = %#v", test.oracle, diagnostics)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) == 0 || effects[0].DiscardEntireHand != test.want {
				t.Fatalf("Parse(%q) DiscardEntireHand = %#v, want %v", test.oracle, effects, test.want)
			}
			if test.want && !effects[0].Exact {
				t.Fatalf("Parse(%q) Exact = false, want true", test.oracle)
			}
		})
	}
}
