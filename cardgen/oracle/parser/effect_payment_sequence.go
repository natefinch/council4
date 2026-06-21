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
	// An intervening-if condition may precede the optional payment ("... if you
	// gained life this turn, you may pay X life, ..."); the trigger frame owns
	// that condition, so drop it before matching the "you may" payment offer.
	paymentTokens = stripLeadingInterveningIfPayment(paymentTokens)
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
	var tokenPTBinding EffectDynamicAmountKind
	switch lifeCost, lifeBinding, lifeOK := controllerPayLifeDynamicCost(paymentSentence, paymentTokens[2:len(paymentTokens)-1]); {
	case payKeywordManaCostOK(paymentTokens):
		manaCost, _, _ := parseKeywordManaCost(paymentTokens, 3)
		payment.ManaCost = manaCost
	case lifeOK:
		// "you may pay X life, where X is <dynamic>" pays a rules-derived amount
		// of life as a resolution cost. The same dynamic sizes a variable "X/X"
		// token created by the consequence, so carry the binding through.
		payment.AdditionalCost = lifeCost
		tokenPTBinding = lifeBinding
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
	// Bind a variable "X/X" token's size to the same dynamic amount the pay-life
	// cost uses, so the consequence creates a token sized to the life paid.
	if tokenPTBinding != EffectDynamicAmountNone && effect.TokenPTVariableX {
		effect.TokenPTDynamic = tokenPTBinding
	}

	paymentSentence.PaymentPrelude = &effect.Payment
	consequenceSentence.Effects[0] = effect
	ability.Optional = false
	ability.OptionalSpan = shared.Span{}
}

// stripLeadingInterveningIfPayment drops a leading intervening-if condition
// ("if you gained life this turn, you may pay ...") from an optional-payment
// sentence's tokens, returning the tokens from the payment offer onward. It
// strips only when the tokens open with "if", a top-level comma separates the
// condition from the remainder, and the remainder opens with "you may"; in every
// other shape it returns the tokens unchanged so non-conditional payments and
// unrelated wording are untouched.
func stripLeadingInterveningIfPayment(tokens []shared.Token) []shared.Token {
	if len(tokens) == 0 || !effectWordsAt(tokens, 0, "if") {
		return tokens
	}
	comma := shared.TopLevelIndex(tokens, shared.Comma)
	if comma <= 0 || comma+1 >= len(tokens) {
		return tokens
	}
	rest := tokens[comma+1:]
	if !effectWordsAt(rest, 0, "you", "may") {
		return tokens
	}
	return rest
}

// payKeywordManaCostOK reports whether an optional-payment sentence's tokens form
// the "you may pay {mana}." offer: a "pay" verb at index 2 followed by a keyword
// mana cost that runs to the trailing period.
func payKeywordManaCostOK(paymentTokens []shared.Token) bool {
	if !effectWordsAt(paymentTokens, 2, "pay") {
		return false
	}
	_, paymentEnd, ok := parseKeywordManaCost(paymentTokens, 3)
	return ok && paymentEnd == len(paymentTokens)-1
}

// controllerPayLifeDynamicCost recognizes a "pay X life, where X is <dynamic>"
// resolution cost (the tokens between "you may" and the trailing period), such
// as Tivash's "pay X life, where X is the amount of life you gained this turn".
// It returns a single-component pay-life Cost carrying the recognized dynamic
// amount plus the bound dynamic kind, or ok=false for any other shape.
func controllerPayLifeDynamicCost(sentence *Sentence, costTokens []shared.Token) (*Cost, EffectDynamicAmountKind, bool) {
	if len(costTokens) < 8 ||
		!equalWord(costTokens[0], "pay") ||
		!equalWord(costTokens[1], "X") ||
		!equalWord(costTokens[2], "life") ||
		costTokens[3].Kind != shared.Comma ||
		!effectWordsAt(costTokens, 4, "where", "X", "is") {
		return nil, EffectDynamicAmountNone, false
	}
	subject, ok := parseDynamicLifeChangedThisTurnSubject(costTokens, 7)
	if !ok || subject.end != len(costTokens) {
		return nil, EffectDynamicAmountNone, false
	}
	dynamic, ok := payLifeDynamicForKind(subject.amount.DynamicKind)
	if !ok {
		return nil, EffectDynamicAmountNone, false
	}
	span := shared.SpanOf(costTokens)
	text := shared.SliceSpan(sentence.Text, costRelativeSpan(span, sentence.Span.Start.Offset))
	cost := &Cost{
		Span: span,
		Text: text,
		Components: []CostComponent{{
			Kind:           CostComponentPayLife,
			Span:           span,
			Text:           text,
			PayLifeDynamic: dynamic,
		}},
	}
	return cost, subject.amount.DynamicKind, true
}

// payLifeDynamicForKind maps a recognized effect dynamic-amount kind onto the
// pay-life cost vocabulary, for amounts a life cost can resolve. It fails closed
// for kinds with no pay-life representation.
func payLifeDynamicForKind(kind EffectDynamicAmountKind) (PayLifeDynamicAmount, bool) {
	switch kind {
	case EffectDynamicAmountLifeGainedThisTurn:
		return PayLifeDynamicLifeGainedThisTurn, true
	default:
		return PayLifeDynamicAmountNone, false
	}
}

// a "you may pay {mana}. If you do, <body>." triggered ability whose consequence
// body begins with a non-controller-context effect (such as the Extort drain
// "each opponent loses 1 life and you gain that much life"). The controller
// recognizer above owns consequences that begin with a controller-context,
// verb-initial effect; this recognizer captures the complementary family without
// rewriting the consequence tokens, so the full multi-effect body still lowers
// through the standard content path while the folded payment gates it. Only the
// mana form is recognized here; non-mana optional costs remain the controller
// recognizer's domain.
func recognizeOptionalManaPaymentBenefitSequence(ability *Ability) {
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
		!effectWordsAt(paymentTokens, 0, "you", "may", "pay") ||
		paymentTokens[len(paymentTokens)-1].Kind != shared.Period {
		return
	}
	manaCost, paymentEnd, manaOK := parseKeywordManaCost(paymentTokens, 3)
	if !manaOK || paymentEnd != len(paymentTokens)-1 {
		return
	}

	consequenceTokens := semanticEffectTokens(consequenceSentence.Tokens)
	if len(consequenceTokens) < 5 ||
		!effectWordsAt(consequenceTokens, 0, "if", "you", "do") ||
		consequenceTokens[3].Kind != shared.Comma ||
		len(consequenceSentence.Effects) == 0 {
		return
	}
	firstEffect := &consequenceSentence.Effects[0]
	if firstEffect.Context == EffectContextController ||
		firstEffect.Payment.Form != EffectPaymentFormUnknown {
		return
	}
	boundary, ok := conditionBoundaryAt(ability.ConditionBoundaries, consequenceTokens[0].Span.Start)
	if !ok || boundary.Kind != ConditionIntroIf {
		return
	}

	firstEffect.Payment = EffectPaymentSyntax{
		Span:                   shared.SpanOf(paymentTokens),
		Form:                   EffectPaymentFormMayPayThenIfDo,
		Payer:                  EffectPaymentPayerController,
		ManaCost:               manaCost,
		SuccessConditionNodeID: boundary.NodeID,
	}
	paymentSentence.PaymentPrelude = &firstEffect.Payment
	ability.Optional = false
	ability.OptionalSpan = shared.Span{}
}

// parseControllerPaymentAdditionalCost recognizes the non-mana cost phrase of a// "you may <cost>. If you do, ..." controller payment (the tokens between "you
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
