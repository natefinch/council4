package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/cost"
)

// recognizeResolvingCopyPaymentGate folds the payment-gated tail of the
// copy-chain family onto its copy consequence effect. The family is a resolving
// spell that performs a base effect on one target, then offers the affected
// target's controller a mana payment; if they pay, they may copy the spell and
// choose a new target for the copy, so the payment-gated copy chains iteratively
// off each new target:
//
//	Return target creature to its owner's hand. Then that creature's controller
//	may pay {U}{U}. If the player does, they may copy this spell and may choose a
//	new target for that copy. (String of Disappearances)
//
//	Chain Lightning deals 3 damage to any target. Then that player or that
//	permanent's controller may pay {R}{R}. If the player does, they may copy this
//	spell and may choose a new target for that copy. (Chain Lightning)
//
// The payment offer sentence carries no effect of its own — only the raw "that
// ... controller may pay {mana}." tokens — so this recognizer extracts the mana
// cost and the affected-target-controller payer and folds them onto the copy
// effect's Payment, linking the "If the player does" gate through the payment's
// SuccessConditionNodeID. Downstream lowering reads the folded payment and gate
// to build the resolution Pay instruction and its result-gated copy. The offer
// sentence records the payment as its PaymentPrelude so its tokens stay covered.
//
// It fails closed on every other wording (the unconditional copy-chain siblings
// carry the copy in the base sentence and need no payment fold; non-mana payment
// siblings are not matched here), keeping the copy-chain family text-blind at the
// lowering layer.
func recognizeResolvingCopyPaymentGate(ability *Ability) {
	if ability.Kind != AbilitySpell || len(ability.Sentences) < 3 {
		return
	}
	consequenceIndex := lastSemanticSentenceIndex(ability.Sentences)
	if consequenceIndex < 2 {
		return
	}
	consequenceSentence := &ability.Sentences[consequenceIndex]
	if len(consequenceSentence.Effects) != 1 {
		return
	}
	copyEffect := consequenceSentence.Effects[0]
	if copyEffect.Kind != EffectCopyStackObject ||
		!copyEffect.CopyMayChooseNewTargets ||
		copyEffect.Payment.Form != "" {
		return
	}
	consequenceTokens := semanticEffectTokens(consequenceSentence.Tokens)
	if len(consequenceTokens) < 6 ||
		!effectWordsAt(consequenceTokens, 0, "if", "the", "player", "does") ||
		consequenceTokens[4].Kind != shared.Comma {
		return
	}
	boundary, ok := conditionBoundaryAt(ability.ConditionBoundaries, consequenceTokens[0].Span.Start)
	if !ok || boundary.Kind != ConditionIntroIf {
		return
	}

	paymentSentence := &ability.Sentences[consequenceIndex-1]
	if len(paymentSentence.Effects) != 0 {
		return
	}
	paymentTokens := semanticEffectTokens(paymentSentence.Tokens)
	paymentSpanTokens := paymentTokens
	// A leading "Then" sequences the offer after the base effect ("... Then that
	// creature's controller may pay ..."); it is not the "Then you may" form
	// stripLeadingThen handles, so drop it here while keeping it in the covered
	// payment span.
	if len(paymentTokens) > 0 && equalWord(paymentTokens[0], "then") {
		paymentTokens = paymentTokens[1:]
	}
	manaCost, ok := matchAffectedControllerPayOffer(paymentTokens)
	if !ok {
		return
	}

	copyEffect.Payment = EffectPaymentSyntax{
		Span:                   shared.SpanOf(paymentSpanTokens),
		Form:                   EffectPaymentFormMayPayThenIfDo,
		Payer:                  EffectPaymentPayerAffectedTargetController,
		ManaCost:               manaCost,
		SuccessConditionNodeID: boundary.NodeID,
	}
	consequenceSentence.Effects[0] = copyEffect
	paymentSentence.PaymentPrelude = &consequenceSentence.Effects[0].Payment
}

// matchAffectedControllerPayOffer matches the copy-chain payment offer "that
// <noun>'s controller may pay {mana}." or the two-subject "that player or that
// <noun>'s controller may pay {mana}." and returns the folded mana cost. It
// fails closed on any other wording so only the exact affected-target-controller
// offer folds its payment onto the copy consequence.
func matchAffectedControllerPayOffer(tokens []shared.Token) (cost.Mana, bool) {
	if len(tokens) < 5 || !equalWord(tokens[0], "that") {
		return nil, false
	}
	payIndex := -1
	for i := 1; i+1 < len(tokens); i++ {
		if equalWord(tokens[i], "may") && equalWord(tokens[i+1], "pay") {
			payIndex = i
			break
		}
	}
	if payIndex < 2 || !equalWord(tokens[payIndex-1], "controller") {
		return nil, false
	}
	if !validAffectedControllerSubject(tokens[:payIndex-1]) {
		return nil, false
	}
	manaCost, end, ok := parseKeywordManaCost(tokens, payIndex+2)
	if !ok || end != len(tokens)-1 || tokens[end].Kind != shared.Period {
		return nil, false
	}
	return manaCost, true
}

// validAffectedControllerSubject reports whether the subject tokens preceding
// "controller" name an affected target's controller in one of the two copy-chain
// wordings: "that <noun>'s" (the single-affected form) or "that player or that
// <noun>'s" (the damage form, where the affected target may be a player or a
// permanent). The trailing possessive noun is left unconstrained so any affected
// permanent noun ("creature's", "permanent's") matches.
func validAffectedControllerSubject(tokens []shared.Token) bool {
	switch len(tokens) {
	case 2:
		return equalWord(tokens[0], "that")
	case 5:
		return equalWord(tokens[0], "that") &&
			equalWord(tokens[1], "player") &&
			equalWord(tokens[2], "or") &&
			equalWord(tokens[3], "that")
	default:
		return false
	}
}
