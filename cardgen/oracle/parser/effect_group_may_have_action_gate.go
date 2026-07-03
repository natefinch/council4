package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// groupMayHaveChooserScope reports the group chooser scope of a multiplayer "may
// have" offer from the offer sentence's leading tokens: "any player may have ..."
// offers every player (EffectContextEachPlayer) and "any opponent may have ..."
// offers every opponent (EffectContextEachOpponent). It fails closed for every
// other subject, so a single-chooser "target opponent may have ..." (owned by
// recognizeNonControllerMayHaveActionGate) never reaches the group family. The
// grant's context is otherwise unset for these offers, so this recognizer sets
// it to encode the scope lowerGroupMayHaveActionGate maps to a player group.
func groupMayHaveChooserScope(tokens []shared.Token) (EffectContextKind, bool) {
	for i := 0; i+3 < len(tokens); i++ {
		if !equalWord(tokens[i], "any") {
			continue
		}
		if !equalWord(tokens[i+2], "may") || !equalWord(tokens[i+3], "have") {
			continue
		}
		switch {
		case equalWord(tokens[i+1], "player"):
			return EffectContextEachPlayer, true
		case equalWord(tokens[i+1], "opponent"):
			return EffectContextEachOpponent, true
		}
	}
	return EffectContextUnknown, false
}

// groupMayHaveActionActorIsSelf reports whether a "may have" caused action's
// actor is the offer's own source object — the source card by name ("... may have
// Browbeat deal ..."), the "this object" reference, or the "it" pronoun that
// binds a source-triggered offer ("any opponent may have it deal ...", Vexing
// Devil). The runtime resolves an unset damage source to the resolving object's
// source, so an unset source deals exactly this actor's damage.
func groupMayHaveActionActorIsSelf(action *EffectSyntax) bool {
	for _, reference := range action.SubjectReferences {
		switch {
		case reference.Kind == ReferenceSelfName,
			reference.Kind == ReferenceThisObject,
			reference.Kind == ReferencePronoun && reference.Pronoun == PronounIt:
			return true
		}
	}
	return false
}

// groupMayHaveActionDamagesEachChooser reports whether a "may have" caused action
// is a fixed-magnitude "deal N damage to them" whose recipient "them" is each
// accepting group member. It fails closed on any other recipient or a
// non-fixed amount, so the group gate lowers only the damage-to-accepters shape.
func groupMayHaveActionDamagesEachChooser(action *EffectSyntax) bool {
	if action.Kind != EffectDealDamage || !action.Amount.Known || action.Amount.Value <= 0 {
		return false
	}
	recipient, ok := damageRecipientTokens(action.Tokens)
	if !ok || len(recipient) != 1 || !equalWord(recipient[0], "them") {
		return false
	}
	return true
}

// groupResolvingGateWidth reports the token width, including the trailing comma,
// of a multiplayer resolving gate at the start of a consequence sentence, and
// whether it names the affirmative or negative branch. A multiplayer "may have"
// offer's consequence resolves on whether at least one player accepted:
//
//   - "If a player does, ..." names the affirmative branch (someone accepted),
//     mapping to ConditionPredicatePriorInstructionAccepted (Vexing Devil,
//     Longhorn Firebeast).
//   - "If no one does, ..." and "If no player does, ..." name the negative
//     branch (nobody accepted), mapping to
//     ConditionPredicatePriorInstructionNotAccepted (Browbeat, Book Burning).
//
// It fails closed (width 0) on every other wording.
func groupResolvingGateWidth(tokens []shared.Token) (int, ConditionPredicateKind, bool) {
	switch {
	case len(tokens) >= 5 &&
		effectWordsAt(tokens, 0, "if", "a", "player", "does") &&
		tokens[4].Kind == shared.Comma:
		return 5, ConditionPredicatePriorInstructionAccepted, true
	case len(tokens) >= 5 &&
		effectWordsAt(tokens, 0, "if", "no", "one", "does") &&
		tokens[4].Kind == shared.Comma:
		return 5, ConditionPredicatePriorInstructionNotAccepted, true
	case len(tokens) >= 5 &&
		effectWordsAt(tokens, 0, "if", "no", "player", "does") &&
		tokens[4].Kind == shared.Comma:
		return 5, ConditionPredicatePriorInstructionNotAccepted, true
	}
	return 0, ConditionPredicateKind(""), false
}

