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
	if len(paymentTokens) < 4 ||
		!effectWordsAt(paymentTokens, 0, "you", "may") ||
		paymentTokens[len(paymentTokens)-1].Kind != shared.Period {
		return
	}

	consequenceTokens := semanticEffectTokens(consequenceSentence.Tokens)
	if len(consequenceTokens) < 5 ||
		!effectWordsAt(consequenceTokens, 0, "if", "you", "do") ||
		consequenceTokens[3].Kind != shared.Comma ||
		len(consequenceSentence.Effects) == 0 {
		return
	}
	boundary, ok := conditionBoundaryAt(ability.ConditionBoundaries, consequenceTokens[0].Span.Start)
	if !ok || boundary.Kind != ConditionIntroIf {
		return
	}

	payment := EffectPaymentSyntax{
		Span:                   shared.SpanOf(paymentTokens),
		Form:                   EffectPaymentFormMayPayThenIfDo,
		Payer:                  EffectPaymentPayerController,
		SuccessConditionNodeID: boundary.NodeID,
	}
	switch manaCost, paymentEnd, manaOK := parseKeywordManaCost(paymentTokens, 3); {
	case effectWordsAt(paymentTokens, 2, "pay") && manaOK && paymentEnd == len(paymentTokens)-1:
		payment.ManaCost = manaCost
	case len(consequenceSentence.Effects) > 1:
		// A non-mana optional cost ("you may sacrifice a land", "you may discard
		// a card") leaves its own cost effect in the payment sentence, so a
		// single-effect consequence is already lowered by the ordered optional
		// path. Only fold the non-mana cost as a resolution payment when the
		// consequence is a multi-effect body (such as a search) that the ordered
		// path cannot merge into one gated instruction.
		parsed, ok := parseControllerPaymentAdditionalCost(paymentSentence, paymentTokens[2:len(paymentTokens)-1], ability)
		if !ok {
			return
		}
		payment.AdditionalCost = parsed
	default:
		return
	}

	// The consequence sentence may describe a single effect ("draw a card") or a
	// multi-effect body ("search ..., put ..., then shuffle") that downstream
	// lowering merges into one instruction. Fold the payment onto the first
	// consequence effect, isolating just its own tokens (dropping the leading
	// "if you do," and any following sibling effects) so it parses exactly while
	// the remaining effects stay intact for the standalone consequence lowering.
	firstEffectEnd := len(consequenceTokens)
	if len(consequenceSentence.Effects) > 1 {
		nextStart := consequenceSentence.Effects[1].ClauseSpan.Start.Offset
		for idx := 4; idx < len(consequenceTokens); idx++ {
			if consequenceTokens[idx].Span.Start.Offset >= nextStart {
				firstEffectEnd = idx
				break
			}
		}
	}
	firstEffectTokens := consequenceTokens[4:firstEffectEnd]
	if len(firstEffectTokens) == 0 {
		return
	}

	effect := consequenceSentence.Effects[0]
	if effect.Context != EffectContextController ||
		effect.VerbSpan.Start != firstEffectTokens[0].Span.Start {
		return
	}
	effect.Tokens = cloneTokens(firstEffectTokens)
	effect.ClauseSpan = shared.SpanOf(firstEffectTokens)
	effect.Negated = false
	effect.HasUnrecognizedSibling = false
	effect.Payment = payment
	effect.Exact = exactEffectSyntax(&effect)
	if !effect.Exact {
		return
	}

	paymentSentence.PaymentPrelude = &effect.Payment
	consequenceSentence.Effects[0] = effect
	ability.Optional = false
	ability.OptionalSpan = shared.Span{}
}

// parseControllerPaymentAdditionalCost recognizes the non-mana cost phrase of a
// "you may <cost>. If you do, ..." controller payment (the tokens between "you
// may" and the trailing period), such as "sacrifice a land" or "discard a card."
// It reuses the activated-ability cost grammar and fails closed unless every
// recognized component is a non-mana, non-loyalty payment the downstream layers
// can lower.
func parseControllerPaymentAdditionalCost(sentence *Sentence, costTokens []shared.Token, ability *Ability) (*Cost, bool) {
	if len(costTokens) == 0 {
		return nil, false
	}
	span := shared.SpanOf(costTokens)
	phrase := Phrase{
		Span:   span,
		Text:   shared.SliceSpan(sentence.Text, costRelativeSpan(span, sentence.Span.Start.Offset)),
		Tokens: cloneTokens(costTokens),
	}
	parsed := parseCost(phrase, ability.Kind, ability.Atoms)
	if len(parsed.Components) == 0 {
		return nil, false
	}
	for i := range parsed.Components {
		switch parsed.Components[i].Kind {
		case CostComponentUnknown,
			CostComponentMana,
			CostComponentLoyalty,
			CostComponentTap,
			CostComponentUntap:
			return nil, false
		default:
		}
	}
	return &parsed, true
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
