package parser

import (
	"slices"
	"strconv"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/cost"
)

// recognizedUnlessCost carries an "<effect> unless you <non-mana cost>"
// controller resolution payment folded out of a single sentence. unlessIndex is
// the token index of the "unless" introducer; the cost verbs that follow it are
// excluded from effect segmentation so the gated effect (e.g. the source
// sacrifice) parses as a single payment-bearing effect rather than a spurious
// two-effect sequence.
type recognizedUnlessCost struct {
	ok          bool
	payment     EffectPaymentSyntax
	unlessIndex int
}

// recognizeUnlessTargetPlayerPayLife recognizes a trailing "unless target
// player/opponent pays N life" payment and folds it onto the preceding effect.
// The targeted player remains an ordinary ability target; the payment carries a
// fixed pay-life additional cost so lowering can offer it to that target and gate
// the preceding effect on the payment result.
func recognizeUnlessTargetPlayerPayLife(sentence Sentence, tokens []shared.Token) recognizedUnlessCost {
	for i := 0; i+6 < len(tokens); i++ {
		if !equalWord(tokens[i], "unless") ||
			!equalWord(tokens[i+1], "target") ||
			(!equalWord(tokens[i+2], "player") && !equalWord(tokens[i+2], "opponent")) ||
			!equalWord(tokens[i+3], "pays") ||
			tokens[i+4].Kind != shared.Integer ||
			!equalWord(tokens[i+5], "life") ||
			tokens[i+6].Kind != shared.Period ||
			i+7 != len(tokens) {
			continue
		}
		amount, err := strconv.Atoi(tokens[i+4].Text)
		if err != nil || amount <= 0 {
			return recognizedUnlessCost{}
		}
		costTokens := tokens[i+4 : i+6]
		costSpan := shared.SpanOf(costTokens)
		costText := shared.SliceSpan(sentence.Text, costRelativeSpan(costSpan, sentence.Span.Start.Offset))
		additionalCost := Cost{
			Span: costSpan,
			Text: costText,
			Components: []CostComponent{{
				Kind:        CostComponentPayLife,
				Span:        costSpan,
				Text:        costText,
				AmountValue: amount,
				AmountKnown: true,
			}},
		}
		return recognizedUnlessCost{
			ok:          true,
			unlessIndex: i,
			payment: EffectPaymentSyntax{
				Span:           shared.SpanOf(tokens[i : i+6]),
				Form:           EffectPaymentFormUnless,
				Payer:          EffectPaymentPayerTargetPlayer,
				AdditionalCost: &additionalCost,
			},
		}
	}
	return recognizedUnlessCost{}
}

// recognizeUnlessControllerAdditionalCost recognizes the trailing "unless you
// <cost>" controller payment of a single sentence whose cost is a non-mana
// additional cost ("discard a card", "sacrifice another creature", "exile a card
// from your graveyard"). The mana and life "unless you pay ..." forms stay with
// parseEffectPayment, so this fails closed when the clause begins with "pay". The
// recognized cost is folded onto the gated effect's Payment as an AdditionalCost
// and its verbs are dropped from segmentation by the caller.
func recognizeUnlessControllerAdditionalCost(sentence Sentence, tokens []shared.Token, atoms Atoms) recognizedUnlessCost {
	for i := 0; i+2 < len(tokens); i++ {
		if !equalWord(tokens[i], "unless") || !equalWord(tokens[i+1], "you") {
			continue
		}
		// The mana/life "unless you pay ..." forms are owned by
		// parseEffectPayment; leave them untouched.
		if equalWord(tokens[i+2], "pay") {
			return recognizedUnlessCost{}
		}
		end := len(tokens)
		if tokens[end-1].Kind == shared.Period {
			end--
		}
		costTokens := tokens[i+2 : end]
		if len(costTokens) == 0 {
			return recognizedUnlessCost{}
		}
		span := shared.SpanOf(costTokens)
		phrase := Phrase{
			Span:   span,
			Text:   shared.SliceSpan(sentence.Text, costRelativeSpan(span, sentence.Span.Start.Offset)),
			Tokens: cloneTokens(costTokens),
		}
		parsed := parseCost(phrase, AbilityActivated, atoms)
		if !controllerAdditionalCostComponents(parsed) {
			return recognizedUnlessCost{}
		}
		return recognizedUnlessCost{
			ok:          true,
			unlessIndex: i,
			payment: EffectPaymentSyntax{
				Span:           shared.SpanOf(tokens[i:end]),
				Form:           EffectPaymentFormUnless,
				Payer:          EffectPaymentPayerController,
				AdditionalCost: &parsed,
			},
		}
	}
	return recognizedUnlessCost{}
}

