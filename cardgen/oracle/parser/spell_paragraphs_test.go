package parser

import "testing"

func TestResolvingAlternativeParagraphRetainsThreeLabelScopes(t *testing.T) {
	source := "Tap target creature. Put a +1/+1 counter on target creature.\n" +
		"Morbid — Put three +1/+1 counters on that creature instead if a creature died this turn.\n" +
		"Formidable — If you have 5 or less life, draw a card."
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 || len(document.Abilities) != 1 {
		t.Fatalf("paragraph scope: abilities=%d diagnostics=%+v", len(document.Abilities), diagnostics)
	}
	ability := &document.Abilities[0]
	if len(ability.ParagraphLabelSpans()) != 4 || !DocumentCoverage(document).Complete {
		t.Fatal("both recognized labels must retain complete original lexical coverage")
	}
	effects := conditionEffects(ability.Sentences)
	if len(effects) != 4 || effects[2].Replacement.ReplacedClauseID() != effects[1].ClauseID {
		t.Fatal("independent first clause displaced alternative ownership")
	}
	seen := map[int]bool{}
	for _, effect := range effects {
		if effect.ClauseID == 0 || seen[effect.ClauseID] {
			t.Fatal("paragraph clauses aliased identity")
		}
		seen[effect.ClauseID] = true
	}
	if len(ability.ConditionSegments) != 2 ||
		ability.ConditionSegments[0].Ownership.ClauseIDs[0] == ability.ConditionSegments[1].Ownership.ClauseIDs[0] {
		t.Fatal("third paragraph predicate aliased earlier alternative")
	}
}

func TestResolvingAlternativeParagraphScope(t *testing.T) {
	source := "Put a +1/+1 counter on target creature.\nMorbid — Put three +1/+1 counters on that creature instead if a creature died this turn."
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	if len(document.Abilities) != 1 {
		for _, ability := range document.Abilities {
			t.Logf("kind=%v plain=%v conditional=%v body=%q sentences=%d", ability.Kind,
				plainResolvingParagraph(ability), conditionalInsteadParagraph(ability),
				joinedEffectText(TokensInSpan(ability.Tokens, ability.BodySpan)), len(ability.Sentences))
		}
		t.Fatalf("paragraphs=%d", len(document.Abilities))
	}
	ability := &document.Abilities[0]
	if len(ability.ParagraphLabelSpans()) != 2 {
		t.Fatal("recognized original label and separator ownership lost")
	}
	for _, span := range ability.ParagraphLabelSpans() {
		found := false
		for _, mustCover := range ability.CoverageSpans() {
			found = found || mustCover.Start.Offset >= span.Start.Offset && mustCover.End.Offset <= span.End.Offset
		}
		if !found {
			t.Fatal("label was removed from must-cover original lexical spans")
		}
	}
	if coverage := DocumentCoverage(document); !coverage.Complete {
		t.Fatalf("original source coverage=%+v", coverage)
	}
	effects := conditionEffects(ability.Sentences)
	if len(effects) != 2 || effects[0].ClauseID == effects[1].ClauseID ||
		effects[1].Replacement.ReplacedClauseID() != effects[0].ClauseID {
		t.Fatal("replacement owner is not a distinct parser clause identity")
	}
}
