package parser

import "testing"

func TestPaidCostSubjectOwnership(t *testing.T) {
	for _, test := range []struct {
		name, text string
		bound      bool
		consumers  int
	}{
		{"independent consumers", "Sacrifice a creature: Draw a card if the sacrificed creature was red. You gain 1 life if the sacrificed creature was black.", true, 2},
		{"spell additional cost", "As an additional cost to cast this spell, discard a card.\nDraw a card. If the discarded card was a land card, you gain 1 life.", true, 1},
		{"past numeric subject", "Sacrifice a creature: If the sacrificed creature had toughness 4 or greater, draw a card.", true, 1},
		{"past possessive numeric subject", "Discard a card: If the discarded card's mana value was 4 or less, draw a card.", true, 1},
		{"not an action result", "Sacrifice a creature: Sacrifice another creature. If the sacrificed creature was red, draw a card.", false, 1},
		{"repeated discard action", "Discard a card: Discard a card. Repeat this process once. If the discarded card was a land card, draw a card.", false, 1},
		{"repeated sacrifice action", "Sacrifice a creature: Sacrifice another creature. Repeat this process once. If the sacrificed creature was red, draw a card.", false, 1},
		{"plural payment", "Discard two cards: If the discarded card was a land card, draw a card.", false, 1},
		{"competing components", "Sacrifice a creature, Sacrifice an artifact: If the sacrificed creature was red, draw a card.", false, 1},
		{"alternative component", "Discard a card or sacrifice a creature: If the discarded card was a land card, draw a card.", false, 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			document, diagnostics := Parse(test.text, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("parse diagnostics: %v", diagnostics)
			}
			consumerIDs := map[int]bool{}
			var producer PaidCostProducer
			found := 0
			for _, ability := range document.Abilities {
				for _, reference := range ability.SemanticReferences {
					if reference.Kind != ReferencePaidCostSubject {
						continue
					}
					found++
					binding := reference.PaidCost
					if binding == nil || binding.Known != test.bound {
						t.Fatalf("binding = %#v, want known=%v", binding, test.bound)
					}
					if binding.ConsumerNodeID != reference.NodeID {
						t.Fatal("unbound grammar lost its exact consumer identity")
					}
					if !test.bound {
						continue
					}
					if binding.ConsumerNodeID != reference.NodeID || consumerIDs[binding.ConsumerNodeID] ||
						binding.Producer.ClauseID <= 0 || binding.Producer.ComponentNodeID <= 0 {
						t.Fatal("cost ownership lost exact consumer or component identity")
					}
					consumerIDs[binding.ConsumerNodeID] = true
					if producer.ClauseID != 0 && producer != binding.Producer {
						t.Fatal("independent conditions did not share the same actual cost component")
					}
					producer = binding.Producer
					component := document.Abilities[producer.ClauseID-1].CostSyntax.Components[producer.ComponentNodeID-1]
					if component.PaidSubject == nil || *component.PaidSubject != producer {
						t.Fatal("bound consumer lacks an exact actual-cost publisher")
					}
				}
			}
			if found != test.consumers {
				t.Fatalf("subject count = %d, want %d", found, test.consumers)
			}
		})
	}
}

func TestPaidCostBindingPreservesResolvingActionDomain(t *testing.T) {
	document, diagnostics := Parse("Sacrifice a creature. If the sacrificed creature was a Human, draw a card.", Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %v", diagnostics)
	}
	ability := document.Abilities[0]
	condition := ability.ConditionClauses[0]
	if condition.ObjectBinding != ConditionObjectBindingPriorInstructionResult || condition.HasSubjectSpan {
		t.Fatalf("resolving action was rewritten as a cost subject: %#v", condition)
	}
	for _, reference := range ability.SemanticReferences {
		if reference.Kind == ReferencePaidCostSubject {
			t.Fatal("independent action retained a spurious paid-cost reference")
		}
	}
}
