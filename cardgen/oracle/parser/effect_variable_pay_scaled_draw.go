package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/cost"
)

// recognizeControllerVariablePayScaledDraw folds the two-sentence "you may pay
// {X}, where X is less than or equal to the amount of life you gained. If you do,
// draw X cards." body (Well of Lost Dreams) onto its draw consequence. The offer
// lets the controller choose X between zero and the triggering life-change
// quantity, pay that much generic mana, and draw the chosen X cards. Unlike the
// fixed "pay {X}, where X is <count>" forms, the "less than or equal to" wording
// makes X a bounded player choice rather than a determined amount, so the payment
// carries EffectPaymentFormMayPayVariableUpTo with the bound (the triggering
// life-change) in GenericManaAmount, and the draw amount stays the variable X.
// Downstream lowering models it as a PayRepeatedly bounded by that quantity whose
// published count sizes the draw. Any other wording, cost symbol, bound, or
// consequence leaves the unfolded body for downstream diagnostics so it fails
// closed.
func recognizeControllerVariablePayScaledDraw(ability *Ability) {
	if ability.Kind != AbilityTriggered || len(ability.Sentences) != 2 {
		return
	}
	paymentSentence := &ability.Sentences[0]
	consequenceSentence := &ability.Sentences[1]

	paymentTokens := semanticEffectTokens(paymentSentence.Tokens)
	if !effectWordsAt(paymentTokens, 0, "you", "may", "pay") {
		return
	}
	manaCost, costEnd, ok := parseKeywordManaCost(paymentTokens, 3)
	if !ok || len(manaCost) != 1 || manaCost[0] != cost.X {
		return
	}
	if costEnd >= len(paymentTokens) || paymentTokens[costEnd].Kind != shared.Comma {
		return
	}
	boundStart := costEnd + 1
	if !effectWordsAt(paymentTokens, boundStart,
		"where", "X", "is", "less", "than", "or", "equal", "to",
		"the", "amount", "of", "life", "you", "gained") {
		return
	}
	const boundLen = 14
	if boundStart+boundLen != len(paymentTokens)-1 ||
		paymentTokens[len(paymentTokens)-1].Kind != shared.Period {
		return
	}
	boundTokens := paymentTokens[boundStart+8 : boundStart+boundLen]

	consequenceTokens := semanticEffectTokens(consequenceSentence.Tokens)
	if !effectWordsAt(consequenceTokens, 0, "if", "you", "do") ||
		len(consequenceTokens) < 5 ||
		consequenceTokens[3].Kind != shared.Comma ||
		len(consequenceSentence.Effects) != 1 {
		return
	}
	boundary, ok := conditionBoundaryAt(ability.ConditionBoundaries, consequenceTokens[0].Span.Start)
	if !ok || boundary.Kind != ConditionIntroIf {
		return
	}

	effect := consequenceSentence.Effects[0]
	if effect.Kind != EffectDraw ||
		effect.Context != EffectContextController ||
		!effect.Amount.VariableX {
		return
	}

	effectTokens := consequenceTokens[4:]
	if len(effectTokens) == 0 {
		return
	}
	effect.Tokens = cloneTokens(effectTokens)
	effect.ClauseSpan = shared.SpanOf(effectTokens)
	effect.Negated = false
	effect.HasUnrecognizedSibling = false
	effect.Payment = EffectPaymentSyntax{
		Span:  shared.SpanOf(paymentTokens),
		Form:  EffectPaymentFormMayPayVariableUpTo,
		Payer: EffectPaymentPayerController,
		GenericManaAmount: EffectAmountSyntax{
			Span:        shared.SpanOf(boundTokens),
			Text:        joinedEffectText(boundTokens),
			DynamicKind: EffectDynamicAmountTriggeringLifeChange,
			Multiplier:  1,
		},
		SuccessConditionNodeID: boundary.NodeID,
	}
	effect.Exact = exactEffectSyntax(&effect)
	if !effect.Exact {
		return
	}
	consequenceSentence.Effects[0] = effect
	paymentSentence.PaymentPrelude = &consequenceSentence.Effects[0].Payment
	ability.Optional = false
	ability.OptionalSpan = shared.Span{}
}
