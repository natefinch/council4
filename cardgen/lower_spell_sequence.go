package cardgen

import (
	"fmt"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// legacyOrderedEffectSequenceExact reports whether a two-effect body may flow
// into the generic ordered-sequence lowerer. It returns false only for the
// remaining effect-kind pairs the generic lowerer cannot yet sequence faithfully,
// diverting them to the "non-exact legacy effect pair" diagnostic so they fail
// closed rather than lower partially. Pairs the generic lowerer now handles
// correctly (such as two power/toughness modifications joined by "and", or a
// counter placement followed by "then proliferate") are no longer diverted here.
func legacyOrderedEffectSequenceExact(effects []compiler.CompiledEffect) bool {
	if len(effects) != 2 {
		return true
	}
	first, second := effects[0], effects[1]
	if first.Kind == compiler.EffectExile &&
		second.Kind == compiler.EffectReturn &&
		second.DelayedTiming != 0 {
		return referencesBindTo(second.References, compiler.ReferenceBindingPriorInstructionResult, 0)
	}
	return true
}

func lowerLinkedCounterTokenSequence(
	ctx contentCtx,
) (game.AbilityContent, *shared.Diagnostic, bool) {
	if len(ctx.content.Effects) != 2 ||
		ctx.content.Effects[0].Kind != compiler.EffectCounter ||
		ctx.content.Effects[1].Kind != compiler.EffectCreate {
		return game.AbilityContent{}, nil, false
	}
	if content, ok := lowerCounterThenTargetControllerTokenSequence(ctx); ok {
		return content, nil, true
	}
	if content, ok := lowerCounterThenSelfTokenSequence(ctx); ok {
		return content, nil, true
	}
	return game.AbilityContent{},
		unsupportedEffectSequenceDiagnostic(ctx, "structural — unsupported linked counter and token creation"),
		true
}

func lowerOrderedSequenceSpecialCase(
	cardName string,
	ctx contentCtx,
	syntax *parser.Ability,
) (game.AbilityContent, *shared.Diagnostic, bool) {
	if len(ctx.content.Modes) != 0 {
		return game.AbilityContent{},
			unsupportedEffectSequenceDiagnostic(ctx, "structural — sequence carries modal options"),
			true
	}
	if isSacrificeConditionedChosenCardsCategory(ctx.content) {
		if content, ok := lowerSacrificeConditionedReanimationSequence(ctx); ok {
			return content, nil, true
		}
		return game.AbilityContent{},
			unsupportedEffectSequenceDiagnostic(ctx, "structural — unsupported sacrifice-conditioned reanimation"),
			true
	}
	if content, ok := lowerOptionalSacrificeScaledReward(ctx); ok {
		return content, nil, true
	}
	if content, ok := lowerOptionalSacrificeReturnWithCounters(ctx); ok {
		return content, nil, true
	}
	if content, ok := lowerSacrificeThenCountSequence(ctx); ok {
		return content, nil, true
	}
	if content, ok := lowerOptionalSacrificeThenSearchSequence(ctx); ok {
		return content, nil, true
	}
	if content, ok := lowerSacrificeThenSearchSequence(ctx); ok {
		return content, nil, true
	}
	if content, ok := lowerDestroyThenSearchSequence(ctx); ok {
		return content, nil, true
	}
	if content, ok := lowerSacrificeWithInabilityFallback(ctx); ok {
		return content, nil, true
	}
	if content, ok := lowerCounterThenNextMainManaSequence(ctx); ok {
		return content, nil, true
	}
	if content, diagnostic, handled := lowerLinkedCounterTokenSequence(ctx); handled {
		return content, diagnostic, true
	}
	if content, ok := lowerCounterThenExileInstead(ctx); ok {
		return content, nil, true
	}
	if content, ok := lowerCounterThenAlternateDestination(ctx); ok {
		return content, nil, true
	}
	if content, ok := lowerSelfBlinkSequence(ctx); ok {
		return content, nil, true
	}
	if content, ok := lowerDelayedSelfBlinkSequence(ctx); ok {
		return content, nil, true
	}
	if content, ok := lowerMillThenPaidReturnSequence(cardName, ctx); ok {
		return content, nil, true
	}
	if content, ok := lowerMillThenOptionalAmongToHandSequence(ctx); ok {
		return content, nil, true
	}
	if content, ok := lowerMillThenOptionalAmongOneOfEachToBattlefield(ctx); ok {
		return content, nil, true
	}
	if content, ok := lowerMillThenPutAmongToBattlefield(ctx); ok {
		return content, nil, true
	}
	if content, ok := lowerInsteadOfThoseTokensSequence(ctx); ok {
		return content, nil, true
	}
	if content, ok := lowerConditionalDamageAmountReplacementSequence(ctx); ok {
		return content, nil, true
	}
	if content, ok := lowerMonarchExiledCardSplitSequence(ctx); ok {
		return content, nil, true
	}
	for _, target := range ctx.content.Targets {
		if _, ok := counterAbilityTargetSpec(target); ok {
			return game.AbilityContent{},
				unsupportedEffectSequenceDiagnostic(ctx, "structural — counter-spell target"),
				true
		}
	}
	// The combined-shape lowerers do not model per-effect conditions; only
	// attempt them when the sequence carries none, so a condition can never be
	// silently dropped.
	if content, ok := lowerCombinedSequenceShapes(cardName, ctx, syntax); ok {
		return content, nil, true
	}
	if content, ok := lowerLeadingSequenceThenSearch(cardName, ctx, syntax); ok {
		return content, nil, true
	}
	if !legacyOrderedEffectSequenceExact(ctx.content.Effects) {
		return game.AbilityContent{},
			unsupportedEffectSequenceDiagnostic(ctx, "structural — non-exact legacy effect pair"),
			true
	}
	return game.AbilityContent{}, nil, false
}

func lowerOrderedEffectSequence(
	cardName string,
	ctx contentCtx,
	syntax *parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	if content, diagnostic, handled := lowerOrderedSequenceSpecialCase(cardName, ctx, syntax); handled {
		return content, diagnostic
	}
	// Record whether the sequence announced a variable player-group target so a
	// later "those players each <verb>" clause can bind its anaphor to the
	// targeted players (Court of Cunning). The flag propagates to each clause's
	// context through lowerSequenceClauseContent.
	ctx.hasTargetedPlayers = sequenceHasPlayerGroupTarget(ctx.content.Targets)
	// Resolving optionality ("you may X. If you do, Y") is realized by marking
	// the optional effect's instruction Optional + PublishResult and gating the
	// "if you do" effect on that result. planOptionalFlow fails closed unless the
	// optionality forms exactly one supported pair.
	optionalFlow, ok := planOptionalFlow(ctx.content)
	if !ok {
		return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ctx, "structural — unsupported resolving optionality")
	}
	// The affirmative "if you do" clause is consumed as the optional-flow gate,
	// not as an ordinary effect-gate condition, so exclude it from the per-effect
	// condition matching (its predicate is not a supported effect-gate predicate).
	gateConditions := optionalFlowGateConditions(ctx.content.Conditions, optionalFlow)
	// Match each condition to the single effect whose clause span contains it and
	// lower it as an effect gate. Fails closed if any condition is not contained
	// in exactly one effect or is not a supported effect-gate condition.
	effectConditions, matchReason, ok := matchSequenceEffectConditions(ctx.content.Effects, gateConditions)
	if !ok {
		return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ctx, matchReason)
	}
	consumedConditions := 0
	if optionalFlow.enabled {
		consumedConditions++
	}
	if optionalFlow.elseGateCondition >= 0 {
		consumedConditions++
	}
	// Every gate condition is consumed by the matching above (which fails closed
	// unless all conditions matched and lowered). A single condition may gate
	// multiple effects of a shared-sentence group, so count conditions here
	// rather than per gated effect.
	consumedConditions += len(gateConditions)
	// "If <condition>, <create> instead." replaces the immediately preceding
	// effect when the condition holds (an either/or, not an additive effect).
	// The conditional clause is already gated on the condition by
	// effectConditions; gate the replaced (preceding) effect on the negation so
	// exactly one of the two effects runs. insteadGates maps the replaced
	// effect's index to that negated gate.
	insteadGates, ok := sequenceInsteadGates(ctx.content.Effects, effectConditions)
	if !ok {
		return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ctx, "structural — instead replacement not gatable")
	}
	// "Otherwise, <effect>." runs the else branch of the immediately preceding
	// conditional effect. The preceding effect is already gated on its condition;
	// gate the otherwise effect on the negation so exactly one branch resolves.
	otherwiseGates, ok := sequenceOtherwiseGates(ctx.content.Effects, effectConditions, optionalFlow)
	if !ok {
		return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ctx, "structural — otherwise branch not gatable")
	}
	var targets []game.TargetSpec
	var sequence []game.Instruction
	consumedTargets := 0
	consumedKeywords := 0
	consumedReferences := 0
	// oracleSpanToGameIdx maps each oracle target's Span to its first index in
	// the accumulated targets slice, recorded when the target is owned (i.e.
	// added as a new game.TargetSpec by a non-shared clause). This index is
	// looked up when an inherited shared-target clause needs to rebase its
	// sequence: the rebase offset equals the start index of the inherited
	// target rather than always 0, which is wrong when earlier effects already
	// contributed target specs before the then-joined group.
	oracleSpanToGameIdx := make(map[shared.Span]int)
	clauseSyntaxes := splitEffectSyntaxes(syntax, ctx.content.Effects)
	// clauseReasons collects every clause that fails to lower so the whole
	// sequence reports all of its blockers, not just the first. It stays empty
	// for a fully-supported sequence, so the success path below is unchanged.
	var clauseReasons []shared.Diagnostic
	for i := range ctx.content.Effects {
		effect := &ctx.content.Effects[i]
		resolvedEffect, clauseAbility := prepareSequenceClause(ctx, optionalFlow, clauseSyntaxes, i)
		effectAbility := contextForEffect(ctx, &resolvedEffect)
		// Per-effect conditions are handled by the sequence gate (effectConditions),
		// not by the individual effect lowerers, so clear the content-level
		// conditions inherited from the parent context before per-effect lowering.
		effectAbility.content.Conditions = nil
		clauseTargets := effect.Targets
		// A leading condition that shares its effect's sentence (e.g. "If this
		// spell was kicked, draw a card.") contributes its own references (the
		// "this spell" object) inside the effect's clause span, so the compiler
		// attributes them to the effect. Those references belong to the
		// condition, not the effect body, and are credited separately below via
		// conditionReferenceCount; strip them here so the per-effect lowerer sees
		// only the effect's own references. Strip against every condition (not just
		// the effect-gate conditions): an optional-flow gate the sequence consumes
		// itself ("... If that player does, they lose 2 life.") also carries an
		// anaphoric player reference ("that player") inside the consequence's
		// clause span, which would otherwise survive as a phantom second reference
		// and make the consequence's own "they" lowering fail closed.
		clauseRefs := referencesOutsideConditionSpans(effect.References, ctx.content.Conditions)
		ownedReferenceCount := len(clauseRefs)
		// A group counter-placement clause whose group filter carries a "with a
		// <kind> counter on it/them" qualifier introduces a pronoun naming each
		// filtered group member, not the prior clause's target. Drop it before
		// antecedent target binding so it does not bind to a prior target and
		// force the clause onto the single-target placement path; the group
		// lowerer represents the filter through the group selection's counter
		// requirement. ownedReferenceCount above still credits the dropped
		// pronoun as consumed.
		if groupCounterQualifierClause(effect) {
			clauseRefs = counterQualifierFilteredReferences(clauseRefs)
		}
		// Combined-shape and characteristic lowerers read the resolved effect's
		// own reference list (ctx.content.Effects[0].References) rather than the
		// clause-level References, so strip the gate-condition references there
		// too. Otherwise a kicked-condition's "this spell" object survives as a
		// phantom subject reference and the per-effect lowerer fails closed.
		effectAbility.content.Effects[0].References = slices.Clone(clauseRefs)
		var inheritedTargets []compiler.CompiledTarget
		if effect.Context == parser.EffectContextPriorSubject {
			inheritedTargets = priorSubjectTargets(ctx.content.Effects, i)
			clauseRefs = append(clauseRefs, priorSubjectReferences(ctx.content.Effects, i)...)
		}
		inheritedTargets = appendReferenceAntecedentTargets(
			inheritedTargets,
			clauseRefs,
			ctx.content.Targets,
			clauseTargets,
		)
		// Three target-handling modes:
		//   allSharedTargets: only inherited, no own — compound-mill "then draws".
		//   mixedTargets:     inherited + own — "then fights target creature" where
		//                     the inherited subject and a new object both appear.
		//   otherwise:        only own (or none) — normal independent effects.
		allSharedTargets := len(inheritedTargets) > 0 && len(clauseTargets) == 0
		mixedTargets := len(inheritedTargets) > 0 && len(clauseTargets) > 0
		switch {
		case allSharedTargets:
			effectAbility.content.Targets = inheritedTargets
		case mixedTargets:
			combined := make([]compiler.CompiledTarget, 0, len(inheritedTargets)+len(clauseTargets))
			combined = append(combined, inheritedTargets...)
			combined = append(combined, clauseTargets...)
			effectAbility.content.Targets = combined
		default:
			effectAbility.content.Targets = clauseTargets
		}
		effectAbility.content.References = clauseRefs
		localReferences, ok := localizeTargetReferences(
			effectAbility.content.References,
			ctx.content.Targets,
			effectAbility.content.Targets,
		)
		if !ok {
			if len(clauseReasons) > 0 {
				continue
			}
			return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ctx, "structural — clause reference not localizable")
		}
		effectAbility.content.References = localReferences
		effectKeywords := keywordsWithinSpan(ctx.content.Keywords, effect.ClauseSpan)
		// A copy-token effect's "[That token] gains <keyword>." rider is a folded
		// sibling sentence whose keyword and pronoun fall outside the create
		// clause's span; attribute them here so the copy-token lowerer sees the
		// granted keyword and the sequence accounts for both as consumed.
		riderKeywords, riderReferences := tokenCopyGrantRiderAttribution(effect, ctx.content.Keywords, ctx.content.References)
		effectKeywords = append(effectKeywords, riderKeywords...)
		effectAbility.content.References = append(effectAbility.content.References, riderReferences...)
		// In a combined "gets +N/+N and gains <keyword>" sentence, the parser
		// splits the clause into a ModifyPT effect and a sibling Gain effect that
		// share the same sentence Span. The keyword belongs to the Gain sibling; do
		// not attribute it to the ModifyPT iteration so the Gain iteration
		// consumes it instead. This avoids double-counting in sequenceCountsConsumed
		// and lets lowerReferencedFixedModifyPT succeed with Keywords == nil.
		if effect.Kind == compiler.EffectModifyPT &&
			i+1 < len(ctx.content.Effects) &&
			ctx.content.Effects[i+1].Kind == compiler.EffectGain &&
			ctx.content.Effects[i+1].Span == effect.Span {
			effectKeywords = nil
		}
		effectAbility.content.Keywords = effectKeywords
		consumedTargets += len(clauseTargets)
		consumedKeywords += len(effectAbility.content.Keywords)
		consumedReferences += ownedReferenceCount + len(riderReferences)
		// Lower the effect through the shared lowerAbilityContent entry point.
		// allSharedTargets: try with inherited targets; if that fails, retry
		//   with targets cleared (e.g. "then proliferate" rejects any target).
		// mixedTargets: inherited+own combined — no fallback (fail-closed).
		// default: straightforward lowering with own targets only.
		var content game.AbilityContent
		var diagnostic *shared.Diagnostic
		// An "Otherwise," else branch is mutually exclusive with the conditional
		// effect it follows, so an EventPermanent "it" inside it cannot denote a
		// sibling clause's product and may bind the triggering permanent. The
		// first clause likewise has no prior instruction whose product an
		// EventPermanent pronoun could denote, so its "it" must be the triggering
		// permanent ("Whenever ~ attacks, put a +1/+1 counter on it, then ...").
		allowEventPronoun := effect.Connection == parser.EffectConnectionOtherwise || i == 0
		// The linked-object publisher a sequential referenced grant chains off is
		// the immediately preceding instruction; if that instruction is itself
		// gated (per-effect condition, "instead", or "otherwise" negation) it may
		// be skipped at resolution and never record its linked key. Signal that so
		// the grant fails closed instead of chaining off a publisher it cannot
		// rely on.
		publisherGated := i > 0 && sequenceClauseInstructionGated(i-1, effectConditions, insteadGates, otherwiseGates)
		if delayedContent, handled, failed := lowerDelayedSequenceClause(ctx.content.Effects, i, effectAbility, sequence, publisherGated); handled {
			if failed {
				if len(clauseReasons) > 0 {
					continue
				}
				return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ctx, "structural — delayed-target sacrifice not linkable")
			}
			content = delayedContent
		} else if allSharedTargets {
			content, diagnostic = lowerSequenceClauseContent(cardName, ctx, effectAbility.content, effectAbility.optional, &clauseAbility, allowEventPronoun)
			if diagnostic != nil {
				effectAbilityNoTarget := effectAbility
				effectAbilityNoTarget.content.Targets = nil
				content, diagnostic = lowerSequenceClauseContent(cardName, ctx, effectAbilityNoTarget.content, effectAbilityNoTarget.optional, &clauseAbility, allowEventPronoun)
			}
		} else {
			content, diagnostic = lowerSequenceClauseContent(cardName, ctx, effectAbility.content, effectAbility.optional, &clauseAbility, allowEventPronoun)
		}
		if diagnostic != nil ||
			len(content.SharedTargets) != 0 ||
			content.IsModal() ||
			len(content.Modes) != 1 {
			clauseReasons = appendClauseReason(clauseReasons, ctx, diagnostic)
			continue
		}
		mode := content.Modes[0]
		// An inherited target that no prior clause owned (a bare "Choose target
		// ..." sentence with no effect of its own) is first materialized here, so
		// this clause consumes it. Inherited targets already recorded in
		// oracleSpanToGameIdx were owned and counted by an earlier clause.
		if allSharedTargets {
			for _, t := range inheritedTargets {
				if _, owned := oracleSpanToGameIdx[t.Span]; !owned {
					consumedTargets++
				}
			}
		}
		newTargets, ok := applyTargetRemapping(
			mode, allSharedTargets, mixedTargets,
			inheritedTargets, clauseTargets,
			targets, oracleSpanToGameIdx,
		)
		if !ok {
			if len(clauseReasons) > 0 {
				continue
			}
			return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ctx, "structural — inherited target not remappable")
		}
		targets = newTargets
		if category := applySequenceClauseGates(mode.Sequence, i, effectConditions, insteadGates, otherwiseGates); category != "" {
			if len(clauseReasons) > 0 {
				continue
			}
			return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ctx, category)
		}
		if optionalFlow.enabled || optionalFlow.bareIndex >= 0 || optionalFlow.independentOptional {
			if category, ok := applyOptionalFlowEnvelope(optionalFlow, i, mode.Sequence); !ok {
				if len(clauseReasons) > 0 {
					continue
				}
				return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ctx, category)
			}
		}
		// A clause must contribute at least one instruction; an empty lowering
		// would silently drop the effect. Earlier code required exactly one
		// instruction per effect (len(sequence) == len(effects)), but a single
		// supported clause legitimately lowers to multiple instructions — e.g.
		// "up to two target creatures each get +1/+2" expands to one ModifyPT per
		// target, and "Add {R}{R}" expands to one AddMana per pip. Requiring 1:1
		// rejected those compositions even though every clause and every
		// target/reference is fully consumed, so only require non-empty here and
		// rely on the consumed-count checks below to prove nothing was dropped.
		if len(mode.Sequence) == 0 {
			if len(clauseReasons) > 0 {
				continue
			}
			return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ctx, "structural — effect produced no instructions")
		}
		sequence = append(sequence, mode.Sequence...)
	}
	// If any clause failed to lower, the sequence is unsupported. Report every
	// collected clause blocker (the first stays primary, matching the reason a
	// first-failure bail used to return) and skip the post-loop structural checks,
	// which assume a fully-consumed sequence.
	if len(clauseReasons) > 0 {
		return game.AbilityContent{}, combineReasons(clauseReasons)
	}
	// A condition's own object pronoun ("its power" in "draw a card if its power
	// is 3 or greater") sits outside every effect clause span, so it is consumed
	// by the matched condition gate rather than by an effect. Credit those
	// references so the consumed-count check does not see them as dropped. Count
	// against every condition to match the clause-reference stripping above, so an
	// optional-flow gate's own anaphor ("that player" in "If that player does")
	// is credited rather than reported as an unconsumed reference.
	consumedReferences += conditionReferenceCount(ctx.content.References, ctx.content.Conditions)
	// A punisher clause ("each opponent loses N life unless they discard a card")
	// carries a subject pronoun ("they" / "that player") that the parser folds
	// into the EffectPunisherLoseLife effect, so it never lands in the effect's own
	// reference list, yet the compiler still records it as a content-level pronoun
	// reference. The punisher lowerer models each affected group member
	// intrinsically, so that pronoun is fully consumed; credit it here so a
	// sequence of punisher clauses (Court of Ambition's escalating edict) does not
	// see it as a dropped reference.
	consumedReferences += punisherSubjectReferenceCount(ctx.content.References, ctx.content.Effects, ctx.content.Conditions)
	if !sequenceCountsConsumed(ctx, consumedTargets, consumedKeywords, consumedReferences, consumedConditions) {
		return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ctx, "structural — unconsumed targets/references/keywords")
	}
	if !publishCreatedTokenLink(sequence, gateConditions) {
		return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ctx, "structural — created-token gate not linkable")
	}
	if !linkDamageDealtThisWay(sequence) {
		return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ctx, "structural — damage-dealt-this-way drain not linkable")
	}
	if sequenceAmountReferencesPlayerSlotPermanent(sequence, targets) {
		// A payoff amount that reads a permanent's characteristics ("gain life
		// equal to that creature's toughness") but whose object resolves to a
		// player-only target slot is a mis-bound "that creature" antecedent: the
		// referenced permanent (the sacrificed creature of an edict) is not a
		// target, so the amount silently resolves to zero. Fail closed rather than
		// emit a card that does nothing until the sacrificed-creature reference is
		// modeled (linked object).
		return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ctx, "structural — payoff amount references a non-target permanent")
	}
	return game.Mode{Targets: targets, Sequence: sequence}.Ability(), nil
}

// sequenceAmountReferencesPlayerSlotPermanent reports whether any resolving
// payoff instruction scales by a permanent's characteristics through a target
// permanent reference that actually addresses a player-only target slot. Such a
// reference can never resolve to an object, so the amount is silently zero. Only
// the player-payoff primitives that carry a scalable Quantity are inspected.
func sequenceAmountReferencesPlayerSlotPermanent(sequence []game.Instruction, targets []game.TargetSpec) bool {
	for i := range sequence {
		primitive := sequence[i].Primitive
		if value, ok := primitive.(game.GainLife); ok && amountReferencesPlayerSlotPermanent(value.Amount, targets) {
			return true
		}
		if value, ok := primitive.(game.LoseLife); ok && amountReferencesPlayerSlotPermanent(value.Amount, targets) {
			return true
		}
		if value, ok := primitive.(game.Draw); ok && amountReferencesPlayerSlotPermanent(value.Amount, targets) {
			return true
		}
		if value, ok := primitive.(game.Mill); ok && amountReferencesPlayerSlotPermanent(value.Amount, targets) {
			return true
		}
	}
	return false
}

// amountReferencesPlayerSlotPermanent reports whether a dynamic quantity reads a
// target permanent's characteristics through a target slot that admits only a
// player.
func amountReferencesPlayerSlotPermanent(amount game.Quantity, targets []game.TargetSpec) bool {
	dynamic := amount.DynamicAmount()
	if !dynamic.Exists {
		return false
	}
	object := dynamic.Val.Object
	if object.Kind() != game.ObjectReferenceTargetPermanent {
		return false
	}
	return targetSlotIsPlayerOnly(targets, object.TargetIndex())
}

// targetSlotIsPlayerOnly reports whether the flat target slot at index is owned by
// a spec whose explicit Allow admits only a player.
func targetSlotIsPlayerOnly(targets []game.TargetSpec, index int) bool {
	if index < 0 {
		return false
	}
	cumulative := 0
	for i := range targets {
		width := max(targets[i].MaxTargets, 1)
		if index < cumulative+width {
			return targets[i].Allow == game.TargetAllowPlayer
		}
		cumulative += width
	}
	return false
}

// publishCreatedTokenLink wires a resolving "If the token is ..." gate (Yenna,
// Redtooth Regent) to the token a prior clause created. When a gate condition
// binds ReferenceBindingCreatedToken its lowered object reference points at the
// createdTokenLinkKey linked object, so the creating CreateToken instruction must
// publish that link. It sets PublishLinked on the sequence's single CreateToken
// instruction. It returns false (fail closed) when the sequence has no
// CreateToken, more than one CreateToken, or a CreateToken that already
// publishes a different link.
func publishCreatedTokenLink(sequence []game.Instruction, conditions []compiler.CompiledCondition) bool {
	gated := false
	for ci := range conditions {
		if conditions[ci].ObjectBinding == compiler.ReferenceBindingCreatedToken {
			gated = true
			break
		}
	}
	if !gated {
		return true
	}
	createIndex := -1
	for si := range sequence {
		if sequence[si].Primitive.Kind() != game.PrimitiveCreateToken {
			continue
		}
		if createIndex >= 0 {
			return false
		}
		createIndex = si
	}
	if createIndex < 0 {
		return false
	}
	create, ok := sequence[createIndex].Primitive.(game.CreateToken)
	if !ok || (create.PublishLinked != "" && create.PublishLinked != createdTokenLinkKey) {
		return false
	}
	create.PublishLinked = createdTokenLinkKey
	sequence[createIndex].Primitive = create
	return true
}

// linkDamageDealtThisWay wires a "...equal to the (excess) damage dealt this
// way." life gain (Corrupt, Razor Rings) to the damage it scales from. The
// consuming GainLife/LoseLife instruction reads the amount published under
// damageDealtThisWayKey, so the Damage instruction that immediately precedes it
// must publish that amount. It sets PublishResult on that preceding Damage
// instruction. It returns false (fail closed) when a consumer has no preceding
// Damage instruction, or that Damage already publishes a different key.
func linkDamageDealtThisWay(sequence []game.Instruction) bool {
	for i := range sequence {
		if !instructionConsumesDamageDealtThisWay(sequence[i]) {
			continue
		}
		damageIndex := -1
		for j := i - 1; j >= 0; j-- {
			if sequence[j].Primitive.Kind() == game.PrimitiveDamage {
				damageIndex = j
				break
			}
		}
		if damageIndex < 0 {
			return false
		}
		if sequence[damageIndex].PublishResult != "" && sequence[damageIndex].PublishResult != damageDealtThisWayKey {
			return false
		}
		sequence[damageIndex].PublishResult = damageDealtThisWayKey
	}
	return true
}

