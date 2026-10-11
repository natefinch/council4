package parser

import "testing"

func TestPaymentActorResultOwnsCompatibleProducerAcrossIndependentClauses(t *testing.T) {
	for _, tc := range []struct {
		name, text string
		owner      int
	}{
		{"adjacent", "Counter target spell unless its controller pays {3}. If they do, draw a card.", 1},
		{"independent controller", "Counter target spell unless its controller pays {3}. You gain 1 life. If they do, draw a card.", 1},
		{"controller owns controller action", "Counter target spell unless its controller pays {3}. You gain 1 life. If you do, draw a card.", 2},
		{"two payments", "Counter target spell unless its controller pays {3}. Counter target spell unless its controller pays {2}. You gain 1 life. If they do, draw a card.", 2},
		{"explicit player actor", "Target player draws a card. You gain 1 life. If that player does, draw a card.", 1},
		{"nonpayment spell action", "Counter target spell. You gain 1 life. If they do, draw a card.", 0},
		{"plural", "Counter up to two target spells unless their controller pays {3}. If they do, draw a card.", 0},
		{"missing actor target", "You gain 1 life. If they do, draw a card.", 0},
		{"forward actor target", "You gain 1 life. If they do, draw a card. Counter target spell unless its controller pays {3}.", 0},
		{"intervening same actor", "Target player gains 1 life. That player discards a card. If that player does, you gain 2 life.", 0},
		{"intervening payer action", "Counter target spell unless its controller pays {3}. They discard a card. If they do, draw a card.", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			doc, diagnostics := Parse(tc.text, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			found := false
			for _, segment := range doc.Abilities[0].ConditionSegments {
				if segment.ClauseIndex < 0 || segment.ClauseIndex >= len(doc.Abilities[0].ConditionClauses) {
					continue
				}
				clause := doc.Abilities[0].ConditionClauses[segment.ClauseIndex]
				if clause.Predicate != ConditionPredicatePriorInstructionAccepted {
					continue
				}
				found = true
				if segment.Ownership.ResultProducerClauseID != tc.owner {
					t.Fatalf("actor producer=%d want=%d", segment.Ownership.ResultProducerClauseID, tc.owner)
				}
				if tc.owner > 0 && segment.Ownership.ResultSubjectTargetOccurrence < 0 {
					t.Fatal("actor lost exact target occurrence")
				}
			}
			if !found {
				t.Fatal("missing result condition")
			}
		})
	}
}

func TestPaymentActorResultRejectsAmbiguousAndForeignOwnership(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*Ability)
	}{
		{"duplicate actor", func(a *Ability) {
			a.SemanticReferences = append(a.SemanticReferences, a.SemanticReferences[len(a.SemanticReferences)-1])
		}},
		{"duplicate target", func(a *Ability) {
			a.Sentences[0].Targets = append(a.Sentences[0].Targets, a.Sentences[0].Targets[0])
		}},
		{"foreign effect owner", func(a *Ability) { a.Sentences[0].Effects[0].Targets = nil }},
		{"incompatible actor", func(a *Ability) { a.Sentences[0].Effects[0].Payment.Payer = EffectPaymentPayerController }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			doc, diagnostics := Parse("Counter target spell unless its controller pays {3}. You gain 1 life. If they do, draw a card.", Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			a := &doc.Abilities[0]
			tc.mutate(a)
			for ci := range a.ConditionSegments {
				a.ConditionSegments[ci].Ownership.ResultProducerClauseID = 0
			}
			emitResultConditionOwnership(a.Sentences, a.ConditionSegments, a.ConditionClauses, a.SemanticReferences)
			for _, segment := range a.ConditionSegments {
				if segment.ClauseIndex >= 0 && a.ConditionClauses[segment.ClauseIndex].Predicate == ConditionPredicatePriorInstructionAccepted &&
					segment.Ownership.ResultProducerClauseID != 0 {
					t.Fatal("ambiguous or incompatible actor acquired a result producer")
				}
			}
		})
	}
}
