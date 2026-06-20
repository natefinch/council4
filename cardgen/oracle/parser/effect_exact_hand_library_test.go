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
	for _, source := range []string{
		"Draw three cards, then put one card from your hand on top of your library in any order.",
		"Draw three cards, then put two cards from your hand on top of your library in any order.",
		"Draw three cards, then put 3 cards from your hand on top of your library in any order.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			put := handLibraryPutEffect(t, source)
			if !put.Exact || !put.HandLibraryPut.Present {
				t.Fatalf("put = %#v, want exact typed hand-library put", put)
			}
		})
	}
}

func TestHandLibraryPutSyntaxFailsClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Draw three cards, then put two cards from your hand on the bottom of your library in any order.",
		"Draw three cards, then put two cards from your hand on top of your library in a random order.",
		"Draw three cards, then put two cards from your hand on top of your library.",
		"Draw three cards, then put two cards from your hand on top of your library in the same order.",
		"Draw three cards, then put X cards from your hand on top of your library in any order.",
		"Draw three cards, then put two cards from an opponent's hand on top of your library in any order.",
		"Draw three cards, then put two revealed cards from your hand on top of your library in any order.",
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