// instructionConsumesDamageDealtThisWay reports whether instr is a life-gain or
// life-loss whose amount reads the damage published under damageDealtThisWayKey.
func instructionConsumesDamageDealtThisWay(instr game.Instruction) bool {
	var amount game.Quantity
	switch instr.Primitive.Kind() {
	case game.PrimitiveGainLife:
		gain, ok := instr.Primitive.(game.GainLife)
		if !ok {
			return false
		}
		amount = gain.Amount
	case game.PrimitiveLoseLife:
		lose, ok := instr.Primitive.(game.LoseLife)
		if !ok {
			return false
		}
		amount = lose.Amount
	default:
		return false
	}
	dynamic := amount.DynamicAmount()
	if !dynamic.Exists || dynamic.Val.ResultKey != damageDealtThisWayKey {
		return false
	}
	return dynamic.Val.Kind == game.DynamicAmountPreviousEffectResult ||
		dynamic.Val.Kind == game.DynamicAmountPreviousEffectExcessDamage
}

// conditionReferenceCount counts the references whose span falls within one of
// the gate conditions. These are the conditions' own object pronouns, consumed
// by the condition gate rather than by any effect clause.
func conditionReferenceCount(
	references []compiler.CompiledReference,
	conditions []compiler.CompiledCondition,
) int {
	count := 0
	for ri := range references {
		for ci := range conditions {
			if spanCovered(references[ri].Span, []shared.Span{conditions[ci].Span}) {
				count++
				break
			}
		}
	}
	return count
}

// punisherSubjectReferenceCount credits the subject pronoun of each punisher
// clause ("each opponent loses N life unless they discard a card") in an ordered
// sequence. The parser folds that "unless <subject>" pronoun into the
// EffectPunisherLoseLife effect, so it never appears in the effect's own
// reference list, but the compiler still records it as a content-level pronoun
// reference. Because the punisher lowerer models each affected group member
// intrinsically, the pronoun is fully consumed by the effect; count it here so a
// sequence of punisher clauses does not fail the consumed-reference check.
// References already covered by a condition span are excluded so they are not
// double-counted with conditionReferenceCount.
func punisherSubjectReferenceCount(
	references []compiler.CompiledReference,
	effects []compiler.CompiledEffect,
	conditions []compiler.CompiledCondition,
) int {
	count := 0
	for ri := range references {
		if references[ri].Kind != compiler.ReferencePronoun {
			continue
		}
		if spanCovered(references[ri].Span, conditionSpans(conditions)) {
			continue
		}
		for ei := range effects {
			if effects[ei].Kind != compiler.EffectPunisherLoseLife {
				continue
			}
			if spanCovered(references[ri].Span, []shared.Span{effects[ei].ClauseSpan}) {
				count++
				break
			}
		}
	}
	return count
}

// conditionSpans collects the source spans of the given conditions.
func conditionSpans(conditions []compiler.CompiledCondition) []shared.Span {
	spans := make([]shared.Span, len(conditions))
	for ci := range conditions {
		spans[ci] = conditions[ci].Span
	}
	return spans
}

// referencesOutsideConditionSpans returns the references whose source span is
// not covered by any of the given conditions' spans, the complement of the
// references credited by conditionReferenceCount. It separates a gate
// condition's own object references (e.g. the "this spell" in "If this spell was
// kicked, ...") from the effect body's references when both fall inside the same
// effect clause span.
func referencesOutsideConditionSpans(
	references []compiler.CompiledReference,
	conditions []compiler.CompiledCondition,
) []compiler.CompiledReference {
	var outside []compiler.CompiledReference
	for ri := range references {
		within := false
		for ci := range conditions {
			if spanCovered(references[ri].Span, []shared.Span{conditions[ci].Span}) {
				within = true
				break
			}
		}
		if !within {
			outside = append(outside, references[ri])
		}
	}
	return outside
}

// sequenceCountsConsumed reports whether the per-clause lowering consumed every
// target, keyword, reference, and condition the ordered sequence carried. A
// shortfall means a clause was silently dropped, so the sequence fails closed.
func sequenceCountsConsumed(
	ctx contentCtx,
	consumedTargets, consumedKeywords, consumedReferences, consumedConditions int,
) bool {
	return consumedTargets == len(ctx.content.Targets) &&
		consumedKeywords == len(ctx.content.Keywords) &&
		consumedReferences == len(ctx.content.References) &&
		consumedConditions == len(ctx.content.Conditions)
}

func lowerLinkedSearchUntapSequence(ctx contentCtx) (game.AbilityContent, bool) {
	effects := ctx.content.Effects
	if ctx.optional ||
		len(effects) != 4 ||
		len(ctx.content.Conditions) != 1 ||
		len(ctx.content.References) != 2 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	searchEffects := effects[:3]
	if effects[0].Context != parser.EffectContextController ||
		effects[3].Kind != compiler.EffectUntap ||
		effects[3].Context != parser.EffectContextController ||
		effects[3].Connection != parser.EffectConnectionThen ||
		!effects[3].Exact ||
		!exactUnqualifiedLandSelector(effects[3].Selector) ||
		len(effects[3].Targets) != 0 ||
		len(effects[3].References) != 1 ||
		len(effects[1].References) != 1 {
		return game.AbilityContent{}, false
	}
	group, ok := searchGroupSpec(searchEffects)
	if !ok ||
		group.Length != len(searchEffects) ||
		group.Amount != 1 ||
		group.Spec.SourceZone != zone.Library ||
		group.Spec.Destination != zone.Battlefield ||
		!group.Spec.EntersTapped ||
		group.Spec.SplitDestination.Exists ||
		len(group.Spec.Filter.RequiredTypes) != 1 ||
		group.Spec.Filter.RequiredTypes[0] != types.Land ||
		group.Spec.Filter.RequirePermanentCard ||
		len(group.Spec.Filter.SubtypesAny) != 0 ||
		group.Spec.Filter.ManaValue.Exists ||
		group.Spec.MaxManaValueFromX ||
		group.Spec.Filter.Power.Exists ||
		group.Spec.Filter.Toughness.Exists ||
		group.Spec.Reveal ||
		group.Spec.SharedSubtype ||
		len(group.Spec.Filter.Supertypes) != 1 ||
		group.Spec.Filter.Supertypes[0] != types.Basic {
		return game.AbilityContent{}, false
	}
	putRef := effects[1].References[0]
	if putRef.Kind != compiler.ReferencePronoun ||
		putRef.Pronoun != compiler.ReferencePronounIt ||
		putRef.Binding != compiler.ReferenceBindingPriorInstructionResult ||
		putRef.PriorInstruction != 0 {
		return game.AbilityContent{}, false
	}
	ref := effects[3].References[0]
	if ref.Kind != compiler.ReferenceThatObject ||
		ref.Binding != compiler.ReferenceBindingPriorInstructionResult ||
		ref.PriorInstruction != 0 {
		return game.AbilityContent{}, false
	}
	condition := ctx.content.Conditions[0]
	if !exactControllerLandCountCondition(condition) ||
		!spanCovered(condition.Span, []shared.Span{effects[3].Span}) {
		return game.AbilityContent{}, false
	}
	loweredCondition, ok := lowerCondition(condition, conditionContextEffectGate)
	if !ok {
		return game.AbilityContent{}, false
	}
	key := game.LinkedKey("searched-land-1")
	object, ok := lowerObjectReference(ref, referenceLoweringContext{
		PriorInstruction: 0,
		PriorLinkedKey:   key,
	})
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: []game.Instruction{
		{Primitive: game.Search{
			Player:        game.ControllerReference(),
			Spec:          group.Spec,
			Amount:        game.Fixed(1),
			PublishLinked: key,
		}},
		{
			Primitive: game.Untap{Object: object},
			Condition: opt.Val(game.EffectCondition{
				Condition: opt.Val(loweredCondition),
			}),
		},
	}}.Ability(), true
}

func exactControllerLandCountCondition(condition compiler.CompiledCondition) bool {
	selection := condition.Selection
	return condition.Kind == compiler.ConditionIf &&
		condition.Resolving &&
		condition.Predicate == compiler.ConditionPredicateControllerControls &&
		!condition.Negated &&
		condition.Threshold == 4 &&
		len(selection.RequiredTypes) == 1 &&
		selection.RequiredTypes[0] == types.Land &&
		len(selection.Supertypes) == 0 &&
		len(selection.SubtypesAny) == 0 &&
		len(selection.ColorsAny) == 0 &&
		!selection.Colorless &&
		!selection.Multicolored &&
		!selection.TokenOnly &&
		!selection.ExcludeSource &&
		selection.Tapped == compiler.ConditionTriAny &&
		selection.CombatState == compiler.ConditionCombatStateAny &&
		selection.Keyword == parser.KeywordUnknown &&
		!selection.MatchPowerAtLeast &&
		!selection.MatchTotalPowerAtLeast
}

func exactUnqualifiedLandSelector(selector compiler.CompiledSelector) bool {
	return selector.Kind == compiler.SelectorLand &&
		selector.Controller == compiler.ControllerAny &&
		!selector.All &&
		!selector.Another &&
		!selector.Other &&
		!selector.Attacking &&
		!selector.Blocking &&
		!selector.Tapped &&
		!selector.Untapped &&
		selector.Keyword == parser.KeywordUnknown &&
		selector.ExcludedKeyword == parser.KeywordUnknown &&
		!selector.MatchManaValue &&
		!selector.MatchPower &&
		!selector.MatchToughness &&
		!selector.Colorless &&
		!selector.Multicolored &&
		!selector.BasicLandType &&
		!selector.PlayerOrPlaneswalker &&
		selector.Zone == zone.None &&
		(len(selector.RequiredTypesAny()) == 0 ||
			slices.Equal(selector.RequiredTypesAny(), []types.Card{types.Land})) &&
		len(selector.ExcludedTypes()) == 0 &&
		len(selector.Supertypes()) == 0 &&
		len(selector.ExcludedSupertypes()) == 0 &&
		len(selector.ColorsAny()) == 0 &&
		len(selector.ExcludedColors()) == 0 &&
		len(selector.SubtypesAny()) == 0 &&
		len(selector.SourceTypes()) == 0 &&
		len(selector.Alternatives) == 0
}

// lowerCombinedSequenceShapes attempts the special-case combined-shape lowerers
// (single continuous effects spread across two clauses) that do not model
// per-effect conditions. It only runs when the sequence carries no conditions,
// so a condition can never be silently dropped.
func lowerCombinedSequenceShapes(cardName string, ctx contentCtx, syntax *parser.Ability) (game.AbilityContent, bool) {
	if content, ok := lowerShuffleRevealPermanentSequence(ctx); ok {
		return content, true
	}
	if content, ok := lowerRevealUntilSequence(ctx); ok {
		return content, true
	}
	if content, ok := lowerPileSplitSequence(ctx); ok {
		return content, true
	}
	if content, ok := lowerRevealTopPartitionSequence(ctx); ok {
		return content, true
	}
	if content, ok := lowerRemovalManifestSequence(ctx); ok {
		return content, true
	}
	if content, ok := lowerRevealChooseHandDiscardSequence(ctx); ok {
		return content, true
	}
	if content, ok := lowerRevealHandLifeLossSaddledSequence(ctx); ok {
		return content, true
	}
	if len(ctx.content.Conditions) != 0 {
		return game.AbilityContent{}, false
	}
	if content, ok := lowerTemporaryDoublePTKeywordSpell(ctx); ok {
		return content, true
	}
	if content, ok := lowerTemporaryPTKeywordSpell(ctx); ok {
		return content, true
	}
	if content, ok := lowerGroupTemporaryPTKeywordSpell(ctx); ok {
		return content, true
	}
	if content, ok := lowerCyclingCountDamageAndGain(cardName, ctx); ok {
		return content, true
	}
	if content, ok := lowerExcessDamageToControllerSpell(ctx); ok {
		return content, true
	}
	if content, ok := lowerGroupLinkedLifeSpell(ctx); ok {
		return content, true
	}
	if content, ok := lowerLifeLostThisWayDrain(ctx); ok {
		return content, true
	}
	if content, ok := lowerControllerDrawLoseShareXSpell(ctx); ok {
		return content, true
	}
	if content, ok := lowerDrainXLifeSpell(ctx); ok {
		return content, true
	}
	if content, ok := lowerDestroyedThisWaySequence(ctx); ok {
		return content, true
	}
	if content, ok := lowerDiceTableSequence(cardName, ctx, syntax); ok {
		return content, true
	}
	if content, ok := lowerDieRollResultSequence(cardName, ctx, syntax); ok {
		return content, true
	}
	if content, ok := lowerDiscardDrawGreatestThisWaySequence(ctx); ok {
		return content, true
	}
	if content, ok := lowerDiscardDrawThenManaValueDamageSequence(ctx); ok {
		return content, true
	}
	if content, ok := lowerWheelDiscardDrawSequence(ctx); ok {
		return content, true
	}
	if content, ok := lowerTapDownSequence(ctx); ok {
		return content, true
	}
	if content, ok := lowerTapStunSequence(ctx); ok {
		return content, true
	}
	if content, ok := lowerGroupBlinkSequence(ctx); ok {
		return content, true
	}
	if content, ok := lowerMassGroupBlinkSequence(ctx); ok {
		return content, true
	}
	if content, ok := lowerDigSequence(ctx); ok {
		return content, true
	}
	if content, ok := lowerDrawHandLibrarySequence(ctx); ok {
		return content, true
	}
	if content, ok := lowerDrawHandDiscardSequence(ctx); ok {
		return content, true
	}
	if content, ok := lowerDynamicCountDrawThenGroupKeywordSequence(ctx); ok {
		return content, true
	}
	if content, ok := lowerGroupCounterThenGroupKeywordSequence(ctx); ok {
		return content, true
	}
	if content, ok := lowerGroupPumpThenGroupCounterSequence(ctx); ok {
		return content, true
	}
	if content, ok := lowerCreateTokenThenGrantKeywordSequence(ctx); ok {
		return content, true
	}
	if content, ok := lowerCreateTokenThenCountersSequence(ctx); ok {
		return content, true
	}
	if content, ok := lowerCreateTokenThenAttachSequence(ctx); ok {
		return content, true
	}
	if content, ok := lowerManifestDreadThenCountersSequence(ctx); ok {
		return content, true
	}
	if content, ok := lowerCloakThenAttachSequence(ctx); ok {
		return content, true
	}
	return game.AbilityContent{}, false
}

// groupBackReferenceThose reports whether the effect's subject is the plural
// demonstrative "those" — the back-reference wording of "Those creatures gain
// <keyword> until end of turn." after a preceding "for each <group>" count. The
// pronoun denotes the just-counted group; it is reconstructed from that count's
// selection rather than bound to an antecedent target.
func groupBackReferenceThose(refs []compiler.CompiledReference) bool {
	return len(refs) == 1 &&
		refs[0].Kind == compiler.ReferencePronoun &&
		refs[0].Pronoun == compiler.ReferencePronounThose
}

// lowerDynamicCountDrawThenGroupKeywordSequence lowers the ordered pair
// "Draw a card for each <group>. Those creatures gain <keyword> until end of
// turn." (Inspiring Call). The first clause counts a battlefield group; the
// second grants a keyword to that same group. Because nothing between the two
// clauses changes the board, the runtime's group continuous effect snapshots its
// members at resolution, so re-evaluating the count's selection yields exactly
// "those creatures". It reuses the count's resolved group for the grant and
// fails closed for any other shape, non-battlefield count, or unsupported
// keyword.
func lowerDynamicCountDrawThenGroupKeywordSequence(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		ctx.optional {
		return game.AbilityContent{}, false
	}
	drawEffect := ctx.content.Effects[0]
	keywordEffect := ctx.content.Effects[1]
	if drawEffect.Kind != compiler.EffectDraw ||
		keywordEffect.Kind != compiler.EffectGain ||
		!drawEffect.Exact ||
		!keywordEffect.Exact ||
		drawEffect.Negated ||
		keywordEffect.Negated ||
		drawEffect.Optional ||
		keywordEffect.Optional ||
		drawEffect.Context != parser.EffectContextController ||
		drawEffect.Amount.DynamicKind != compiler.DynamicAmountCount ||
		keywordEffect.Duration != compiler.DurationUntilEndOfTurn ||
		keywordEffect.StaticSubject != compiler.StaticSubjectNone ||
		keywordEffect.KeywordGrantChoice ||
		!groupBackReferenceThose(keywordEffect.SubjectReferences) {
		return game.AbilityContent{}, false
	}
	dynamic, ok := lowerDynamicAmount(drawEffect.Amount, game.SourcePermanentReference())
	if !ok || dynamic.Kind != game.DynamicAmountCountSelector || !dynamic.Group.Valid() {
		return game.AbilityContent{}, false
	}
	keywords, abilities, ok := partitionTemporaryKeywords(keywordsWithinSpan(ctx.content.Keywords, keywordEffect.ClauseSpan))
	if !ok || (len(keywords) == 0 && len(abilities) == 0) {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Sequence: []game.Instruction{
			{Primitive: game.Draw{
				Amount: game.Dynamic(dynamic),
				Player: game.ControllerReference(),
			}},
			{Primitive: game.ApplyContinuous{
				ContinuousEffects: []game.ContinuousEffect{{
					Layer:        game.LayerAbility,
					Group:        dynamic.Group,
					AddKeywords:  keywords,
					AddAbilities: abilities,
				}},
				Duration: game.DurationUntilEndOfTurn,
			}},
		},
	}.Ability(), true
}

// lowerGroupCounterThenGroupKeywordSequence lowers the ordered pair "Put a
// +1/+1 counter on each creature you control. Those creatures gain <keyword>
// until end of turn." (Felidar Retreat's second mode). The first clause places
// a fixed counter on a battlefield group; the second grants a keyword to that
// same group. As with the draw-count variant, nothing between the two clauses
// changes the board, so the runtime's group continuous effect snapshots the same
// members the counter placement affected and "those creatures" resolves to that
// group. It reuses the counter clause's resolved group for the grant and fails
// closed for any other shape, non-group recipient, or unsupported keyword.
func lowerGroupCounterThenGroupKeywordSequence(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		ctx.optional {
		return game.AbilityContent{}, false
	}
	counterEffect := ctx.content.Effects[0]
	keywordEffect := ctx.content.Effects[1]
	if counterEffect.Kind != compiler.EffectPut ||
		keywordEffect.Kind != compiler.EffectGain ||
		!counterEffect.Exact ||
		!keywordEffect.Exact ||
		counterEffect.Negated ||
		keywordEffect.Negated ||
		counterEffect.Optional ||
		keywordEffect.Optional ||
		counterEffect.Context != parser.EffectContextController ||
		!counterEffect.CounterKindKnown ||
		!compiler.CounterKindPlacementSupported(counterEffect.CounterKind) ||
		counterEffect.CounterKind.PlayerOnly() ||
		!counterEffect.Amount.Known ||
		counterEffect.Amount.Value < 1 ||
		keywordEffect.Duration != compiler.DurationUntilEndOfTurn ||
		keywordEffect.StaticSubject != compiler.StaticSubjectNone ||
		keywordEffect.KeywordGrantChoice ||
		!groupBackReferenceThose(keywordEffect.SubjectReferences) {
		return game.AbilityContent{}, false
	}
	group, ok := damageGroupRecipient(counterEffect.Selector)
	if !ok {
		return game.AbilityContent{}, false
	}
	keywords, abilities, ok := partitionTemporaryKeywords(keywordsWithinSpan(ctx.content.Keywords, keywordEffect.ClauseSpan))
	if !ok || (len(keywords) == 0 && len(abilities) == 0) {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Sequence: []game.Instruction{
			{Primitive: game.AddCounter{
				Amount:      game.Fixed(counterEffect.Amount.Value),
				Group:       group,
				CounterKind: counterEffect.CounterKind,
			}},
			{Primitive: game.ApplyContinuous{
				ContinuousEffects: []game.ContinuousEffect{{
					Layer:        game.LayerAbility,
					Group:        group,
					AddKeywords:  keywords,
					AddAbilities: abilities,
				}},
				Duration: game.DurationUntilEndOfTurn,
			}},
		},
	}.Ability(), true
}

// groupCounterBackReferencePronoun reports whether a counter-placement
// recipient is the plural group back-reference "each of them" / "each of those
// creatures" / "each of those" — a pronoun ("them"/"they") or demonstrative
// ("those") that denotes the group an immediately preceding clause affected
// rather than a self-contained selection. The single allowed reference carries
// no antecedent target; the group is reconstructed from the preceding clause.
func groupCounterBackReferencePronoun(references []compiler.CompiledReference) bool {
	return len(references) == 1 &&
		references[0].Kind == compiler.ReferencePronoun &&
		(references[0].Pronoun == compiler.ReferencePronounThem ||
			references[0].Pronoun == compiler.ReferencePronounThey ||
			references[0].Pronoun == compiler.ReferencePronounThose)
}

// lowerGroupPumpThenGroupCounterSequence lowers the ordered pair "Other
// creatures you control get +2/+2 until end of turn. Put an indestructible
// counter on each of them." (Summon: Knights of Round's final chapter). The
// first clause pumps a never-resolving controlled creature group until end of
// turn; the second places one fixed counter on every member of that same group,
// named by the plural back-reference "each of them". Because nothing between the
// two clauses changes the board, the runtime's group continuous effect and group
// counter placement both snapshot the same members, so the counter clause's
// "them" resolves to the just-pumped group. It reuses the pump clause's resolved
// group for the counter placement and fails closed for any other shape, a
// non-group pump subject, an unsupported counter kind, or a non-back-reference
// recipient.
func lowerGroupPumpThenGroupCounterSequence(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		ctx.optional {
		return game.AbilityContent{}, false
	}
	pumpEffect := ctx.content.Effects[0]
	counterEffect := ctx.content.Effects[1]
	if pumpEffect.Kind != compiler.EffectModifyPT ||
		counterEffect.Kind != compiler.EffectPut ||
		!pumpEffect.Exact ||
		pumpEffect.Negated ||
		pumpEffect.Optional ||
		pumpEffect.Duration != compiler.DurationUntilEndOfTurn ||
		pumpEffect.StaticSubject == compiler.StaticSubjectNone {
		return game.AbilityContent{}, false
	}
	if counterEffect.Negated ||
		counterEffect.Optional ||
		counterEffect.Context != parser.EffectContextController ||
		counterEffect.Duration != compiler.DurationNone ||
		!counterEffect.CounterKindKnown ||
		!compiler.CounterKindPlacementSupported(counterEffect.CounterKind) ||
		counterEffect.CounterKind.PlayerOnly() ||
		!counterEffect.Amount.Known ||
		counterEffect.Amount.Value < 1 ||
		!groupCounterBackReferencePronoun(counterEffect.References) {
		return game.AbilityContent{}, false
	}
	group, ok := resolvingStaticSubjectGroup(&pumpEffect)
	if !ok {
		return game.AbilityContent{}, false
	}
	continuous, ok := groupModifyPTContinuousEffect(&pumpEffect, group)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Sequence: []game.Instruction{
			{Primitive: game.ApplyContinuous{
				ContinuousEffects: []game.ContinuousEffect{continuous},
				Duration:          game.DurationUntilEndOfTurn,
			}},
			{Primitive: game.AddCounter{
				Amount:      game.Fixed(counterEffect.Amount.Value),
				Group:       group,
				CounterKind: counterEffect.CounterKind,
			}},
		},
	}.Ability(), true
}

// createdTokenLinkKey links a freshly created token to a following "that token
// gains <keyword> until end of turn" grant so the keyword is applied to exactly
// that token. The runtime scopes the key per source object, so a fixed string is
// unambiguous across cards.
const createdTokenLinkKey = "created-token"

// lowerCreateTokenThenGrantKeywordSequence lowers the ordered pair "Create
// [token]. That token gains <keyword> until end of turn." (Loyal Apprentice).
// The first clause creates a token; the second grants it a temporary keyword. The
// grant's "that token" back-reference binds to the create instruction's result,
// realized by publishing the created token under a link key and resolving the
// grant's object reference to that linked token. It supports only the singular
// "that token" back-reference onto a controller-created synthesized token; it
// fails closed for copy/choice tokens, plural back-references, other recipients,
// durations, or unsupported keywords.
func lowerCreateTokenThenGrantKeywordSequence(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		ctx.optional {
		return game.AbilityContent{}, false
	}
	createEffect := ctx.content.Effects[0]
	keywordEffect := ctx.content.Effects[1]
	if createEffect.Kind != compiler.EffectCreate ||
		keywordEffect.Kind != compiler.EffectGain ||
		!createEffect.Exact ||
		!keywordEffect.Exact ||
		createEffect.Negated ||
		keywordEffect.Negated ||
		createEffect.Optional ||
		keywordEffect.Optional ||
		createEffect.Context != parser.EffectContextController ||
		keywordEffect.Context != parser.EffectContextReferencedObject ||
		keywordEffect.Duration != compiler.DurationUntilEndOfTurn ||
		keywordEffect.StaticSubject != compiler.StaticSubjectNone ||
		keywordEffect.KeywordGrantChoice ||
		createEffect.TokenCopyOfTarget ||
		createEffect.TokenCopyOfReference ||
		createEffect.TokenCopyOfAttached ||
		createEffect.TokenCopyOfForEach ||
		createEffect.TokenChoice ||
		!referencesBindTo(keywordEffect.SubjectReferences, compiler.ReferenceBindingPriorInstructionResult, 0) {
		return game.AbilityContent{}, false
	}
	keywords, abilities, ok := partitionTemporaryKeywords(
		keywordsWithinSpan(ctx.content.Keywords, keywordEffect.ClauseSpan))
	if !ok || (len(keywords) == 0 && len(abilities) == 0) {
		return game.AbilityContent{}, false
	}
	createContent, diagnostic := lowerCreateTokenSpellLinked(
		contextForEffect(ctx, &createEffect), createdTokenLinkKey)
	if diagnostic != nil ||
		len(createContent.Modes) != 1 ||
		len(createContent.Modes[0].Sequence) != 1 ||
		len(createContent.Modes[0].Targets) != 0 {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Sequence: []game.Instruction{
			createContent.Modes[0].Sequence[0],
			{Primitive: game.ApplyContinuous{
				Object: opt.Val(game.LinkedObjectReference(string(createdTokenLinkKey))),
				ContinuousEffects: []game.ContinuousEffect{{
					Layer:        game.LayerAbility,
					AddKeywords:  keywords,
					AddAbilities: abilities,
				}},
				Duration: game.DurationUntilEndOfTurn,
			}},
		},
	}.Ability(), true
}

