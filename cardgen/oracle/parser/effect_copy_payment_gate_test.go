package parser

import (
	"testing"
)

// TestParseResolvingCopyPaymentGate proves the payment-gated copy-chain family
// folds its "that ... controller may <cost>." offer onto the copy consequence
// effect as a MayPayThenIfDo payment whose payer is the affected target's
// controller, linked to the "If the player does" gate. The offered cost is
// either mana ("may pay {mana}") or a single non-mana resolution cost ("may
// discard a card", "may sacrifice a land of their choice").
func TestParseResolvingCopyPaymentGate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracle     string
		mana       string
		additional bool
	}{
		{
			name:   "String of Disappearances",
			oracle: "Return target creature to its owner's hand. Then that creature's controller may pay {U}{U}. If the player does, they may copy this spell and may choose a new target for that copy.",
			mana:   "{U}{U}",
		},
		{
			name:   "Chain Lightning",
			oracle: "Chain Lightning deals 3 damage to any target. Then that player or that permanent's controller may pay {R}{R}. If the player does, they may copy this spell and may choose a new target for that copy.",
			mana:   "{R}{R}",
		},
		{
			name:   "Chain Stasis",
			oracle: "You may tap or untap target creature. Then that creature's controller may pay {2}{U}. If the player does, they may copy this spell and may choose a new target for that copy.",
			mana:   "{2}{U}",
		},
		{
			name:       "Chain of Plasma",
			oracle:     "Chain of Plasma deals 3 damage to any target. Then that player or that permanent's controller may discard a card. If the player does, they may copy this spell and may choose a new target for that copy.",
			additional: true,
		},
		{
			name:       "Chain of Vapor",
			oracle:     "Return target nonland permanent to its owner's hand. Then that permanent's controller may sacrifice a land of their choice. If the player does, they may copy this spell and may choose a new target for that copy.",
			additional: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.oracle, Context{CardName: test.name, InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			ability := document.Abilities[0]
			consequence := ability.Sentences[len(ability.Sentences)-1]
			if len(consequence.Effects) != 1 {
				t.Fatalf("consequence effects = %#v", consequence.Effects)
			}
			copyEffect := consequence.Effects[0]
			if copyEffect.Kind != EffectCopyStackObject || !copyEffect.CopyMayChooseNewTargets {
				t.Fatalf("copy effect = %#v", copyEffect)
			}
			payment := copyEffect.Payment
			if payment.Form != EffectPaymentFormMayPayThenIfDo ||
				payment.Payer != EffectPaymentPayerAffectedTargetController {
				t.Fatalf("payment = %#v", payment)
			}
			if test.additional {
				if payment.AdditionalCost == nil || len(payment.ManaCost) != 0 {
					t.Fatalf("expected non-mana AdditionalCost, got payment = %#v (mana %q)", payment, payment.ManaCost.String())
				}
			} else {
				if payment.ManaCost.String() != test.mana || payment.AdditionalCost != nil {
					t.Fatalf("payment = %#v (mana %q)", payment, payment.ManaCost.String())
				}
			}
			offer := ability.Sentences[len(ability.Sentences)-2]
			if offer.PaymentPrelude == nil {
				t.Fatalf("payment offer sentence has no PaymentPrelude: %#v", offer)
			}
			if len(offer.Effects) != 0 {
				t.Fatalf("folded payment offer sentence still carries effects: %#v", offer.Effects)
			}
		})
	}
}

// TestParseResolvingCopyPaymentGateFailsClosed proves the recognizer folds no
// payment onto wordings outside the payment-gated copy-chain family: the
// unconditional copy-chain siblings carry the copy in the base sentence with no
// payment offer and no "If the player does" gate, so their copy effect's payment
// stays unset.
func TestParseResolvingCopyPaymentGateFailsClosed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		oracle string
	}{
		{
			name:   "Chain of Acid unconditional",
			oracle: "Destroy target noncreature permanent. Then that permanent's controller may copy this spell and may choose a new target for that copy.",
		},
		{
			name:   "Barroom Brawl unconditional plural",
			oracle: "Target creature you control fights target creature the opponent to your left controls. Then that player may copy this spell and may choose new targets for the copy.",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.oracle, Context{CardName: test.name, InstantOrSorcery: true})
			ability := document.Abilities[0]
			for si := range ability.Sentences {
				for ei := range ability.Sentences[si].Effects {
					effect := ability.Sentences[si].Effects[ei]
					if effect.Kind != EffectCopyStackObject {
						continue
					}
					if effect.Payment.Payer == EffectPaymentPayerAffectedTargetController {
						t.Fatalf("copy effect unexpectedly folded an affected-controller payment: %#v", effect.Payment)
					}
				}
			}
		})
	}
}
