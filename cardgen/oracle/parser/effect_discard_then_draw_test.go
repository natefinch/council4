package parser

import "testing"

// fusedDiscardEffect returns the single fused discard effect of a parsed
// looter body, failing the test if the body did not fuse into exactly one
// effect.
func fusedDiscardEffect(t *testing.T, source string) EffectSyntax {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 || len(document.Abilities) != 1 {
		t.Fatalf("abilities = %#v diagnostics = %#v, want one ability", document.Abilities, diagnostics)
	}
	sentences := document.Abilities[0].Sentences
	if len(sentences) != 1 || len(sentences[0].Effects) != 1 {
		t.Fatalf("sentences = %#v, want one sentence with one fused effect", sentences)
	}
	return sentences[0].Effects[0]
}

// TestFuseDiscardThenDrawUpToN proves the parser fuses "discard up to two
// cards, then draw that many cards" into a single discard effect bounded at two
// with no draw offset, and clears the ordered-lowering requirement.
func TestFuseDiscardThenDrawUpToN(t *testing.T) {
	effect := fusedDiscardEffect(t, "Discard up to two cards, then draw that many cards.")
	if effect.Kind != EffectDiscard {
		t.Fatalf("effect kind = %v, want EffectDiscard", effect.Kind)
	}
	if !effect.DiscardThenDraw {
		t.Fatal("DiscardThenDraw = false, want true")
	}
	if effect.DiscardThenDrawMax != 2 {
		t.Fatalf("DiscardThenDrawMax = %d, want 2", effect.DiscardThenDrawMax)
	}
	if effect.DiscardThenDrawOffset != 0 {
		t.Fatalf("DiscardThenDrawOffset = %d, want 0", effect.DiscardThenDrawOffset)
	}
	if effect.RequiresOrderedLowering {
		t.Fatal("RequiresOrderedLowering = true, want false after fusion")
	}
}

// TestFuseDiscardThenDrawAnyNumberPlusOffset proves the parser fuses "discard
// any number of cards, then draw that many cards plus one" into a single
// unbounded discard effect carrying the draw offset.
func TestFuseDiscardThenDrawAnyNumberPlusOffset(t *testing.T) {
	effect := fusedDiscardEffect(t, "Discard any number of cards, then draw that many cards plus one.")
	if !effect.DiscardThenDraw {
		t.Fatal("DiscardThenDraw = false, want true")
	}
	if effect.DiscardThenDrawMax != 0 {
		t.Fatalf("DiscardThenDrawMax = %d, want 0 (any number)", effect.DiscardThenDrawMax)
	}
	if effect.DiscardThenDrawOffset != 1 {
		t.Fatalf("DiscardThenDrawOffset = %d, want 1", effect.DiscardThenDrawOffset)
	}
}

// TestFuseDiscardThenDrawSpanCoversBothClauses proves the fused effect's span
// covers both the discard and the consumed draw clause, so the draw tokens are
// reported as covered during lowering completeness checks.
func TestFuseDiscardThenDrawSpanCoversBothClauses(t *testing.T) {
	source := "Discard any number of cards, then draw that many cards."
	effect := fusedDiscardEffect(t, source)
	if effect.Span.Start.Offset != 0 || effect.Span.End.Offset != len(source) {
		t.Fatalf("fused span = %d-%d, want 0-%d covering both clauses",
			effect.Span.Start.Offset, effect.Span.End.Offset, len(source))
	}
}

// TestFuseDiscardThenDrawIgnoresPlainLooter proves an exact "draw then discard"
// loot body, which the ordinary effect vocabulary already supports, is left
// untouched (no fusion, two effects remain).
func TestFuseDiscardThenDrawIgnoresPlainLooter(t *testing.T) {
	source := "Draw two cards, then discard two cards."
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 || len(document.Abilities) != 1 {
		t.Fatalf("abilities = %#v diagnostics = %#v, want one ability", document.Abilities, diagnostics)
	}
	for _, effect := range document.Abilities[0].Sentences[0].Effects {
		if effect.DiscardThenDraw {
			t.Fatal("plain looter effect marked DiscardThenDraw, want untouched")
		}
	}
}