func lowerPonderSequence(ctx contentCtx) (game.AbilityContent, bool) {
	effectCount := len(ctx.content.Effects)
	if ctx.optional ||
		(effectCount != 3 && (!ctx.allowPonderPrefix || effectCount != 2)) ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!matchPonderReorder(&ctx.content.Effects[0], ctx.content.References) ||
		!matchPonderShuffle(&ctx.content.Effects[1]) {
		return game.AbilityContent{}, false
	}
	if effectCount == 3 && !matchPonderDraw(&ctx.content.Effects[2]) {
		return game.AbilityContent{}, false
	}
	sequence := []game.Instruction{
		{Primitive: game.ReorderLibraryTop{
			Player: game.ControllerReference(),
			Amount: game.Fixed(ctx.content.Effects[0].Amount.Value),
		}},
		{
			Primitive: game.ShuffleLibrary{Player: game.ControllerReference()},
			Optional:  true,
		},
	}
	if effectCount == 3 {
		sequence = append(sequence, game.Instruction{
			Primitive: game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(1)},
		})
	}
	return game.Mode{Sequence: sequence}.Ability(), true
}

// lowerStandaloneReorderLibraryTop lowers a lone "Look at the top N cards of
// your library, then put them back in any order." effect — Index, and Sensei's
// Divining Top's first activated ability — into a single ReorderLibraryTop
// instruction. The effect already captures the full look-and-reorder semantics;
// the internal "them" pronoun that refers to the looked-at cards is consumed
// here rather than bound to an external antecedent. It fails closed on any other
// shape (targets, conditions, modes, keywords, or a non-reorder effect).
func lowerStandaloneReorderLibraryTop(ctx contentCtx) (game.AbilityContent, bool) {
	// Invariant: the sole caller dispatches this lowerer only inside the
	// `len(ctx.content.Effects) == 1` branch (lower_spell.go:297), so a length
	// other than one is an upstream bug rather than an unsupported card shape.
	if len(ctx.content.Effects) != 1 {
		panic(fmt.Sprintf("lowerStandaloneReorderLibraryTop: expected a single effect, got %d", len(ctx.content.Effects)))
	}
	if len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		!matchPonderReorder(&ctx.content.Effects[0], ctx.content.References) {
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.ReorderLibraryTop{
			Player: game.ControllerReference(),
			Amount: game.Fixed(ctx.content.Effects[0].Amount.Value),
		},
	}}}.Ability(), true
}

func matchPonderReorder(effect *compiler.CompiledEffect, references []compiler.CompiledReference) bool {
	if effect.Kind != compiler.EffectReorderLibraryTop ||
		!effect.Exact ||
		effect.Optional ||
		effect.Negated ||
		effect.Context != parser.EffectContextController ||
		!effect.Amount.Known ||
		effect.Amount.Value < 1 ||
		len(effect.Targets) != 0 ||
		len(references) != 1 {
		return false
	}
	reference := references[0]
	return reference.Kind == compiler.ReferencePronoun &&
		reference.Pronoun == compiler.ReferencePronounThem &&
		spanCovered(reference.Span, []shared.Span{effect.Span})
}

func matchPonderShuffle(effect *compiler.CompiledEffect) bool {
	return effect.Kind == compiler.EffectShuffle &&
		effect.Exact &&
		effect.Optional &&
		!effect.Negated &&
		effect.Context == parser.EffectContextController &&
		len(effect.Targets) == 0 &&
		len(effect.References) == 0
}

func matchPonderDraw(effect *compiler.CompiledEffect) bool {
	return effect.Kind == compiler.EffectDraw &&
		effect.Exact &&
		!effect.Optional &&
		!effect.Negated &&
		effect.Context == parser.EffectContextController &&
		effect.Amount.Known &&
		effect.Amount.Value == 1 &&
		len(effect.Targets) == 0 &&
		len(effect.References) == 0
}

func lowerShuffleRevealPermanentSequence(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) != 3 ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.Conditions) != 1 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 5 {
		return game.AbilityContent{}, false
	}
	shuffle := ctx.content.Effects[0]
	reveal := ctx.content.Effects[1]
	put := ctx.content.Effects[2]
	if shuffle.Kind != compiler.EffectShuffle ||
		shuffle.Context != parser.EffectContextTarget ||
		shuffle.Player != parser.EffectPlayerTargetOwner ||
		shuffle.CardSource != parser.EffectCardSourceNone ||
		!shuffle.Exact ||
		shuffle.Optional ||
		shuffle.Negated ||
		shuffle.ToZone != zone.Library ||
		len(shuffle.Targets) != 1 ||
		len(shuffle.References) != 2 ||
		!referencesBindTo(shuffle.References, compiler.ReferenceBindingTarget, 0) {
		return game.AbilityContent{}, false
	}
	if reveal.Kind != compiler.EffectReveal ||
		reveal.Context != parser.EffectContextPriorSubject ||
		reveal.Connection != parser.EffectConnectionThen ||
		reveal.Player != parser.EffectPlayerTargetOwner ||
		reveal.CardSource != parser.EffectCardSourceTopOfPlayerLibrary ||
		!reveal.Exact ||
		reveal.Optional ||
		reveal.Negated ||
		reveal.Selector.Kind != compiler.SelectorCard ||
		len(reveal.Targets) != 0 ||
		len(reveal.References) != 1 ||
		!referencesBindTo(reveal.References, compiler.ReferenceBindingTarget, 0) {
		return game.AbilityContent{}, false
	}
	if put.Kind != compiler.EffectPut ||
		put.Context != parser.EffectContextEventPlayer ||
		put.Player != parser.EffectPlayerTargetOwner ||
		put.CardSource != parser.EffectCardSourcePriorInstructionResult ||
		!put.RequirePermanentCard ||
		!put.Exact ||
		put.Optional ||
		put.Negated ||
		put.ToZone != zone.Battlefield ||
		len(put.Targets) != 0 ||
		len(put.References) != 2 ||
		!referencesContainBinding(put.References, compiler.ReferenceBindingTarget, 0) ||
		!referencesContainBinding(put.References, compiler.ReferenceBindingPriorInstructionResult, 1) {
		return game.AbilityContent{}, false
	}
	condition := ctx.content.Conditions[0]
	if condition.Kind != compiler.ConditionIf ||
		condition.Predicate != compiler.ConditionPredicateUnsupported ||
		!spanCovered(condition.Span, []shared.Span{put.ClauseSpan}) {
		return game.AbilityContent{}, false
	}
	target, ok := permanentTargetSpec(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, false
	}

	key := game.LinkedKey("revealed-card-1")
	owner := game.ObjectOwnerReference(game.TargetPermanentReference(0))
	return game.Mode{
		Targets: []game.TargetSpec{target},
		Sequence: []game.Instruction{
			{Primitive: game.ShufflePermanentIntoLibrary{
				Object: game.TargetPermanentReference(0),
			}},
			{Primitive: game.Reveal{
				Amount:        game.Fixed(1),
				Player:        owner,
				PublishLinked: key,
			}},
			{
				Primitive: game.PutOnBattlefield{
					Source:    game.LinkedBattlefieldSource(key),
					Recipient: opt.Val(owner),
				},
				CardCondition: opt.Val(game.CardSelection{
					Card: game.CardReference{
						Kind:   game.CardReferenceLinked,
						LinkID: string(key),
					},
					Selection: game.Selection{RequirePermanentCard: true},
				}),
			},
		},
	}.Ability(), true
}

// lowerRevealHandLifeLossSaddledSequence lowers the Caustic Bronco shape:
// "reveal the top card of your library and put it into your hand. You lose life
// equal to that card's mana value if this creature isn't saddled. Otherwise,
// each opponent loses that much life." It reveals the controller's top library
// card, moves it to hand, then drains life equal to that card's mana value —
// from the controller on the gated branch and from each opponent on the
// negated "otherwise" branch. The per-effect gate condition (the saddled-state
// gate) is matched and lowered through the shared sequence-gate machinery, and
// its negation gates the otherwise branch, so exactly one of the two life-loss
// branches resolves. Both branches read the same revealed card's mana value,
// which is the meaning of "that much life" here (the controller never loses
// life on the saddled branch, so it cannot refer to a prior life loss).
func lowerRevealHandLifeLossSaddledSequence(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) != 4 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 1 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	reveal := ctx.content.Effects[0]
	put := ctx.content.Effects[1]
	loseSelf := ctx.content.Effects[2]
	loseOpponents := ctx.content.Effects[3]
	if reveal.Kind != compiler.EffectReveal ||
		reveal.Context != parser.EffectContextController ||
		reveal.Selector.Kind != compiler.SelectorCard ||
		reveal.Optional || reveal.Negated ||
		len(reveal.References) != 0 {
		return game.AbilityContent{}, false
	}
	if put.Kind != compiler.EffectPut ||
		put.Connection != parser.EffectConnectionAnd ||
		put.Context != parser.EffectContextController ||
		put.ToZone != zone.Hand ||
		put.Optional || put.Negated ||
		len(put.References) != 1 ||
		!referencesBindTo(put.References, compiler.ReferenceBindingPriorInstructionResult, 0) {
		return game.AbilityContent{}, false
	}
	if loseSelf.Kind != compiler.EffectLose ||
		loseSelf.Context != parser.EffectContextController ||
		!loseSelf.LifeObject ||
		loseSelf.Optional || loseSelf.Negated ||
		loseSelf.Amount.DynamicKind != compiler.DynamicAmountSourceManaValue ||
		loseSelf.Amount.Multiplier != 1 ||
		len(loseSelf.References) != 1 {
		return game.AbilityContent{}, false
	}
	if loseOpponents.Kind != compiler.EffectLose ||
		loseOpponents.Connection != parser.EffectConnectionOtherwise ||
		loseOpponents.Context != parser.EffectContextEachOpponent ||
		!loseOpponents.LifeObject ||
		loseOpponents.Optional || loseOpponents.Negated ||
		loseOpponents.Amount.DynamicKind != compiler.DynamicAmountTriggeringLifeChange ||
		len(loseOpponents.References) != 0 {
		return game.AbilityContent{}, false
	}

	// Match and lower the per-effect gate condition (the saddled-state gate on
	// the controller's life-loss branch) and derive its negation for the
	// each-opponent otherwise branch. The single condition must gate exactly the
	// controller branch; any other shape fails closed.
	effectConditions, _, ok := matchSequenceEffectConditions(ctx.content.Effects, ctx.content.Conditions)
	if !ok {
		return game.AbilityContent{}, false
	}
	selfGate, gated := effectConditions[2]
	if !gated || len(effectConditions) != 1 {
		return game.AbilityContent{}, false
	}
	otherwiseGates, ok := sequenceOtherwiseGates(ctx.content.Effects, effectConditions, optionalFlowPlan{elseIndex: -1})
	if !ok {
		return game.AbilityContent{}, false
	}
	opponentsGate, gated := otherwiseGates[3]
	if !gated || len(otherwiseGates) != 1 {
		return game.AbilityContent{}, false
	}

	key := game.LinkedKey("revealed-card-1")
	linked := game.LinkedObjectReference(string(key))
	manaValue, ok := objectCharacteristicAmount(compiler.DynamicAmountSourceManaValue, linked)
	if !ok {
		return game.AbilityContent{}, false
	}

	return game.Mode{
		Sequence: []game.Instruction{
			{Primitive: game.Reveal{
				Amount:        game.Fixed(1),
				Player:        game.ControllerReference(),
				PublishLinked: key,
			}},
			{Primitive: game.MoveCard{
				Card: game.CardReference{
					Kind:   game.CardReferenceLinked,
					LinkID: string(key),
				},
				FromZone:    zone.Library,
				Destination: zone.Hand,
			}},
			{
				Primitive: game.LoseLife{
					Player: game.ControllerReference(),
					Amount: game.Dynamic(manaValue),
				},
				Condition: opt.Val(selfGate),
			},
			{
				Primitive: game.LoseLife{
					PlayerGroup: game.OpponentsReference(),
					Amount:      game.Dynamic(manaValue),
				},
				Condition: opt.Val(opponentsGate),
			},
		},
	}.Ability(), true
}

// lowerRevealUntilSequence lowers the closed "reveal cards from the top of
// <library> until <player> reveal a <type> card, then put those cards into
// <zone>" family (Undercity Informer, Balustrade Spy, Treasure Hunt) into a
// single RevealUntil primitive. The parser marks all three effects with
// RevealUntilThenPut, records the boundary card type on the match reveal's
// selector, and the destination on the put effect's ToZone. This text-blind
// lowerer reads only those typed fields plus the head reveal's player subject;
// any shape mismatch or unmodeled subject fails closed.
func lowerRevealUntilSequence(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) != 3 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	revealUntil := ctx.content.Effects[0]
	matchReveal := ctx.content.Effects[1]
	put := ctx.content.Effects[2]
	if !revealUntil.RevealUntilThenPut ||
		!matchReveal.RevealUntilThenPut ||
		!put.RevealUntilThenPut ||
		revealUntil.Kind != compiler.EffectReveal ||
		matchReveal.Kind != compiler.EffectReveal ||
		put.Kind != compiler.EffectPut {
		return game.AbilityContent{}, false
	}
	if put.ToZone != zone.Graveyard && put.ToZone != zone.Hand {
		return game.AbilityContent{}, false
	}
	until, ok := cardSelectionForSelector(matchReveal.Selector)
	if !ok {
		return game.AbilityContent{}, false
	}
	primitive := game.RevealUntil{
		Until:       until,
		Destination: put.ToZone,
	}
	targets, ok := revealUntilPlayerSubject(ctx, revealUntil.Context, &primitive)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Targets: targets,
		Sequence: []game.Instruction{
			{Primitive: primitive},
		},
	}.Ability(), true
}

// revealUntilPlayerSubject resolves the reveal's player subject from the head
// reveal effect's typed Context, setting the primitive's single Player or group
// PlayerGroup, and returns the player target spec when the subject is a single
// target player. Unmodeled subjects fail closed.
func revealUntilPlayerSubject(
	ctx contentCtx,
	context parser.EffectContextKind,
	primitive *game.RevealUntil,
) ([]game.TargetSpec, bool) {
	switch context {
	case parser.EffectContextController:
		if len(ctx.content.Targets) != 0 {
			return nil, false
		}
		primitive.Player = game.ControllerReference()
		return nil, true
	case parser.EffectContextEachOpponent, parser.EffectContextEachOtherPlayer:
		if len(ctx.content.Targets) != 0 {
			return nil, false
		}
		primitive.PlayerGroup = game.OpponentsReference()
		return nil, true
	case parser.EffectContextEachPlayer:
		if len(ctx.content.Targets) != 0 {
			return nil, false
		}
		primitive.PlayerGroup = game.AllPlayersReference()
		return nil, true
	case parser.EffectContextTarget, parser.EffectContextPriorSubject:
		if len(ctx.content.Targets) != 1 {
			return nil, false
		}
		targetSpec, ok := playerTargetSpec(ctx.content.Targets[0])
		if !ok {
			return nil, false
		}
		primitive.Player = game.TargetPlayerReference(0)
		return []game.TargetSpec{targetSpec}, true
	default:
		// Other contexts are not supported player recipients; fail closed.
	}
	return nil, false
}

// lowerRevealChooseHandDiscardSequence lowers the targeted hand-disruption
// family "Target player reveals their hand. You choose a [filter] card from it.
// That player discards that card.[ You lose N life.]" (Coercion, Duress,
// Thoughtseize, Inquisition of Kozilek) into a single ChooseDiscardFromHand
// primitive (optionally followed by a controller LoseLife for the Thoughtseize
// rider). The parser marks the reveal and discard halves with
// RevealChooseDiscard and folds the filter onto the discard's HandChoiceDiscard;
// this text-blind lowerer reads only those typed fields plus the lone target
// player. Any shape mismatch fails closed so the general sequence path is
// untouched.
func lowerRevealChooseHandDiscardSequence(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	effects := ctx.content.Effects
	if len(effects) != 2 && len(effects) != 3 {
		return game.AbilityContent{}, false
	}
	reveal := effects[0]
	discard := effects[1]
	if !reveal.RevealChooseDiscard ||
		!discard.RevealChooseDiscard ||
		reveal.Kind != compiler.EffectReveal ||
		discard.Kind != compiler.EffectDiscard ||
		!discard.HandChoiceDiscard.Present ||
		reveal.Context != parser.EffectContextTarget ||
		discard.Context != parser.EffectContextReferencedPlayer {
		return game.AbilityContent{}, false
	}
	targetSpec, ok := playerTargetSpec(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, false
	}
	primitive := game.ChooseDiscardFromHand{
		Player:          game.TargetPlayerReference(0),
		ExcludeCreature: discard.HandChoiceDiscard.ExcludeCreature,
		ExcludeLand:     discard.HandChoiceDiscard.ExcludeLand,
	}
	if selection, ok := handChoiceDiscardSelection(discard.HandChoiceDiscard); ok {
		primitive.Selection = selection
	}
	if discard.HandChoiceDiscard.HasMaxManaValue {
		primitive.MaxManaValue = opt.Val(discard.HandChoiceDiscard.MaxManaValue)
	}
	sequence := []game.Instruction{{Primitive: primitive}}
	if len(effects) == 3 {
		lose, ok := revealChooseDiscardLifeLoss(effects[2])
		if !ok {
			return game.AbilityContent{}, false
		}
		sequence = append(sequence, game.Instruction{Primitive: lose})
	}
	return game.Mode{
		Targets:  []game.TargetSpec{targetSpec},
		Sequence: sequence,
	}.Ability(), true
}

// revealChooseDiscardLifeLoss lowers the optional trailing "You lose N life."
// rider (Thoughtseize) into a controller LoseLife of a fixed positive amount.
// Any non-fixed or non-controller life change fails closed.
func revealChooseDiscardLifeLoss(effect compiler.CompiledEffect) (game.LoseLife, bool) {
	if effect.Kind != compiler.EffectLose ||
		effect.Context != parser.EffectContextController ||
		!effect.LifeObject ||
		effect.Negated ||
		!effect.Amount.Known ||
		effect.Amount.Value < 1 {
		return game.LoseLife{}, false
	}
	return game.LoseLife{
		Amount: game.Fixed(effect.Amount.Value),
		Player: game.ControllerReference(),
	}, true
}

// handChoiceDiscardSelection builds the canonical card filter for a positive
// reveal-choose-discard descriptor ("a creature card", "an artifact or creature
// card", "a nonbasic land card"). It returns ok=false for the zero filter (the
// any-card / pure noncreature-nonland forms already modeled by the primitive's
// exclude flags) so those continue to emit no Selection. The Selection composes
// conjunctively with the primitive's ExcludeCreature/ExcludeLand flags and
// mana-value bound, so the supertype exclusion and required-type union here are
// the only fields it carries.
func handChoiceDiscardSelection(filter parser.HandChoiceDiscardSyntax) (game.Selection, bool) {
	if len(filter.RequiredTypesAny) == 0 && filter.ExcludedSupertype == "" {
		return game.Selection{}, false
	}
	selection := game.Selection{
		RequiredTypesAny: slices.Clone(filter.RequiredTypesAny),
	}
	if filter.ExcludedSupertype != "" {
		selection.ExcludedSupertype = filter.ExcludedSupertype
	}
	return selection, true
}

func referencesContainBinding(references []compiler.CompiledReference, binding compiler.ReferenceBinding, prior int) bool {
	for i := range references {
		if references[i].Binding != binding {
			continue
		}
		if binding == compiler.ReferenceBindingTarget && references[i].Occurrence == prior {
			return true
		}
		if binding == compiler.ReferenceBindingPriorInstructionResult && references[i].PriorInstruction == prior {
			return true
		}
	}
	return false
}

// lowerRemovalManifestSequence lowers the ordered pair "<Exile/Destroy> target
// creature. Its controller manifests [dread / the top card of their library]."
// (Reality Shift, Unwanted Remake) into a removal of the single target followed
// by a manifest performed by that target's controller. The manifesting player is
// bound to the controller of the removed permanent (resolved through last-known
// information after it leaves the battlefield), so cards are manifested from that
// player's library. It accepts only the controller-subject removal paired with a
// referenced-controller manifest whose "Its" reference resolves to the lone
// target; every other shape (mass removal, multiple targets, or any added clause)
// fails closed so the general sequence path is untouched.
func lowerRemovalManifestSequence(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) != 2 ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	removal := ctx.content.Effects[0]
	manifest := ctx.content.Effects[1]
	if removal.Context != parser.EffectContextController ||
		!removal.Exact ||
		removal.Optional ||
		removal.Negated ||
		removal.Duration != compiler.DurationNone ||
		len(removal.Targets) != 1 ||
		len(removal.References) != 0 {
		return game.AbilityContent{}, false
	}
	dread := manifest.Kind == compiler.EffectManifestDread
	if manifest.Kind != compiler.EffectManifest && !dread ||
		manifest.Context != parser.EffectContextReferencedObjectController ||
		manifest.Optional ||
		manifest.Negated ||
		len(manifest.Targets) != 0 ||
		!referencesContainBinding(manifest.References, compiler.ReferenceBindingTarget, 0) {
		return game.AbilityContent{}, false
	}
	target, ok := permanentTargetSpec(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, false
	}
	removed := game.TargetPermanentReference(0)
	var removalPrimitive game.Primitive
	switch removal.Kind {
	case compiler.EffectExile:
		removalPrimitive = game.Exile{Object: removed}
	case compiler.EffectDestroy:
		removalPrimitive = game.Destroy{Object: removed}
	default:
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Targets: []game.TargetSpec{target},
		Sequence: []game.Instruction{
			{Primitive: removalPrimitive},
			{Primitive: game.Manifest{Dread: dread, Player: game.ObjectControllerReference(removed)}},
		},
	}.Ability(), true
}

// lowerTapDownSequence lowers the "tap then stun" sequence — "Tap <target
// permanent>. <It / That permanent> doesn't untap during its controller's next
// untap step." — into a tap of the single target followed by a SkipNextUntap on
// that same permanent. It accepts only the parser-exact singular prior-subject
// "next untap step" clause whose references all resolve to the tapped target;
// every other shape (multi-target, plural "those creatures", "next two untap
// steps", or any added clause) fails closed so the general sequence path is
// untouched.
func lowerTapDownSequence(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 || ctx.optional ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	tap := ctx.content.Effects[0]
	stun := ctx.content.Effects[1]
	if tap.Kind != compiler.EffectTap || tap.Negated || tap.Optional || !tap.Exact ||
		tap.Context != parser.EffectContextController ||
		len(tap.References) != 0 || len(tap.Targets) != 1 {
		return game.AbilityContent{}, false
	}
	if stun.Kind != compiler.EffectUntap || !stun.Negated || stun.Optional || !stun.Exact ||
		stun.Context != parser.EffectContextReferencedObject ||
		len(stun.Targets) != 0 {
		return game.AbilityContent{}, false
	}
	// Every content-level reference must be the stun clause's prior-subject
	// reference to the tapped permanent (target 0); reject anything else so no
	// reference is silently dropped.
	for _, ref := range ctx.content.References {
		if ref.Binding != compiler.ReferenceBindingTarget || ref.Occurrence != 0 {
			return game.AbilityContent{}, false
		}
	}
	targetSpec, ok := permanentTargetSpec(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{
			{Primitive: game.Tap{Object: game.TargetPermanentReference(0)}},
			{Primitive: game.SkipNextUntap{Object: game.TargetPermanentReference(0)}},
		},
	}.Ability(), true
}

// lowerTapStunSequence lowers the multi-target "tap then stun" sequence — "Tap
// up to two target creatures. Those creatures don't untap during their
// controller's next untap step." — into one Tap per target slot followed by one
// SkipNextUntap per target slot, all addressing the same multi-target permanent
// spec. It generalizes lowerTapDownSequence to the plural "those creatures"
// prior-subject form, which the parser leaves as an EffectContextUnknown stun
// clause whose anaphora ("those creatures", "their") are ambiguous between the
// several chosen targets; lowerTapDownSequence's singular
// EffectContextReferencedObject gate rejects exactly that form. The runtime
// Tap/SkipNextUntap handlers no-op on an unresolved target slot, so an "up to N"
// tap-stun safely affects only the chosen targets. Every other shape (added
// clauses, a multi-step "next two untap steps" window — which the parser splits
// into three effects — mass "all creatures", non-target references, or any
// reference outside the stun clause) fails closed so the general sequence path
// is untouched.
func lowerTapStunSequence(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 || ctx.optional ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	tap := ctx.content.Effects[0]
	stun := ctx.content.Effects[1]
	if tap.Kind != compiler.EffectTap || tap.Negated || tap.Optional || !tap.Exact ||
		tap.Context != parser.EffectContextController ||
		len(tap.References) != 0 || len(tap.Targets) != 1 {
		return game.AbilityContent{}, false
	}
	if stun.Kind != compiler.EffectUntap || !stun.Negated || stun.Optional || !stun.Exact ||
		stun.Context != parser.EffectContextUnknown ||
		len(stun.Targets) != 0 {
		return game.AbilityContent{}, false
	}
	// Every content reference must be the stun clause's plural anaphor back to
	// the tapped permanents (target 0) — "those creatures" and the "their"
	// possessive in "their controller's next untap step". Require each reference
	// to fall within the stun clause span and resolve to target 0, so no
	// reference that would need its own instruction is silently dropped.
	for _, ref := range ctx.content.References {
		if ref.Occurrence != 0 ||
			!spanCovered(ref.Span, []shared.Span{stun.Span}) ||
			(ref.Binding != compiler.ReferenceBindingTarget &&
				ref.Binding != compiler.ReferenceBindingAmbiguous) {
			return game.AbilityContent{}, false
		}
	}
	targetSpec, ok := permanentTargetSpecWithCardinality(ctx.content.Targets[0])
	if !ok || targetSpec.MaxTargets < 1 {
		return game.AbilityContent{}, false
	}
	sequence := make([]game.Instruction, 0, 2*targetSpec.MaxTargets)
	for i := range targetSpec.MaxTargets {
		sequence = append(sequence, game.Instruction{
			Primitive: game.Tap{Object: game.TargetPermanentReference(i)},
		})
	}
	for i := range targetSpec.MaxTargets {
		sequence = append(sequence, game.Instruction{
			Primitive: game.SkipNextUntap{Object: game.TargetPermanentReference(i)},
		})
	}
	return game.Mode{
		Targets:  []game.TargetSpec{targetSpec},
		Sequence: sequence,
	}.Ability(), true
}

