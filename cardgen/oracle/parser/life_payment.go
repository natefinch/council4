package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// EffectLifePaymentKind distinguishes life payments from ordinary life loss.
type EffectLifePaymentKind uint8

// Life payments may require a player's acceptance as well as affordability.
const (
	EffectLifePaymentNone EffectLifePaymentKind = iota
	EffectLifePaymentRequired
	EffectLifePaymentOptional
)

func effectLifePayment(kind EffectKind, optional bool, tokens []shared.Token, index int) EffectLifePaymentKind {
	if kind != EffectLose || !payLifeVerbAt(tokens, index) {
		return EffectLifePaymentNone
	}
	if optional {
		return EffectLifePaymentOptional
	}
	return EffectLifePaymentRequired
}
