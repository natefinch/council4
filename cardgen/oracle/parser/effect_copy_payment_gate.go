package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/cost"
)

// recognizeResolvingCopyPaymentGate folds the payment-gated tail of the
// copy-chain family onto its copy consequence effect. The family is a resolving
// spell that performs a base effect on one target, then offers the affected
// target's controller a payment; if they pay, they may copy the spell and choose
// a new target for the copy, so the payment-gated copy chains iteratively off
// each new target. The offered cost is either mana or a non-mana resolution cost:
//
//	Return target creature to its owner's hand. Then that creature's controller
//	may pay {U}{U}. If the player does, they may copy this spell and may choose a
//	new target for that copy. (String of Disappearances — mana)
//
//	Chain Lightning deals 3 damage to any target. Then that player or that
//	permanent's controller may pay {R}{R}. If the player does, they may copy this
//	spell and may choose a new target for that copy. (Chain Lightning — mana)
//
//	Chain of Plasma deals 3 damage to any target. Then that player or that
//	permanent's controller may discard a card. If the player does, they may copy
//	this spell and may choose a new target for that copy. (Chain of Plasma —
//	non-mana: discard a card)
//
//	Return target nonland permanent to its owner's hand. Then that permanent's
//	controller may sacrifice a land of their choice. If the player does, they may
//	copy this spell and may choose a new target for that copy. (Chain of Vapor —
//	non-mana: sacrifice a land)
//
// The payment offer sentence carries no effect of its own — only the raw "that
// ... controller may <cost>." tokens — so this recognizer extracts the cost (a
// mana payment or a non-mana AdditionalCost) and the affected-target-controller
// payer and folds them onto the copy effect's Payment, linking the "If the player
// does" gate through the payment's SuccessConditionNodeID. Downstream lowering
// reads the folded payment and gate to build the resolution Pay instruction and
// its result-gated copy. The offer sentence records the payment as its
// PaymentPrelude so its tokens stay covered.
//
// It fails closed on every other wording (the unconditional copy-chain siblings
// carry the copy in the base sentence and need no payment fold), keeping the
// copy-chain family text-blind at the lowering layer.
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
	// A mana offer ("may pay {mana}") carries no effect of its own; a non-mana
	// offer ("may discard a card", "may sacrifice a land of their choice") parses
	// its cost verb as a lone optional effect on the payment sentence. Allow that
	// single cost effect through here and clear it below once the offer is folded,
	// so the copy-chain path sees only the base effect and the copy consequence.
	if len(paymentSentence.Effects) > 1 {
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
	offer, ok := matchAffectedControllerPayOffer(paymentSentence, paymentTokens, ability)
	if !ok {
		return
	}
	// The mana offer must carry no effect; the non-mana offer must carry exactly
	// the one cost effect being folded. Any other pairing is an unexpected shape,
	// so fail closed rather than drop an unrelated effect.
	if (offer.manaCost != nil) == (offer.additionalCost != nil) ||
		(offer.manaCost != nil && len(paymentSentence.Effects) != 0) ||
		(offer.additionalCost != nil && len(paymentSentence.Effects) != 1) {
		return
	}

	copyEffect.Payment = EffectPaymentSyntax{
		Span:                   shared.SpanOf(paymentSpanTokens),
		Form:                   EffectPaymentFormMayPayThenIfDo,
		Payer:                  EffectPaymentPayerAffectedTargetController,
		ManaCost:               offer.manaCost,
		AdditionalCost:         offer.additionalCost,
		SuccessConditionNodeID: boundary.NodeID,
	}
	// The folded non-mana cost effect is now represented by the copy's Payment, so
	// drop it from the payment sentence; its tokens stay covered by the
	// PaymentPrelude set below.
	paymentSentence.Effects = nil
	consequenceSentence.Effects[0] = copyEffect
	paymentSentence.PaymentPrelude = &consequenceSentence.Effects[0].Payment
}

// affectedControllerPayOffer is the folded cost of a copy-chain payment offer:
// exactly one of manaCost ("may pay {mana}") or additionalCost (a non-mana cost
// such as "may discard a card" or "may sacrifice a land of their choice") is set.
type affectedControllerPayOffer struct {
	manaCost       cost.Mana
	additionalCost *Cost
}

// matchAffectedControllerPayOffer matches the copy-chain payment offer "that
// <noun>'s controller may <cost>." or the two-subject "that player or that
// <noun>'s controller may <cost>." and returns the folded cost. The cost is
// either a mana payment ("may pay {mana}", Chain Lightning) or a non-mana
// resolution cost ("may discard a card", Chain of Plasma; "may sacrifice a land
// of their choice", Chain of Vapor / Chain of Silence). It fails closed on any
// other wording so only the exact affected-target-controller offer folds its
// payment onto the copy consequence.
func matchAffectedControllerPayOffer(sentence *Sentence, tokens []shared.Token, ability *Ability) (affectedControllerPayOffer, bool) {
	if len(tokens) < 5 || !equalWord(tokens[0], "that") {
		return affectedControllerPayOffer{}, false
	}
	mayIndex := -1
	for i := 1; i < len(tokens); i++ {
		if equalWord(tokens[i], "may") && equalWord(tokens[i-1], "controller") {
			mayIndex = i
			break
		}
	}
	if mayIndex < 2 {
		return affectedControllerPayOffer{}, false
	}
	if !validAffectedControllerSubject(tokens[:mayIndex-1]) {
		return affectedControllerPayOffer{}, false
	}
	costTokens := tokens[mayIndex+1:]
	if len(costTokens) == 0 || costTokens[len(costTokens)-1].Kind != shared.Period {
		return affectedControllerPayOffer{}, false
	}
	if equalWord(costTokens[0], "pay") {
		manaCost, end, ok := parseKeywordManaCost(costTokens, 1)
		if !ok || end != len(costTokens)-1 {
			return affectedControllerPayOffer{}, false
		}
		return affectedControllerPayOffer{manaCost: manaCost}, true
	}
	// A non-mana offer ("discard a card", "sacrifice a land of their choice")
	// reuses the shared controller-payment cost grammar, which fails closed on any
	// cost the resolution payment cannot carry. A trailing "of their choice"
	// qualifier ("sacrifice a land of their choice") only restates that the paying
	// player picks which permanent to sacrifice — the default for a self-paid
	// sacrifice cost — so drop it before parsing the bare cost phrase.
	costPhrase := stripOfTheirChoice(costTokens[:len(costTokens)-1])
	additionalCost, ok := parseControllerPaymentAdditionalCost(sentence, costPhrase, ability)
	if !ok {
		return affectedControllerPayOffer{}, false
	}
	return affectedControllerPayOffer{additionalCost: additionalCost}, true
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