// lowerControlledGroupSkipUntapEffect lowers the mass self-stun "<group> you
// control don't untap during your next untap step." (Rhonas's Last Stand's "Lands
// you control ...", and the parallel creatures/permanents/artifacts wordings) into
// a single group SkipNextUntap over the controlled-permanent group. The affected
// group is the source controller's own permanents (recorded in StaticSubject by
// the parser) and the window is that controller's own next untap step, so the
// clause carries no target or reference. It accepts only the parser-exact negated
// untap effect whose StaticSubject is a controlled-permanent group with no
// duration; every other shape (a targeted player's permanents, a multi-step
// window, a color or subtype filter) fails closed so the general paths are
// untouched.
func lowerControlledGroupSkipUntapEffect(ctx contentCtx) (game.AbilityContent, bool) {
	// Invariant: the sole caller dispatches this lowerer only inside the
	// `len(ctx.content.Effects) == 1` branch (lower_spell.go), so a length other
	// than one is an upstream bug rather than an unsupported card shape.
	if len(ctx.content.Effects) != 1 {
		panic(fmt.Sprintf("lowerControlledGroupSkipUntapEffect: expected a single effect, got %d", len(ctx.content.Effects)))
	}
	if ctx.optional ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.References) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectUntap || !effect.Negated || effect.Optional || !effect.Exact ||
		effect.Duration != compiler.DurationNone || effect.DelayedTiming != 0 ||
		len(effect.Targets) != 0 || len(effect.References) != 0 {
		return game.AbilityContent{}, false
	}
	switch effect.StaticSubject {
	case compiler.StaticSubjectControlledLands,
		compiler.StaticSubjectControlledCreatures,
		compiler.StaticSubjectControlledPermanents,
		compiler.StaticSubjectControlledArtifacts:
	default:
		return game.AbilityContent{}, false
	}
	group, ok := resolvingStaticSubjectGroup(&effect)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Sequence: []game.Instruction{
			{Primitive: game.SkipNextUntap{Group: group}},
		},
	}.Ability(), true
}

// lowerStandaloneStunEffect lowers the standalone targeted stun "Target
// <permanent> doesn't untap during its controller's next untap step." (Sleeper
// Dart, House Guildmage, Skyline Cascade) into a single SkipNextUntap on the
// effect's own permanent target. Unlike the tap-down sequence, no tap precedes
// the stun: the spell or ability only denies the target its next untap. It
// accepts only the parser-exact single negated-untap effect carrying one
// permanent target and one possessive "its controller's" reference that resolves
// to that target; every other shape (added clauses, mass or plural wording, a
// non-target reference, or a multi-step window) fails closed so the general
// paths are untouched.
func lowerStandaloneStunEffect(ctx contentCtx) (game.AbilityContent, bool) {
	// Invariant: the sole caller dispatches this lowerer only inside the
	// `len(ctx.content.Effects) == 1` branch (lower_spell.go:297), so a length
	// other than one is an upstream bug rather than an unsupported card shape.
	if len(ctx.content.Effects) != 1 {
		panic(fmt.Sprintf("lowerStandaloneStunEffect: expected a single effect, got %d", len(ctx.content.Effects)))
	}
	if ctx.optional ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Conditions) != 0 {
		return game.AbilityContent{}, false
	}
	stun := ctx.content.Effects[0]
	if stun.Kind != compiler.EffectUntap || !stun.Negated || stun.Optional || !stun.Exact ||
		stun.Context != parser.EffectContextTarget ||
		stun.Duration != compiler.DurationNone || stun.DelayedTiming != 0 ||
		len(stun.Targets) != 1 {
		return game.AbilityContent{}, false
	}
	// Every content reference must be the stun clause's possessive "its" pointing
	// at the targeted permanent (target 0); reject anything else so no reference
	// is silently dropped.
	if len(ctx.content.References) == 0 {
		return game.AbilityContent{}, false
	}
	for _, ref := range ctx.content.References {
		if ref.Binding != compiler.ReferenceBindingTarget || ref.Occurrence != 0 {
			return game.AbilityContent{}, false
		}
	}
	targetSpec, ok := permanentTargetSpec(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{
			{Primitive: game.SkipNextUntap{Object: game.TargetPermanentReference(0)}},
		},
	}.Ability(), true
}

// lowerInheritedSubjectStunEffect lowers a prior-subject "doesn't untap during
// its controller's next untap step" clause that an ordered sequence feeds in as
// a standalone sub-clause whose permanent is an inherited target rather than the
// clause's own ("Tap target creature. Its controller mills two cards. That
// creature doesn't untap during its controller's next untap step." — Glacial
// Grasp; "Tap up to two target creatures. Those creatures don't untap during
// their controller's next untap step. Scry 1." — Sudden Storm). The dedicated
// two-effect lowerTapDownSequence/lowerTapStunSequence shapes only fire when the
// tap and stun are the body's only effects, so any added clause (a mill, a scry,
// or a prevent) diverts the body to the generic ordered-sequence path, which
// rebases the just-tapped subject onto this clause's content targets and lowers
// it here as one SkipNextUntap per inherited target slot. It accepts only the
// parser-exact negated-untap clause that owns no target of its own (every
// permanent is the inherited subject) and whose references all resolve to that
// inherited target; every other shape (an own target, a multi-step window, a
// duration, or a stray reference) fails closed so lowerStandaloneStunEffect and
// the general paths are untouched.
func lowerInheritedSubjectStunEffect(ctx contentCtx) (game.AbilityContent, bool) {
	// Invariant: the sole caller dispatches this lowerer only inside the
	// `len(ctx.content.Effects) == 1` branch (lower_spell.go:297), so a length
	// other than one is an upstream bug rather than an unsupported card shape.
	if len(ctx.content.Effects) != 1 {
		panic(fmt.Sprintf("lowerInheritedSubjectStunEffect: expected a single effect, got %d", len(ctx.content.Effects)))
	}
	if ctx.optional ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Conditions) != 0 {
		return game.AbilityContent{}, false
	}
	stun := ctx.content.Effects[0]
	if stun.Kind != compiler.EffectUntap || !stun.Negated || stun.Optional || !stun.Exact ||
		(stun.Context != parser.EffectContextReferencedObject &&
			stun.Context != parser.EffectContextUnknown) ||
		stun.Duration != compiler.DurationNone || stun.DelayedTiming != 0 ||
		len(stun.Targets) != 0 {
		return game.AbilityContent{}, false
	}
	// Every content reference is the stun clause's anaphor back to the inherited
	// subject — the demonstrative subject ("That creature"/"Those creatures") and
	// the possessive controller pronoun ("its"/"their") — and must resolve to the
	// inherited target (target 0); reject anything else so no reference is
	// silently dropped.
	if len(ctx.content.References) == 0 {
		return game.AbilityContent{}, false
	}
	for _, ref := range ctx.content.References {
		if (ref.Binding != compiler.ReferenceBindingTarget &&
			ref.Binding != compiler.ReferenceBindingAmbiguous) ||
			ref.Occurrence != 0 {
			return game.AbilityContent{}, false
		}
	}
	targetSpec, ok := permanentTargetSpecWithCardinality(ctx.content.Targets[0])
	if !ok || targetSpec.MaxTargets < 1 {
		return game.AbilityContent{}, false
	}
	sequence := make([]game.Instruction, 0, targetSpec.MaxTargets)
	for i := range targetSpec.MaxTargets {
		sequence = append(sequence, game.Instruction{
			Primitive: game.SkipNextUntap{Object: game.TargetPermanentReference(i)},
		})
	}
	return game.Mode{
		Targets:  []game.TargetSpec{targetSpec},
		Sequence: sequence,
	}.Ability(), true
}

// lowerEventSubjectStunEffect lowers a prior-subject "doesn't untap during its
// controller's next untap step" clause whose stunned permanent is a triggering
// combat event's creature rather than a chosen target ("Whenever this creature
// blocks a creature, that creature doesn't untap during its controller's next
// untap step." — Labyrinth Minotaur; "Whenever this creature attacks, it doesn't
// untap during its controller's next untap step." — Apes of Rath). It also
// lowers the stun half of the combat-damage "tap that creature and it doesn't
// untap during its controller's next untap step." cycle (Kashi-Tribe Warriors,
// Mercurial Kite), which reaches this lowerer as the second clause of the
// ordered tap-then-stun sequence. The stunned creature is named by the clause's
// single subject reference, an event permanent (the creature this one dealt
// combat damage to / blocked / the attacking source); SkipNextUntap addresses
// that creature directly and the runtime denies its controller's next untap, so
// the clause's possessive "its controller" reference is consumed without its own
// instruction. It accepts only the parser-exact negated-untap clause that owns
// no target and whose every reference binds to an event or source permanent;
// every other shape (an own target, a multi-step window, a duration, or a
// reference needing its own instruction) fails closed so the target and inherited
// stun paths are untouched.
func lowerEventSubjectStunEffect(ctx contentCtx) (game.AbilityContent, bool) {
	// Invariant: the sole caller dispatches this lowerer only inside the
	// `len(ctx.content.Effects) == 1` branch (lower_spell.go:297), so a length
	// other than one is an upstream bug rather than an unsupported card shape.
	if len(ctx.content.Effects) != 1 {
		panic(fmt.Sprintf("lowerEventSubjectStunEffect: expected a single effect, got %d", len(ctx.content.Effects)))
	}
	if ctx.optional ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Conditions) != 0 {
		return game.AbilityContent{}, false
	}
	stun := ctx.content.Effects[0]
	if stun.Kind != compiler.EffectUntap || !stun.Negated || stun.Optional || !stun.Exact ||
		stun.Context != parser.EffectContextReferencedObject ||
		stun.Duration != compiler.DurationNone || stun.DelayedTiming != 0 ||
		len(stun.Targets) != 0 ||
		len(stun.SubjectReferences) != 1 {
		return game.AbilityContent{}, false
	}
	referenceCtx := referenceLoweringContext{AllowEvent: true, AllowSource: true}
	object, ok := lowerObjectReference(stun.SubjectReferences[0], referenceCtx)
	if !ok {
		return game.AbilityContent{}, false
	}
	// Every clause reference is the subject anaphor or the possessive "its
	// controller" pronoun; require each to resolve as an event or source object
	// so no reference that would need its own instruction is silently dropped.
	if len(ctx.content.References) == 0 {
		return game.AbilityContent{}, false
	}
	for _, ref := range ctx.content.References {
		if _, ok := lowerObjectReference(ref, referenceCtx); !ok {
			return game.AbilityContent{}, false
		}
	}
	return game.Mode{
		Sequence: []game.Instruction{
			{Primitive: game.SkipNextUntap{Object: object}},
		},
	}.Ability(), true
}

// lowerStandaloneSourceStunEffect lowers the self-source stun "This <permanent>
// doesn't untap during your next untap step." (the dual lands Mogg Hollows /
// Rootwater Depths and Arbalest Elite) into a single SkipNextUntap on the
// source itself. The stunned permanent is the resolving source — there is no
// target and no tap — so the instruction references the source permanent
// directly. It accepts only the parser-exact single negated-untap effect whose
// context is the source and whose one reference binds to the source ("This
// land"/"This creature"); every other shape (added clauses, a target, mass or
// plural wording, a non-source reference, or a multi-step window) fails closed.
// Multi-effect abilities reach this lowerer per clause through the ordered
// sequence path, so "{T}: Add {R} or {G}. This land doesn't untap ..." lowers
// the appended stun as its own instruction.
func lowerStandaloneSourceStunEffect(ctx contentCtx) (game.AbilityContent, bool) {
	// Invariant: the sole caller dispatches this lowerer only inside the
	// `len(ctx.content.Effects) == 1` branch (lower_spell.go:297), so a length
	// other than one is an upstream bug rather than an unsupported card shape.
	if len(ctx.content.Effects) != 1 {
		panic(fmt.Sprintf("lowerStandaloneSourceStunEffect: expected a single effect, got %d", len(ctx.content.Effects)))
	}
	if ctx.optional ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Conditions) != 0 {
		return game.AbilityContent{}, false
	}
	stun := ctx.content.Effects[0]
	if stun.Kind != compiler.EffectUntap || !stun.Negated || stun.Optional || !stun.Exact ||
		stun.Context != parser.EffectContextSource ||
		stun.Duration != compiler.DurationNone || stun.DelayedTiming != 0 ||
		len(stun.Targets) != 0 {
		return game.AbilityContent{}, false
	}
	// The clause's single reference must be the self "This <permanent>" subject
	// binding to the source; reject anything else so no reference is dropped.
	if len(ctx.content.References) != 1 ||
		!referencesBindTo(ctx.content.References, compiler.ReferenceBindingSource, 0) {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Sequence: []game.Instruction{
			{Primitive: game.SkipNextUntap{Object: game.SourcePermanentReference()}},
		},
	}.Ability(), true
}

// blink uses to name the several exiled cards. It distinguishes the multi-target
// flicker from the singular "it"/"that card" single-target blink, which the
// per-clause path lowers on its own.
func referencesIncludeThose(refs []compiler.CompiledReference) bool {
	for _, ref := range refs {
		if ref.Kind == compiler.ReferencePronoun && ref.Pronoun == compiler.ReferencePronounThose {
			return true
		}
	}
	return false
}

// lowerGroupBlinkSequence lowers the multi-target "blink" (flicker) sequence —
// "Exile <N> target <permanents> you control, then return those cards to the
// battlefield under [your|their owner's] control." (Illusionist's Stratagem) and
// its delayed "… at the beginning of the next end step." variant (Eerie
// Interlude-style). It exiles each chosen permanent under its own linked key and
// returns each from exile, so the cards leave and re-enter the battlefield as new
// objects (re-triggering enters abilities). The single-target form ("… then
// return it …") keeps its singular "it"/"that card" back-reference and is left to
// the per-clause blink path; this lowerer requires the plural "those" demonstrative
// and a multi-target cardinality so the two never overlap.
//
// Both the immediate (", then return …") and delayed ("… at the beginning of the
// next end step") return timings are modeled, as are the "under your control" and
// "under their owner's control" controller riders and a fixed "with a <kind>
// counter on it" entry-counter rider. Every other shape — singular back-reference,
// single-target cardinality, negated or optional clauses, added references, color
// or type entry choices — fails closed so the general sequence path is untouched.
func lowerGroupBlinkSequence(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 || ctx.optional ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	exile := ctx.content.Effects[0]
	ret := ctx.content.Effects[1]
	if exile.Kind != compiler.EffectExile || exile.Negated || exile.Optional || !exile.Exact ||
		exile.Context != parser.EffectContextController ||
		exile.DelayedTiming != 0 ||
		len(exile.References) != 0 || len(exile.Targets) != 1 {
		return game.AbilityContent{}, false
	}
	target := ctx.content.Targets[0]
	if targetCardinalityIsOne(target) {
		return game.AbilityContent{}, false
	}
	delayed, entryCounters, ok := blinkReturnRider(ret)
	if !ok {
		return game.AbilityContent{}, false
	}
	// Every content reference must be the return clause's plural back-reference to
	// the exiled cards (prior instruction 0). Requiring the "those" demonstrative
	// keeps the singular single-target blink on its own path.
	if !referencesBindTo(ctx.content.References, compiler.ReferenceBindingPriorInstructionResult, 0) ||
		!referencesIncludeThose(ctx.content.References) {
		return game.AbilityContent{}, false
	}
	consumed := ctx
	consumed.content.References = nil
	consumed.content.Targets = nil
	if consumed.content.Unconsumed() {
		return game.AbilityContent{}, false
	}
	targetSpec, ok := permanentTargetSpecAllowingUnbounded(target, true)
	if !ok || targetSpec.MaxTargets < 1 {
		return game.AbilityContent{}, false
	}
	makePut := func(key game.LinkedKey) game.PutOnBattlefield {
		put := game.PutOnBattlefield{
			Source:        game.LinkedBattlefieldSource(key),
			EntryTapped:   ret.EntersTapped,
			EntryCounters: entryCounters,
		}
		if ret.UnderYourControl {
			put.Recipient = opt.Val(game.ControllerReference())
		}
		return put
	}
	// The unbounded "any number of target" count cannot unroll a fixed slot per
	// target, so capture every chosen permanent under one linked key with a
	// single exile and return the whole group with one put — the runtime
	// all-target-permanents reference and linked-battlefield source carry the
	// group across the exile and return.
	if targetCardinalityIsUnbounded(target) {
		const key = game.LinkedKey("group-blink")
		exileInstr := game.Instruction{Primitive: game.Exile{
			Object:         game.AllTargetPermanentsReference(0),
			ExileLinkedKey: key,
		}}
		putInstr := game.Instruction{Primitive: makePut(key)}
		var sequence []game.Instruction
		if delayed {
			sequence = []game.Instruction{exileInstr, {Primitive: game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
				Timing:  game.DelayedAtBeginningOfNextEndStep,
				Content: game.Mode{Sequence: []game.Instruction{putInstr}}.Ability(),
			}}}}
		} else {
			sequence = []game.Instruction{exileInstr, putInstr}
		}
		return game.Mode{
			Targets:  []game.TargetSpec{targetSpec},
			Sequence: sequence,
		}.Ability(), true
	}
	sequence := make([]game.Instruction, 0, 2*targetSpec.MaxTargets)
	keys := make([]game.LinkedKey, targetSpec.MaxTargets)
	for i := range targetSpec.MaxTargets {
		key := game.LinkedKey(fmt.Sprintf("group-blink-%d", i))
		keys[i] = key
		sequence = append(sequence, game.Instruction{Primitive: game.Exile{
			Object:         game.TargetPermanentReference(i),
			ExileLinkedKey: key,
		}})
	}
	puts := make([]game.Instruction, 0, targetSpec.MaxTargets)
	for i := range targetSpec.MaxTargets {
		puts = append(puts, game.Instruction{Primitive: makePut(keys[i])})
	}
	if delayed {
		sequence = append(sequence, game.Instruction{Primitive: game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
			Timing:  game.DelayedAtBeginningOfNextEndStep,
			Content: game.Mode{Sequence: puts}.Ability(),
		}}})
	} else {
		sequence = append(sequence, puts...)
	}
	return game.Mode{
		Targets:  []game.TargetSpec{targetSpec},
		Sequence: sequence,
	}.Ability(), true
}

// blinkReturnRider validates a group-blink return clause and reports whether the
// return is delayed to the next end step (versus an immediate "then return")
// along with any fixed entry-counter rider. It accepts only the battlefield
// return form with no color/type entry choice and a fixed, known, positive
// counter of a known kind, failing closed for every other return shape so the
// group and mass blink lowerers share one return contract.
func blinkReturnRider(ret compiler.CompiledEffect) (delayed bool, entryCounters []game.CounterPlacement, ok bool) {
	switch {
	case ret.Connection == parser.EffectConnectionThen && ret.DelayedTiming == 0:
		delayed = false
	case ret.DelayedTiming == game.DelayedAtBeginningOfNextEndStep:
		delayed = true
	default:
		return false, nil, false
	}
	if ret.Kind != compiler.EffectReturn || ret.Negated ||
		ret.ToZone != zone.Battlefield ||
		ret.EntersColorChoice || ret.EntersTypeChoice || ret.EntersWithCounters {
		return false, nil, false
	}
	if ret.CounterKindKnown {
		if !ret.Amount.Known || ret.Amount.Value < 1 {
			return false, nil, false
		}
		entryCounters = []game.CounterPlacement{{Kind: ret.CounterKind, Amount: ret.Amount.Value}}
	}
	return delayed, entryCounters, true
}

// lowerMassGroupBlinkSequence lowers the untargeted mass "blink" (flicker)
// sequence — "Exile each <permanent> you control. Return those cards to the
// battlefield under their owner's control [at the beginning of the next end
// step]." (Ghostway) — into one group exile that captures every exiled permanent
// under a single linked key and one return that brings the whole group back. It
// requires the exile to be an exact mass ("each"/"all") group selection with no
// targets so the cards leave and re-enter as new objects. Both the immediate
// and delayed return timings, the controller riders, and a fixed entry-counter
// rider are modeled; every other shape — a targeted exile, a singular
// back-reference, negated or optional clauses, an unexpressible mass selector —
// fails closed so the targeted group path and the general sequence path are
// untouched.
func lowerMassGroupBlinkSequence(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 || ctx.optional ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	exile := ctx.content.Effects[0]
	ret := ctx.content.Effects[1]
	if exile.Kind != compiler.EffectExile || exile.Negated || exile.Optional || !exile.Exact ||
		exile.Context != parser.EffectContextController ||
		exile.DelayedTiming != 0 ||
		!exile.Selector.All ||
		len(exile.References) != 0 || len(exile.Targets) != 0 {
		return game.AbilityContent{}, false
	}
	selection, ok := massGroupSelection(exile.Selector)
	if !ok {
		return game.AbilityContent{}, false
	}
	delayed, entryCounters, ok := blinkReturnRider(ret)
	if !ok {
		return game.AbilityContent{}, false
	}
	if !referencesBindTo(ctx.content.References, compiler.ReferenceBindingPriorInstructionResult, 0) ||
		!referencesIncludeThose(ctx.content.References) {
		return game.AbilityContent{}, false
	}
	consumed := ctx
	consumed.content.References = nil
	if consumed.content.Unconsumed() {
		return game.AbilityContent{}, false
	}
	const key = game.LinkedKey("group-blink")
	exileInstr := game.Instruction{Primitive: game.Exile{
		Group:          game.BattlefieldGroup(selection),
		ExileLinkedKey: key,
	}}
	put := game.PutOnBattlefield{
		Source:        game.LinkedBattlefieldSource(key),
		EntryTapped:   ret.EntersTapped,
		EntryCounters: entryCounters,
	}
	if ret.UnderYourControl {
		put.Recipient = opt.Val(game.ControllerReference())
	}
	putInstr := game.Instruction{Primitive: put}
	var sequence []game.Instruction
	if delayed {
		sequence = []game.Instruction{exileInstr, {Primitive: game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
			Timing:  game.DelayedAtBeginningOfNextEndStep,
			Content: game.Mode{Sequence: []game.Instruction{putInstr}}.Ability(),
		}}}}
	} else {
		sequence = []game.Instruction{exileInstr, putInstr}
	}
	return game.Mode{Sequence: sequence}.Ability(), true
}

// lowerPileSplitSequence lowers the closed pile-split family "reveal the top N
// cards of your library[ and separate them into two piles]. An opponent
// {separates those cards into|chooses one of} two piles. Put {one|that} pile
// into your hand and the other into your graveyard." (Fact or Fiction, Steam
// Augury, Sphinx of Uthuun) into a single PileSplit primitive. The parser marks
// the reveal and put effects with PileSplitSequence and records the separate and
// choose roles, the non-kept destination, the revealed amount, and the
// zero-effect middle sentence span on the put effect. This text-blind lowerer
// reads only those typed fields; any shape mismatch or unmodeled destination
// fails closed.
func lowerPileSplitSequence(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 || ctx.optional ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	reveal := ctx.content.Effects[0]
	put := ctx.content.Effects[1]
	if reveal.Kind != compiler.EffectReveal || !reveal.PileSplitSequence ||
		put.Kind != compiler.EffectPut || !put.PileSplitSequence ||
		reveal.Context != parser.EffectContextController ||
		put.Context != parser.EffectContextController ||
		put.PileSplitAmount < 1 {
		return game.AbilityContent{}, false
	}
	if put.PileSplitOtherZone != zone.Graveyard && put.PileSplitOtherZone != zone.Library {
		return game.AbilityContent{}, false
	}
	// The pile-split clauses' only references are the put clause's "those cards"/
	// "the other" anaphors back to the revealed cards, which the PileSplit
	// primitive models directly. Every content reference must fall within the
	// reveal, middle, or put spans so no reference needing its own instruction is
	// dropped.
	for ri := range ctx.content.References {
		if !spanCovered(ctx.content.References[ri].Span, []shared.Span{reveal.Span, put.PileSplitMiddleSpan, put.Span}) {
			return game.AbilityContent{}, false
		}
	}
	return game.Mode{
		Sequence: []game.Instruction{
			{Primitive: game.PileSplit{
				Player:            game.ControllerReference(),
				Amount:            game.Fixed(put.PileSplitAmount),
				SeparatorOpponent: put.PileSplitSeparatorOpponent,
				ChooserOpponent:   put.PileSplitChooserOpponent,
				Kept:              zone.Hand,
				Other:             put.PileSplitOtherZone,
			}},
		},
	}.Ability(), true
}

// lowerRevealTopPartitionSequence lowers the closed "Reveal the top N cards of
// your library. Put all <type> cards revealed this way into your hand and the
// rest <remainder>." family (Borborygmos Enraged, Sift Through Sands, the tribal
// "reveal and gather" creatures) into a single RevealTopPartition primitive. The
// parser marks the Reveal and Put effects with RevealTopPartition, leaves the
// fixed count on the Reveal's Amount and the typed filter on the Put's Selector,
// and records the remainder on RevealPartitionRemainder. This text-blind lowerer
// reads only those typed fields; any shape mismatch or unmodeled filter fails
// closed.
func lowerRevealTopPartitionSequence(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 || ctx.optional ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	reveal := ctx.content.Effects[0]
	put := ctx.content.Effects[1]
	if reveal.Kind != compiler.EffectReveal || !reveal.RevealTopPartition ||
		put.Kind != compiler.EffectPut || !put.RevealTopPartition ||
		reveal.Context != parser.EffectContextController ||
		put.Context != parser.EffectContextController ||
		!reveal.Amount.Known || reveal.Amount.Value < 1 {
		return game.AbilityContent{}, false
	}
	// The reveal-partition clauses' only references are the put clause's "revealed
	// this way" / "from among them" anaphors back to the revealed cards, which the
	// RevealTopPartition primitive models directly. Every content reference must
	// fall within the reveal or put spans so no reference needing its own
	// instruction is dropped.
	for ri := range ctx.content.References {
		if !spanCovered(ctx.content.References[ri].Span, []shared.Span{reveal.Span, put.Span}) {
			return game.AbilityContent{}, false
		}
	}
	selection, ok := cardSelectionForSelector(put.Selector)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Sequence: []game.Instruction{
			{Primitive: game.RevealTopPartition{
				Player:    game.ControllerReference(),
				Amount:    game.Fixed(reveal.Amount.Value),
				Selection: selection,
				Remainder: digRemainder(put.RevealPartitionRemainder),
			}},
		},
	}.Ability(), true
}

