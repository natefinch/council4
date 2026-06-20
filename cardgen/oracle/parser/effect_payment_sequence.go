package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// recognizeControllerMandatoryPaymentSequence folds the exact two-sentence
// "pay {N}. If you don't, you lose the game." Pact form onto its consequence
// effect. Unlike the optional "you may pay" recognizers the payment is
// mandatory wording; downstream lowering offers it as a resolution payment
// whose failure triggers the consequence (CR 104.3a). The first sentence
// remains explicit typed syntax for source coverage.
func recognizeControllerMandatoryPaymentSequence(ability *Ability) {
	if ability.Kind != AbilityTriggered || len(ability.Sentences) < 2 {
		return
	}
	for i := 2; i < len(ability.Sentences); i++ {
		if len(semanticEffectTokens(ability.Sentences[i].Tokens)) != 0 {
			return
		}
	}
	paymentSentence := &ability.Sentences[0]
	consequenceSentence := &ability.Sentences[1]
	paymentTokens := semanticEffectTokens(paymentSentence.Tokens)
	if len(paymentTokens) < 3 ||
		!effectWordsAt(paymentTokens, 0, "pay") {
		return
	}
	manaCost, paymentEnd, ok := parseKeywordManaCost(paymentTokens, 1)
	if !ok || paymentEnd != len(paymentTokens)-1 || paymentTokens[paymentEnd].Kind != shared.Period {
		return
	}

	consequenceTokens := semanticEffectTokens(consequenceSentence.Tokens)
	if len(consequenceTokens) < 5 ||
		!effectWordsAt(consequenceTokens, 0, "if", "you", "don't") ||
		consequenceTokens[3].Kind != shared.Comma ||
		len(consequenceSentence.Effects) != 1 {
		return
	}
	boundary, ok := conditionBoundaryAt(ability.ConditionBoundaries, consequenceTokens[0].Span.Start)
	if !ok || boundary.Kind != ConditionIntroIf {
		return
	}

	effect := consequenceSentence.Effects[0]
	consequence := consequenceTokens[4:]
	if effect.Kind != EffectLoseGame ||
		effect.Context != EffectContextController {
		return
	}
	effect.Tokens = cloneTokens(consequence)
	effect.ClauseSpan = shared.SpanOf(consequence)
	effect.Negated = false
	effect.HasUnrecognizedSibling = false
	effect.Payment = EffectPaymentSyntax{
		Span:                   shared.SpanOf(paymentTokens),
		Form:                   EffectPaymentFormMayPayThenIfDoesNot,
		Payer:                  EffectPaymentPayerController,
		ManaCost:               manaCost,
		FailureConditionNodeID: boundary.NodeID,
	}
	effect.Exact = exactEffectSyntax(&effect)
	if !effect.Exact {
		return
	}

	paymentSentence.PaymentPrelude = &effect.Payment
	consequenceSentence.Effects[0] = effect
}

func recognizeControllerOptionalPaymentSequence(ability *Ability) {
	if ability.Kind != AbilityTriggered || len(ability.Sentences) < 2 {
		return
	}
	for i := 2; i < len(ability.Sentences); i++ {
		if len(semanticEffectTokens(ability.Sentences[i].Tokens)) != 0 {
			return
		}
	}
	paymentSentence := &ability.Sentences[0]
	consequenceSentence := &ability.Sentences[1]
	paymentTokens := semanticEffectTokens(paymentSentence.Tokens)
	if len(paymentTokens) < 5 ||
		!effectWordsAt(paymentTokens, 0, "you", "may", "pay") {
		return
	}
	manaCost, paymentEnd, ok := parseKeywordManaCost(paymentTokens, 3)
	if !ok || paymentEnd != len(paymentTokens)-1 || paymentTokens[paymentEnd].Kind != shared.Period {
		return
	}

	consequenceTokens := semanticEffectTokens(consequenceSentence.Tokens)
	if len(consequenceTokens) < 7 ||
		!effectWordsAt(consequenceTokens, 0, "if", "you", "do") ||
		consequenceTokens[3].Kind != shared.Comma ||
		len(consequenceSentence.Effects) != 1 {
		return
	}
	boundary, ok := conditionBoundaryAt(ability.ConditionBoundaries, consequenceTokens[0].Span.Start)
	if !ok || boundary.Kind != ConditionIntroIf {
		return
	}

	effect := consequenceSentence.Effects[0]
	consequence := consequenceTokens[4:]
	if effect.Context != EffectContextController ||
		effect.VerbSpan.Start != consequence[0].Span.Start {
		return
	}
	effect.Tokens = cloneTokens(consequence)
	effect.ClauseSpan = shared.SpanOf(consequence)
	effect.Negated = false
	effect.HasUnrecognizedSibling = false
	effect.Payment = EffectPaymentSyntax{
		Span:                   shared.SpanOf(paymentTokens),
		Form:                   EffectPaymentFormMayPayThenIfDo,
		Payer:                  EffectPaymentPayerController,
		ManaCost:               manaCost,
		SuccessConditionNodeID: boundary.NodeID,
	}
	effect.Exact = exactEffectSyntax(&effect)
	if !effect.Exact {
		return
	}

	paymentSentence.PaymentPrelude = &effect.Payment
	consequenceSentence.Effects[0] = effect
	ability.Optional = false
	ability.OptionalSpan = shared.Span{}
}