// controllerAdditionalCostComponents reports whether a parsed cost is a usable
// non-mana resolution payment: it has at least one component and every component
// is a non-mana, non-loyalty, non-tap payment the downstream layers can lower.
func controllerAdditionalCostComponents(parsed Cost) bool {
	if len(parsed.Components) == 0 {
		return false
	}
	for i := range parsed.Components {
		switch parsed.Components[i].Kind {
		case CostComponentUnknown,
			CostComponentMana,
			CostComponentLoyalty,
			CostComponentTap,
			CostComponentUntap:
			return false
		default:
		}
	}
	return true
}

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

// resolvingDoConsequenceAt reports whether the consequence sentence of an
// optional-payment ability opens with a resolving "the prior optional action
// happened" gate. Two wordings express it: the inline "If you do," rider and the
// reflexive "When you do," triggered preamble. Both gate the trailing effect on
// a preceding "you may pay" offer having been accepted, and the parser records
// each as the same ConditionIntroIf boundary (see conditionIntroAt's reflexive
// case), so the payment recognizers fold either form onto the offer identically.
func resolvingDoConsequenceAt(tokens []shared.Token) bool {
	return effectWordsAt(tokens, 0, "if", "you", "do") ||
		effectWordsAt(tokens, 0, "when", "you", "do")
}