// of your library. Put M of them into your hand and the rest into your
// graveyard." — into a single Dig primitive: the controller looks at the top N
// cards, takes M into hand, and the remainder goes to their graveyard. It
// accepts only the parser-exact two-effect form whose look count exceeds its
// take count and that carries no targets, references, keywords, or optionality;
// every other shape (library-bottom remainder, variable counts, added clauses)
// fails closed so the general sequence path is untouched.
func lowerDigSequence(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 || ctx.optional ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	look := ctx.content.Effects[0]
	put := ctx.content.Effects[1]
	if look.Kind != compiler.EffectDig || !look.Exact || look.Optional || look.Negated ||
		look.Context != parser.EffectContextController ||
		!look.Amount.Known || len(look.Targets) != 0 || len(look.References) != 0 {
		return game.AbilityContent{}, false
	}
	if put.Kind != compiler.EffectPut || !put.Exact || !put.Dig.Put || put.Optional || put.Negated ||
		put.Context != parser.EffectContextController ||
		!put.Amount.Known || len(put.Targets) != 0 {
		return game.AbilityContent{}, false
	}
	// The only reference the exact impulse clauses carry is the put clause's
	// "them"/"those cards" anaphor back to the looked-at cards, which the Dig
	// primitive models directly. Every content reference must be one of the put
	// clause's references so no reference needing its own instruction is dropped.
	for ri := range ctx.content.References {
		if !spanCovered(ctx.content.References[ri].Span, []shared.Span{put.Span}) {
			return game.AbilityContent{}, false
		}
	}
	lookCount := look.Amount.Value
	takeCount := put.Amount.Value
	if takeCount < 1 || lookCount <= takeCount {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Sequence: []game.Instruction{
			{Primitive: game.Dig{
				Player:    game.ControllerReference(),
				Look:      game.Fixed(lookCount),
				Take:      game.Fixed(takeCount),
				Remainder: digRemainder(put.Dig.Remainder),
			}},
		},
	}.Ability(), true
}

// digRemainder maps the parser's recorded impulse remainder destination to the
// runtime Dig remainder. The library-bottom rider variants ("in any order" / "in
// a random order") share one runtime placement; only the graveyard default
// differs.
func digRemainder(remainder parser.DigRemainderKind) game.DigRemainder {
	switch remainder {
	case parser.DigRemainderLibraryBottom,
		parser.DigRemainderLibraryBottomAny,
		parser.DigRemainderLibraryBottomRandom:
		return game.DigRemainderLibraryBottom
	default:
		return game.DigRemainderGraveyard
	}
}

// lowerDrawHandLibrarySequence lowers "Draw N cards, then put M cards from your
// hand on top of your library in any order." The runtime MoveCard choice sees
// the post-draw hand and preserves the selected option order as library order.
func lowerDrawHandLibrarySequence(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 || ctx.optional ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Conditions) != 0 {
		return game.AbilityContent{}, false
	}
	draw := ctx.content.Effects[0]
	put := ctx.content.Effects[1]
	if draw.Kind != compiler.EffectDraw || !draw.Exact || draw.Optional || draw.Negated ||
		draw.Context != parser.EffectContextController ||
		!draw.Amount.Known || draw.Amount.Value < 1 ||
		put.Kind != compiler.EffectPut || !put.Exact || !put.HandLibraryPut.Present ||
		put.Optional || put.Negated ||
		put.Context != parser.EffectContextController ||
		put.Connection != parser.EffectConnectionThen ||
		!put.Amount.Known || put.Amount.Value < 1 {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Sequence: []game.Instruction{
			{Primitive: game.Draw{
				Player: game.ControllerReference(),
				Amount: game.Fixed(draw.Amount.Value),
			}},
			{Primitive: game.MoveCard{
				Player:      game.ControllerReference(),
				Amount:      game.Fixed(put.Amount.Value),
				FromZone:    zone.Hand,
				Destination: zone.Library,
			}},
		},
	}.Ability(), true
}

// lowerDrawHandDiscardSequence lowers the exact controller sequence "Draw N
// cards, then discard M cards." The typed discard marker excludes targeted,
// opponent, random, typed-card, and variable-cardinality discard forms.
func lowerDrawHandDiscardSequence(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 || ctx.optional ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Conditions) != 0 {
		return game.AbilityContent{}, false
	}
	draw := ctx.content.Effects[0]
	discard := ctx.content.Effects[1]
	if draw.Kind != compiler.EffectDraw || !draw.Exact || draw.Optional || draw.Negated ||
		draw.Context != parser.EffectContextController ||
		!draw.Amount.Known || draw.Amount.Value < 1 ||
		discard.Kind != compiler.EffectDiscard || !discard.Exact || !discard.HandDiscard.Present ||
		discard.Optional || discard.Negated ||
		discard.Context != parser.EffectContextController ||
		discard.Connection != parser.EffectConnectionThen ||
		!discard.Amount.Known || discard.Amount.Value < 1 {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Sequence: []game.Instruction{
			{Primitive: game.Draw{
				Player: game.ControllerReference(),
				Amount: game.Fixed(draw.Amount.Value),
			}},
			{Primitive: game.Discard{
				Player:   game.ControllerReference(),
				Amount:   game.Fixed(discard.Amount.Value),
				AtRandom: discard.HandDiscard.AtRandom,
			}},
		},
	}.Ability(), true
}

// applyEffectConditionGate attaches an effect-gate condition to every
// instruction a gated effect produced. It returns false (fail closed) if the
// effect produced no instructions, or if any instruction already carries a
// condition, so a gate can never be silently dropped or double-applied.
func applyEffectConditionGate(sequence []game.Instruction, condition *game.EffectCondition) bool {
	if len(sequence) == 0 {
		return false
	}
	for k := range sequence {
		if sequence[k].Condition.Exists {
			return false
		}
		sequence[k].Condition = opt.Val(*condition)
	}
	return true
}

// applySequenceClauseGates applies the per-effect condition gate, the "instead"
// negated gate, and the "otherwise" negated gate that apply to clause i. It
// returns an empty string on success, or a diagnostic category when a gate
// cannot be applied.
func applySequenceClauseGates(
	sequence []game.Instruction,
	i int,
	effectConditions, insteadGates, otherwiseGates map[int]game.EffectCondition,
) string {
	if gate, gated := effectConditions[i]; gated && !applyEffectConditionGate(sequence, &gate) {
		return "structural — per-effect condition gate not applicable"
	}
	if gate, gated := insteadGates[i]; gated && !applyEffectConditionGate(sequence, &gate) {
		return "structural — instead negated gate not applicable"
	}
	if gate, gated := otherwiseGates[i]; gated && !applyEffectConditionGate(sequence, &gate) {
		return "structural — otherwise negated gate not applicable"
	}
	return ""
}

// sequenceClauseInstructionGated reports whether the sequence clause at index i
// will carry a per-effect condition gate — a matched effect condition, an
// "instead" negated gate, or an "otherwise" negated gate. A gated clause's
// instruction is skipped at resolution whenever its condition does not hold, so
// a later clause must not chain a linked-object reference off it.
func sequenceClauseInstructionGated(
	i int,
	effectConditions, insteadGates, otherwiseGates map[int]game.EffectCondition,
) bool {
	if _, gated := effectConditions[i]; gated {
		return true
	}
	if _, gated := insteadGates[i]; gated {
		return true
	}
	if _, gated := otherwiseGates[i]; gated {
		return true
	}
	return false
}

// matchSequenceEffectConditions maps each compiled condition to the single
// effect whose clause span contains it and lowers it as an effect gate. It
// returns the lowered EffectCondition keyed by effect index. ok is false (fail
// closed) if any condition is not contained in exactly one effect, if two
// conditions land on the same effect, or if a condition is not a supported
// effect-gate condition. On failure it also returns a closed, enumerable
// blocker category (see the effectGateCategory* constants) so the support
// report can break the otherwise-opaque per-effect-condition reason into
// actionable sub-categories rather than one large bucket.
func matchSequenceEffectConditions(
	effects []compiler.CompiledEffect,
	conditions []compiler.CompiledCondition,
) (map[int]game.EffectCondition, string, bool) {
	if len(conditions) == 0 {
		return nil, "", true
	}
	result := make(map[int]game.EffectCondition, len(conditions))
	for ci := range conditions {
		condition := conditions[ci]
		var matched []int
		for ei := range effects {
			if spanCovered(condition.Span, []shared.Span{effects[ei].Span}) {
				matched = append(matched, ei)
			}
		}
		if len(matched) == 0 {
			return nil, effectGateCategoryNoClause, false
		}
		// A condition span-covered by more than one effect must be a leading
		// condition on a shared-sentence "then" group (the effects share an
		// identical sentence span), e.g. "If you control X, draw a card, then
		// discard a card." Such a condition gates every effect in the group.
		// Any other multi-match shape (a mid-sentence condition, or effects with
		// differing spans) fails closed.
		if len(matched) > 1 && !leadingGroupCondition(condition, effects, matched) {
			return nil, effectGateCategoryMultiClause, false
		}
		lowered, ok := lowerCondition(condition, conditionContextEffectGate)
		if !ok {
			return nil, effectGateRejectCategory(condition), false
		}
		for _, ei := range matched {
			if _, exists := result[ei]; exists {
				return nil, effectGateCategoryMultiCondition, false
			}
			// A per-effect gate that tests the source permanent's own
			// characteristics ("if it's tapped") while gating a target-scoped
			// clause means the condition's "it" antecedent was resolved to the
			// source rather than the clause's target — the source is not the
			// subject the printed condition names. Binding the gate to the wrong
			// object would silently test the wrong permanent, so fail closed.
			if effects[ei].Context == parser.EffectContextTarget &&
				conditionTestsSourceObjectCharacteristics(lowered) {
				return nil, effectGateCategoryPredicate, false
			}
			result[ei] = game.EffectCondition{
				Condition: opt.Val(lowered),
			}
		}
	}
	return result, "", true
}

// conditionTestsSourceObjectCharacteristics reports whether a lowered effect-gate
// condition tests the source permanent's own characteristics (a Selection match on
// a source object reference), the signature of an "it"/"that" antecedent that was
// resolved to the source rather than to the clause's own target subject.
func conditionTestsSourceObjectCharacteristics(condition game.Condition) bool {
	if !condition.Object.Exists || !condition.ObjectMatches.Exists {
		return false
	}
	switch condition.Object.Val.Kind() {
	case game.ObjectReferenceSourcePermanent,
		game.ObjectReferenceSourceAttachedPermanent,
		game.ObjectReferenceSourceCard:
		return true
	default:
		return false
	}
}

// Closed blocker categories for an ordered sequence whose per-effect condition
// could not be matched and lowered as an effect gate. Each is a stable,
// enumerable diagnostic detail consumed by the support report's
// ordered-sequence sub-category breakdown. effectGateCategoryUnrecognizedPrefix
// is followed by the recognized condition wording so the report can rank which
// unrecognized conditions block the most cards; the other categories are exact.
const (
	effectGateCategoryNoClause           = "structural — per-effect condition has no containing clause"
	effectGateCategoryMultiClause        = "structural — per-effect condition spans multiple clauses"
	effectGateCategoryMultiCondition     = "structural — multiple conditions gate one clause"
	effectGateCategoryKind               = "structural — per-effect condition kind not gateable"
	effectGateCategoryPredicate          = "structural — per-effect condition predicate not gateable"
	effectGateCategoryLowering           = "structural — per-effect condition lowering failed"
	effectGateCategoryUnrecognizedPrefix = "structural — per-effect condition unrecognized: "
)

// effectGateRejectCategory classifies why lowerCondition rejected a condition in
// the effect-gate context, returning one of the closed effectGateCategory
// constants. When the predicate was never recognized (the compiler emitted
// ConditionPredicateUnsupported), it appends the recognized condition wording so
// the support report can rank unrecognized conditions by how many cards they
// block. The wording is diagnostic metadata only; lowering never reads it back.
func effectGateRejectCategory(condition compiler.CompiledCondition) string {
	if !conditionKindAllowedInContext(condition, conditionContextEffectGate) {
		return effectGateCategoryKind
	}
	if !conditionPredicateAllowedInContext(condition.Predicate, conditionContextEffectGate) {
		if condition.Predicate == compiler.ConditionPredicateUnsupported {
			return effectGateCategoryUnrecognizedPrefix + strings.TrimSpace(condition.Text)
		}
		return effectGateCategoryPredicate
	}
	return effectGateCategoryLowering
}

// leadingGroupCondition reports whether the matched effects form a shared-
// sentence "then" group (every matched effect has the same sentence span) whose
// leading clause begins at or after the condition. A leading condition on such a
// group gates the entire group rather than a single clause.
func leadingGroupCondition(
	condition compiler.CompiledCondition,
	effects []compiler.CompiledEffect,
	matched []int,
) bool {
	groupSpan := effects[matched[0]].Span
	minClauseStart := effects[matched[0]].ClauseSpan.Start.Offset
	for _, ei := range matched {
		if effects[ei].Span != groupSpan {
			return false
		}
		if effects[ei].ClauseSpan.Start.Offset < minClauseStart {
			minClauseStart = effects[ei].ClauseSpan.Start.Offset
		}
	}
	return condition.Span.Start.Offset <= minClauseStart
}

// sequenceInsteadGates builds, for each effect carrying an "instead"
// replacement, a negated gate on the immediately preceding effect. The "instead"
// effect replaces the prior effect when its own condition holds, so gating the
// prior effect on the negation makes exactly one of the two run. It fails closed
// (ok=false) when an "instead" effect has no preceding effect, is not gated by a
// condition, its condition cannot be negated, or two replacements target the
// same preceding effect.
func sequenceInsteadGates(
	effects []compiler.CompiledEffect,
	effectConditions map[int]game.EffectCondition,
) (map[int]game.EffectCondition, bool) {
	var gates map[int]game.EffectCondition
	for j := range effects {
		if effects[j].Replacement.Kind != parser.EffectReplacementInstead {
			continue
		}
		if j == 0 {
			return nil, false
		}
		condition, gated := effectConditions[j]
		if !gated {
			return nil, false
		}
		negated, ok := negatedEffectCondition(&condition)
		if !ok {
			return nil, false
		}
		if _, exists := gates[j-1]; exists {
			return nil, false
		}
		if gates == nil {
			gates = make(map[int]game.EffectCondition)
		}
		gates[j-1] = negated
	}
	return gates, true
}

// sequenceOtherwiseGates builds, for each effect introduced by "Otherwise,", a
// negated gate derived from the immediately preceding effect's condition. The
// otherwise effect is the else branch of that conditional effect, so gating it
// on the negation makes exactly one of the two branches resolve. It fails closed
// (ok=false) when an otherwise effect has no preceding effect, the preceding
// effect carries no gate condition, that condition cannot be negated, or two
// otherwise effects target the same preceding effect.
func sequenceOtherwiseGates(
	effects []compiler.CompiledEffect,
	effectConditions map[int]game.EffectCondition,
	optionalFlow optionalFlowPlan,
) (map[int]game.EffectCondition, bool) {
	var gates map[int]game.EffectCondition
	for j := range effects {
		if effects[j].Connection != parser.EffectConnectionOtherwise {
			continue
		}
		// An "Otherwise" effect that is the else branch of a resolving-optional
		// flow ("you may X. If you do, Y. Otherwise, Z.") is gated on the optional
		// result by the optional-flow envelope, not on a preceding effect's
		// condition, so skip it here.
		if optionalFlow.gatesElse(j) {
			continue
		}
		if j == 0 {
			return nil, false
		}
		condition, gated := effectConditions[j-1]
		if !gated {
			return nil, false
		}
		negated, ok := negatedEffectCondition(&condition)
		if !ok {
			return nil, false
		}
		if _, exists := gates[j]; exists {
			return nil, false
		}
		if gates == nil {
			gates = make(map[int]game.EffectCondition)
		}
		gates[j] = negated
	}
	return gates, true
}

// negatedEffectCondition returns the logical negation of an effect-gate
// condition by inverting its wrapped shared Condition. It fails closed for
// permanent-type effect gates (which carry no wrapped Condition) because those
// are not part of the supported "instead" replacement forms.
func negatedEffectCondition(condition *game.EffectCondition) (game.EffectCondition, bool) {
	if !condition.Condition.Exists {
		return game.EffectCondition{}, false
	}
	inner := condition.Condition.Val
	inner.Negate = !inner.Negate
	return game.EffectCondition{
		Text:      condition.Text,
		Object:    condition.Object,
		Condition: opt.Val(inner),
	}, true
}

func localizeTargetReferences(
	references []compiler.CompiledReference,
	allTargets []compiler.CompiledTarget,
	localTargets []compiler.CompiledTarget,
) ([]compiler.CompiledReference, bool) {
	localized := append([]compiler.CompiledReference(nil), references...)
	for i := range localized {
		if localized[i].Binding != compiler.ReferenceBindingTarget {
			continue
		}
		if localized[i].Occurrence < 0 || localized[i].Occurrence >= len(allTargets) {
			return nil, false
		}
		targetSpan := allTargets[localized[i].Occurrence].Span
		local := slices.IndexFunc(localTargets, func(target compiler.CompiledTarget) bool {
			return target.Span == targetSpan
		})
		if local < 0 {
			return nil, false
		}
		localized[i].Occurrence = local
	}
	return localized, true
}

func appendReferenceAntecedentTargets(
	inherited []compiler.CompiledTarget,
	references []compiler.CompiledReference,
	allTargets []compiler.CompiledTarget,
	clauseTargets []compiler.CompiledTarget,
) []compiler.CompiledTarget {
	for _, reference := range references {
		if reference.Binding != compiler.ReferenceBindingTarget ||
			reference.Occurrence < 0 ||
			reference.Occurrence >= len(allTargets) {
			continue
		}
		target := allTargets[reference.Occurrence]
		if !oracleTargetSpanIn(target.Span, clauseTargets) &&
			!oracleTargetSpanIn(target.Span, inherited) {
			inherited = append(inherited, target)
		}
	}
	return inherited
}

func lowerDelayedTargetReturn(
	effectIndex int,
	ctx contentCtx,
	sequence []game.Instruction,
) (game.ModifyPT, game.AbilityContent, bool) {
	// Invariant: ctx is the contextForEffect-narrowed per-clause effectAbility
	// built at lowerOrderedEffectSequence and threaded through
	// lowerDelayedSequenceClause, so content.Effects always holds exactly one
	// clause effect; any other length is an upstream bug.
	if len(ctx.content.Effects) != 1 {
		panic(fmt.Sprintf("lowerDelayedTargetReturn: expected a single effect, got %d", len(ctx.content.Effects)))
	}
	if effectIndex == 0 ||
		len(sequence) != effectIndex ||
		ctx.content.Effects[0].Kind != compiler.EffectReturn ||
		ctx.content.Effects[0].DelayedTiming != game.DelayedAtBeginningOfNextEndStep ||
		ctx.content.Effects[0].Negated ||
		ctx.optional ||
		ctx.content.Effects[0].Context != parser.EffectContextController ||
		ctx.content.Effects[0].ToZone != zone.Hand ||
		ctx.content.Effects[0].CounterKindKnown ||
		!referencesBindTo(ctx.content.References, compiler.ReferenceBindingTarget, 0) {
		return game.ModifyPT{}, game.AbilityContent{}, false
	}
	previous := sequence[effectIndex-1].Primitive
	if previous.Kind() != game.PrimitiveModifyPT {
		return game.ModifyPT{}, game.AbilityContent{}, false
	}
	modify, ok := previous.(game.ModifyPT)
	if !ok ||
		modify.Object.Kind() != game.ObjectReferenceTargetPermanent ||
		modify.PublishLinked != "" {
		return game.ModifyPT{}, game.AbilityContent{}, false
	}
	consumed := ctx
	consumed.content.References = nil
	consumed.content.Targets = nil
	if consumed.content.Unconsumed() {
		return game.ModifyPT{}, game.AbilityContent{}, false
	}
	key := game.LinkedKey(fmt.Sprintf("delayed-target-%d", effectIndex))
	object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{
		TargetLinkedKey: key,
	})
	if !ok {
		return game.ModifyPT{}, game.AbilityContent{}, false
	}
	modify.PublishLinked = key
	delayed := game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
		Timing: game.DelayedAtBeginningOfNextEndStep,
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.Bounce{
			Object: object,
		}}}}.Ability(),
	}}
	return modify, game.Mode{Sequence: []game.Instruction{{Primitive: delayed}}}.Ability(), true
}

// isDelayedTargetSacrificeEffect reports whether effect is a delayed
// "sacrifice it/that creature at the beginning of the next end step" clause whose
// subject is the permanent targeted by an earlier effect in the same sequence.
func isDelayedTargetSacrificeEffect(effect *compiler.CompiledEffect) bool {
	return effect.Kind == compiler.EffectSacrifice &&
		effect.DelayedTiming == game.DelayedAtBeginningOfNextEndStep &&
		!effect.Negated &&
		effect.Context == parser.EffectContextController &&
		!effect.CounterKindKnown &&
		referencesBindTo(effect.References, compiler.ReferenceBindingTarget, 0)
}

// lowerDelayedSequenceClause attempts the linked clause shapes (delayed
// sacrifice, delayed return-to-hand, delayed blink-return, and immediate
// blink-return) that capture an earlier target and resolve it at a later step or
// in the same resolution. When the clause matches one of these shapes it rewrites
// the publishing instruction in sequence and returns the linked-effect content
// with handled set. failed reports a matched-but-unlinkable sacrifice clause so
// the caller can fail closed. handled is false when no linked shape applies and
// the caller should lower the clause normally.
func lowerDelayedSequenceClause(
	effects []compiler.CompiledEffect,
	effectIndex int,
	ctx contentCtx,
	sequence []game.Instruction,
	publisherGated bool,
) (content game.AbilityContent, handled, failed bool) {
	if isDelayedTargetSacrificeEffect(&effects[effectIndex]) {
		publisher, delayed, ok := lowerDelayedTargetSacrifice(effectIndex, ctx, sequence)
		if !ok {
			return game.AbilityContent{}, true, true
		}
		sequence[len(sequence)-1].Primitive = publisher
		return delayed, true, false
	}
	if modify, delayed, ok := lowerDelayedTargetReturn(effectIndex, ctx, sequence); ok {
		sequence[len(sequence)-1].Primitive = modify
		return delayed, true, false
	}
	if modify, delayed, ok := lowerDelayedCombatDamageDrawTrigger(effectIndex, ctx, sequence); ok {
		sequence[len(sequence)-1].Primitive = modify
		return delayed, true, false
	}
	if add, delayed, ok := lowerDelayedAttacksMonarchGrant(effectIndex, ctx, sequence); ok {
		sequence[len(sequence)-1].Primitive = add
		return delayed, true, false
	}
	if delayed, ok := lowerDelayedCommanderMonarchDiesTrigger(effectIndex, ctx, sequence); ok {
		return delayed, true, false
	}
	if publisher, delayed, ok := lowerDelayedTargetExile(effectIndex, ctx, sequence); ok {
		sequence[len(sequence)-1].Primitive = publisher
		return delayed, true, false
	}
	if publisher, grant, ok := lowerSequentialReferencedKeywordGrant(effectIndex, ctx, sequence, publisherGated); ok {
		sequence[len(sequence)-1].Primitive = publisher
		return grant, true, false
	}
	if publisher, replacement, ok := lowerSequentialLeaveBattlefieldExileReplacement(effectIndex, ctx, sequence); ok {
		sequence[len(sequence)-1].Primitive = publisher
		return replacement, true, false
	}
	if publisher, placement, ok := lowerSequentialReanimationCounterPlacement(effectIndex, ctx, sequence); ok {
		sequence[len(sequence)-1].Primitive = publisher
		return placement, true, false
	}
	if publisher, grant, ok := lowerSequentialReanimationTypeColorGrant(effectIndex, ctx, sequence); ok {
		sequence[len(sequence)-1].Primitive = publisher
		return grant, true, false
	}
	if exile, delayed, ok := lowerDelayedBlinkReturn(effects, effectIndex, ctx, sequence); ok {
		sequence[len(sequence)-1].Primitive = exile
		return delayed, true, false
	}
	if exile, returnContent, ok := lowerImmediateBlinkReturn(effects, effectIndex, ctx, sequence); ok {
		sequence[len(sequence)-1].Primitive = exile
		return returnContent, true, false
	}
	if lowered, ok := lowerCharacteristicLifeRider(effects, effectIndex, ctx, sequence); ok {
		if lowered.priorPrimitive != nil {
			sequence[len(sequence)-1].Primitive = lowered.priorPrimitive
		}
		if lowered.priorResult != "" {
			sequence[len(sequence)-1].PublishResult = lowered.priorResult
		}
		return lowered.content, true, false
	}
	if content, ok := lowerThatMuchLifeBackref(ctx, effectIndex, sequence); ok {
		return content, true, false
	}
	return game.AbilityContent{}, false, false
}

type characteristicLifeRiderLowering struct {
	content        game.AbilityContent
	priorPrimitive game.Primitive
	priorResult    game.ResultKey
}

