package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

type EffectLifePaymentKind uint8

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
