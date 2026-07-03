package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// negativeResolvingGateWidth reports the token width, including the trailing
// comma, of a non-controller negative resolving gate at the start of a
// consequence sentence, or 0 when the tokens do not open with such a gate. The
// gate names the failure branch of a preceding non-controller optional action
// ("<player> may <action>. If they don't, <consequence>") in its three
// equivalent player-subject spellings — "if they don't", "if the player
// doesn't", and "if that player doesn't". They all name the same branch (the
// non-controller player did not perform the optional action), differing only in
// how they restate that player.
func negativeResolvingGateWidth(tokens []shared.Token) int {
	switch {
	case len(tokens) >= 4 &&
		effectWordsAt(tokens, 0, "if", "they", "don't") &&
		tokens[3].Kind == shared.Comma:
		return 4
	case len(tokens) >= 5 &&
		effectWordsAt(tokens, 0, "if", "the", "player", "doesn't") &&
		tokens[4].Kind == shared.Comma:
		return 5
	case len(tokens) >= 5 &&
		effectWordsAt(tokens, 0, "if", "that", "player", "doesn't") &&
		tokens[4].Kind == shared.Comma:
		return 5
	}
	return 0
}

// nonControllerOptionalActionContext reports whether an effect context names a
// single non-controller player who performs an optional action: the target
// opponent ("target opponent may sacrifice ..."), the triggering event's player
// ("that player may sacrifice ..."), or the defending player. The ability's
// controller ("you") is deliberately excluded; a controller "you may ..." offer
// is the affirmative optional-flow family the shared optional-flow planner
// already owns.
func nonControllerOptionalActionContext(context EffectContextKind) bool {
	switch context {
	case EffectContextTarget,
		EffectContextReferencedPlayer,
		EffectContextEventPlayer,
		EffectContextDefendingPlayer:
		return true
	default:
		return false
	}
}

// recognizeEventPlayerOptionalActionGate types the non-controller negative
// resolving gate that follows a non-controller optional action offer:
//
//	At the beginning of your end step, target opponent may sacrifice two
//	nonland, nontoken permanents of their choice. If they don't, you draw two
//	cards. (Rakdos, Patron of Chaos)
//
// The first sentence's optional action is performed by a non-controller player
// (the target opponent, the event player, the defending player); the second
// sentence's "if they don't" gate names the branch taken when that player
// declines the action. Unlike the event-player payment recognizers, which fold
// a "may pay" offer onto its consequence, this leaves both the action and the
// consequence as their own effects and appends only a
// ConditionPredicatePriorInstructionNotAccepted clause spanning the gate. That
// typed clause is the resolving-failure mirror of the affirmative
// "if they do"/"if the player does" gate that recognizePriorInstructionCondition
// already types: it links the consequence to the preceding optional action so
// the compiler and the text-blind lowering can gate the consequence on the
// action having been declined, without either reading the gate's wording.
//
// It runs after the payment recognizers, so a non-controller "may pay" offer
// (Smothering Tithe) has already folded its payment and never reaches here — its
// first sentence carries a PaymentPrelude and its action effect a payment form.
// It also clears the spurious negation the trailing "don't" left on the
// consequence effect, which belongs to the gate rather than to the drawn or
// created effect. It fails closed on any other shape.
func recognizeEventPlayerOptionalActionGate(ability *Ability) {
	if ability.Kind != AbilityTriggered || len(ability.Sentences) < 2 {
		return
	}
	for i := 2; i < len(ability.Sentences); i++ {
		if len(semanticEffectTokens(ability.Sentences[i].Tokens)) != 0 {
			return
		}
	}
	actionSentence := &ability.Sentences[0]
	consequenceSentence := &ability.Sentences[1]
	if actionSentence.PaymentPrelude != nil || len(actionSentence.Effects) != 1 {
		return
	}
	action := actionSentence.Effects[0]
	if !action.Optional ||
		action.Negated ||
		action.Payment.Form != EffectPaymentFormUnknown ||
		!nonControllerOptionalActionContext(action.Context) {
		return
	}

	consequenceTokens := semanticEffectTokens(consequenceSentence.Tokens)
	gateWidth := negativeResolvingGateWidth(consequenceTokens)
	if gateWidth == 0 || len(consequenceSentence.Effects) == 0 {
		return
	}
	boundary, ok := conditionBoundaryAt(ability.ConditionBoundaries, consequenceTokens[0].Span.Start)
	if !ok || boundary.Kind != ConditionIntroIf {
		return
	}

	// The gate span covers "if ... don't" without the trailing comma, matching
	// the span the event-player payment recognizer records for its own gate so
	// the condition-segment coverage scan links this clause the same way.
	conditionSpan := shared.SpanOf(consequenceTokens[:gateWidth-1])
	ability.ConditionClauses = append(ability.ConditionClauses, ConditionClause{
		Span:      conditionSpan,
		Intro:     ConditionIntroIf,
		Predicate: ConditionPredicatePriorInstructionNotAccepted,
	})

	for i := range consequenceSentence.Effects {
		if consequenceSentence.Effects[i].Negated {
			consequenceSentence.Effects[i].Negated = false
			consequenceSentence.Effects[i].Exact = exactEffectSyntax(&consequenceSentence.Effects[i])
		}
	}
}
