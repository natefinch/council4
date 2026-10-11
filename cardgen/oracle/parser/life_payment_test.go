package parser

import "testing"

func TestLifePaymentIsNotLifeLoss(t *testing.T) {
	for _, tc := range []struct {
		text string
		want EffectLifePaymentKind
	}{
		{"You may pay 2 life.", EffectLifePaymentOptional},
		{"Pay 2 life.", EffectLifePaymentRequired},
		{"You may lose 2 life.", EffectLifePaymentNone},
		{"You lose 2 life.", EffectLifePaymentNone},
		{"You gain 2 life.", EffectLifePaymentNone},
	} {
		t.Run(tc.text, func(t *testing.T) {
			document, diagnostics := Parse(tc.text, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 || len(document.Abilities) != 1 ||
				len(document.Abilities[0].Sentences[0].Effects) != 1 {
				t.Fatalf("diagnostics=%#v document=%#v", diagnostics, document)
			}
			if got := document.Abilities[0].Sentences[0].Effects[0].LifePayment; got != tc.want {
				t.Fatalf("payment kind=%v, want %v", got, tc.want)
			}
		})
	}
}
