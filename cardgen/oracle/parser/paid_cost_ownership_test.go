package parser

import "testing"

func TestPaidCostSpellOwnershipScopes(t *testing.T) {
	for _, test := range []struct {
		name, text string
		bound      bool
	}{
		{"earlier sacrifice paragraph", "As an additional cost to cast this spell, sacrifice a creature.\nSacrifice another creature.\nIf the sacrificed creature was red, draw a card.", false},
		{"earlier discard paragraph", "As an additional cost to cast this spell, discard a card.\nDiscard another card.\nIf the discarded card was a land card, draw a card.", false},
		{"earlier numeric sacrifice paragraph", "As an additional cost to cast this spell, sacrifice a creature.\nSacrifice another creature.\nIf the sacrificed creature had toughness 4 or greater, draw a card.", false},
		{"negated ambiguous sacrifice", "As an additional cost to cast this spell, sacrifice a creature.\nSacrifice another creature.\nIf the sacrificed creature was not red, draw a card.", false},
		{"earlier modal paragraph", "As an additional cost to cast this spell, sacrifice a creature.\nChoose one \u2014\n\u2022 Sacrifice another creature.\n\u2022 You gain 1 life.\nIf the sacrificed creature was red, draw a card.", false},
		{"earlier paragraph modal consumer", "As an additional cost to cast this spell, sacrifice a creature.\nSacrifice another creature.\nChoose one \u2014\n\u2022 If the sacrificed creature was red, draw a card.\n\u2022 You gain 1 life.", false},
		{"alternative before consumer", "As an additional cost to cast this spell, sacrifice a creature.\nYou may sacrifice a creature rather than pay this spell's mana cost.\nIf the sacrificed creature was red, draw a card.", false},
		{"alternative after consumer", "As an additional cost to cast this spell, sacrifice a creature.\nIf the sacrificed creature was red, draw a card.\nYou may sacrifice a creature rather than pay this spell's mana cost.", false},
		{"discard alternative", "As an additional cost to cast this spell, discard a card.\nYou may discard a card rather than pay this spell's mana cost.\nIf the discarded card was a land card, draw a card.", false},
		{"alternative cached replacement", "As an additional cost to cast this spell, sacrifice a creature.\nYou may sacrifice a creature rather than pay this spell's mana cost.\nDraw a card. If the sacrificed creature had toughness 4 or greater, draw two cards instead.", false},
		{"alternative alone is not publisher", "You may sacrifice a creature rather than pay this spell's mana cost.\nIf the sacrificed creature was red, draw a card.", false},
		{"later additional ambiguity", "As an additional cost to cast this spell, sacrifice a creature.\nIf the sacrificed creature was red, draw a card.\nAs an additional cost to cast this spell, sacrifice an artifact.", false},
		{"different alternative domain", "As an additional cost to cast this spell, discard a card.\nYou may sacrifice a creature rather than pay this spell's mana cost.\nIf the discarded card was a land card, draw a card.", true},
		{"unconditional spell paragraph", "As an additional cost to cast this spell, sacrifice a creature.\nYou gain 1 life.\nIf the sacrificed creature was red, draw a card.", true},
		{"separate activation", "Sacrifice a creature: Sacrifice another creature.\nSacrifice a creature: If the sacrificed creature was red, draw a card.", true},
		{"separate triggered ability", "Whenever you sacrifice a creature, sacrifice another creature.\nSacrifice a creature: If the sacrificed creature was red, draw a card.", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			document, diagnostics := Parse(test.text, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics: %v", diagnostics)
			}
			found := 0
			for _, ability := range document.Abilities {
				references := append([]Reference(nil), ability.SemanticReferences...)
				if ability.Modal != nil {
					for _, mode := range ability.Modal.Options {
						references = append(references, mode.SemanticReferences...)
					}
				}
				for _, reference := range references {
					if reference.Kind != ReferencePaidCostSubject {
						continue
					}
					found++
					if reference.PaidCost == nil || reference.PaidCost.Known != test.bound {
						t.Fatalf("binding = %#v, want known=%v", reference.PaidCost, test.bound)
					}
					if test.bound {
						producer := document.Abilities[reference.PaidCost.Producer.ClauseID-1]
						if producer.Kind == AbilitySpellAlternativeCost {
							t.Fatal("optional alternative became a required cost publisher")
						}
					}
				}
			}
			if found != 1 {
				t.Fatalf("paid-cost reference count = %d, want 1", found)
			}
		})
	}
}

func TestPaidCostPassiveSpellOwnershipScopes(t *testing.T) {
	for _, text := range []string{
		"As an additional cost to cast this spell, sacrifice a creature.\nSacrifice another creature.\nIf a Saproling was sacrificed this way, draw a card.",
		"As an additional cost to cast this spell, sacrifice a creature.\nIf a Saproling was sacrificed this way, draw a card.\nYou may sacrifice a creature rather than pay this spell's mana cost.",
	} {
		document, diagnostics := Parse(text, Context{InstantOrSorcery: true})
		if len(diagnostics) != 0 {
			t.Fatalf("diagnostics: %v", diagnostics)
		}
		found := false
		for _, ability := range document.Abilities {
			for _, condition := range ability.ConditionClauses {
				if condition.ThisWaySelection != nil {
					found = true
					if condition.Predicate != ConditionPredicateResultThisWay {
						t.Fatal("ambiguous passive antecedent became a paid-cost predicate")
					}
				}
			}
			for _, reference := range ability.SemanticReferences {
				if reference.Kind == ReferencePaidCostSubject && reference.PaidCost.Known {
					t.Fatal("ambiguous passive antecedent gained a cost publisher")
				}
			}
		}
		if !found {
			t.Fatal("passive antecedent lost its unresolved action-result grammar")
		}
	}
}
