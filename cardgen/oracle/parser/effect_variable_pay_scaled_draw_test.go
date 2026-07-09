package parser

import (
	"testing"
)

// TestParseControllerVariablePayScaledDraw proves Well of Lost Dreams folds its
// "you may pay {X}, where X is less than or equal to the amount of life you
// gained." offer onto the "If you do, draw X cards." consequence as an
// EffectPaymentFormMayPayVariableUpTo payment. The payment is paid by the
// controller, carries the triggering life-change quantity as its bound in
// GenericManaAmount and no fixed ManaCost, links to the "If you do" gate through
// SuccessConditionNodeID, and leaves the draw amount as the chosen variable X.
func TestParseControllerVariablePayScaledDraw(t *testing.T) {
	t.Parallel()
	oracle := "Whenever you gain life, you may pay {X}, where X is less than or equal to the amount of life you gained. If you do, draw X cards."
	document, diagnostics := Parse(oracle, Context{CardName: "Well of Lost Dreams"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := document.Abilities[0]
	if ability.Kind != AbilityTriggered || ability.Optional {
		t.Fatalf("ability = kind %v optional %v", ability.Kind, ability.Optional)
	}
	if len(ability.Sentences) != 2 {
		t.Fatalf("sentences = %d, want 2", len(ability.Sentences))
	}
	offer := ability.Sentences[0]
	consequence := ability.Sentences[1]
	if offer.PaymentPrelude == nil {
		t.Fatalf("payment offer sentence has no PaymentPrelude: %#v", offer)
	}
	if len(offer.Effects) != 0 {
		t.Fatalf("payment offer sentence carries effects: %#v", offer.Effects)
	}
	if len(consequence.Effects) != 1 {
		t.Fatalf("consequence effects = %#v", consequence.Effects)
	}
	draw := consequence.Effects[0]
	if draw.Kind != EffectDraw || !draw.Exact || draw.HasUnrecognizedSibling {
		t.Fatalf("draw effect = %#v", draw)
	}
	if !draw.Amount.VariableX {
		t.Fatalf("draw amount is not the chosen variable X: %#v", draw.Amount)
	}
	payment := draw.Payment
	if payment.Form != EffectPaymentFormMayPayVariableUpTo ||
		payment.Payer != EffectPaymentPayerController {
		t.Fatalf("payment form/payer = %#v", payment)
	}
	if len(payment.ManaCost) != 0 {
		t.Fatalf("payment carries a fixed mana cost, want none: %q", payment.ManaCost.String())
	}
	if payment.GenericManaAmount.DynamicKind != EffectDynamicAmountTriggeringLifeChange {
		t.Fatalf("payment bound = %#v, want triggering life-change", payment.GenericManaAmount)
	}
	boundaries := ability.ConditionBoundaries
	if len(boundaries) != 1 || boundaries[0].Kind != ConditionIntroIf {
		t.Fatalf("condition boundaries = %#v, want one ConditionIntroIf", boundaries)
	}
	if payment.SuccessConditionNodeID != boundaries[0].NodeID {
		t.Fatalf("payment SuccessConditionNodeID = %d, want %d (the 'if you do' gate boundary)",
			payment.SuccessConditionNodeID, boundaries[0].NodeID)
	}
}

// TestParseControllerVariablePayScaledDrawFailsClosed proves the recognizer folds
// no payment onto adjacent wordings it must not claim: a fixed "where X is <count>"
// payment is not the bounded "less than or equal to" offer, an event-player "that
// player may pay {X}" offer is not the controller offer, and a non-draw
// consequence is outside the scaled-draw family. None should carry the
// EffectPaymentFormMayPayVariableUpTo form.
func TestParseControllerVariablePayScaledDrawFailsClosed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		oracle string
	}{
		{
			name:   "fixed X not bounded choice",
			oracle: "Whenever you gain life, you may pay {X}, where X is the amount of life you gained. If you do, draw X cards.",
		},
		{
			name:   "event player not controller",
			oracle: "Whenever a player gains life, that player may pay {X}, where X is less than or equal to the amount of life they gained. If they do, they draw X cards.",
		},
		{
			name:   "no less-than-or-equal bound",
			oracle: "Whenever you gain life, you may pay {X}. If you do, draw X cards.",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.oracle, Context{CardName: test.name})
			for _, ability := range document.Abilities {
				for _, sentence := range ability.Sentences {
					for _, effect := range sentence.Effects {
						if effect.Payment.Form == EffectPaymentFormMayPayVariableUpTo {
							t.Fatalf("wording unexpectedly folded a MayPayVariableUpTo payment: %#v", effect.Payment)
						}
					}
				}
			}
		})
	}
}
