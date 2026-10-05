package parser

import "testing"

func TestPaidCostModalCompetingProducers(t *testing.T) {
	for _, test := range []struct {
		name, choose, cost, action, condition string
		bound                                 bool
		passive                               bool
	}{
		{"co-selected sacrifice", "one or both", "sacrifice a creature", "Sacrifice another creature",
			"the sacrificed creature was red", false, false},
		{"exclusive sacrifice", "one", "sacrifice a creature", "Sacrifice another creature",
			"the sacrificed creature was red", true, false},
		{"co-selected discard", "one or both", "discard a card", "Discard a card",
			"the discarded card was a land card", false, false},
		{"exclusive discard", "one", "discard a card", "Discard a card",
			"the discarded card was a land card", true, false},
		{"co-selected passive sacrifice", "one or both", "sacrifice a creature", "Sacrifice another creature",
			"a Saproling was sacrificed this way", false, true},
		{"exclusive passive sacrifice", "one", "sacrifice a creature", "Sacrifice another creature",
			"a Saproling was sacrificed this way", true, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			doc, diagnostics := Parse("As an additional cost to cast this spell, "+test.cost+
				".\nChoose "+test.choose+" \u2014\n\u2022 "+test.action+".\n\u2022 If "+test.condition+", draw a card.",
				Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("parse diagnostics: %v", diagnostics)
			}
			var mode *Mode
			for _, ability := range doc.Abilities {
				if ability.Modal != nil {
					mode = &ability.Modal.Options[1]
				}
			}
			if mode == nil {
				t.Fatal("modal fixture not parsed")
			}
			if test.passive && !test.bound {
				if mode.ConditionClauses[0].Predicate != ConditionPredicateResultThisWay {
					t.Fatal("co-selectable effect producer was stolen by a passive cost rewrite")
				}
				return
			}
			found := false
			for _, reference := range mode.SemanticReferences {
				if reference.Kind == ReferencePaidCostSubject {
					found = true
					if reference.PaidCost.Known != test.bound {
						t.Fatalf("bound=%v, want %v", reference.PaidCost.Known, test.bound)
					}
				}
			}
			if !found {
				t.Fatalf("missing typed paid-subject reference: conditions %+v", mode.ConditionClauses)
			}
		})
	}
}
