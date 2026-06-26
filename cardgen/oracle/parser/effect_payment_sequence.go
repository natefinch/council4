package parser

import (
	"strconv"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/cost"
)

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
	// Locate the consequence ("if you do, ...") and the payment offer. The
	// consequence is the last sentence carrying effect tokens; the payment offer
	// is the sentence immediately before it. Any sentences before the payment are
	// mandatory effects performed before the optional payment ("mill three cards.
	// Then you may pay ... If you do, ...") and are left intact for the ordered
	// sequence lowering; the two-sentence form is the special case where the
	// payment offer is the first sentence and there are no leading effects.
	consequenceIndex := lastSemanticSentenceIndex(ability.Sentences)
	if consequenceIndex < 1 {
		return
	}
	paymentIndex := consequenceIndex - 1
	paymentSentence := &ability.Sentences[paymentIndex]
	consequenceSentence := &ability.Sentences[consequenceIndex]
	paymentTokens := semanticEffectTokens(paymentSentence.Tokens)
	// An intervening-if condition may precede the optional payment ("... if you
	// gained life this turn, you may pay X life, ..."); the trigger frame owns
	// that condition, so drop it before matching the "you may" payment offer. A
	// leading "Then" sequences the offer after the preceding effect ("... Then
	// you may pay ...") and is likewise dropped.
	paymentTokens = stripLeadingInterveningIfPayment(paymentTokens)
	// Capture the payment span before dropping a leading "Then" so the connector
	// token stays covered by a recognized semantic span; the trigger frame owns
	// any intervening-if already removed above.
	paymentSpanTokens := paymentTokens
	paymentTokens = stripLeadingThen(paymentTokens)
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
		Span:                   shared.SpanOf(paymentSpanTokens),
		Form:                   EffectPaymentFormMayPayThenIfDo,
		Payer:                  EffectPaymentPayerController,
		SuccessConditionNodeID: boundary.NodeID,
	}
	var tokenPTBinding EffectDynamicAmountKind
	switch lifeCost, lifeBinding, lifeOK := controllerPayLifeDynamicCost(paymentSentence, paymentTokens[2:len(paymentTokens)-1]); {
	case payKeywordManaCostOK(paymentTokens):
		manaCost, _, _ := parseKeywordManaCost(paymentTokens, 3)
		payment.ManaCost = manaCost
	case combinedPayOK(paymentSentence, paymentTokens):
		// "you may pay {mana} and N life." pays a combined mana + life resolution
		// cost, with the mana captured in ManaCost and the fixed life portion as
		// the AdditionalCost.
		manaCost, lifeCostCombined, _ := controllerPayManaAndLifeCost(paymentSentence, paymentTokens)
		payment.ManaCost = manaCost
		payment.AdditionalCost = lifeCostCombined
	case combinedPayManaAndAdditionalCostOK(paymentSentence, paymentTokens, ability):
		// "you may pay {mana} and <non-mana cost>." (such as Conspiracy
		// Theorist's "pay {1} and discard a card.") pays a combined mana +
		// non-mana resolution cost, with the mana captured in ManaCost and the
		// trailing cost phrase ("discard a card") as the AdditionalCost. This
		// generalizes the combined mana + life case above to any non-mana cost
		// the resolution payment can carry.
		manaCost, additionalCost, _ := controllerPayManaAndAdditionalCost(paymentSentence, paymentTokens, ability)
		payment.ManaCost = manaCost
		payment.AdditionalCost = additionalCost
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
	if effect.Context != EffectContextController {
		return
	}
	// The folded effect must begin exactly at the isolated consequence tokens so
	// it parses standalone. A verb-first controller clause ("draw a card") starts
	// at its verb, while a subject-led controller clause ("you gain 1 life")
	// starts at the "You" subject with the verb following; both begin at
	// firstEffectTokens[0], so the verb sits at or after that token. A verb
	// preceding the isolated tokens would mean the clause start was mis-located,
	// so reject it. exactEffectSyntax below validates the recomposed clause.
	if effect.VerbSpan.Start.Offset < firstEffectTokens[0].Span.Start.Offset {
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

// lastSemanticSentenceIndex returns the index of the last sentence carrying any
// semantic (non-parenthetical, non-quoted) tokens, or -1 when every sentence is
// empty of them. It locates the consequence sentence of an optional-payment
// ability while skipping trailing reminder-only or empty sentences.
func lastSemanticSentenceIndex(sentences []Sentence) int {
	for i := len(sentences) - 1; i >= 0; i-- {
		if len(semanticEffectTokens(sentences[i].Tokens)) != 0 {
			return i
		}
	}
	return -1
}

// stripLeadingThen drops a leading "then" connector from a payment offer's tokens
// ("... Then you may pay ..."), returning the tokens from "you" onward. It strips
// only when the tokens open with "then" immediately followed by "you may"; every
// other shape is returned unchanged so non-sequenced offers are untouched.
func stripLeadingThen(tokens []shared.Token) []shared.Token {
	if len(tokens) < 3 || !equalWord(tokens[0], "then") {
		return tokens
	}
	rest := tokens[1:]
	if !effectWordsAt(rest, 0, "you", "may") {
		return tokens
	}
	return rest
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

// combinedPayOK reports whether an optional-payment offer's tokens form the
// combined "you may pay {mana} and N life." cost.
func combinedPayOK(sentence *Sentence, paymentTokens []shared.Token) bool {
	_, _, ok := controllerPayManaAndLifeCost(sentence, paymentTokens)
	return ok
}

// controllerPayManaAndLifeCost recognizes the combined "pay {mana} and N life"
// resolution cost of an optional payment offer ("you may pay {1} and 3 life.").
// paymentTokens are the offer's semantic tokens beginning at "you may", so "pay"
// is at index 2 and the keyword mana cost at index 3. It returns the mana portion
// as a cost.Mana plus a single-component fixed pay-life Cost for the "N life"
// portion, or ok=false for any other shape (a missing mana part, a non-integer or
// non-positive life amount, or any token past "life" before the period).
func controllerPayManaAndLifeCost(sentence *Sentence, paymentTokens []shared.Token) (cost.Mana, *Cost, bool) {
	if !effectWordsAt(paymentTokens, 2, "pay") {
		return nil, nil, false
	}
	manaCost, idx, ok := parseKeywordManaCost(paymentTokens, 3)
	if !ok || idx+3 != len(paymentTokens)-1 {
		return nil, nil, false
	}
	if !equalWord(paymentTokens[idx], "and") ||
		paymentTokens[idx+1].Kind != shared.Integer ||
		!equalWord(paymentTokens[idx+2], "life") ||
		paymentTokens[idx+3].Kind != shared.Period {
		return nil, nil, false
	}
	amount, err := strconv.Atoi(paymentTokens[idx+1].Text)
	if err != nil || amount <= 0 {
		return nil, nil, false
	}
	lifeTokens := paymentTokens[idx+1 : idx+3]
	span := shared.SpanOf(lifeTokens)
	text := shared.SliceSpan(sentence.Text, costRelativeSpan(span, sentence.Span.Start.Offset))
	lifeCost := &Cost{
		Span: span,
		Text: text,
		Components: []CostComponent{{
			Kind:        CostComponentPayLife,
			Span:        span,
			Text:        text,
			AmountValue: amount,
			AmountKnown: true,
		}},
	}
	return manaCost, lifeCost, true
}

// combinedPayManaAndAdditionalCostOK reports whether an optional-payment offer's
// tokens form the combined "you may pay {mana} and <non-mana cost>." cost.
func combinedPayManaAndAdditionalCostOK(sentence *Sentence, paymentTokens []shared.Token, ability *Ability) bool {
	_, _, ok := controllerPayManaAndAdditionalCost(sentence, paymentTokens, ability)
	return ok
}

// controllerPayManaAndAdditionalCost recognizes the combined "pay {mana} and
// <non-mana cost>" resolution cost of an optional payment offer (such as
// Conspiracy Theorist's "you may pay {1} and discard a card."). paymentTokens are
// the offer's semantic tokens beginning at "you may", so "pay" is at index 2 and
// the keyword mana cost at index 3. It returns the mana portion as a cost.Mana
// plus the parsed non-mana Cost for the phrase after "and" (reusing the same
// activated-cost grammar that already lowers "discard a card", "sacrifice a
// land", and similar resolution payments), or ok=false for any other shape (a
// missing mana part, a missing "and" connector, an empty trailing phrase, or a
// trailing phrase that parses to a mana, loyalty, or tap/untap component).
func controllerPayManaAndAdditionalCost(sentence *Sentence, paymentTokens []shared.Token, ability *Ability) (cost.Mana, *Cost, bool) {
	if !effectWordsAt(paymentTokens, 2, "pay") {
		return nil, nil, false
	}
	manaCost, idx, ok := parseKeywordManaCost(paymentTokens, 3)
	if !ok || idx >= len(paymentTokens)-1 || !equalWord(paymentTokens[idx], "and") {
		return nil, nil, false
	}
	costTokens := paymentTokens[idx+1 : len(paymentTokens)-1]
	additionalCost, ok := parseControllerPaymentAdditionalCost(sentence, costTokens, ability)
	if !ok {
		return nil, nil, false
	}
	return manaCost, additionalCost, true
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
	lifeCost := &Cost{
		Span: span,
		Text: text,
		Components: []CostComponent{{
			Kind:           CostComponentPayLife,
			Span:           span,
			Text:           text,
			PayLifeDynamic: dynamic,
		}},
	}
	return lifeCost, subject.amount.DynamicKind, true
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