// lowerCharacteristicLifeRider lowers a life-gain or life-loss clause whose
// amount is a permanent's own power, toughness, or mana value ("… gains life
// equal to its power", "… loses life equal to its toughness", "… lose life equal
// to that permanent's mana value") where that permanent is the subject acted on
// by an earlier clause in the same ordered sequence. It backs the most-played
// versions of this shape — Swords to Plowshares ("Exile target creature. Its
// controller gains life equal to its power."), Chastise ("Destroy target
// attacking creature. You gain life equal to its power."), Feed the Swarm
// ("Destroy target creature or enchantment an opponent controls. You lose life
// equal to that permanent's mana value."), and Divine Offering ("Destroy target
// artifact. You gain life equal to its mana value.").
//
// The clause carries an "its power"/"its toughness"/"its mana value" referent the
// executable backend resolves through the object that the prior clause targeted
// or exiled, using last-known information when that permanent has left the
// battlefield. Two recipient forms are modeled: the spell's controller ("You gain
// …") and the acted-on permanent's controller ("Its controller gains …"). The
// amount referent binds either directly to the inherited target ("its power" when
// "you" already took no binding) or to the prior instruction's result, in which
// case the preceding exile is rewritten to publish the exiled object under a
// linked key so the amount reads its last-known power or toughness.
//
// The mana-value form is restricted further: its referent must be either the
// target permanent the immediately preceding clause destroyed or the fresh
// permanent created by an exact single-creature graveyard return under the
// controller's control. The latter publishes both the moved permanent and the
// move result so a replacement that diverts the card suppresses the rider.
//
// It returns the lowered content plus a rewritten prior primitive and result key
// when linking requires them. It returns handled=false (so the caller lowers the
// clause normally and ultimately fails closed) for every clause outside this
// exact shape.
func lowerCharacteristicLifeRider(
	effects []compiler.CompiledEffect,
	effectIndex int,
	ctx contentCtx,
	sequence []game.Instruction,
) (characteristicLifeRiderLowering, bool) {
	// Invariant: ctx is the contextForEffect-narrowed per-clause effectAbility
	// built at lowerOrderedEffectSequence and threaded through
	// lowerDelayedSequenceClause, so content.Effects always holds exactly one
	// clause effect; any other length is an upstream bug.
	if len(ctx.content.Effects) != 1 {
		panic(fmt.Sprintf("lowerCharacteristicLifeRider: expected a single effect, got %d", len(ctx.content.Effects)))
	}
	if effectIndex == 0 ||
		len(sequence) != effectIndex {
		return characteristicLifeRiderLowering{}, false
	}
	effect := &ctx.content.Effects[0]
	if (effect.Kind != compiler.EffectGain && effect.Kind != compiler.EffectLose) ||
		!effect.LifeObject ||
		!effect.Exact ||
		effect.Negated ||
		ctx.optional ||
		effect.Amount.Known ||
		effect.Amount.DynamicForm != compiler.DynamicAmountEqual ||
		effect.Amount.Multiplier != 1 ||
		(effect.Amount.DynamicKind != compiler.DynamicAmountSourcePower &&
			effect.Amount.DynamicKind != compiler.DynamicAmountSourceToughness &&
			effect.Amount.DynamicKind != compiler.DynamicAmountSourceManaValue) ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 {
		return characteristicLifeRiderLowering{}, false
	}
	amountRef, subjectRefs, ok := sourcePowerReferences(effect)
	if !ok {
		return characteristicLifeRiderLowering{}, false
	}
	player, ok := lifeRiderRecipient(ctx, effect, subjectRefs)
	if !ok {
		return characteristicLifeRiderLowering{}, false
	}
	amountLowering, ok := lifeRiderAmountObject(
		effects,
		amountRef,
		effectIndex,
		sequence,
		effect.Amount.DynamicKind == compiler.DynamicAmountSourceManaValue,
	)
	if !ok {
		return characteristicLifeRiderLowering{}, false
	}
	if effect.Amount.DynamicKind == compiler.DynamicAmountSourceManaValue {
		reanimation := reanimationManaValueAntecedent(effects, effectIndex)
		if (reanimation && amountLowering.priorResult == "") ||
			(!reanimation && !priorClauseDestroys(sequence, effectIndex, amountLowering.object)) {
			return characteristicLifeRiderLowering{}, false
		}
	}
	dynamic, ok := objectCharacteristicAmount(effect.Amount.DynamicKind, amountLowering.object)
	if !ok {
		return characteristicLifeRiderLowering{}, false
	}
	var primitive game.Primitive
	switch effect.Kind {
	case compiler.EffectGain:
		primitive = game.GainLife{Amount: game.Dynamic(dynamic), Player: player}
	case compiler.EffectLose:
		primitive = game.LoseLife{Amount: game.Dynamic(dynamic), Player: player}
	default:
		return characteristicLifeRiderLowering{}, false
	}
	instruction := game.Instruction{Primitive: primitive}
	if amountLowering.priorResult != "" {
		instruction.ResultGate = opt.Val(game.InstructionResultGate{
			Key:       amountLowering.priorResult,
			Succeeded: game.TriTrue,
		})
	}
	return characteristicLifeRiderLowering{
		content:        game.Mode{Sequence: []game.Instruction{instruction}}.Ability(),
		priorPrimitive: amountLowering.priorPrimitive,
		priorResult:    amountLowering.priorResult,
	}, true
}

// priorClauseDestroys reports whether the instruction immediately preceding the
// life rider is a single-target Destroy of exactly the permanent whose mana value
// the rider reads. A mana-value rider must read its subject's last-known mana
// value, which is only recorded when an earlier clause moved that permanent off
// the battlefield. Requiring the prior clause to destroy the same target
// permanent keeps the shape to the "Destroy target permanent. <recipient>
// gains/loses life equal to that permanent's mana value." staples and fails
// closed for graveyard-return riders, whose referent is a card-zone target with
// no battlefield mana value to read.
func priorClauseDestroys(sequence []game.Instruction, effectIndex int, object game.ObjectReference) bool {
	destroy, ok := sequence[effectIndex-1].Primitive.(game.Destroy)
	if !ok || destroy.Group.Valid() {
		return false
	}
	return destroy.Object == object
}

// lifeRiderRecipient resolves the player who gains or loses life for a
// characteristic life rider. "You gain/lose life" (controller context, no
// subject reference) yields the spell's controller; "Its controller gains/loses
// life" (referenced-object-controller context) yields the controller of the
// inherited antecedent permanent. Any other context or leftover subject
// reference fails closed.
func lifeRiderRecipient(
	ctx contentCtx,
	effect *compiler.CompiledEffect,
	subjectRefs []compiler.CompiledReference,
) (game.PlayerReference, bool) {
	switch effect.Context {
	case parser.EffectContextController:
		if len(subjectRefs) != 0 {
			return game.PlayerReference{}, false
		}
		return game.ControllerReference(), true
	case parser.EffectContextReferencedObjectController:
		recipientCtx := ctx
		recipientCtx.content.References = subjectRefs
		return referencedControllerPlayerRef(recipientCtx)
	case parser.EffectContextReferencedObjectOwner:
		recipientCtx := ctx
		recipientCtx.content.References = subjectRefs
		return referencedOwnerPlayerRef(recipientCtx)
	default:
		return game.PlayerReference{}, false
	}
}

// lifeRiderAmountObject resolves the permanent whose characteristic the rider
// reads. A target-bound referent ("its power" where "its" is the inherited
// target) resolves to that target permanent. A prior-instruction-result referent
// ("Its controller gains life equal to its power", where the recipient already
// consumed the target binding) resolves to the object exiled by the immediately
// preceding clause. Exile is rewritten to publish last-known information.
// An exact graveyard-to-battlefield move publishes the fresh permanent and its
// success result, so the rider reads the entered object only when the move reached
// the battlefield. Every other binding fails closed.
type lifeRiderAmountLowering struct {
	object         game.ObjectReference
	priorPrimitive game.Primitive
	priorResult    game.ResultKey
}

func lifeRiderAmountObject(
	effects []compiler.CompiledEffect,
	amountRef compiler.CompiledReference,
	effectIndex int,
	sequence []game.Instruction,
	manaValue bool,
) (lifeRiderAmountLowering, bool) {
	switch amountRef.Binding {
	case compiler.ReferenceBindingTarget:
		// "lose life equal to that card's mana value" after an exact return to the
		// battlefield under its owner's control ("Return target creature card from
		// your graveyard to the battlefield") binds the amount to the returned
		// card's target. The card becomes a fresh permanent on entry, so the rider
		// must read the entered permanent's last-known mana value rather than the
		// stale graveyard target. Route through the reanimation publish path, which
		// rewrites the move to record the entered permanent and gates the life
		// change on the move reaching the battlefield. An "under your control"
		// return binds "that card" to the prior instruction result instead and is
		// handled below; its target-bound "its mana value" variant stays fail
		// closed because that pronoun does not anchor the card-to-permanent move.
		if manaValue &&
			reanimationManaValueAntecedent(effects, effectIndex) &&
			!effects[effectIndex-1].UnderYourControl {
			put, ok := sequence[effectIndex-1].Primitive.(game.PutOnBattlefield)
			if !ok {
				return lifeRiderAmountLowering{}, false
			}
			return reanimationManaValuePublish(put, amountRef, effectIndex, true)
		}
		obj, ok := lowerObjectReference(amountRef, referenceLoweringContext{AllowTarget: true})
		if !ok {
			return lifeRiderAmountLowering{}, false
		}
		// A power/toughness rider whose referent binds the target of a graveyard
		// reanimation reads the stale graveyard card rather than the fresh
		// permanent; only the mana-value publish path tracks the card-to-permanent
		// move, so non-mana-value characteristics fail closed here.
		if !manaValue && reanimationManaValueAntecedent(effects, effectIndex) {
			return lifeRiderAmountLowering{}, false
		}
		return lifeRiderAmountLowering{object: obj}, true
	case compiler.ReferenceBindingPriorInstructionResult:
		if amountRef.PriorInstruction != effectIndex-1 {
			return lifeRiderAmountLowering{}, false
		}
		exile, ok := sequence[effectIndex-1].Primitive.(game.Exile)
		if ok {
			if exile.Group.Valid() ||
				exile.Object.Kind() != game.ObjectReferenceTargetPermanent ||
				exile.ExileLinkedKey != "" {
				return lifeRiderAmountLowering{}, false
			}
			key := game.LinkedKey(fmt.Sprintf("life-rider-%d", effectIndex))
			obj, ok := lowerObjectReference(amountRef, referenceLoweringContext{
				PriorInstruction: effectIndex - 1,
				PriorLinkedKey:   key,
			})
			if !ok {
				return lifeRiderAmountLowering{}, false
			}
			exile.ExileLinkedKey = key
			return lifeRiderAmountLowering{object: obj, priorPrimitive: exile}, true
		}
		put, ok := sequence[effectIndex-1].Primitive.(game.PutOnBattlefield)
		if !ok || !manaValue || !reanimationManaValueAntecedent(effects, effectIndex) {
			return lifeRiderAmountLowering{}, false
		}
		return reanimationManaValuePublish(put, amountRef, effectIndex, false)
	default:
		return lifeRiderAmountLowering{}, false
	}
}

// reanimationManaValuePublish rewrites the immediately preceding reanimation
// PutOnBattlefield so it records the entered permanent under a linked key and
// returns the linked object whose mana value the life rider reads, gated on the
// move reaching the battlefield. targetBound selects whether the amount
// reference binds the move's target card ("that card's mana value") or the prior
// instruction's published result.
func reanimationManaValuePublish(
	put game.PutOnBattlefield,
	amountRef compiler.CompiledReference,
	effectIndex int,
	targetBound bool,
) (lifeRiderAmountLowering, bool) {
	if put.PublishLinked != "" {
		return lifeRiderAmountLowering{}, false
	}
	card, ok := put.Source.CardRef()
	if !ok || card.Kind != game.CardReferenceTarget {
		return lifeRiderAmountLowering{}, false
	}
	key := game.LinkedKey(fmt.Sprintf("life-rider-%d", effectIndex))
	refCtx := referenceLoweringContext{}
	if targetBound {
		refCtx.TargetLinkedKey = key
	} else {
		refCtx.PriorInstruction = effectIndex - 1
		refCtx.PriorLinkedKey = key
	}
	obj, ok := lowerObjectReference(amountRef, refCtx)
	if !ok {
		return lifeRiderAmountLowering{}, false
	}
	put.PublishLinked = key
	resultKey := game.ResultKey(fmt.Sprintf("life-rider-move-%d", effectIndex))
	return lifeRiderAmountLowering{
		object:         obj,
		priorPrimitive: put,
		priorResult:    resultKey,
	}, true
}

func reanimationManaValueAntecedent(effects []compiler.CompiledEffect, effectIndex int) bool {
	if effectIndex == 0 || effectIndex >= len(effects) {
		return false
	}
	effect := effects[effectIndex-1]
	if (effect.Kind != compiler.EffectPut && effect.Kind != compiler.EffectReturn) ||
		!effect.Exact ||
		effect.Negated ||
		effect.FromZone != zone.Graveyard ||
		effect.ToZone != zone.Battlefield ||
		effect.EntersTapped ||
		effect.CounterKindKnown ||
		effect.Amount.Known ||
		len(effect.Targets) != 1 ||
		len(effect.References) != 0 {
		return false
	}
	target := effect.Targets[0]
	spec, ok := cardInZoneTargetSpec(target, zone.Graveyard)
	if !ok ||
		spec.MinTargets != 1 ||
		spec.MaxTargets != 1 ||
		!spec.Selection.Exists {
		return false
	}
	selection := spec.Selection.Val
	if !slices.Equal(selection.RequiredTypes, []types.Card{types.Creature}) {
		return false
	}
	// "from your graveyard" restricts the reanimation target to the controller's
	// own graveyard, leaving Controller == ControllerYou; "from a graveyard"
	// leaves it ControllerAny. Either ownership is a valid reanimation antecedent
	// for the mana-value rider, which reads the entered permanent regardless of
	// whose graveyard it came from. Clear the ownership constraint before the
	// no-other-constraints emptiness check and fail closed for any other owner.
	if selection.Controller != game.ControllerAny && selection.Controller != game.ControllerYou {
		return false
	}
	selection.RequiredTypes = nil
	selection.Controller = game.ControllerAny
	return selection.Empty()
}

// lowerDelayedTargetSacrifice lowers a delayed "sacrifice it at the beginning of
// the next end step" clause that refers to the permanent targeted by the
// immediately preceding effect (e.g. "Target creature you control gains flying
// until end of turn. Sacrifice it at the beginning of the next end step."). The
// preceding instruction publishes the resolved target under a linked key and the
// delayed trigger sacrifices that linked object, so the captured permanent is
// sacrificed rather than the source. It returns the rewritten publishing
// primitive and the delayed-trigger content, or false to fail closed.
func lowerDelayedTargetSacrifice(
	effectIndex int,
	ctx contentCtx,
	sequence []game.Instruction,
) (game.Primitive, game.AbilityContent, bool) {
	// Invariant: ctx is the contextForEffect-narrowed per-clause effectAbility
	// built at lowerOrderedEffectSequence and threaded through
	// lowerDelayedSequenceClause, so content.Effects always holds exactly one
	// clause effect; any other length is an upstream bug.
	if len(ctx.content.Effects) != 1 {
		panic(fmt.Sprintf("lowerDelayedTargetSacrifice: expected a single effect, got %d", len(ctx.content.Effects)))
	}
	if effectIndex == 0 ||
		len(sequence) != effectIndex ||
		!isDelayedTargetSacrificeEffect(&ctx.content.Effects[0]) ||
		ctx.optional ||
		!referencesBindTo(ctx.content.References, compiler.ReferenceBindingTarget, 0) {
		return nil, game.AbilityContent{}, false
	}
	key := game.LinkedKey(fmt.Sprintf("delayed-sacrifice-%d", effectIndex))
	publisher, ok := publishLinkedTargetPermanent(sequence[effectIndex-1].Primitive, key)
	if !ok {
		return nil, game.AbilityContent{}, false
	}
	consumed := ctx
	consumed.content.References = nil
	consumed.content.Targets = nil
	if consumed.content.Unconsumed() {
		return nil, game.AbilityContent{}, false
	}
	object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{
		TargetLinkedKey: key,
	})
	if !ok {
		return nil, game.AbilityContent{}, false
	}
	delayed := game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
		Timing: game.DelayedAtBeginningOfNextEndStep,
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.Sacrifice{
			Object: object,
		}}}}.Ability(),
	}}
	return publisher, game.Mode{Sequence: []game.Instruction{{Primitive: delayed}}}.Ability(), true
}

// isDelayedTargetExileEffect reports whether effect is a delayed "exile it/that
// creature at the beginning of the next end step" clause whose subject is the
// permanent targeted by an earlier effect in the same sequence (the temporary-
// reanimation cleanup "Return target creature card ... Exile it at the beginning
// of the next end step." — Whip of Erebos and kin).
func isDelayedTargetExileEffect(effect *compiler.CompiledEffect) bool {
	return effect.Kind == compiler.EffectExile &&
		effect.DelayedTiming == game.DelayedAtBeginningOfNextEndStep &&
		!effect.Negated &&
		effect.Context == parser.EffectContextController &&
		!effect.CounterKindKnown &&
		referencesBindTo(effect.References, compiler.ReferenceBindingTarget, 0)
}

// lowerDelayedTargetExile lowers a delayed "Exile it at the beginning of the next
// end step." clause that exiles the permanent an earlier clause put onto the
// battlefield or pumped. It captures that permanent under a linked key (rewriting
// the earlier publishing instruction) and schedules a delayed end-step trigger
// that exiles the linked object, mirroring lowerDelayedTargetSacrifice. It
// returns ok=false (so the caller lowers the clause normally) for any shape it
// cannot link, preserving existing behavior for unlinkable predecessors.
func lowerDelayedTargetExile(
	effectIndex int,
	ctx contentCtx,
	sequence []game.Instruction,
) (game.Primitive, game.AbilityContent, bool) {
	// Invariant: ctx is the contextForEffect-narrowed per-clause effectAbility
	// built at lowerOrderedEffectSequence and threaded through
	// lowerDelayedSequenceClause, so content.Effects always holds exactly one
	// clause effect; any other length is an upstream bug.
	if len(ctx.content.Effects) != 1 {
		panic(fmt.Sprintf("lowerDelayedTargetExile: expected a single effect, got %d", len(ctx.content.Effects)))
	}
	if effectIndex == 0 ||
		len(sequence) != effectIndex ||
		!isDelayedTargetExileEffect(&ctx.content.Effects[0]) ||
		ctx.optional ||
		!referencesBindTo(ctx.content.References, compiler.ReferenceBindingTarget, 0) {
		return nil, game.AbilityContent{}, false
	}
	key := game.LinkedKey(fmt.Sprintf("delayed-exile-%d", effectIndex))
	publisher, ok := publishLinkedTargetPermanent(sequence[effectIndex-1].Primitive, key)
	if !ok {
		return nil, game.AbilityContent{}, false
	}
	consumed := ctx
	consumed.content.References = nil
	consumed.content.Targets = nil
	if consumed.content.Unconsumed() {
		return nil, game.AbilityContent{}, false
	}
	object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{
		TargetLinkedKey: key,
	})
	if !ok {
		return nil, game.AbilityContent{}, false
	}
	delayed := game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
		Timing: game.DelayedAtBeginningOfNextEndStep,
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.Exile{
			Object: object,
		}}}}.Ability(),
	}}
	return publisher, game.Mode{Sequence: []game.Instruction{{Primitive: delayed}}}.Ability(), true
}

// isSequentialReferencedKeywordGrantEffect reports whether effect is an exact
// no-duration "it gains <keyword>" clause whose subject is the permanent an
// earlier clause in the same sequence acted on (the reanimation companion grant
// "Return target creature card ... to the battlefield. It gains haste." — Whip of
// Erebos, Apprentice Necromancer, Puppeteer Clique). The keyword lasts as long as
// the permanent remains on the battlefield, so the duration is absent rather than
// until end of turn.
func isSequentialReferencedKeywordGrantEffect(effect *compiler.CompiledEffect) bool {
	_, durationOK := sequentialReferencedKeywordGrantDuration(effect.Duration)
	return effect.Kind == compiler.EffectGain &&
		effect.Exact &&
		!effect.Negated &&
		!effect.KeywordGrantChoice &&
		effect.Context == parser.EffectContextReferencedObject &&
		durationOK &&
		effect.StaticSubject == compiler.StaticSubjectNone &&
		referencesBindTo(effect.References, compiler.ReferenceBindingTarget, 0)
}

// sequentialReferencedKeywordGrantDuration maps the compiled duration of a
// keyword grant bound to a permanent an earlier sequence clause acted on to its
// runtime EffectDuration. A no-duration grant lasts as long as the permanent
// remains on the battlefield (DurationPermanent), which composes with the
// temporary reanimation pattern that exiles the permanent at end of turn ("It
// gains haste. Exile it at the beginning of the next end step.", Whip of
// Erebos). The bounded "until end of turn" and "until your next turn" forms
// ("Return target creature card from your graveyard to the battlefield. It gains
// haste until your next turn.", Bond of Revival) expire on their own. Any other
// duration returns ok=false so richer grants stay fail-closed.
func sequentialReferencedKeywordGrantDuration(duration compiler.DurationKind) (game.EffectDuration, bool) {
	switch duration {
	case compiler.DurationNone:
		return game.DurationPermanent, true
	case compiler.DurationUntilEndOfTurn:
		return game.DurationUntilEndOfTurn, true
	case compiler.DurationUntilYourNextTurn:
		return game.DurationUntilYourNextTurn, true
	default:
		return game.DurationPermanent, false
	}
}

// lowerSequentialReferencedKeywordGrant lowers an "it gains <keyword>[ until end
// of turn| until your next turn]." clause that grants a keyword to the permanent
// an earlier clause in the same sequence put onto the battlefield or otherwise
// acted on. "It" binds to that earlier permanent, which (for a reanimation) is a
// freshly created object that the targeted graveyard card became, so a plain
// target-permanent reference cannot resolve it. The earlier publishing
// instruction is rewritten to record the permanent under a linked key, and the
// keyword grant reads that linked object, mirroring lowerDelayedTargetExile's
// capture. The grant lasts permanently (no duration), until end of turn, or
// until the controller's next turn per sequentialReferencedKeywordGrantDuration.
// It returns the rewritten publishing primitive and the grant content, or false
// to decline the linked-capture lowering so the caller lowers the clause
// normally. It declines when the immediately preceding publishing instruction is
// itself gated by a per-effect condition (publisherGated): a gated publisher is
// skipped whenever its condition does not hold, so its linked key is never
// recorded and this grant's linked-object reference would resolve to nothing.
// That is exactly the mutually-exclusive "If it's your turn, ... gains trample.
// Otherwise, it gains first strike." shape (Stolen Vitality, Rapier Wit), where
// the else branch would chain off the "your turn" branch it can never co-occur
// with. Declining routes the clause through normal reference lowering, which for
// a plain targeted subject binds the target permanent directly (correct — the
// same permanent the linked key would have captured) and otherwise fails the
// clause closed. Only a created/reanimated subject ("it" naming a freshly made
// object that a plain target reference cannot denote) would bind the wrong
// permanent through the fallback; no such created-object subject appears in a
// two-branch publisher-gated shape in the corpus, so every generated card that
// reaches this decline binds correctly.
func lowerSequentialReferencedKeywordGrant(
	effectIndex int,
	ctx contentCtx,
	sequence []game.Instruction,
	publisherGated bool,
) (game.Primitive, game.AbilityContent, bool) {
	// Invariant: ctx is the contextForEffect-narrowed per-clause effectAbility
	// built at lowerOrderedEffectSequence and threaded through
	// lowerDelayedSequenceClause, so content.Effects always holds exactly one
	// clause effect; any other length is an upstream bug.
	if len(ctx.content.Effects) != 1 {
		panic(fmt.Sprintf("lowerSequentialReferencedKeywordGrant: expected a single effect, got %d", len(ctx.content.Effects)))
	}
	if effectIndex == 0 ||
		publisherGated ||
		len(sequence) != effectIndex ||
		ctx.optional ||
		len(ctx.content.Keywords) == 0 ||
		!isSequentialReferencedKeywordGrantEffect(&ctx.content.Effects[0]) {
		return nil, game.AbilityContent{}, false
	}
	keywords, abilities, ok := partitionTemporaryKeywords(ctx.content.Keywords)
	if !ok {
		return nil, game.AbilityContent{}, false
	}
	duration, ok := sequentialReferencedKeywordGrantDuration(ctx.content.Effects[0].Duration)
	if !ok {
		return nil, game.AbilityContent{}, false
	}
	key := game.LinkedKey(fmt.Sprintf("gain-keyword-%d", effectIndex))
	publisher, ok := publishLinkedTargetPermanent(sequence[effectIndex-1].Primitive, key)
	if !ok {
		return nil, game.AbilityContent{}, false
	}
	consumed := ctx
	consumed.content.References = nil
	consumed.content.Keywords = nil
	consumed.content.Targets = nil
	if consumed.content.Unconsumed() {
		return nil, game.AbilityContent{}, false
	}
	object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{
		TargetLinkedKey: key,
	})
	if !ok {
		return nil, game.AbilityContent{}, false
	}
	grant := game.ApplyContinuous{
		Object: opt.Val(object),
		ContinuousEffects: []game.ContinuousEffect{{
			Layer:        game.LayerAbility,
			AddKeywords:  keywords,
			AddAbilities: abilities,
		}},
		Duration: duration,
	}
	return publisher, game.Mode{Sequence: []game.Instruction{{Primitive: grant}}}.Ability(), true
}

// isLeaveBattlefieldExileReplacementEffect reports whether effect is the exact
// leaves-the-battlefield exile replacement "If it would leave the battlefield,
// exile it instead of putting it anywhere else." applied to the permanent an
// earlier clause in the same sequence acted on (Whip of Erebos). "It" binds to
// that earlier permanent.
func isLeaveBattlefieldExileReplacementEffect(effect *compiler.CompiledEffect) bool {
	return effect.Kind == compiler.EffectExileIfLeaveBattlefield &&
		effect.Exact &&
		!effect.Negated &&
		effect.Context == parser.EffectContextReferencedObject &&
		referencesBindTo(effect.References, compiler.ReferenceBindingTarget, 0)
}

// lowerSequentialLeaveBattlefieldExileReplacement lowers an "If it would leave
// the battlefield, exile it instead of putting it anywhere else." clause that
// redirects any zone change off the battlefield to exile for the permanent an
// earlier clause in the same sequence put onto the battlefield or otherwise
// acted on (Whip of Erebos's reanimated creature). "It" binds to that earlier
// permanent, which (for a reanimation) is a freshly created object a plain
// target-permanent reference cannot resolve, so the lowering reuses the linked
// key under which an earlier clause already recorded the permanent, or rewrites
// the immediately-prior instruction to publish it. The created replacement is a
// CreateReplacement bound to that linked object. It returns the (possibly
// rewritten) prior publishing primitive and the replacement content, or false to
// fail closed so the caller lowers the clause normally.
func lowerSequentialLeaveBattlefieldExileReplacement(
	effectIndex int,
	ctx contentCtx,
	sequence []game.Instruction,
) (game.Primitive, game.AbilityContent, bool) {
	// Invariant: ctx is the contextForEffect-narrowed per-clause effectAbility
	// built at lowerOrderedEffectSequence and threaded through
	// lowerDelayedSequenceClause, so content.Effects always holds exactly one
	// clause effect; any other length is an upstream bug.
	if len(ctx.content.Effects) != 1 {
		panic(fmt.Sprintf("lowerSequentialLeaveBattlefieldExileReplacement: expected a single effect, got %d", len(ctx.content.Effects)))
	}
	if effectIndex == 0 ||
		len(sequence) != effectIndex ||
		ctx.optional ||
		!isLeaveBattlefieldExileReplacementEffect(&ctx.content.Effects[0]) {
		return nil, game.AbilityContent{}, false
	}
	key, publisher, ok := reuseOrPublishLinkedPermanent(effectIndex, sequence)
	if !ok {
		return nil, game.AbilityContent{}, false
	}
	consumed := ctx
	consumed.content.References = nil
	consumed.content.Targets = nil
	if consumed.content.Unconsumed() {
		return nil, game.AbilityContent{}, false
	}
	object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{
		TargetLinkedKey: key,
	})
	if !ok {
		return nil, game.AbilityContent{}, false
	}
	create := game.CreateReplacement{
		Object: object,
		Replacement: &game.ReplacementEffect{
			MatchEvent:    game.EventZoneChanged,
			MatchFromZone: true,
			FromZone:      zone.Battlefield,
			ReplaceToZone: zone.Exile,
		},
	}
	return publisher, game.Mode{Sequence: []game.Instruction{{Primitive: create}}}.Ability(), true
}

