package parser

import "testing"

func handLibraryPutEffect(t *testing.T, source string) EffectSyntax {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true, CardName: "Test"})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("Parse(%q) abilities = %#v", source, document.Abilities)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	for i := range effects {
		if effects[i].Kind == EffectPut {
			return effects[i]
		}
	}
	t.Fatalf("Parse(%q) effects = %#v, want EffectPut", source, effects)
	return EffectSyntax{}
}

func TestExactHandLibraryPutSyntax(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		source string
		bottom bool
	}{
		{"Draw three cards, then put one card from your hand on top of your library in any order.", false},
		{"Draw three cards, then put two cards from your hand on top of your library in any order.", false},
		{"Draw three cards, then put 3 cards from your hand on top of your library in any order.", false},
		// The trailing "in any order" qualifier and its complete omission are
		// equivalent (CR 401.4 already grants the player-chosen-order
		// permission whenever multiple cards go to the same library position
		// at once) -- see HandLibraryPutSyntax's doc comment.
		{"Draw three cards, then put two cards from your hand on top of your library.", false},
		{"Draw three cards, then put a card from your hand on the bottom of your library.", true},
		{"Draw three cards, then put two cards from your hand on the bottom of your library in any order.", true},
		{"Draw three cards, then put two cards from your hand on the bottom of your library.", true},
		{"Put X cards from your hand on top of your library.", false},
		{"Put five cards from your hand on the bottom of your library.", true},
		{"Put 12 cards from your hand on top of your library.", false},
	} {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			put := handLibraryPutEffect(t, test.source)
			if !put.Exact || !put.HandLibraryPut.Present || put.HandLibraryPut.Bottom != test.bottom {
				t.Fatalf("put = %#v, want exact typed hand-library put with Bottom=%v", put, test.bottom)
			}
		})
	}
}

func TestHandLibraryPutSyntaxFailsClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		// "In a random order" and "in the same order" name a genuinely
		// different ordering rule than the player's free choice CR 401.4
		// grants by default (see HandLibraryPutSyntax's doc comment), so
		// neither is accepted even though the destination and amount are
		// otherwise exact.
		"Draw three cards, then put two cards from your hand on top of your library in a random order.",
		"Draw three cards, then put two cards from your hand on top of your library in the same order.",
		"Draw three cards, then put two cards from an opponent's hand on top of your library in any order.",
		"Draw three cards, then put two revealed cards from your hand on top of your library in any order.",
		"Put zero cards from your hand on top of your library.",
		"Put up to X cards from your hand on top of your library.",
		"Put X cards from your hand on top of your library, where X is the number of cards in your hand.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			put := handLibraryPutEffect(t, source)
			if put.Exact && put.HandLibraryPut.Present {
				t.Fatalf("put = %#v, unexpectedly recognized unsupported wording", put)
			}

		})
	}
}

func TestHandLibraryPutParameterSpans(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		text, parameter, spanText string
	}{
		{"Put zero cards from your hand on top of your library.", "amount", "zero cards"},
		{"Put up to two cards from your hand on top of your library.", "amount", "up to two cards"},
		{"Put X cards from your hand on top of your library in a random order.", "library order", "in a random order"},
		{"Put two cards from your hand on top of your library in the same order.", "library order", "in the same order"},
		{"Put two cards from an opponent's hand on top of your library.", "source owner", "from an opponent's hand"},
		{"Put two revealed cards from your hand on top of your library.", "card selection", "revealed"},
		{"Put two cards from your hand on top of an opponent's library.", "destination owner", "on top of an opponent's"},
	} {
		t.Run(tt.parameter+"/"+tt.spanText, func(t *testing.T) {
			t.Parallel()
			put := handLibraryPutEffect(t, tt.text)
			payload := put.HandLibraryPut
			if put.Exact || payload.Present || payload.UnsupportedParameter != tt.parameter {
				t.Fatalf("put exact=%v payload=%+v", put.Exact, payload)
			}
			got := tt.text[payload.ParameterSpan.Start.Offset:payload.ParameterSpan.End.Offset]
			if got != tt.spanText {
				t.Fatalf("parameter span = %q, want %q", got, tt.spanText)
			}
		})
	}
}