// recognizeEventPlayerOptionalPaymentSequence folds the exact two-sentence
// "that player may pay {N}. If the player doesn't, EFFECT." form onto its
// consequence effect. The first sentence remains explicit typed syntax for
// source coverage; downstream compilation receives one payment-bearing effect.
func recognizeEventPlayerOptionalPaymentSequence(ability *Ability) {
	if ability.Kind != AbilityTriggered || len(ability.Sentences) < 2 {
		return
	}
	for i := 2; i < len(ability.Sentences); i++ {
		if len(semanticEffectTokens(ability.Sentences[i].Tokens)) != 0 {
			return
		}
	}
	paymentSentence := &ability.Sentences[0]
	consequenceSentence := &ability.Sentences[1]
	paymentTokens := semanticEffectTokens(paymentSentence.Tokens)
	if len(paymentTokens) < 6 ||
		!effectWordsAt(paymentTokens, 0, "that", "player", "may", "pay") {
		return
	}
	manaCost, paymentEnd, ok := parseKeywordManaCost(paymentTokens, 4)
	if !ok || paymentEnd != len(paymentTokens)-1 || paymentTokens[paymentEnd].Kind != shared.Period {
		return
	}

	consequenceTokens := semanticEffectTokens(consequenceSentence.Tokens)
	if len(consequenceTokens) < 8 ||
		!effectWordsAt(consequenceTokens, 0, "if", "the", "player", "doesn't") ||
		consequenceTokens[4].Kind != shared.Comma ||
		len(consequenceSentence.Effects) != 1 {
		return
	}
	boundary, ok := conditionBoundaryAt(ability.ConditionBoundaries, consequenceTokens[0].Span.Start)
	if !ok || boundary.Kind != ConditionIntroIf {
		return
	}

	effect := consequenceSentence.Effects[0]
	consequence := consequenceTokens[6:]
	if effect.Context != EffectContextController ||
		effect.VerbSpan.Start != consequence[0].Span.Start {
		return
	}
	effect.Tokens = cloneTokens(consequence)
	effect.ClauseSpan = shared.SpanOf(consequence)
	effect.Negated = false
	effect.HasUnrecognizedSibling = false
	effect.Payment = EffectPaymentSyntax{
		Span:                   shared.SpanOf(paymentTokens),
		Form:                   EffectPaymentFormMayPayThenIfDoesNot,
		Payer:                  EffectPaymentPayerEventPlayer,
		ManaCost:               manaCost,
		FailureConditionNodeID: boundary.NodeID,
	}
	effect.Exact = exactEffectSyntax(&effect)
	if !effect.Exact {
		return
	}

	conditionSpan := shared.SpanOf(consequenceTokens[:4])
	ability.ConditionClauses = append(ability.ConditionClauses, ConditionClause{
		Span:      conditionSpan,
		Intro:     ConditionIntroIf,
		Predicate: ConditionPredicateEventPlayerDoesNotPay,
	})
	paymentSentence.PaymentPrelude = &effect.Payment
	consequenceSentence.Effects[0] = effect
}