// recognizeGroupMayHaveActionGate types the resolving gate that follows a
// multiplayer "may have" causative offer:
//
//	Any player may have Browbeat deal 5 damage to them. If no one does, target
//	player draws three cards. (Browbeat)
//
//	Any player may have Book Burning deal 6 damage to them. If no one does,
//	target player mills six cards. (Book Burning)
//
//	When this creature enters, any opponent may have it deal 4 damage to them.
//	If a player does, sacrifice this creature. (Vexing Devil)
//
// The first sentence offers a source-actor "deal N damage to them", where each
// player in the group (every player, or every opponent) is offered the damage
// in turn and "them" is each accepting player. The consequence resolves on
// whether at least one player accepted: "If a player does" is the affirmative
// branch and "If no one does" / "If no player does" the negative branch.
//
// The parser leaves the structural "have" grant and the caused action as their
// own effects so the text-blind lowering re-lowers them compositionally. This
// recognizer only makes two corrections the downstream stages cannot infer
// without reading wording:
//
//   - It sets the grant's context to the group chooser scope
//     (EffectContextEachPlayer or EffectContextEachOpponent), which
//     lowerGroupMayHaveActionGate maps to a player group.
//   - It appends the ConditionPredicatePriorInstruction{Accepted,NotAccepted}
//     clause that links the consequence to the offer's collective decision and
//     strips the gate's own subject ("no one"/"a player"/"no player") from the
//     consequence effects, along with any spurious negation the negative gate
//     leaked, so each consequence lowers as its own standalone effect.
//
// It fails closed on any other shape (for example a third semantic sentence, as
// in Breaking Point's regeneration rider).
func recognizeGroupMayHaveActionGate(ability *Ability) {
	if len(ability.Sentences) < 2 {
		return
	}
	for i := 2; i < len(ability.Sentences); i++ {
		if len(semanticEffectTokens(ability.Sentences[i].Tokens)) != 0 {
			return
		}
	}
	actionSentence := &ability.Sentences[0]
	consequenceSentence := &ability.Sentences[1]
	if actionSentence.PaymentPrelude != nil || len(actionSentence.Effects) != 2 {
		return
	}
	have := &actionSentence.Effects[0]
	action := &actionSentence.Effects[1]
	if have.Kind != EffectGrantKeyword || !have.Optional || have.Negated {
		return
	}
	if action.Span != have.Span || action.Negated {
		return
	}
	scope, ok := groupMayHaveChooserScope(semanticEffectTokens(actionSentence.Tokens))
	if !ok {
		return
	}
	if !groupMayHaveActionActorIsSelf(action) || !groupMayHaveActionDamagesEachChooser(action) {
		return
	}

	consequenceTokens := semanticEffectTokens(consequenceSentence.Tokens)
	if len(consequenceSentence.Effects) == 0 {
		return
	}
	gateWidth, predicate, ok := groupResolvingGateWidth(consequenceTokens)
	if !ok {
		return
	}
	boundary, ok := conditionBoundaryAt(ability.ConditionBoundaries, consequenceTokens[0].Span.Start)
	if !ok || boundary.Kind != ConditionIntroIf {
		return
	}

	have.Context = scope
	have.Exact = exactEffectSyntax(have)

	gateSpan := shared.SpanOf(consequenceTokens[:gateWidth-1])
	ability.ConditionClauses = append(ability.ConditionClauses, ConditionClause{
		Span:      gateSpan,
		Intro:     ConditionIntroIf,
		Predicate: predicate,
	})
	for i := range consequenceSentence.Effects {
		effect := &consequenceSentence.Effects[i]
		changed := false
		if effect.Negated {
			// A negative gate ("If no one does, ...") can leak a spurious
			// negation onto the consequence effect; the gate itself, not the
			// consequence, carries the negation.
			effect.Negated = false
			changed = true
		}
		// The gate's own subject ("no one"/"a player"/"no player" in "If no one
		// does, ...") lives in the gate span and leaks into the consequence
		// effect's references. Drop the gate-span references so the consequence
		// reconstructs and lowers as its own standalone effect.
		if refs := referencesOutsideGateSpan(effect.References, gateSpan); len(refs) != len(effect.References) {
			effect.References = refs
			changed = true
		}
		if subs := referencesOutsideGateSpan(effect.SubjectReferences, gateSpan); len(subs) != len(effect.SubjectReferences) {
			effect.SubjectReferences = subs
			changed = true
		}
		if changed {
			effect.Exact = exactEffectSyntax(effect)
		}
	}
}