// lowerSequentialReanimationCounterPlacement lowers a counter-placement clause
// ("Put a +1/+1 counter on it." / "Put a +1/+1 counter or a loyalty counter on
// it.", Elspeth Conquers Death chapter III) whose "it" denotes the permanent a
// preceding reanimation clause in the same sequence returned to the battlefield.
// A reanimated permanent is a fresh object a plain target-permanent reference
// cannot resolve, so the clause binds to the linked key under which the
// immediately-prior battlefield-entry instruction publishes the entered
// permanent. It returns the rewritten publishing primitive and the AddCounter
// content, or false to fail closed so the caller lowers the clause normally. It
// applies only when the prior instruction is a single-source battlefield entry,
// where the generic target-permanent reference would be invalid; every other
// prior shape is left to the generic referenced-counter path.
func lowerSequentialReanimationCounterPlacement(
	effectIndex int,
	ctx contentCtx,
	sequence []game.Instruction,
) (game.Primitive, game.AbilityContent, bool) {
	// Invariant: ctx is the contextForEffect-narrowed per-clause effectAbility
	// built at lowerOrderedEffectSequence and threaded through
	// lowerDelayedSequenceClause, so content.Effects always holds exactly one
	// clause effect; any other length is an upstream bug.
	if len(ctx.content.Effects) != 1 {
		panic(fmt.Sprintf("lowerSequentialReanimationCounterPlacement: expected a single effect, got %d", len(ctx.content.Effects)))
	}
	if effectIndex == 0 ||
		len(sequence) != effectIndex ||
		ctx.optional ||
		len(ctx.content.References) != 1 {
		return nil, game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectPut ||
		!effect.Exact ||
		effect.Negated ||
		effect.Context != parser.EffectContextController ||
		!effect.Amount.Known ||
		effect.Amount.Value <= 0 ||
		!referencesBindTo(ctx.content.References, compiler.ReferenceBindingTarget, 0) {
		return nil, game.AbilityContent{}, false
	}
	if sequence[effectIndex-1].Primitive.Kind() != game.PrimitivePutOnBattlefield {
		return nil, game.AbilityContent{}, false
	}
	kindChoices, ok := referencedCounterKindChoices(effect)
	if !ok {
		return nil, game.AbilityContent{}, false
	}
	consumed := ctx
	consumed.content.References = nil
	consumed.content.Targets = nil
	if consumed.content.Unconsumed() {
		return nil, game.AbilityContent{}, false
	}
	key, publisher, ok := reuseOrPublishLinkedPermanent(effectIndex, sequence)
	if !ok {
		return nil, game.AbilityContent{}, false
	}
	object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{
		TargetLinkedKey: key,
	})
	if !ok {
		return nil, game.AbilityContent{}, false
	}
	add := game.AddCounter{
		Amount: game.Fixed(effect.Amount.Value),
		Object: object,
	}
	if len(kindChoices) != 0 {
		add.KindChoices = kindChoices
	} else {
		add.CounterKind = effect.CounterKind
	}
	return publisher, game.Mode{Sequence: []game.Instruction{{Primitive: add}}}.Ability(), true
}

// sequence recorded under a linked key so a later linked effect can bind to it.
// It scans backward for the most recent already-published linked key and reuses
// it (returning the immediately-prior primitive unchanged); when none exists it
// rewrites the immediately-prior instruction to publish its acted-on permanent
// under a fresh key. It returns the key, the primitive to store at the prior
// instruction slot, and false when the prior instruction cannot be linked.
// It scans backward for the most recent already-published linked key and reuses
// it (returning the immediately-prior primitive unchanged); when none exists it
// rewrites the immediately-prior instruction to publish its acted-on permanent
// under a fresh key. It returns the key, the primitive to store at the prior
// instruction slot, and false when the prior instruction cannot be linked.
func reuseOrPublishLinkedPermanent(effectIndex int, sequence []game.Instruction) (game.LinkedKey, game.Primitive, bool) {
	for i := effectIndex - 1; i >= 0; i-- {
		if key := game.PublishedLinkedKey(sequence[i].Primitive); key != "" {
			return key, sequence[effectIndex-1].Primitive, true
		}
	}
	key := game.LinkedKey(fmt.Sprintf("leave-bf-exile-%d", effectIndex))
	publisher, ok := publishLinkedTargetPermanent(sequence[effectIndex-1].Primitive, key)
	if !ok {
		return "", nil, false
	}
	return key, publisher, true
}

// primitive that targets a permanent so it records that permanent under key for a
// later linked effect. It returns the rewritten primitive, or false when the
// primitive does not target a permanent or already publishes a linked object.
func publishLinkedTargetPermanent(primitive game.Primitive, key game.LinkedKey) (game.Primitive, bool) {
	if primitive.Kind() == game.PrimitiveModifyPT {
		modify, ok := primitive.(game.ModifyPT)
		if !ok ||
			modify.Object.Kind() != game.ObjectReferenceTargetPermanent ||
			modify.PublishLinked != "" {
			return nil, false
		}
		modify.PublishLinked = key
		return modify, true
	}
	if primitive.Kind() == game.PrimitiveApplyContinuous {
		apply, ok := primitive.(game.ApplyContinuous)
		if !ok ||
			!apply.Object.Exists ||
			(apply.Object.Val.Kind() != game.ObjectReferenceTargetPermanent &&
				apply.Object.Val.Kind() != game.ObjectReferenceLinkedObject) ||
			apply.PublishLinked != "" {
			return nil, false
		}
		apply.PublishLinked = key
		return apply, true
	}
	// A single-source battlefield entry (reanimation, "Return target creature
	// card from your graveyard to the battlefield") produces one fresh permanent
	// the linked effect can capture, so it publishes the entered permanent under
	// key. A multi-source (Sources) entry produces several and is left unlinkable.
	if primitive.Kind() == game.PrimitivePutOnBattlefield {
		put, ok := primitive.(game.PutOnBattlefield)
		if !ok ||
			len(put.Sources) != 0 ||
			put.PublishLinked != "" {
			return nil, false
		}
		if _, ok := put.Source.CardRef(); !ok {
			return nil, false
		}
		put.PublishLinked = key
		return put, true
	}
	return nil, false
}

// lowerDelayedBlinkReturn lowers the delayed single-target "Exile target
// <permanent>. … return that card to the battlefield … at the beginning of the
// next end step" flicker (blink) clause. The return clause is the second effect
// of a two-step sequence whose object back-references the exiled card (a
// ReferenceBindingPriorInstructionResult "it"/"that card"). Unlike
// lowerImmediateBlinkReturn the card returns at the next end step, so the
// put-onto-battlefield instruction is wrapped in a delayed trigger.
//
// The "under your control" controller rider, the "tapped" entry rider, and a
// fixed "with a <kind> counter on it" entry-counter rider are modeled, matching
// the immediate, self, and group blink lowerers. The same shape with no rider
// (Turn to Mist) and the counter form (Otherworldly Journey, Long Road Home)
// both lower. Every other shape — plural/group exiles, non-target exiles, color
// or type entry choices, unknown counter forms, or unconsumed clause content —
// fails closed so the general sequence path is untouched.
func lowerDelayedBlinkReturn(
	effects []compiler.CompiledEffect,
	effectIndex int,
	ctx contentCtx,
	sequence []game.Instruction,
) (game.Exile, game.AbilityContent, bool) {
	// Invariant: ctx is the contextForEffect-narrowed per-clause effectAbility
	// built at lowerOrderedEffectSequence and threaded through
	// lowerDelayedSequenceClause, so content.Effects always holds exactly one
	// clause effect; any other length is an upstream bug.
	if len(ctx.content.Effects) != 1 {
		panic(fmt.Sprintf("lowerDelayedBlinkReturn: expected a single effect, got %d", len(ctx.content.Effects)))
	}
	returnEffect := ctx.content.Effects[0]
	if effectIndex == 0 ||
		len(sequence) != effectIndex ||
		effects[effectIndex-1].Kind != compiler.EffectExile ||
		effects[effectIndex-1].DelayedTiming != 0 ||
		returnEffect.Kind != compiler.EffectReturn ||
		returnEffect.DelayedTiming != game.DelayedAtBeginningOfNextEndStep ||
		returnEffect.Negated ||
		returnEffect.ToZone != zone.Battlefield ||
		returnEffect.EntersColorChoice ||
		returnEffect.EntersTypeChoice ||
		returnEffect.EntersWithCounters {
		return game.Exile{}, game.AbilityContent{}, false
	}
	if !referencesBindTo(ctx.content.References, compiler.ReferenceBindingPriorInstructionResult, effectIndex-1) {
		return game.Exile{}, game.AbilityContent{}, false
	}
	// "with a <kind> counter on it" rider: only fixed, known, positive counts of a
	// known kind are modeled; every other counter form fails closed.
	var entryCounters []game.CounterPlacement
	if returnEffect.CounterKindKnown {
		if !returnEffect.Amount.Known || returnEffect.Amount.Value < 1 {
			return game.Exile{}, game.AbilityContent{}, false
		}
		entryCounters = []game.CounterPlacement{{
			Kind:   returnEffect.CounterKind,
			Amount: returnEffect.Amount.Value,
		}}
	}
	// References validated — clear before fail-closed check.
	consumed := ctx
	consumed.content.References = nil
	if consumed.content.Unconsumed() {
		return game.Exile{}, game.AbilityContent{}, false
	}
	exile, ok := sequence[effectIndex-1].Primitive.(game.Exile)
	if !ok ||
		exile.Group.Valid() ||
		exile.Object.Kind() != game.ObjectReferenceTargetPermanent ||
		exile.ExileLinkedKey != "" {
		return game.Exile{}, game.AbilityContent{}, false
	}
	key := game.LinkedKey(fmt.Sprintf("delayed-blink-%d", effectIndex))
	if _, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{
		PriorInstruction: effectIndex - 1,
		PriorLinkedKey:   key,
	}); !ok {
		return game.Exile{}, game.AbilityContent{}, false
	}
	exile.ExileLinkedKey = key
	put := game.PutOnBattlefield{
		Source:           game.LinkedBattlefieldSource(key),
		EntryTapped:      returnEffect.EntersTapped,
		EntryTransformed: returnEffect.EntersTransformed,
		EntryCounters:    entryCounters,
	}
	if returnEffect.UnderYourControl {
		put.Recipient = opt.Val(game.ControllerReference())
	}
	delayed := game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
		Timing:  game.DelayedAtBeginningOfNextEndStep,
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: put}}}.Ability(),
	}}
	return exile, game.Mode{Sequence: []game.Instruction{{Primitive: delayed}}}.Ability(), true
}

// lowerSelfBlinkSequence lowers the self-blink "Exile this creature, then return
// it to the battlefield [tapped] under [its owner's|your] control [with a <kind>
// counter on it]." (Flickering Spirit, Ojutai Exemplars, Magus of the Bridge).
// Unlike the target blink the exiled object is the source permanent itself, which
// the compiler co-references through "it"/"its" bound to the source. The parser
// deliberately leaves the standalone "Exile this creature" inexact (nonsensical
// in isolation, e.g. on a spell), so the exile-then-return pair is recognized
// wholesale here rather than through per-effect lowering. It fails closed for any
// shape it does not fully model. The delayed sibling
// (lowerDelayedSelfBlinkSequence) shares the same return contract.
func lowerSelfBlinkSequence(ctx contentCtx) (game.AbilityContent, bool) {
	returnEffect, entryCounters, ok := selfBlinkReturnRiders(ctx)
	if !ok ||
		returnEffect.Connection != parser.EffectConnectionThen ||
		returnEffect.DelayedTiming != 0 {
		return game.AbilityContent{}, false
	}
	key := game.LinkedKey("self-blink")
	exile := game.Exile{Object: game.SourcePermanentReference(), ExileLinkedKey: key}
	put := selfBlinkPutOnBattlefield(key, returnEffect, entryCounters)
	return game.Mode{Sequence: []game.Instruction{
		{Primitive: exile},
		{Primitive: put},
	}}.Ability(), true
}

// lowerImmediateBlinkReturn lowers the immediate "Exile target <permanent>, then
// return it/that card to the battlefield [tapped] under [its owner's|your]
// control [with a +1/+1 counter on it]" flicker (blink) clause. The return clause
// is the second effect of a two-step sequence whose object back-references the
// exiled card (a ReferenceBindingPriorInstructionResult "it"/"its"/"that card").
// Unlike lowerDelayedBlinkReturn the card returns during the same resolution, so
// the put-onto-battlefield instruction is emitted directly rather than wrapped in
// a delayed trigger. It rewrites the preceding exile instruction to remember the
// exiled object under a linked key, and returns that rewritten exile plus the
// put-onto-battlefield content. It returns false (fail closed) for any shape it
// does not fully model — plural/group exiles, non-target exiles, unknown counter
// forms, or unconsumed clause content.
func lowerImmediateBlinkReturn(
	effects []compiler.CompiledEffect,
	effectIndex int,
	ctx contentCtx,
	sequence []game.Instruction,
) (game.Exile, game.AbilityContent, bool) {
	// Invariant: ctx is the contextForEffect-narrowed per-clause effectAbility
	// built at lowerOrderedEffectSequence and threaded through
	// lowerDelayedSequenceClause, so content.Effects always holds exactly one
	// clause effect; any other length is an upstream bug.
	if len(ctx.content.Effects) != 1 {
		panic(fmt.Sprintf("lowerImmediateBlinkReturn: expected a single effect, got %d", len(ctx.content.Effects)))
	}
	returnEffect := ctx.content.Effects[0]
	if effectIndex == 0 ||
		len(sequence) != effectIndex ||
		effects[effectIndex-1].Kind != compiler.EffectExile ||
		effects[effectIndex-1].DelayedTiming != 0 ||
		returnEffect.Kind != compiler.EffectReturn ||
		// Only the ", then return …" connective form lowers immediately. A return
		// whose clause omits "then" (e.g. a leading "At the beginning of the next
		// end step, return …" whose delayed timing the parser does not capture in
		// this position) is rejected so a delayed blink is never resolved at once.
		returnEffect.Connection != parser.EffectConnectionThen ||
		returnEffect.DelayedTiming != 0 ||
		returnEffect.Negated ||
		returnEffect.ToZone != zone.Battlefield ||
		returnEffect.EntersColorChoice ||
		returnEffect.EntersTypeChoice ||
		returnEffect.EntersWithCounters {
		return game.Exile{}, game.AbilityContent{}, false
	}
	if !referencesBindTo(ctx.content.References, compiler.ReferenceBindingPriorInstructionResult, effectIndex-1) {
		return game.Exile{}, game.AbilityContent{}, false
	}
	// "with a <kind> counter on it" rider: only fixed, known, positive counts of a
	// known kind are modeled; every other counter form fails closed.
	var entryCounters []game.CounterPlacement
	if returnEffect.CounterKindKnown {
		if !returnEffect.Amount.Known || returnEffect.Amount.Value < 1 {
			return game.Exile{}, game.AbilityContent{}, false
		}
		entryCounters = []game.CounterPlacement{{
			Kind:   returnEffect.CounterKind,
			Amount: returnEffect.Amount.Value,
		}}
	}
	consumed := ctx
	consumed.content.References = nil
	if consumed.content.Unconsumed() {
		return game.Exile{}, game.AbilityContent{}, false
	}
	exile, ok := sequence[effectIndex-1].Primitive.(game.Exile)
	if !ok ||
		exile.Group.Valid() ||
		exile.Object.Kind() != game.ObjectReferenceTargetPermanent ||
		exile.ExileLinkedKey != "" {
		return game.Exile{}, game.AbilityContent{}, false
	}
	key := game.LinkedKey(fmt.Sprintf("blink-%d", effectIndex))
	if _, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{
		PriorInstruction: effectIndex - 1,
		PriorLinkedKey:   key,
	}); !ok {
		return game.Exile{}, game.AbilityContent{}, false
	}
	exile.ExileLinkedKey = key
	put := game.PutOnBattlefield{
		Source:           game.LinkedBattlefieldSource(key),
		EntryTapped:      returnEffect.EntersTapped,
		EntryTransformed: returnEffect.EntersTransformed,
		EntryCounters:    entryCounters,
	}
	if returnEffect.UnderYourControl {
		put.Recipient = opt.Val(game.ControllerReference())
	}
	return exile, game.Mode{Sequence: []game.Instruction{{Primitive: put}}}.Ability(), true
}

// joinedTokenText reconstructs the source text from a token slice, inserting
// spaces between tokens where appropriate (following oracle punctuation rules).
// This mirrors the unexported compiler.joinedSourceText function.
func joinedTokenText(tokens []shared.Token) string {
	if len(tokens) == 0 {
		return ""
	}
	var b strings.Builder
	for i, tok := range tokens {
		if i > 0 && joinedTokenNeedsSpace(tokens[i-1], tok) { //nolint:gosec // i>0 guarantees valid index
			_ = b.WriteByte(' ')
		}
		_, _ = b.WriteString(tok.Text)
	}
	return b.String()
}

// upperFirst returns s with its first byte uppercased. It is safe for ASCII
// oracle text where the first character is always a plain letter.
func upperFirst(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[size:]
}

// sharedTargetRebaseOffset returns the accumulated-targets start index for the
// first inherited target oracle span, by looking it up in oracleSpanToGameIdx.
// The offset is used to rebase the sequence of an inherited shared-target
// clause (e.g. the "then draws" in "mills …, then draws …") so that its
// local target index 0 maps to the correct position in the already-accumulated
// targets slice, even when an earlier unrelated effect already contributed
// target specs at indices 0, 1, etc.
//
// Returns (0, false) if inherited is empty or the first span has no entry in
// the map (caller should treat this as fail-closed).
// oracleTargetSpanIn reports whether any target in list has the given span.
func oracleTargetSpanIn(span shared.Span, list []compiler.CompiledTarget) bool {
	for _, t := range list {
		if t.Span == span {
			return true
		}
	}
	return false
}

func sharedTargetRebaseOffset(inherited []compiler.CompiledTarget, spanToIdx map[shared.Span]int) (int, bool) {
	if len(inherited) == 0 {
		return 0, false
	}
	idx, ok := spanToIdx[inherited[0].Span]
	return idx, ok
}

// applyTargetRemapping sequences mode's target references to the correct
// accumulated game indices and updates the targets slice and oracleSpanToGameIdx
// map accordingly. It handles three cases:
//   - allSharedTargets: uniform rebase to the inherited target's recorded index.
//   - mixedTargets: non-uniform per-local-index remap for inherited+owned targets.
//   - default: uniform rebase starting at len(accum).
//
// Returns the updated accum slice (false if any remapping step fails).
func applyTargetRemapping(
	mode game.Mode,
	allSharedTargets, mixedTargets bool,
	inherited, owned []compiler.CompiledTarget,
	accum []game.TargetSpec,
	spanToIdx map[shared.Span]int,
) ([]game.TargetSpec, bool) {
	m := mode
	switch {
	case len(m.Targets) > 0 && allSharedTargets:
		if rebaseOffset, ok := sharedTargetRebaseOffset(inherited, spanToIdx); ok {
			if !rebaseTargetedSequence(m.Sequence, rebaseOffset, cardTargetSpecsBefore(accum, rebaseOffset)) {
				return nil, false
			}
			break
		}
		// The inherited target is declared by a bare "Choose target ..."
		// sentence that owns no effect of its own, so this clause is the first
		// to materialize it ("Choose target creature you control. It deals
		// damage equal to its power to each other creature."). Treat the clause
		// as the target's owner: rebase its sequence to the next accumulated
		// index, append its target specs, and record their indices so any later
		// shared clause rebases to the same target.
		if len(m.Targets) != len(inherited) {
			return nil, false
		}
		gameStartIdx := len(accum)
		if !rebaseTargetedSequence(m.Sequence, gameStartIdx, cardTargetSpecsBefore(accum, gameStartIdx)) {
			return nil, false
		}
		for j, t := range inherited {
			spanToIdx[t.Span] = gameStartIdx + j
		}
		accum = append(accum, m.Targets...)
	case len(m.Targets) == 0 && allSharedTargets:
		// A shared-target clause that owns no target spec still embeds the
		// inherited antecedent's clause-local index in its primitives (e.g. a
		// CreateToken whose Recipient is the controller of the inherited target).
		// When the antecedent is not the first accumulated game target, rebase
		// that index so the reference is not silently left pointing at game target
		// 0. When the offset is zero (antecedent is the first game target),
		// existing shared clauses are left exactly as-is.
		if rebaseOffset, ok := sharedTargetRebaseOffset(inherited, spanToIdx); ok && rebaseOffset != 0 {
			if !rebaseTargetedSequence(m.Sequence, rebaseOffset, cardTargetSpecsBefore(accum, rebaseOffset)) {
				return nil, false
			}
		}
	case len(m.Targets) > 0 && mixedTargets:
		if len(m.Targets) != len(inherited)+len(owned) {
			return nil, false
		}
		localToGame := make([]int, len(m.Targets))
		for j, t := range inherited {
			idx, ok := spanToIdx[t.Span]
			if !ok {
				return nil, false
			}
			localToGame[j] = idx
		}
		gameStartForOwn := len(accum)
		for j, ot := range owned {
			localToGame[len(inherited)+j] = gameStartForOwn + j
			spanToIdx[ot.Span] = gameStartForOwn + j
		}
		if !remapTargetedSequence(m.Sequence, localToGame) {
			return nil, false
		}
		accum = append(accum, m.Targets[len(inherited):]...)
	case len(m.Targets) > 0:
		gameStartIdx := len(accum)
		if !rebaseTargetedSequence(m.Sequence, gameStartIdx, cardTargetSpecsBefore(accum, gameStartIdx)) {
			return nil, false
		}
		for j, ot := range owned {
			if j < len(m.Targets) {
				spanToIdx[ot.Span] = gameStartIdx + j
			}
		}
		accum = append(accum, m.Targets...)
	default:
	}
	return accum, true
}

// cardTargetSpecsBefore counts how many of the first n accumulated target specs
// allow card targets. This is the card-reference rebase base, because the runtime
// numbers card-target references among card targets only rather than among all
// targets.
func cardTargetSpecsBefore(specs []game.TargetSpec, n int) int {
	n = min(n, len(specs))
	count := 0
	for i := range n {
		if specs[i].Allow&game.TargetAllowCard != 0 {
			count++
		}
	}
	return count
}

func joinedTokenNeedsSpace(prev, cur shared.Token) bool {
	if cur.Kind == shared.Comma || cur.Kind == shared.Period || cur.Kind == shared.Colon ||
		cur.Kind == shared.Semicolon || cur.Kind == shared.RightParen ||
		cur.Kind == shared.Apostrophe || prev.Kind == shared.Apostrophe ||
		prev.Kind == shared.LeftParen || prev.Kind == shared.Quote || cur.Kind == shared.Quote {
		return false
	}
	if prev.Kind == shared.Plus || prev.Kind == shared.Minus || prev.Kind == shared.Slash ||
		cur.Kind == shared.Plus || cur.Kind == shared.Minus || cur.Kind == shared.Slash ||
		prev.Kind == shared.Asterisk || cur.Kind == shared.Asterisk {
		return false
	}
	return true
}

func lowerCyclingCountDamageAndGain(_ string, ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 ||
		ctx.content.Effects[0].Kind != compiler.EffectDealDamage ||
		ctx.content.Effects[1].Kind != compiler.EffectGain ||
		ctx.content.Effects[0].Context != parser.EffectContextSource ||
		ctx.content.Effects[1].Context != parser.EffectContextController ||
		ctx.content.Effects[1].Connection != parser.EffectConnectionAnd ||
		ctx.content.Effects[0].Negated ||
		ctx.content.Effects[1].Negated ||
		len(ctx.content.Targets) != 1 ||
		ctx.content.Targets[0].Cardinality.Min != 1 ||
		ctx.content.Targets[0].Cardinality.Max != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!singleSelfReference(ctx.content.References) {
		return game.AbilityContent{}, false
	}
	amountEffect := ctx.content.Effects[1].Amount
	if amountEffect.DynamicKind == compiler.DynamicAmountNone ||
		amountEffect.DynamicForm != compiler.DynamicAmountWhereX ||
		!ctx.content.Effects[0].Amount.VariableX {
		return game.AbilityContent{}, false
	}
	dynamic, ok := lowerDynamicAmount(amountEffect, game.SourcePermanentReference())
	if !ok {
		return game.AbilityContent{}, false
	}
	target, ok := damageTargetSpec(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, false
	}
	amount := game.Dynamic(dynamic)
	return game.Mode{
		Targets: []game.TargetSpec{target},
		Sequence: []game.Instruction{
			{Primitive: game.Damage{
				Amount:    amount,
				Recipient: game.AnyTargetDamageRecipient(0),
			}},
			{Primitive: game.GainLife{
				Amount: amount,
				Player: game.ControllerReference(),
			}},
		},
	}.Ability(), true
}