func recognizeControllerOptionalPaymentSequence(ability *Ability) {
	if ability.Kind != AbilityTriggered || len(ability.Sentences) < 2 {
		return
	}
	// Locate the consequence ("if you do, ...") and the payment offer. The
	// consequence is the sentence that opens with the resolving "if you do,"/"when
	// you do," gate; the payment offer is the sentence immediately before it. Any
	// sentences before the payment are mandatory effects performed before the
	// optional payment ("mill three cards. Then you may pay ... If you do, ...")
	// and are left intact for the ordered sequence lowering. Any sentences after
	// the consequence continue its gated body ("... put two +1/+1 counters ... on
	// target attacking creature. It becomes an Angel ..." — Guide of Souls) and
	// likewise stay intact; the reflexive repackaging in lowering carries the
	// whole gated body, so the payment is folded only onto the first consequence
	// effect. The two-sentence form is the special case where the payment offer is
	// the first sentence and there are no leading effects.
	consequenceIndex := resolvingDoConsequenceIndex(ability.Sentences)
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
		!resolvingDoConsequenceAt(consequenceTokens) ||
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
	case controllerPayEnergyOK(paymentTokens):
		// "you may pay {E}{E}." pays a pure energy resolution cost (the Kaladesh
		// energy cycle's attack and enter riders, such as Thriving Rats' "you
		// may pay {E}{E}. If you do, put a +1/+1 counter on it."). Unlike the
		// "sacrifice"/"discard" optional costs, "pay {E}" is not an effect verb,
		// so the ordered optional path cannot fold it; capture it directly as the
		// AdditionalCost for both single- and multi-effect consequences.
		energyCost, ok := parseControllerPaymentAdditionalCost(paymentSentence, paymentTokens[2:len(paymentTokens)-1], ability)
		if !ok {
			return
		}
		payment.AdditionalCost = energyCost
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

// resolvingDoConsequenceIndex returns the index of the sentence that opens the
// resolving consequence of an optional-payment ability — the first sentence
// (after the payment offer) whose semantic tokens begin with the "if you do,"/
// "when you do," gate. This is the sentence carrying the effect the payment is
// folded onto. A single-sentence consequence is the last semantic sentence, so
// this coincides with lastSemanticSentenceIndex for those abilities; a
// multi-sentence consequence ("... put counters ... on target attacking
// creature. It becomes an Angel ..." — Guide of Souls) keeps the trailing
// sentences as part of the same gated body rather than mistaking the last one
// for the consequence. Returns -1 when no sentence opens with the gate.
func resolvingDoConsequenceIndex(sentences []Sentence) int {
	for i := 1; i < len(sentences); i++ {
		tokens := semanticEffectTokens(sentences[i].Tokens)
		if len(tokens) != 0 && resolvingDoConsequenceAt(tokens) {
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

// controllerPayEnergyOK reports whether an optional-payment offer's tokens form
// the pure energy cost "you may pay {E}{E}." paymentTokens are the offer's
// semantic tokens beginning at "you may", so "pay" is at index 2 and the energy
// symbols run up to the trailing period. A missing "pay", an empty symbol run, or
// any non-energy token before the period fails closed so the mana, combined, and
// life payment forms keep their own "pay" offers.
func controllerPayEnergyOK(paymentTokens []shared.Token) bool {
	if !effectWordsAt(paymentTokens, 2, "pay") || len(paymentTokens) < 5 {
		return false
	}
	symbols := paymentTokens[3 : len(paymentTokens)-1]
	return len(symbols) > 0 && allEnergySymbols(symbols)
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
		!resolvingDoConsequenceAt(consequenceTokens) ||
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
	// A single-effect consequence ("If you do, target player loses 1 life.") was
	// demoted to non-exact because the unrecognized "you may pay" offer counted as
	// an unrecognized sibling of the lone consequence effect (see the
	// currentEffects==1 demotion in classifyEffectSyntax). Now that the offer is
	// folded into the payment, the offer is recognized, so the consequence is no
	// longer accompanied by an unrecognized sibling; restore its exact syntax so a
	// targeted rider ("target player loses 1 life", "this enchantment deals 1
	// damage to any target") lowers as a gated single effect. Multi-effect
	// consequences keep their per-effect classification untouched.
	if len(consequenceSentence.Effects) == 1 {
		firstEffect.HasUnrecognizedSibling = false
		firstEffect.Exact = exactEffectSyntax(firstEffect)
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

// recognizeDefendingPlayerOptionalPaymentSequence folds the exact two-sentence
// "defending player may pay {N}. If that player doesn't, <consequence>." form
// (Shrouded Serpent) onto its consequence effect. It is the defending-player
// counterpart of recognizeEventPlayerOptionalPaymentSequence: the player being
// attacked is offered the payment, and when they decline, the consequence
// resolves. The payment records Payer EffectPaymentPayerDefendingPlayer and the
// failure-gate NodeID, and the parser appends the shared
// ConditionPredicateDefendingPlayerDoesNotPay clause so the downstream lowering
// pairs the offer with the gate.
//
// The consequence body is re-isolated to everything after the "if that player
// doesn't," introducer and re-checked for exactness with its references and
// subject references narrowed to that body span, so the "that player" payer
// reference of the gate does not leak into the folded effect. Any consequence
// the parser cannot reconstruct exactly on its own fails closed, leaving the
// unfolded two-sentence form for downstream diagnostics.
func recognizeDefendingPlayerOptionalPaymentSequence(ability *Ability) {
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
		!effectWordsAt(paymentTokens, 0, "defending", "player", "may", "pay") {
		return
	}
	manaCost, paymentEnd, ok := parseKeywordManaCost(paymentTokens, 4)
	if !ok || paymentEnd != len(paymentTokens)-1 || paymentTokens[paymentEnd].Kind != shared.Period {
		return
	}

	consequenceTokens := semanticEffectTokens(consequenceSentence.Tokens)
	if len(consequenceTokens) < 6 ||
		!effectWordsAt(consequenceTokens, 0, "if", "that", "player", "doesn't") ||
		consequenceTokens[4].Kind != shared.Comma ||
		len(consequenceSentence.Effects) != 1 {
		return
	}
	boundary, ok := conditionBoundaryAt(ability.ConditionBoundaries, consequenceTokens[0].Span.Start)
	if !ok || boundary.Kind != ConditionIntroIf {
		return
	}

	// The consequence body is everything after the "if that player doesn't,"
	// introducer. It is subject-led (a source-context "this creature can't be
	// blocked this turn."), so the body begins at its own subject; isolate it and
	// narrow the effect's references to that span so the gate's "that player"
	// payer reference is dropped before exactness is re-checked.
	consequence := consequenceTokens[5:]
	if len(consequence) == 0 {
		return
	}
	bodySpan := shared.SpanOf(consequence)
	effect := consequenceSentence.Effects[0]
	if effect.ClauseSpan.Start.Offset > bodySpan.Start.Offset {
		return
	}
	effect.Tokens = cloneTokens(consequence)
	effect.ClauseSpan = bodySpan
	effect.References = referencesWithinSpan(effect.References, bodySpan)
	effect.SubjectReferences = referencesWithinSpan(effect.SubjectReferences, bodySpan)
	effect.Negated = false
	effect.HasUnrecognizedSibling = false
	effect.Payment = EffectPaymentSyntax{
		Span:                   shared.SpanOf(paymentTokens),
		Form:                   EffectPaymentFormMayPayThenIfDoesNot,
		Payer:                  EffectPaymentPayerDefendingPlayer,
		ManaCost:               manaCost,
		FailureConditionNodeID: boundary.NodeID,
	}
	effect.Exact = exactEffectSyntax(&effect)
	if !effect.Exact {
		return
	}

	ability.ConditionClauses = append(ability.ConditionClauses, ConditionClause{
		Span:      shared.SpanOf(consequenceTokens[:4]),
		Intro:     ConditionIntroIf,
		Predicate: ConditionPredicateDefendingPlayerDoesNotPay,
	})
	paymentSentence.PaymentPrelude = &effect.Payment
	consequenceSentence.Effects[0] = effect
}

// referencesWithinSpan returns the references whose source span is fully covered
// by span, dropping any reference that lies outside it.
func referencesWithinSpan(references []Reference, span shared.Span) []Reference {
	var result []Reference
	for _, reference := range references {
		if spanCovers(span, reference.Span) {
			result = append(result, reference)
		}
	}
	return result
}

// two-sentence "that player may pay {N}. If the player does, EFFECT." form onto
// its consequence effect, the resolving-success mirror of
// recognizeEventPlayerOptionalPaymentSequence's "If the player doesn't" failure
// gate. The event player is offered the payment; when the player pays, the
// consequence resolves. The affirmative "If the player does" gate is already a
// ConditionPredicatePriorInstructionAccepted clause (recognizePriorInstructionCondition),
// so this recognizer does not append a duplicate condition — it only folds the
// payment onto the consequence effect (Form MayPayThenIfDo, payer the event
// player) and records the shared success-gate NodeID so the compiler pairs the
// payment offer with the gate. The consequence may be verb-first ("untap the
// creature") or subject-led ("they draw a card"); either way it is folded whole
// and the downstream lowering gates it on the published payment success. It
// fails closed on any other wording, leaving the "doesn't" failure gate to the
// recognizer above.
func recognizeEventPlayerOptionalPaymentAffirmativeSequence(ability *Ability) {
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
	if len(consequenceTokens) < 6 ||
		!effectWordsAt(consequenceTokens, 0, "if", "the", "player", "does") ||
		consequenceTokens[4].Kind != shared.Comma ||
		len(consequenceSentence.Effects) != 1 {
		return
	}
	boundary, ok := conditionBoundaryAt(ability.ConditionBoundaries, consequenceTokens[0].Span.Start)
	if !ok || boundary.Kind != ConditionIntroIf {
		return
	}

	// The consequence body is everything after the "if the player does,"
	// introducer (a verb-first "untap the creature" or a subject-led "they draw
	// a card"). Isolate it so the folded effect parses exactly on its own; the
	// effect's verb must sit at or after the body start, matching the affirmative
	// controller recognizer's verb-position guard.
	consequence := consequenceTokens[5:]
	if len(consequence) == 0 {
		return
	}
	effect := consequenceSentence.Effects[0]
	if effect.VerbSpan.Start.Offset < consequence[0].Span.Start.Offset {
		return
	}
	effect.Tokens = cloneTokens(consequence)
	effect.ClauseSpan = shared.SpanOf(consequence)
	effect.Negated = false
	effect.HasUnrecognizedSibling = false
	effect.Payment = EffectPaymentSyntax{
		Span:                   shared.SpanOf(paymentTokens),
		Form:                   EffectPaymentFormMayPayThenIfDo,
		Payer:                  EffectPaymentPayerEventPlayer,
		ManaCost:               manaCost,
		SuccessConditionNodeID: boundary.NodeID,
	}
	effect.Exact = exactEffectSyntax(&effect)
	if !effect.Exact {
		return
	}
	paymentSentence.PaymentPrelude = &effect.Payment
	consequenceSentence.Effects[0] = effect
}

// recognizeEventPlayerPerCreatureUntapPayment folds the two-sentence "that
// player may choose any number of tapped <filter> creatures they control and pay
// {N} for each creature chosen this way. If the player does, untap those
// creatures." form (Dream Tides, Magnetic Mountain, Thelon's Curse) onto its
// untap consequence. It is the per-creature member of the event-player payment
// family: the upkeep player pays the fixed cost once per creature they choose
// from the folded selection and those creatures untap. The affirmative "If the
// player does" gate is already a ConditionPredicatePriorInstructionAccepted
// clause, so this recognizer only attaches the payment (Form
// EffectPaymentFormPerChosenCreature, payer the event player, the parsed creature
// filter travelling on the payment) to the untap effect and records the shared
// success-gate NodeID. It fails closed on any other wording, leaving the
// unfolded two-sentence form for downstream diagnostics.
func recognizeEventPlayerPerCreatureUntapPayment(ability *Ability) {
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
	if !effectWordsAt(paymentTokens, 0, "that", "player", "may", "choose", "any", "number", "of") {
		return
	}
	const selectionStart = 7
	// The creature filter runs from "any number of" up to the "and pay" join.
	andIdx := -1
	for i := selectionStart; i+1 < len(paymentTokens); i++ {
		if equalWord(paymentTokens[i], "and") && equalWord(paymentTokens[i+1], "pay") {
			andIdx = i
			break
		}
	}
	if andIdx <= selectionStart {
		return
	}
	selectionTokens := paymentTokens[selectionStart:andIdx]
	selection := parseSelection(selectionTokens, ability.Atoms)
	if selection.Kind != SelectionCreature ||
		!selection.Tapped ||
		selection.Controller != SelectionControllerAny ||
		selection.Other ||
		selection.Another {
		return
	}
	manaCost, costEnd, ok := parseKeywordManaCost(paymentTokens, andIdx+2)
	if !ok ||
		paymentManaCostHasVariable(manaCost) ||
		!effectWordsAt(paymentTokens, costEnd, "for", "each", "creature", "chosen", "this", "way") ||
		costEnd+6 != len(paymentTokens)-1 ||
		paymentTokens[len(paymentTokens)-1].Kind != shared.Period {
		return
	}

	consequenceTokens := semanticEffectTokens(consequenceSentence.Tokens)
	if !effectWordsAt(consequenceTokens, 0, "if", "the", "player", "does") ||
		len(consequenceTokens) < 5 ||
		consequenceTokens[4].Kind != shared.Comma ||
		!effectWordsAt(consequenceTokens, 5, "untap", "those", "creatures") ||
		len(consequenceSentence.Effects) != 1 {
		return
	}
	boundary, ok := conditionBoundaryAt(ability.ConditionBoundaries, consequenceTokens[0].Span.Start)
	if !ok || boundary.Kind != ConditionIntroIf {
		return
	}

	effect := consequenceSentence.Effects[0]
	if effect.Kind != EffectUntap {
		return
	}
	folded := selection
	effect.HasUnrecognizedSibling = false
	effect.Payment = EffectPaymentSyntax{
		Span:                   shared.SpanOf(paymentTokens),
		Form:                   EffectPaymentFormPerChosenCreature,
		Payer:                  EffectPaymentPayerEventPlayer,
		ManaCost:               manaCost,
		SuccessConditionNodeID: boundary.NodeID,
		PerCreatureSelection:   &folded,
	}
	paymentSentence.PaymentPrelude = &effect.Payment
	consequenceSentence.Effects[0] = effect
}

// paymentManaCostHasVariable reports whether a parsed fixed mana cost carries an
// {X} symbol, which the per-creature untap payment cannot represent as a fixed
// per-creature cost and so fails closed on.
func paymentManaCostHasVariable(manaCost cost.Mana) bool {
	return slices.Contains(manaCost, cost.X)
}
