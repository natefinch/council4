package parser

import (
	"testing"
)

// TestParseControllerOptionalEnergyPayment proves the "you may pay {E}{E}. If you
// do, ..." optional payment offer (the Kaladesh energy cycle's attack and enter
// riders, such as Thriving Rats) folds its pure energy cost onto the consequence
// effect as a single-component energy AdditionalCost.
func TestParseControllerOptionalEnergyPayment(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Whenever this creature attacks, you may pay {E}{E}. If you do, put a +1/+1 counter on it.",
		Context{CardName: "Thriving Rats"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := document.Abilities[0]
	if len(ability.Sentences) < 2 {
		t.Fatalf("sentences = %#v", ability.Sentences)
	}
	effect := ability.Sentences[1].Effects[0]
	payment := effect.Payment
	if payment.Form != EffectPaymentFormMayPayThenIfDo ||
		payment.Payer != EffectPaymentPayerController ||
		len(payment.ManaCost) != 0 ||
		payment.AdditionalCost == nil {
		t.Fatalf("payment = %#v", payment)
	}
	if len(payment.AdditionalCost.Components) != 1 {
		t.Fatalf("additional cost components = %#v", payment.AdditionalCost.Components)
	}
	component := payment.AdditionalCost.Components[0]
	if component.Kind != CostComponentEnergy ||
		!component.AmountKnown ||
		component.AmountValue != 2 {
		t.Fatalf("energy component = %#v", component)
	}
}

// TestParseControllerOptionalEnergyPaymentSingleSymbol proves the single-symbol
// "you may pay {E}." offer recognizes a one-energy AdditionalCost.
func TestParseControllerOptionalEnergyPaymentSingleSymbol(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Whenever this creature attacks, you may pay {E}. If you do, draw a card.",
		Context{CardName: "Aetherstream Leopard"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := document.Abilities[0]
	if len(ability.Sentences) < 2 {
		t.Fatalf("sentences = %#v", ability.Sentences)
	}
	payment := ability.Sentences[1].Effects[0].Payment
	if payment.AdditionalCost == nil ||
		len(payment.AdditionalCost.Components) != 1 ||
		payment.AdditionalCost.Components[0].Kind != CostComponentEnergy ||
		payment.AdditionalCost.Components[0].AmountValue != 1 {
		t.Fatalf("payment = %#v", payment)
	}
}
