package parser

import "testing"

// priorSubjectLifeClauseExact parses a compound sentence whose second effect is
// a life change with an elided ("inherited") subject and reports whether that
// second effect round-tripped to an exact, lowerable production.
func priorSubjectLifeClauseExact(t *testing.T, source string) bool {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("Parse(%q) shape = %#v", source, document.Abilities)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) < 2 {
		t.Fatalf("Parse(%q) effects = %#v, want at least 2", source, effects)
	}
	last := effects[len(effects)-1]
	if last.Context != EffectContextPriorSubject {
		t.Fatalf("Parse(%q) last effect context = %v, want EffectContextPriorSubject", source, last.Context)
	}
	return last.Exact
}

// TestExactPriorSubjectLifeClauseAccepts proves that a life change whose subject
// is shared with (inherited from) the prior effect in a compound sentence
// reconstructs byte-exactly from its bare third-person verb, enabling the drain
// "draw/discard/mill ... and lose N life" sequence to lower.
func TestExactPriorSubjectLifeClauseAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		"Target player draws two cards and loses 2 life.",
		"Target player draws a card and loses 1 life.",
		"Target player discards two cards and loses 2 life.",
		"Target player draws three cards and loses 3 life.",
		"Target player draws X cards and loses X life.",
		"Target player mills two cards, draws two cards, and loses 2 life.",
	}
	for _, source := range accepted {
		if !priorSubjectLifeClauseExact(t, source) {
			t.Errorf("priorSubjectLifeClauseExact(%q) = false, want true", source)
		}
	}
}

// TestExactPriorSubjectLifeClauseFailsClosed proves the elided-subject life
// clause is not marked exact when its amount uses the "where X is ..." form. A
// trailing "where X is ..." defines a single X shared by every effect, but the
// parser binds the clause to only one effect, so the sibling's amount would be
// unbound; the whole sequence must fail closed.
func TestExactPriorSubjectLifeClauseFailsClosed(t *testing.T) {
	t.Parallel()
	rejected := []string{
		"Target player draws X cards and loses X life, where X is the number of artifacts you control.",
	}
	for _, source := range rejected {
		if priorSubjectLifeClauseExact(t, source) {
			t.Errorf("priorSubjectLifeClauseExact(%q) = true, want false", source)
		}
	}
}
