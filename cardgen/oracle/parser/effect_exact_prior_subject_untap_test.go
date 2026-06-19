package parser

import "testing"

// priorSubjectNextUntapClauseExact parses a two-sentence tap-down spell whose
// second sentence is a negated untap with an inherited ("prior subject")
// subject — "<It / That permanent> doesn't untap during its controller's next
// untap step." — and reports whether that clause round-tripped to an exact,
// lowerable production.
func priorSubjectNextUntapClauseExact(t *testing.T, source string) bool {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("Parse(%q) abilities = %#v", source, document.Abilities)
	}
	var untap *EffectSyntax
	for i := range document.Abilities[0].Sentences {
		effects := document.Abilities[0].Sentences[i].Effects
		for j := range effects {
			if effects[j].Kind == EffectUntap && effects[j].Negated {
				untap = &effects[j]
			}
		}
	}
	if untap == nil {
		t.Fatalf("Parse(%q) found no negated untap effect", source)
	}
	return untap.Exact
}

// TestExactPriorSubjectNextUntapClauseAccepts proves that a singular "doesn't
// untap during its controller's next untap step" clause whose subject is
// inherited from the prior tap reconstructs byte-exactly, enabling the
// tap-down (stun) sequence to lower.
func TestExactPriorSubjectNextUntapClauseAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		"Tap target creature. It doesn't untap during its controller's next untap step.",
		"Tap target creature an opponent controls. That creature doesn't untap during its controller's next untap step.",
		"Tap target land an opponent controls. That land doesn't untap during its controller's next untap step.",
		"Tap target nonland permanent an opponent controls. That permanent doesn't untap during its controller's next untap step.",
		"Tap up to two target creatures. Those creatures don't untap during their controller's next untap step.",
	}
	for _, source := range accepted {
		if !priorSubjectNextUntapClauseExact(t, source) {
			t.Errorf("priorSubjectNextUntapClauseExact(%q) = false, want true", source)
		}
	}
}

// TestExactPriorSubjectNextUntapClauseFailsClosed proves the inherited-subject
// negated-untap clause stays inexact for shapes the SkipNextUntap primitive
// cannot model: a multi-step "next two untap steps" window, an open-ended "for
// as long as you control" duration, and a wrong-player "your next untap step".
func TestExactPriorSubjectNextUntapClauseFailsClosed(t *testing.T) {
	t.Parallel()
	rejected := []string{
		"Tap target creature. It doesn't untap during its controller's next two untap steps.",
		"Tap target creature. It doesn't untap during its controller's untap step for as long as you control this creature.",
		"Tap target creature. It doesn't untap during your next untap step.",
	}
	for _, source := range rejected {
		if priorSubjectNextUntapClauseExact(t, source) {
			t.Errorf("priorSubjectNextUntapClauseExact(%q) = true, want false", source)
		}
	}
}
