package parser

import "testing"

func TestCounterResultOwnership(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		text string
		want []int
	}{
		{"Counter target spell. If that spell is countered this way, draw a card.", []int{1}},
		{"Counter target spell unless its controller pays {3}. You gain 2 life. If that spell is countered this way, draw a card.", []int{1}},
		{"You gain 2 life. Counter target spell. If that spell was countered this way, draw a card.", []int{2}},
		{"Exile target creature. Counter target spell. If that spell is countered this way, draw a card.", []int{2}},
		{"Counter target spell. If that spell is countered this way, draw a card. Counter target spell. If that spell is countered this way, you gain 2 life.", []int{1, 3}},
		{"Counter target spell. Exile target creature. If that spell is countered this way, draw a card.", []int{0}},
		{"Counter up to two target spells. If that spell is countered this way, draw a card.", []int{0}},
		{"If that spell is countered this way, draw a card. Counter target spell.", []int{0}},
	} {
		t.Run(tt.text, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(tt.text, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			ability, found := document.Abilities[0], 0
			for _, segment := range ability.ConditionSegments {
				if segment.ClauseIndex < 0 ||
					ability.ConditionClauses[segment.ClauseIndex].Predicate != ConditionPredicateCounterSucceeded {
					continue
				}
				if found >= len(tt.want) || segment.Ownership.ResultProducerClauseID != tt.want[found] ||
					segment.Ownership.ResultSubjectReferenceNodeID < 0 {
					t.Fatalf("counter condition ownership = %#v; want producers %v", segment.Ownership, tt.want)
				}
				found++
			}
			if found != len(tt.want) {
				t.Fatalf("counter conditions = %d, want %d", found, len(tt.want))
			}
		})
	}
}

func TestCounterSucceededGrammarRejectsOtherOutcomes(t *testing.T) {
	t.Parallel()
	for _, body := range []string{
		"that spell is exiled this way",
		"a spell is countered this way",
		"that ability is countered this way",
		"that spell is not countered this way",
		"that spell is countered",
		"that spell is countered by an opponent this way",
	} {
		t.Run(body, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse("Counter target spell. If "+body+", draw a card.", Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			for _, clause := range document.Abilities[0].ConditionClauses {
				if clause.Predicate == ConditionPredicateCounterSucceeded {
					t.Fatalf("unexpected counter predicate for %q", body)
				}
			}
		})
	}
}
