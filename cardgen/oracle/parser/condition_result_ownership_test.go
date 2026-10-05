package parser

import (
	"reflect"
	"testing"
)

func TestResultConditionOwnsProducerAndConsumers(t *testing.T) {
	for _, tt := range []struct {
		text      string
		producers []int
		owners    [][]int
		elseOwner int
	}{
		{"Discard a card. You gain 2 life. If a land card was discarded this way, draw a card. Scry 1.", []int{1}, [][]int{{3}}, 0},
		{"Discard a card. If a land card was discarded this way, draw a card. If a creature card was discarded this way, you gain 2 life.", []int{1, 1}, [][]int{{2}, {3}}, 0},
		{"You may discard a card. If you do, draw a card. You may sacrifice a creature. If you do, you gain 2 life.", []int{1, 3}, [][]int{{2}, {4}}, 0},
		{"Discard a card. If a land card was discarded this way, draw a card. Otherwise, you gain 2 life.", []int{1}, [][]int{{2}}, 2},
		{"You may discard a card. If a land card was discarded this way, draw a card. If you don't, you gain 2 life.", []int{1, 1}, [][]int{{2}, {3}}, 0},
		{"You may discard a card. If you do, you may sacrifice a creature. If you don't, you gain 2 life.", []int{1, 2}, [][]int{{2}, {3}}, 0},
		{"If a land card was discarded this way, draw a card. Discard a card.", []int{0}, [][]int{{1}}, 0},
	} {
		t.Run(tt.text, func(t *testing.T) {
			document, diagnostics := Parse(tt.text, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			ability := document.Abilities[0]
			if len(ability.ConditionSegments) != len(tt.producers) {
				t.Fatalf("conditions = %#v", ability.ConditionSegments)
			}
			for ci, segment := range ability.ConditionSegments {
				if segment.Ownership.ResultProducerClauseID != tt.producers[ci] ||
					!reflect.DeepEqual(segment.Ownership.ClauseIDs, tt.owners[ci]) {
					t.Fatalf("condition %d ownership = %#v", ci, segment.Ownership)
				}
			}
			for _, sentence := range ability.Sentences {
				for _, effect := range sentence.Effects {
					if effect.Connection == EffectConnectionOtherwise && effect.ResultElseOfClauseID != tt.elseOwner {
						t.Fatalf("Otherwise owner = %d, want %d", effect.ResultElseOfClauseID, tt.elseOwner)
					}
				}
			}
		})
	}
}

func TestOtherwiseOwnsSubjectContinuation(t *testing.T) {
	document, diagnostics := Parse("Discard a card. If a land card was discarded this way, you gain 2 life. Otherwise, put a +1/+1 counter on target creature. It gains trample until end of turn. You gain 1 life.", Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	effects := conditionEffects(document.Abilities[0].Sentences)
	for i, effect := range effects {
		want := 0
		if i == 2 || i == 3 {
			want = 2
		}
		if effect.ResultElseOfClauseID != want {
			t.Fatalf("clause %d else owner = %d, want %d", effect.ClauseID, effect.ResultElseOfClauseID, want)
		}
	}
}

func TestRepeatedOptionalActionOwnership(t *testing.T) {
	for _, tt := range []struct {
		text      string
		wantGroup []int
	}{
		{"Repeat the following process two times. You may draw a card and gain 2 life.", []int{1, 2}},
		{"Repeat the following process two times. You may draw a card.", []int{1}},
		{"You may put a land card from your hand onto the battlefield. If you do, draw a card and repeat this process.", []int{1}},
	} {
		t.Run(tt.text, func(t *testing.T) {
			document, diagnostics := Parse(tt.text, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			effects := conditionEffects(document.Abilities[0].Sentences)
			if len(effects) == 0 || !reflect.DeepEqual(effects[0].OptionalActionClauseIDs, tt.wantGroup) {
				t.Fatalf("optional group = %#v, want %v", effects, tt.wantGroup)
			}
		})
	}
}

func TestQuantifiedResultGrammar(t *testing.T) {
	for _, tt := range []struct {
		body       string
		controller bool
		count      int
		comparison ConditionComparison
		supported  bool
	}{
		{"you discarded a land card this way", true, 0, ConditionComparisonNone, true},
		{"you discard a land card this way", true, 0, ConditionComparisonNone, true},
		{"you sacrifice an Island this way", true, 0, ConditionComparisonNone, true},
		{"you exiled two instant or sorcery cards this way", true, 2, ConditionComparisonNone, true},
		{"two or more creature cards were milled this way", false, 2, ConditionComparisonAtLeast, true},
		{"three cards were discarded this way", false, 3, ConditionComparisonNone, true},
		{"cards were exiled this way", false, 0, ConditionComparisonNone, true},
		{"that card was discarded this way", false, 0, ConditionComparisonNone, false},
		{"you discarded cards that share a color this way", false, 0, ConditionComparisonNone, false},
		{"a card was revealed this way", false, 0, ConditionComparisonNone, false},
		{"you discarded no cards this way", false, 0, ConditionComparisonNone, false},
		{"zero cards were discarded this way", false, 0, ConditionComparisonNone, false},
		{"two or fewer cards were discarded this way", false, 0, ConditionComparisonNone, false},
	} {
		t.Run(tt.body, func(t *testing.T) {
			document, diagnostics := Parse("Discard three cards. If "+tt.body+", draw a card.", Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			clauses := document.Abilities[0].ConditionClauses
			if !tt.supported && len(clauses) == 0 {
				return
			}
			if len(clauses) != 1 {
				t.Fatalf("clauses = %#v", clauses)
			}
			clause := clauses[0]
			if (clause.Predicate == ConditionPredicateResultThisWay) != tt.supported {
				t.Fatalf("predicate = %v, want supported=%t", clause.Predicate, tt.supported)
			}
			if tt.supported && (clause.ThisWayController != tt.controller || clause.ThisWayCount != tt.count ||
				clause.ThisWayCountComparison != tt.comparison) {
				t.Fatalf("payload = %#v", clause)
			}
		})
	}
}
