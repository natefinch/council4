package parser

import "testing"

func TestParseTargetOpponentPayLifeUnlessEventCardReturn(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Whenever another creature you own dies, return it to your hand unless target opponent pays 3 life.",
		Context{CardName: "Athreos, God of Passage"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := document.Abilities[0]
	if len(ability.Sentences) != 1 || len(ability.Sentences[0].Effects) != 1 {
		t.Fatalf("sentences/effects = %#v", ability.Sentences)
	}
	effect := ability.Sentences[0].Effects[0]
	if effect.Kind != EffectReturn {
		t.Fatalf("effect kind = %v, want return", effect.Kind)
	}
	if !effect.Exact {
		t.Fatalf(
			"return not exact: from=%v to=%v targets=%#v payment=%#v refs=%#v clause=%q",
			effect.FromZone,
			effect.ToZone,
			effect.Targets,
			effect.Payment,
			effect.References,
			exactEffectClauseText(&effect),
		)
	}
	if len(effect.Targets) != 1 ||
		effect.Targets[0].Selection.Kind != SelectionOpponent ||
		!effect.Targets[0].Exact {
		t.Fatalf("targets = %#v", effect.Targets)
	}
	payment := effect.Payment
	if payment.Form != EffectPaymentFormUnless ||
		payment.Payer != EffectPaymentPayerTargetPlayer ||
		payment.AdditionalCost == nil ||
		len(payment.AdditionalCost.Components) != 1 {
		t.Fatalf("payment = %#v", payment)
	}
	component := payment.AdditionalCost.Components[0]
	if component.Kind != CostComponentPayLife ||
		!component.AmountKnown ||
		component.AmountValue != 3 {
		t.Fatalf("payment component = %#v", component)
	}
	if len(payment.ManaCost) != 0 {
		t.Fatalf("unexpected mana payment = %#v", payment.ManaCost)
	}
	if len(ability.ConditionClauses) != 0 {
		t.Fatalf("conditions = %#v", ability.ConditionClauses)
	}
}