// lowerExcessDamageToControllerSpell lowers the "<Spell> deals N damage to target
// creature. Excess damage is dealt to that creature's controller instead." family
// (Flame Spill, Pigment Storm, Gandalf's Sanction): a fixed or X amount of damage
// dealt to one target creature, with the overflow beyond what was lethal to the
// creature redirected to that creature's controller (or owner). It emits one
// target spec and a single Damage instruction whose ExcessRecipient redirects the
// excess to the resolved controller/owner, so the creature is dealt only its
// lethal damage and the remainder is dealt to the player as one damage event
// (CR self-replacement). It fails closed for every other shape, leaving the
// ordinary single-target and sequence paths unchanged.
func lowerExcessDamageToControllerSpell(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 ||
		ctx.content.Effects[0].Kind != compiler.EffectDealDamage ||
		ctx.content.Effects[1].Kind != compiler.EffectDealDamage ||
		ctx.content.Effects[0].Negated || ctx.content.Effects[1].Negated ||
		ctx.content.Effects[0].Divided || ctx.content.Effects[1].Divided ||
		ctx.content.Effects[0].DamageRecipient.Reference != parser.DamageRecipientReferenceNone ||
		len(ctx.content.Effects[0].DamageRiders) != 0 ||
		len(ctx.content.Targets) != 1 ||
		ctx.content.Targets[0].Cardinality.Min != 1 ||
		ctx.content.Targets[0].Cardinality.Max != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	recipientKind := ctx.content.Effects[1].DamageRecipient.Reference
	if recipientKind != parser.DamageRecipientReferenceController &&
		recipientKind != parser.DamageRecipientReferenceOwner {
		return game.AbilityContent{}, false
	}
	if ctx.content.Effects[1].Amount.DynamicKind != compiler.DynamicAmountExcessDamageDealtThisWay {
		return game.AbilityContent{}, false
	}
	primary := ctx.content.Effects[0]
	if primary.Amount.Known && primary.Amount.Value < 1 {
		return game.AbilityContent{}, false
	}
	amount, ok := lowerSingleTargetDamageAmount(ctx, primary)
	if !ok {
		return game.AbilityContent{}, false
	}
	target, ok := damageTargetSpec(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, false
	}
	recipient, ok := referencedDamageRecipientPlayer(ctx, recipientKind)
	if !ok {
		return game.AbilityContent{}, false
	}
	source := primaryDamageSource(ctx.content.References)
	return game.Mode{
		Targets: []game.TargetSpec{target},
		Sequence: []game.Instruction{{
			Primitive: game.Damage{
				Amount:          rebindRecipientControlledCountAmount(rebindRecipientHandSizeAmount(amount, game.TargetPlayerReference(0)), game.TargetPlayerReference(0)),
				Recipient:       game.AnyTargetDamageRecipient(0),
				DamageSource:    source,
				ExcessRecipient: game.PlayerDamageRecipient(recipient),
			},
		}},
	}.Ability(), true
}

// lowerGroupLinkedLifeSpell handles linked two-effect patterns of the form
// "Each opponent loses N life and you gain [N | that much] life."
// It emits LoseLife with PublishResult "life-change" followed by GainLife.
// For "that much", the GainLife amount uses DynamicAmountPreviousEffectResult.
func lowerGroupLinkedLifeSpell(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 ||
		ctx.content.Effects[0].Kind != compiler.EffectLose ||
		ctx.content.Effects[1].Kind != compiler.EffectGain ||
		ctx.content.Effects[0].Context != parser.EffectContextEachOpponent ||
		ctx.content.Effects[1].Context != parser.EffectContextController ||
		ctx.content.Effects[1].Connection != parser.EffectConnectionAnd ||
		ctx.content.Effects[0].Negated ||
		ctx.content.Effects[1].Negated ||
		!ctx.content.Effects[0].Amount.Known ||
		ctx.content.Effects[0].Amount.Value < 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 {
		return game.AbilityContent{}, false
	}
	loseAmount := game.Fixed(ctx.content.Effects[0].Amount.Value)

	// Determine the gain amount: fixed if effects[1] has a known value, dynamic ("that much") otherwise.
	var gainAmount game.Quantity
	switch {
	case ctx.content.Effects[1].Amount.Known && ctx.content.Effects[1].Amount.Value > 0:
		gainAmount = game.Fixed(ctx.content.Effects[1].Amount.Value)
	case !ctx.content.Effects[1].Amount.Known:
		gainAmount = game.Dynamic(game.DynamicAmount{
			Kind:      game.DynamicAmountPreviousEffectResult,
			ResultKey: "life-change",
		})
	default:
		return game.AbilityContent{}, false
	}

	return game.Mode{
		Sequence: []game.Instruction{
			{
				Primitive:     game.LoseLife{PlayerGroup: game.OpponentsReference(), Amount: loseAmount},
				PublishResult: "life-change",
			},
			{
				Primitive: game.GainLife{Player: game.ControllerReference(), Amount: gainAmount},
			},
		},
	}.Ability(), true
}

// lowerDestroyedThisWaySequence handles the mass-destroy payoff pattern
// "Destroy all <group>. <You gain N life | You lose N life | Draw a card> for
// each <permanent> destroyed this way." (Fumigate, Multani's Decree, Paraselene,
// Righteous Fury, Rain of Daggers, Death Begets Life). The first clause is an
// exact untargeted mass destroy; the payoff clause's amount is the "for each
// <noun> destroyed this way" dynamic form. It emits a group Destroy that
// publishes the number of permanents it destroyed under "destroyed-this-way"
// followed by the payoff instruction whose amount reads that published count
// (scaled by the per-permanent multiplier), so the controller gains, loses, or
// draws exactly that many. It fails closed unless every guard holds, so targeted
// mass destroys and richer wordings keep failing the round-trip.
func lowerDestroyedThisWaySequence(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 {
		return game.AbilityContent{}, false
	}
	destroy := &ctx.content.Effects[0]
	payoff := &ctx.content.Effects[1]
	if destroy.Kind != compiler.EffectDestroy ||
		!destroy.Exact ||
		!destroy.Selector.All ||
		destroy.Negated || destroy.Optional || ctx.optional ||
		destroy.Context != parser.EffectContextController ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 {
		return game.AbilityContent{}, false
	}
	if payoff.Amount.DynamicKind != compiler.DynamicAmountDestroyedThisWay ||
		payoff.Amount.DynamicForm != compiler.DynamicAmountForEach ||
		payoff.Amount.Multiplier < 1 ||
		!payoff.Exact ||
		payoff.Negated || payoff.Optional ||
		payoff.Context != parser.EffectContextController {
		return game.AbilityContent{}, false
	}
	selection, ok := massGroupSelection(destroy.Selector)
	if !ok {
		return game.AbilityContent{}, false
	}
	const resultKey = game.ResultKey("destroyed-this-way")
	amount := game.Dynamic(game.DynamicAmount{
		Kind:       game.DynamicAmountPreviousEffectResult,
		ResultKey:  resultKey,
		Multiplier: payoff.Amount.Multiplier,
	})
	var payoffPrimitive game.Primitive
	switch payoff.Kind {
	case compiler.EffectGain:
		payoffPrimitive = game.GainLife{Player: game.ControllerReference(), Amount: amount}
	case compiler.EffectLose:
		payoffPrimitive = game.LoseLife{Player: game.ControllerReference(), Amount: amount}
	case compiler.EffectDraw:
		payoffPrimitive = game.Draw{Player: game.ControllerReference(), Amount: amount}
	default:
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Sequence: []game.Instruction{
			{
				Primitive:     game.Destroy{Group: game.BattlefieldGroup(selection)},
				PublishResult: resultKey,
			},
			{Primitive: payoffPrimitive},
		},
	}.Ability(), true
}

// lowerDiceTableSequence handles the die-roll outcome-table form "Roll a d<N>.
// <lo>—<hi> | <effect> ..." (Bag of Tricks et al.). The compiler flattens the
// table into a leading exact controller-scoped EffectRollDie followed by each
// row's effects, every row effect stamped with its inclusive result interval
// [DiceRowMin, DiceRowMax]. This emits a RollDie that publishes its rolled value
// under dieRollResultKey, then lowers each contiguous same-interval row group
// through the standard content path and gates every resulting instruction on the
// rolled value falling in that interval, so exactly the matching row resolves.
// It fails closed unless the first effect is the exact roll, every later effect
// belongs to a row, the content carries no shared targets/conditions/modes/
// references, and each row lowers to a single non-modal untargeted mode whose
// instructions carry no existing result gate.
func lowerDiceTableSequence(cardName string, ctx contentCtx, syntax *parser.Ability) (game.AbilityContent, bool) {
	effects := ctx.content.Effects
	if len(effects) < 2 {
		return game.AbilityContent{}, false
	}
	roll := &effects[0]
	if roll.Kind != compiler.EffectRollDie ||
		roll.DieSides < 2 ||
		!roll.Exact ||
		roll.DiceRow ||
		roll.Negated || roll.Optional || ctx.optional ||
		roll.Context != parser.EffectContextController ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 {
		return game.AbilityContent{}, false
	}
	for i := 1; i < len(effects); i++ {
		if !effects[i].DiceRow {
			return game.AbilityContent{}, false
		}
	}
	publisher := game.Instruction{Primitive: game.RollDie{Sides: roll.DieSides}}
	var branches []resultGatedBranch
	for start := 1; start < len(effects); {
		end := start + 1
		for end < len(effects) &&
			effects[end].DiceRowMin == effects[start].DiceRowMin &&
			effects[end].DiceRowMax == effects[start].DiceRowMax {
			end++
		}
		rowRange := game.IntRange{Min: effects[start].DiceRowMin, Max: effects[start].DiceRowMax}
		rowCtx := ctx
		rowCtx.content = compiler.AbilityContent{Effects: effects[start:end]}
		rowContent, diagnostic := lowerContent(cardName, rowCtx, syntax)
		if diagnostic != nil ||
			rowContent.IsModal() ||
			len(rowContent.SharedTargets) != 0 ||
			len(rowContent.Modes[0].Targets) != 0 {
			return game.AbilityContent{}, false
		}
		branches = append(branches, resultGatedBranch{
			predicate: game.InstructionResultGate{AmountRange: opt.Val(rowRange)},
			sequence:  rowContent.Modes[0].Sequence,
		})
		start = end
	}
	sequence, ok := assembleResultGatedBranches(publisher, dieRollResultKey, branches)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: sequence}.Ability(), true
}

// lowerDieRollResultSequence handles the dice pattern "Roll a d<N>. <payoff>"
// where one or more payoff effects scale by "the result" (Ancient Copper Dragon
// and the Ancient Dragon dice cycle). It emits a RollDie that publishes its
// rolled value under dieRollResultKey, then lowers every effect after the roll
// through the standard content path and prepends the publishing instruction.
// Because the payoff is lowered by the general machinery, any supported payoff
// composes: artifact-token creation (Treasure), creature-token creation (Faerie
// Dragon), card draw, and riders such as "no maximum hand size". Each payoff's
// "equal to the result" amount lowers (via lowerDynamicAmount) to a
// previous-effect-result read of the published die value. It fails closed unless
// the first effect is an exact controller-scoped die roll, the sequence carries
// no content-level targets/conditions/modes/references/keywords, and at least
// one payoff effect actually reads the die result.
func lowerDieRollResultSequence(cardName string, ctx contentCtx, syntax *parser.Ability) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) < 2 {
		return game.AbilityContent{}, false
	}
	roll := &ctx.content.Effects[0]
	if roll.Kind != compiler.EffectRollDie ||
		roll.DieSides < 2 ||
		!roll.Exact ||
		roll.Negated || roll.Optional || ctx.optional ||
		roll.Context != parser.EffectContextController ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Keywords) != 0 {
		return game.AbilityContent{}, false
	}
	if !sequenceReadsDieRollResult(ctx.content.Effects[1:]) {
		return game.AbilityContent{}, false
	}
	// Lower the payoff (every effect after the roll) through the standard content
	// path so any supported payoff composes; its "equal to the result" amount
	// resolves to a previous-effect-result read keyed to the die roll. Then
	// prepend the publishing RollDie instruction.
	payoffCtx := ctx
	payoffCtx.content.Effects = ctx.content.Effects[1:]
	payoffContent, diagnostic := lowerContent(cardName, payoffCtx, syntax)
	if diagnostic != nil || payoffContent.IsModal() ||
		len(payoffContent.Modes) != 1 || len(payoffContent.SharedTargets) != 0 {
		return game.AbilityContent{}, false
	}
	payoffContent.Modes[0].Sequence = append(
		[]game.Instruction{{
			Primitive:     game.RollDie{Sides: roll.DieSides},
			PublishResult: dieRollResultKey,
		}},
		payoffContent.Modes[0].Sequence...,
	)
	return payoffContent, true
}

// sequenceReadsDieRollResult reports whether any effect in the payoff scales by
// the die-roll result, so a RollDie is only emitted when its published value is
// consumed.
func sequenceReadsDieRollResult(effects []compiler.CompiledEffect) bool {
	for i := range effects {
		if effects[i].Amount.DynamicKind == compiler.DynamicAmountDieRollResult {
			return true
		}
	}
	return false
}

// lowerLifeLostThisWayDrain handles the two-effect drain pattern
// "Each opponent loses <amount> life. You gain life equal to the life lost this
// way." The two clauses are separate sentences (the joining "that much" case is
// owned by lowerGroupLinkedLifeSpell), and the gain clause's amount is the
// explicit "equal to the life lost this way" dynamic form. It emits a group
// LoseLife that publishes its total under "life-change" followed by a GainLife
// whose amount reads that published result, so the controller gains exactly the
// life lost. It fails closed unless every guard holds, including a lose amount
// that lowers to a supported quantity (a fixed value, the spell's X, or an
// "equal to ..." count) and an exact lose clause.
func lowerLifeLostThisWayDrain(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 {
		return game.AbilityContent{}, false
	}
	lose := &ctx.content.Effects[0]
	gain := &ctx.content.Effects[1]
	if lose.Kind != compiler.EffectLose ||
		gain.Kind != compiler.EffectGain ||
		lose.Context != parser.EffectContextEachOpponent ||
		gain.Context != parser.EffectContextController ||
		lose.Negated || gain.Negated || ctx.optional ||
		!lose.Exact || !gain.Exact ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 {
		return game.AbilityContent{}, false
	}
	if gain.Amount.DynamicKind != compiler.DynamicAmountLifeLostThisWay ||
		gain.Amount.DynamicForm != compiler.DynamicAmountEqual {
		return game.AbilityContent{}, false
	}
	loseAmount, ok := drainLoseAmount(lose)
	if !ok {
		return game.AbilityContent{}, false
	}
	gainAmount := game.Dynamic(game.DynamicAmount{
		Kind:      game.DynamicAmountPreviousEffectResult,
		ResultKey: "life-change",
	})
	return game.Mode{
		Sequence: []game.Instruction{
			{
				Primitive:     game.LoseLife{PlayerGroup: game.OpponentsReference(), Amount: loseAmount},
				PublishResult: "life-change",
			},
			{
				Primitive: game.GainLife{Player: game.ControllerReference(), Amount: gainAmount},
			},
		},
	}.Ability(), true
}

// lowerDiscardDrawThenManaValueDamageSequence handles Summon: Kujata chapter III
// "Discard a card, then draw two cards. When you discard a card this way, this
// creature deals damage equal to that card's mana value to each opponent." The
// parser leaves the reflexive "When you discard a card this way" preamble
// in-sentence, so the chapter flattens to four effects: the controller's
// single-card discard, a fixed draw chained with "then", the reflexive
// restatement of that same discard, and the source-dealt damage to each opponent
// whose amount equals "that card's mana value". The restatement effect carries no
// independent action — the controller discards exactly once — so it is collapsed.
//
// It emits a Discard that publishes the discarded card under a linked key,
// followed by the fixed Draw, followed by a source Damage to each opponent whose
// amount reads the published card's mana value. With an empty hand nothing is
// discarded, the linked key stays empty, and the mana-value amount resolves to
// zero, matching the reflexive trigger that never fires. It fails closed unless
// every guard holds.
func lowerDiscardDrawThenManaValueDamageSequence(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) != 4 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	discard := &ctx.content.Effects[0]
	draw := &ctx.content.Effects[1]
	restatement := &ctx.content.Effects[2]
	damage := &ctx.content.Effects[3]
	if !singleCardControllerDiscardAction(discard) || !reflexiveDiscardRestatement(restatement) {
		return game.AbilityContent{}, false
	}
	if draw.Kind != compiler.EffectDraw ||
		(draw.Context != parser.EffectContextController && draw.Context != parser.EffectContextPriorSubject) ||
		draw.Connection != parser.EffectConnectionThen ||
		draw.Negated || draw.Optional ||
		!draw.Exact ||
		!draw.Amount.Known || draw.Amount.Value < 1 ||
		draw.Amount.DynamicKind != compiler.DynamicAmountNone ||
		draw.Amount.VariableX || draw.Amount.RangeKnown ||
		len(draw.References) != 0 {
		return game.AbilityContent{}, false
	}
	if damage.Kind != compiler.EffectDealDamage ||
		damage.Context != parser.EffectContextSource ||
		damage.Negated || damage.Optional || damage.Divided ||
		damage.Amount.DynamicKind != compiler.DynamicAmountSourceManaValue ||
		damage.Amount.DynamicForm != compiler.DynamicAmountEqual ||
		damage.Amount.Multiplier != 1 ||
		damage.Selector.Kind != compiler.SelectorOpponent || damage.Selector.Other ||
		len(damage.DamageRecipient.GroupSelectors) != 0 {
		return game.AbilityContent{}, false
	}
	if !damageSourceIsSourcePermanent(damage.References) ||
		len(damage.References) != 2 ||
		damage.References[1].Kind != compiler.ReferenceThatObject {
		return game.AbilityContent{}, false
	}

	const linkKey = game.LinkedKey("discarded-card-mana-value")
	manaValue, ok := objectCharacteristicAmount(
		compiler.DynamicAmountSourceManaValue,
		game.LinkedObjectReference(string(linkKey)),
	)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Sequence: []game.Instruction{
			{Primitive: game.Discard{
				Amount:        game.Fixed(1),
				Player:        game.ControllerReference(),
				PublishLinked: linkKey,
			}},
			{Primitive: game.Draw{
				Amount: game.Fixed(draw.Amount.Value),
				Player: game.ControllerReference(),
			}},
			{Primitive: game.Damage{
				Amount:       game.Dynamic(manaValue),
				Recipient:    game.PlayerGroupDamageRecipient(game.OpponentsReference()),
				DamageSource: opt.Val(game.SourcePermanentReference()),
			}},
		},
	}.Ability(), true
}

// singleCardControllerDiscardAction reports whether the effect is the
// controller's own unfiltered single-card discard action ("Discard a card."): an
// exact one-card discard from hand with no typed card filter, no entire-hand or
// at-random rider, no target/reference, and no delayed timing or duration. It is
// the discard whose discarded card
// lowerDiscardDrawThenManaValueDamageSequence remembers under a linked key.
func singleCardControllerDiscardAction(effect *compiler.CompiledEffect) bool {
	return singleCardControllerDiscard(effect) &&
		effect.HandDiscard.Present &&
		effect.Exact
}

// reflexiveDiscardRestatement reports whether the effect is the reflexive
// restatement of the discard action that the parser leaves in-sentence ("When
// you discard a card this way, ..."). It is a controller single-card discard
// that, unlike the action itself, carries neither the HandDiscard structure nor
// the exact flag, so it never matches a genuine "Discard a card." action and is
// safely collapsed.
func reflexiveDiscardRestatement(effect *compiler.CompiledEffect) bool {
	return singleCardControllerDiscard(effect) &&
		!effect.HandDiscard.Present &&
		!effect.Exact
}

// singleCardControllerDiscard reports whether the effect is the controller's own
// unfiltered single-card discard ("Discard a card."): a one-card discard with no
// typed card filter, no entire-hand or at-random rider, no target/reference, and
// no delayed timing or duration.
func singleCardControllerDiscard(effect *compiler.CompiledEffect) bool {
	return effect.Kind == compiler.EffectDiscard &&
		effect.Context == parser.EffectContextController &&
		!effect.DiscardEntireHand &&
		!effect.HandDiscard.AtRandom &&
		!effect.Negated &&
		!effect.Optional &&
		effect.DelayedTiming == 0 &&
		effect.Duration == compiler.DurationNone &&
		effect.Amount.Known &&
		effect.Amount.Value == 1 &&
		!effect.Amount.RangeKnown &&
		!effect.Amount.VariableX &&
		effect.Amount.DynamicKind == compiler.DynamicAmountNone &&
		!discardSelectorImposesCardFilter(effect.Selector) &&
		len(effect.References) == 0
}

// lowerDiscardDrawGreatestThisWaySequence handles the Windfall pattern
// "Each player discards their hand, then draws cards equal to the greatest
// number of cards a player discarded this way." The discard clause is an exact
// each-player "discard their hand" and the draw clause inherits the each-player
// subject ("then draws ...") with the "greatest number of cards a player
// discarded this way" dynamic amount. It emits a group Discard that publishes
// the greatest per-player discard count under "discarded-this-way" followed by a
// group Draw whose amount reads that published result, so every player draws
// that maximum. It fails closed unless every guard holds.
func lowerDiscardDrawGreatestThisWaySequence(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 {
		return game.AbilityContent{}, false
	}
	discard := &ctx.content.Effects[0]
	draw := &ctx.content.Effects[1]
	if discard.Kind != compiler.EffectDiscard ||
		draw.Kind != compiler.EffectDraw ||
		!discard.DiscardEntireHand ||
		discard.Context != parser.EffectContextEachPlayer ||
		draw.Context != parser.EffectContextPriorSubject && draw.Context != parser.EffectContextEachPlayer ||
		discard.Negated || draw.Negated || discard.Optional || draw.Optional || ctx.optional ||
		!discard.Exact || !draw.Exact ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	if draw.Amount.DynamicKind != compiler.DynamicAmountGreatestDiscardedThisWay ||
		draw.Amount.DynamicForm != compiler.DynamicAmountEqual {
		return game.AbilityContent{}, false
	}
	const resultKey = game.ResultKey("discarded-this-way")
	return game.Mode{
		Sequence: []game.Instruction{
			{
				Primitive:     game.Discard{EntireHand: true, PlayerGroup: game.AllPlayersReference()},
				PublishResult: resultKey,
			},
			{
				Primitive: game.Draw{
					PlayerGroup: game.AllPlayersReference(),
					Amount: game.Dynamic(game.DynamicAmount{
						Kind:      game.DynamicAmountPreviousEffectResult,
						ResultKey: resultKey,
					}),
				},
			},
		},
	}.Ability(), true
}

// lowerWheelDiscardDrawSequence handles the "wheel" pattern "<subject> discards
// their hand, then draws N cards" (Wheel of Fortune, Wheel of Misfortune, Magus
// of the Wheel, and single-player "You discard your hand, then draw seven
// cards"). The discard clause is an exact whole-hand discard whose subject is
// every player ("Each player discards their hand") or the controller ("You
// discard your hand"); the draw clause inherits that subject ("then draws ...")
// with a fixed card count or the spell's X. It emits a whole-hand Discard
// followed by a Draw, both scoped to the same player group or controller. It
// fails closed unless every guard holds, so dynamic "this way" wheels (handled
// by lowerDiscardDrawGreatestThisWaySequence) and any richer wording keep
// failing the round-trip.
func lowerWheelDiscardDrawSequence(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 {
		return game.AbilityContent{}, false
	}
	discard := &ctx.content.Effects[0]
	draw := &ctx.content.Effects[1]
	fixedAmount := draw.Amount.Known && draw.Amount.Value >= 1 &&
		draw.Amount.DynamicKind == compiler.DynamicAmountNone
	variableX := draw.Amount.VariableX && !draw.Amount.Known &&
		draw.Amount.DynamicKind == compiler.DynamicAmountNone
	if discard.Kind != compiler.EffectDiscard ||
		draw.Kind != compiler.EffectDraw ||
		!discard.DiscardEntireHand ||
		!discard.Exact ||
		discard.Negated || draw.Negated || discard.Optional || draw.Optional || ctx.optional ||
		(!fixedAmount && !variableX) ||
		len(ctx.content.Targets) != 0 ||
		!wheelReferencesAllTheir(ctx.content.References) ||
		len(ctx.content.Conditions) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	if draw.Context != parser.EffectContextPriorSubject && draw.Context != discard.Context {
		return game.AbilityContent{}, false
	}
	amount := game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX})
	if fixedAmount {
		amount = game.Fixed(draw.Amount.Value)
	}
	discardPrimitive := game.Discard{EntireHand: true}
	drawPrimitive := game.Draw{Amount: amount}
	switch discard.Context {
	case parser.EffectContextEachPlayer:
		discardPrimitive.PlayerGroup = game.AllPlayersReference()
		drawPrimitive.PlayerGroup = game.AllPlayersReference()
	case parser.EffectContextController:
		discardPrimitive.Player = game.ControllerReference()
		drawPrimitive.Player = game.ControllerReference()
	default:
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Sequence: []game.Instruction{
			{Primitive: discardPrimitive},
			{Primitive: drawPrimitive},
		},
	}.Ability(), true
}

// wheelReferencesAllTheir reports whether every content reference is the
// possessive "their" pronoun of an each-player whole-hand discard ("Each player
// discards their hand"). That pronoun is already absorbed by the EntireHand
// discard, so it is harmless; any other reference (a target, an event subject, a
// stray pronoun) makes the wheel fail closed.
func wheelReferencesAllTheir(references []compiler.CompiledReference) bool {
	for i := range references {
		if references[i].Pronoun != compiler.ReferencePronounTheir {
			return false
		}
	}
	return true
}

// drainLoseAmount lowers the life-loss amount of an "Each opponent loses
// <amount> life" drain clause to a runtime quantity. It accepts a fixed value, a
// spell's X ("loses X life"), or a dynamic "equal to ..." / "where X is ..."
// count that lowerDynamicAmount recognizes (for example "where X is your
// devotion to black"); it fails closed for every other amount form.
func drainLoseAmount(effect *compiler.CompiledEffect) (game.Quantity, bool) {
	amount := effect.Amount
	switch {
	case amount.DynamicKind == compiler.DynamicAmountNone && !amount.VariableX &&
		amount.Known && amount.Value >= 1:
		return game.Fixed(amount.Value), true
	case amount.DynamicKind == compiler.DynamicAmountNone && amount.VariableX && !amount.Known:
		return game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX}), true
	case amount.DynamicKind != compiler.DynamicAmountNone &&
		(amount.DynamicForm == compiler.DynamicAmountEqual ||
			amount.DynamicForm == compiler.DynamicAmountWhereX):
		dynamic, ok := lowerDynamicAmount(amount, game.SourcePermanentReference())
		if !ok {
			return game.Quantity{}, false
		}
		return game.Dynamic(dynamic), true
	default:
		return game.Quantity{}, false
	}
}
