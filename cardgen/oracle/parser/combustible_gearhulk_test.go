package parser

import "testing"

func TestParseReferencedCardsTotalManaValue(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"When this creature enters, target opponent may have you draw three cards. If the player doesn't, you mill three cards, then this creature deals damage to that player equal to the total mana value of those cards.",
		Context{CardName: "Combustible Gearhulk"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := document.Abilities[0]
	damage := ability.Sentences[1].Effects[1]
	if damage.Kind != EffectDealDamage ||
		damage.Amount.DynamicKind != EffectDynamicAmountReferencedCardsTotalManaValue ||
		damage.Amount.ReferenceNodeID < 0 ||
		!damage.Exact {
		t.Fatalf("damage effect = %#v", damage)
	}
	if len(ability.ConditionClauses) != 1 ||
		ability.ConditionClauses[0].Predicate != ConditionPredicatePriorInstructionNotAccepted {
		t.Fatalf("condition clauses = %#v", ability.ConditionClauses)
	}
}
