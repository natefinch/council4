package parser

import "testing"

func TestPaidCostReferencesBelongOnlyToPredicateConsumers(t *testing.T) {
	for _, text := range []string{
		"Sacrifice a creature: You gain life equal to the sacrificed creature's power.",
		"Sacrifice a creature: Target player mills cards equal to the sacrificed creature's toughness.",
		"As an additional cost to cast this spell, sacrifice a creature.\nDraw cards equal to the sacrificed creature's mana value.",
		"Choose one \u2014\n\u2022 Sacrifice a creature. You gain life equal to the sacrificed creature's toughness.\n\u2022 Draw a card.",
	} {
		document, diagnostics := Parse(text, Context{InstantOrSorcery: true})
		if len(diagnostics) != 0 {
			t.Fatalf("diagnostics: %v", diagnostics)
		}
		for _, ability := range document.Abilities {
			references := append([]Reference(nil), ability.SemanticReferences...)
			if ability.Modal != nil {
				for _, mode := range ability.Modal.Options {
					references = append(references, mode.SemanticReferences...)
				}
			}
			for _, reference := range references {
				if reference.Kind == ReferencePaidCostSubject {
					t.Fatal("legacy characteristic amount acquired an unowned predicate reference")
				}
			}
		}
	}
}

func TestPaidCostPredicateDoesNotClaimNeighboringAmount(t *testing.T) {
	document, diagnostics := Parse("Sacrifice a creature: You gain life equal to the sacrificed creature's power. If the sacrificed creature had toughness 4 or greater, draw a card.",
		Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics: %v", diagnostics)
	}
	references := document.Abilities[0].SemanticReferences
	if len(references) != 1 || references[0].Kind != ReferencePaidCostSubject ||
		references[0].PaidCost == nil || !references[0].PaidCost.Known {
		t.Fatalf("predicate references = %#v, want the single exact condition subject", references)
	}
	condition := document.Abilities[0].ConditionClauses[0]
	if condition.SubjectRefID != references[0].NodeID || condition.SubjectSpan != references[0].Span {
		t.Fatal("predicate consumer inherited the neighboring amount's identity")
	}
}

func TestPaidCostPredicatePreservesOtherTypedOperand(t *testing.T) {
	document, diagnostics := Parse("Sacrifice a creature: This creature deals damage equal to its power to target creature. If the sacrificed creature had toughness 4 or greater, draw a card.",
		Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics: %v", diagnostics)
	}
	ability := document.Abilities[0]
	found := false
	for _, sentence := range ability.Sentences {
		for _, effect := range sentence.Effects {
			for _, reference := range ability.SemanticReferences {
				if reference.Pronoun == PronounIts && reference.NodeID == effect.Amount.ReferenceNodeID &&
					reference.Span == effect.Amount.ReferenceSpan {
					found = true
					if reference.Kind == ReferencePaidCostSubject {
						t.Fatal("typed source operand was stolen by paid predicate ownership")
					}
				}
			}
		}
	}
	if !found {
		t.Fatal("typed characteristic operand lost its exact reference identity")
	}
}
