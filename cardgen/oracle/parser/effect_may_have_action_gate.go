package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// nonControllerMayHaveChooserContext reports whether an effect context names a
// single non-controller player who is offered a "may have" causative choice: the
// target opponent ("target opponent may have you draw ...") or the defending
// player ("defending player may have you draw ..."). The controller context ("you
// may have target player lose ...") is excluded here because a controller "may
// have" gate already types naturally — its caused action keeps its own actor and
// its "If you do" link is recognized by recognizePriorInstructionCondition — so it
// needs none of the non-controller corrections below; lowerMayHaveActionGate
// lowers both choosers.
func nonControllerMayHaveChooserContext(context EffectContextKind) bool {
	switch context {
	case EffectContextTarget, EffectContextDefendingPlayer:
		return true
	default:
		return false
	}
}

// mayHaveCausativeActorIsController reports whether the causative "may have"
// verb's actor is the controller ("... may have you draw ..."), returning false
// for a source-actor form ("... may have <SourceName> deal ..."). The second
// bool reports whether a "may have" adjacency was found at all.
func mayHaveCausativeActorIsController(tokens []shared.Token) (controller bool, found bool) {
	for i := 0; i+2 < len(tokens); i++ {
		if equalWord(tokens[i], "may") && equalWord(tokens[i+1], "have") {
			return equalWord(tokens[i+2], "you"), true
		}
	}
	return false, false
}

// mayHaveActionSubjectIsSource reports whether the caused action's subject names
// the source object ("<SourceName> deal ..." / "it deal ..."), the source-actor
// form of a "may have" causative offer.
func mayHaveActionSubjectIsSource(action *EffectSyntax) bool {
	for _, reference := range action.SubjectReferences {
		if reference.Kind == ReferenceSelfName || reference.Kind == ReferenceThisObject {
			return true
		}
	}
	return false
}

// recognizeNonControllerMayHaveActionGate types the resolving gate that follows
// a non-controller "may have" causative offer:
//
//	When this creature enters, target opponent may have you create two Lander
//	tokens. If they don't, put two +1/+1 counters on this creature.
//	(Terrapact Intimidator)
//
//	Target opponent may have Risk Factor deal 4 damage to them. If that player
//	doesn't, you draw three cards. (Risk Factor)
//
//	Whenever this creature attacks, defending player may have you draw a card.
//	If they do, untap this creature and remove it from combat. (Shakedown Heavy)
//
// The first sentence offers a causative action ("<chooser> may have <actor>
// <action>") whose decision is made by a non-controller player (the target
// opponent or defending player) even though the caused action's own actor is the
// controller ("you draw/create") or the source ("Risk Factor deal ... to them").
// The parser leaves both the structural "have" grant and the caused action as
// their own effects so the text-blind lowering strips the "have" and re-lowers
// the action compositionally; this recognizer only makes two corrections the
// downstream stages cannot infer without reading wording:
//
//   - It retags a controller-actor caused action ("... may have you draw three
//     cards") from the sentence-subject chooser context to
//     EffectContextController, because the drawer/creator is the controller, not
//     the chooser. A source-actor action ("... may have Risk Factor deal 4 to
//     them") already carries its source subject and target recipient and is left
//     untouched.
//   - For the negative branch ("If they don't"/"If that player doesn't"/"If the
//     player doesn't") it appends the ConditionPredicatePriorInstructionNotAccepted
//     clause that links the consequence to the offer having been declined and
//     clears the spurious negation the trailing "don't"/"doesn't" left on the
//     consequence effect. The affirmative branch ("If they do") is already typed
//     ConditionPredicatePriorInstructionAccepted by recognizePriorInstructionCondition.
//
// It fails closed on any other shape.
func recognizeNonControllerMayHaveActionGate(ability *Ability) {
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
	if have.Kind != EffectGrantKeyword ||
		!have.Optional ||
		have.Negated ||
		!nonControllerMayHaveChooserContext(have.Context) {
		return
	}
	if action.Span != have.Span || action.Negated {
		return
	}

	actionTokens := semanticEffectTokens(actionSentence.Tokens)
	actorIsController, found := mayHaveCausativeActorIsController(actionTokens)
	if !found {
		return
	}
	switch {
	case actorIsController:
		if action.Context != EffectContextController {
			// The caused action's actor is the controller ("you draw/create");
			// the parser attributed it to the sentence-subject chooser. Retag it
			// to the controller so the recipient is the controller, not the
			// chooser. Its spurious chooser targets and the "you" subject
			// reference are the chooser's, not the action's; the lowering scopes
			// the caused action to its own bindings, so drop them here.
			action.Context = EffectContextController
			action.Targets = nil
			action.SubjectTargets = nil
			action.SubjectReferences = nil
			action.Exact = exactEffectSyntax(action)
		}
	case mayHaveActionSubjectIsSource(action):
		if action.Context != EffectContextSource {
			// The caused action's actor is the source ("Risk Factor deal 4
			// damage to them"); the parser attributed it to the sentence-subject
			// chooser. Retag it to the source so the caused effect resolves as
			// the source acting on its own target recipient ("them" = the
			// chooser), which the single-effect lowering already models.
			action.Context = EffectContextSource
			action.Exact = exactEffectSyntax(action)
		}
	default:
		return
	}

	consequenceTokens := semanticEffectTokens(consequenceSentence.Tokens)
	if len(consequenceSentence.Effects) == 0 {
		return
	}
	if gateWidth := negativeResolvingGateWidth(consequenceTokens); gateWidth != 0 {
		boundary, ok := conditionBoundaryAt(ability.ConditionBoundaries, consequenceTokens[0].Span.Start)
		if !ok || boundary.Kind != ConditionIntroIf {
			return
		}
		gateSpan := shared.SpanOf(consequenceTokens[:gateWidth-1])
		ability.ConditionClauses = append(ability.ConditionClauses, ConditionClause{
			Span:      gateSpan,
			Intro:     ConditionIntroIf,
			Predicate: ConditionPredicatePriorInstructionNotAccepted,
		})
		for i := range consequenceSentence.Effects {
			effect := &consequenceSentence.Effects[i]
			changed := false
			if effect.Negated {
				// The trailing "don't"/"doesn't" of the gate ("If they don't,
				// ...") leaks a spurious negation onto the consequence effect;
				// the gate itself, not the consequence, carries the negation.
				effect.Negated = false
				changed = true
			}
			// The gate's own subject ("they"/"that player"/"the player" in "If
			// they don't, ...") lives in the gate span and leaks into the
			// consequence effect's references, displacing its real subject
			// ("put ... on this creature"). Drop the gate-span references so the
			// consequence reconstructs and lowers as its own standalone effect.
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
}

// referencesOutsideGateSpan returns the references whose source span is not
// covered by the resolving gate's clause span, dropping the gate subject
// ("they"/"that player") that leaks into a consequence effect.
func referencesOutsideGateSpan(references []Reference, gateSpan shared.Span) []Reference {
	outside := references[:0:0]
	for _, reference := range references {
		if !spanCovers(gateSpan, reference.Span) {
			outside = append(outside, reference)
		}
	}
	return outside
}
