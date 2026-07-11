package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func lowerEnterTrigger(
	cardName string,
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.TriggeredAbility, *shared.Diagnostic) {
	if ability.Trigger == nil {
		return game.TriggeredAbility{}, executableDiagnostic(
			ability,
			"unsupported triggered ability",
			"the executable source backend requires a semantic self trigger pattern",
		)
	}
	pattern, supportedEvent := lowerTriggerPattern(&ability.Trigger.Pattern)
	eventKind := pattern.Event
	summary := "unsupported triggered ability"
	effectSummary := "unsupported triggered ability effect"
	detail := "the executable source backend supports only recognized semantic self triggers with supported effects"
	switch ability.Trigger.Pattern.Event {
	case compiler.TriggerEventPermanentEnteredBattlefield:
		summary = "unsupported enter trigger"
		effectSummary = "unsupported enter trigger effect"
		detail = "the executable source backend supports only recognized semantic self-enter triggers with supported effects"
	case compiler.TriggerEventPermanentDied:
		summary = "unsupported dies trigger"
		effectSummary = "unsupported dies trigger effect"
		detail = "the executable source backend supports only recognized semantic self-dies triggers with supported effects"
	default:
	}
	intervening, supportedCondition := lowerSelfInterveningCondition(eventKind, ability.Trigger)
	if !supportedSelfTriggerKind(eventKind, ability.Trigger.Pattern.Kind) ||
		!supportedEvent ||
		!supportedCondition {
		return game.TriggeredAbility{}, executableDiagnostic(ability, summary, detail)
	}
	if triggerContentUnsupported(ability) {
		return game.TriggeredAbility{}, executableDiagnostic(ability, effectSummary, detail)
	}
	prepared, ok := prepareTriggerBody(ability, syntax)
	if !ok {
		return game.TriggeredAbility{}, executableDiagnostic(ability, effectSummary, detail)
	}
	body, bodySyntax, triggerOptional := prepared.body, prepared.syntax, prepared.optional
	content, diagnostic := lowerTriggerBodyContent(cardName, body.Content, body.Optional, &bodySyntax, pattern)
	if diagnostic != nil {
		return game.TriggeredAbility{}, diagnostic
	}
	triggerType, ok := lowerTriggerKind(ability.Trigger.Pattern.Kind)
	if !ok {
		return game.TriggeredAbility{}, executableDiagnostic(ability, summary, detail)
	}
	return game.TriggeredAbility{
		Text: ability.Text,
		Trigger: game.TriggerCondition{
			Type:                                   triggerType,
			Pattern:                                pattern,
			InterveningIf:                          interveningIfText(ability.Trigger),
			InterveningCondition:                   intervening.condition,
			InterveningIfEventPermanentHadCounters: intervening.hadCounters,
			InterveningIfEventPermanentHadNoCounterKind:                     intervening.hadNoCounterKind,
			InterveningIfEventPermanentHadCounterKind:                       intervening.hadCounterKind,
			InterveningIfEventPermanentWasKicked:                            intervening.wasKicked,
			InterveningIfEventPermanentWasCast:                              intervening.wasCast,
			InterveningIfEventPermanentWasCastByController:                  intervening.wasCastByController,
			InterveningIfEventPermanentWasCastFromControllerHand:            intervening.wasCastFromCtrlHand,
			InterveningIfEventPermanentEnteredOrCastFromGraveyard:           intervening.enteredOrCastFromGY,
			InterveningIfEventPermanentEnteredOrCastFromControllerGraveyard: intervening.enteredOrCastFromCGY,
		},
		Optional: triggerOptional,
		Content:  content,
	}, nil
}

func lowerLifeDamageTrigger(
	cardName string,
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.TriggeredAbility, *shared.Diagnostic) {
	if ability.Trigger == nil {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability",
			"the executable source backend requires a semantic life or damage trigger")
	}
	pattern, ok := lowerTriggerPattern(&ability.Trigger.Pattern)
	if !ok ||
		(pattern.Event != game.EventLifeGained &&
			pattern.Event != game.EventLifeLost &&
			pattern.Event != game.EventDamageDealt) {
		if ability.Trigger.Pattern.OneOrMore {
			if diagnostic := triggerBodyDiagnostic(cardName, ability, syntax); diagnostic != nil {
				return game.TriggeredAbility{}, diagnostic
			}
		}
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability",
			"the executable source backend does not support this semantic life or damage trigger pattern")
	}
	triggerType, ok := lowerTriggerKind(ability.Trigger.Pattern.Kind)
	if !ok || (triggerType != game.TriggerWhen && triggerType != game.TriggerWhenever) {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability",
			"life and damage triggers require When or Whenever")
	}
	intervening, ok := lowerAtInterveningCondition(ability.Trigger)
	if !ok {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability",
			"the executable source backend does not support this semantic life or damage trigger condition")
	}
	// A modal "choose one —" body routes through the shared modal-content
	// lowering, exactly like spell-cast and permanent-zone-change triggers; each
	// mode lowers as an independent already-supported effect and the runtime
	// presents the modes when the ability resolves.
	if modalTriggerBody(ability) {
		content, diagnostic := lowerModalTriggerBody(cardName, ability, syntax, pattern.Event)
		if diagnostic != nil {
			return game.TriggeredAbility{}, diagnostic
		}
		return game.TriggeredAbility{
			Text: ability.Text,
			Trigger: game.TriggerCondition{
				Type:                 triggerType,
				Pattern:              pattern,
				InterveningIf:        interveningIfText(ability.Trigger),
				InterveningCondition: intervening,
			},
			Optional: ability.Optional,
			Content:  content,
		}, nil
	}
	if triggerContentUnsupported(ability) {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability effect",
			"the executable source backend does not support this life or damage trigger body")
	}
	prepared, ok := prepareTriggerBody(ability, syntax)
	if !ok {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability effect",
			"the executable source backend does not support this life or damage trigger body")
	}
	body, bodySyntax, triggerOptional := prepared.body, prepared.syntax, prepared.optional
	content, diagnostic := lowerTriggerBodyContent(
		cardName,
		body.Content,
		body.Optional,
		&bodySyntax,
		pattern,
	)
	if diagnostic != nil {
		return game.TriggeredAbility{}, diagnostic
	}
	return game.TriggeredAbility{
		Text: ability.Text,
		Trigger: game.TriggerCondition{
			Type:                 triggerType,
			Pattern:              pattern,
			InterveningIf:        interveningIfText(ability.Trigger),
			InterveningCondition: intervening,
		},
		Optional: triggerOptional,
		Content:  content,
	}, nil
}

func lowerEventCardEffect(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 1 {
		return game.AbilityContent{}, false
	}
	if ctx.triggerOneOrMore {
		return game.AbilityContent{}, false
	}
	if !referencesBindTo(ctx.content.References, compiler.ReferenceBindingEventCard, 0) {
		return game.AbilityContent{}, false
	}
	eventCard, ok := lowerCardReference(ctx.content.References[0], referenceLoweringContext{AllowEvent: true})
	if !ok {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Negated || effect.DelayedTiming != 0 {
		return game.AbilityContent{}, false
	}
	switch effect.Kind {
	case compiler.EffectReturn:
		switch effect.ToZone {
		case zone.Hand:
			if (!effect.Exact && !referencesContainKind(ctx.content.References, compiler.ReferenceThatObject)) ||
				ctx.optional {
				return game.AbilityContent{}, false
			}
		case zone.Battlefield:
			if ctx.optional {
				return game.AbilityContent{}, false
			}
		default:
			return game.AbilityContent{}, false
		}
	case compiler.EffectExile:
	case compiler.EffectPut:
		if effect.ToZone != zone.Battlefield || ctx.optional {
			return game.AbilityContent{}, false
		}
	case compiler.EffectCast:
		if effect.FromZone != zone.Graveyard ||
			effect.Duration != compiler.DurationUntilYourNextTurn ||
			!effect.CastAsAdventure ||
			len(ctx.content.References) != 1 {
			return game.AbilityContent{}, false
		}
	default:
		return game.AbilityContent{}, false
	}
	consumed := ctx
	consumed.content.References = nil
	if consumed.content.Unconsumed() {
		return game.AbilityContent{}, false
	}
	switch effect.Kind {
	case compiler.EffectReturn:
		if effect.ToZone == zone.Battlefield {
			put := game.PutOnBattlefield{
				Source:           game.CardBattlefieldSource(eventCard),
				EntryTapped:      effect.EntersTapped,
				EntryTransformed: effect.EntersTransformed,
			}
			if effect.CounterKindKnown {
				// Only a fixed +1/+1 counter placement is representable as an
				// entry-counter rider. Fail closed on any other counter rider so
				// the return is never emitted with the counter silently dropped.
				if effect.CounterKind != counter.PlusOnePlusOne || !effect.Amount.Known || effect.Amount.Value < 1 {
					return game.AbilityContent{}, false
				}
				put.EntryCounters = []game.CounterPlacement{{Kind: counter.PlusOnePlusOne, Amount: effect.Amount.Value}}
			}
			if effect.ReturnAsEnchantment {
				put.ContinuousEffects = []game.ContinuousEffect{{
					Layer:    game.LayerType,
					SetTypes: []types.Card{types.Enchantment},
				}}
			}
			return game.Mode{Sequence: []game.Instruction{{
				Primitive: put,
			}}}.Ability(), true
		}
		return game.Mode{Sequence: []game.Instruction{{
			Primitive: game.MoveCard{
				Card:        eventCard,
				FromZone:    zone.Graveyard,
				Destination: zone.Hand,
			},
		}}}.Ability(), true
	case compiler.EffectCast:
		return game.Mode{Sequence: []game.Instruction{{
			Primitive: game.GrantCastPermission{
				Card:     eventCard,
				FromZone: zone.Graveyard,
				Face:     game.FaceAlternate,
				Duration: game.DurationUntilEndOfYourNextTurn,
			},
		}}}.Ability(), true
	case compiler.EffectPut:
		return game.Mode{Sequence: []game.Instruction{{
			Primitive: game.PutOnBattlefield{
				Source:      game.CardBattlefieldSource(eventCard),
				EntryTapped: effect.EntersTapped,
			},
		}}}.Ability(), true
	case compiler.EffectExile:
		move := game.MoveCard{
			Card:        eventCard,
			FromZone:    zone.Graveyard,
			Destination: zone.Exile,
		}
		if effect.CounterKindKnown {
			move.Counter = opt.Val(effect.CounterKind)
		}
		return game.Mode{Sequence: []game.Instruction{{
			Primitive: move,
		}}}.Ability(), true
	default:
		return game.AbilityContent{}, false
	}
}

// lowerEventCardBatchReanimation lowers a coalesced one-or-more zone-change
// trigger's "put/return them onto the battlefield" clause into a batch
// reanimation of exactly the triggering cards. The plural "them" binds to the
// whole simultaneous batch (ReferenceBindingEventCard under a OneOrMore
// trigger), which the singular CardReferenceEvent of lowerEventCardEffect cannot
// represent, so it emits a MassReturnFromGraveyard restricted to the trigger
// batch (Hedge Shredder).
func lowerEventCardBatchReanimation(ctx contentCtx) (game.AbilityContent, bool) {
	if !ctx.triggerOneOrMore || ctx.triggerToZone != zone.Graveyard || ctx.optional {
		return game.AbilityContent{}, false
	}
	if len(ctx.content.Effects) != 1 {
		return game.AbilityContent{}, false
	}
	if !referencesBindTo(ctx.content.References, compiler.ReferenceBindingEventCard, 0) {
		return game.AbilityContent{}, false
	}
	reference := ctx.content.References[0]
	if reference.Kind != compiler.ReferencePronoun ||
		reference.Pronoun != compiler.ReferencePronounThem {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Negated || effect.DelayedTiming != 0 || effect.ToZone != zone.Battlefield {
		return game.AbilityContent{}, false
	}
	switch effect.Kind {
	case compiler.EffectPut, compiler.EffectReturn:
	default:
		return game.AbilityContent{}, false
	}
	consumed := ctx
	consumed.content.References = nil
	if consumed.content.Unconsumed() {
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.MassReturnFromGraveyard{
			Player:           game.ControllerReference(),
			Destination:      zone.Battlefield,
			EntryTapped:      effect.EntersTapped,
			FromTriggerBatch: true,
		},
	}}}.Ability(), true
}

type enterInterveningCondition struct {
	condition            opt.V[game.Condition]
	hadCounters          bool
	hadNoCounterKind     opt.V[counter.Kind]
	hadCounterKind       opt.V[counter.Kind]
	wasKicked            bool
	wasCast              bool
	wasCastByController  bool
	wasCastFromCtrlHand  bool
	enteredOrCastFromGY  bool
	enteredOrCastFromCGY bool
}

func lowerSelfInterveningCondition(
	eventKind game.EventKind,
	trigger *compiler.CompiledTrigger,
) (enterInterveningCondition, bool) {
	if trigger != nil && trigger.Condition != nil {
		if condition, ok := lowerCondition(*trigger.Condition, conditionContextInterveningTrigger); ok {
			return enterInterveningCondition{condition: opt.Val(condition)}, true
		}
		if trigger.Condition.Predicate == compiler.ConditionPredicateEventSubjectHadCounters {
			if trigger.Condition.ObjectBinding != compiler.ReferenceBindingEventPermanent {
				return enterInterveningCondition{}, false
			}
			return enterInterveningCondition{hadCounters: true}, true
		}
	}
	switch eventKind {
	case game.EventPermanentEnteredBattlefield:
		return lowerEnterInterveningCondition(trigger)
	case game.EventPermanentDied:
		return lowerDiesInterveningCondition(trigger)
	default:
		return enterInterveningCondition{}, trigger == nil || trigger.Condition == nil
	}
}

func supportedSelfTriggerKind(eventKind game.EventKind, kind compiler.TriggerKind) bool {
	switch eventKind {
	case game.EventPermanentEnteredBattlefield,
		game.EventPermanentDied,
		game.EventZoneChanged,
		game.EventPermanentTurnedFaceUp,
		game.EventPermanentSacrificed,
		game.EventObjectBecameTarget,
		// The fragile attacker/blocker idiom "When this creature attacks or
		// blocks, <delayed self-disposal> at end of combat" (Fog Elemental,
		// Cinder Wall) introduces its self attack/block trigger with "When".
		// "When" and "Whenever" are mechanically identical for a triggered
		// ability (CR 603.1); a creature is declared as an attacker or blocker
		// at most once per combat, so the single-shot reading matches.
		game.EventAttackerDeclared,
		game.EventBlockerDeclared:
		return kind == compiler.TriggerWhen || kind == compiler.TriggerWhenever
	case game.EventPermanentMutated,
		game.EventAttackerBecameBlocked,
		game.EventAttackerBecameUnblocked,
		game.EventDamageDealt,
		game.EventPermanentTapped,
		game.EventPermanentUntapped,
		game.EventCountersAdded:
		return kind == compiler.TriggerWhenever
	default:
		return kind == compiler.TriggerWhen
	}
}

func lowerEnterInterveningCondition(trigger *compiler.CompiledTrigger) (enterInterveningCondition, bool) {
	if trigger == nil || trigger.Condition == nil {
		return enterInterveningCondition{}, true
	}
	condition := trigger.Condition
	if condition.Kind != compiler.ConditionIf || !condition.Intervening {
		return enterInterveningCondition{}, false
	}
	switch condition.Predicate {
	case compiler.ConditionPredicateEventSubjectWasKicked:
		return enterInterveningCondition{wasKicked: true}, true
	case compiler.ConditionPredicateEventSubjectWasCast:
		return enterInterveningCondition{wasCast: true}, true
	case compiler.ConditionPredicateEventSubjectWasCastByController:
		return enterInterveningCondition{wasCastByController: true}, true
	case compiler.ConditionPredicateEventSubjectWasCastFromControllerHand:
		return enterInterveningCondition{wasCastFromCtrlHand: true}, true
	case compiler.ConditionPredicateEventSubjectEnteredOrCastFromGraveyard:
		return enterInterveningCondition{enteredOrCastFromGY: true}, true
	case compiler.ConditionPredicateEventSubjectEnteredOrCastFromControllerGraveyard:
		return enterInterveningCondition{enteredOrCastFromCGY: true}, true
	default:
	}
	lowered, ok := lowerCondition(*condition, conditionContextInterveningTrigger)
	if !ok {
		return enterInterveningCondition{}, false
	}
	return enterInterveningCondition{
		condition: opt.Val(lowered),
	}, true
}

func lowerDiesInterveningCondition(trigger *compiler.CompiledTrigger) (enterInterveningCondition, bool) {
	if trigger == nil || trigger.Condition == nil {
		return enterInterveningCondition{}, true
	}
	condition := trigger.Condition
	if condition.Kind != compiler.ConditionIf || !condition.Intervening {
		return enterInterveningCondition{}, false
	}
	return eventSubjectCounterInterveningCondition(condition)
}

// eventSubjectCounterInterveningCondition maps the dying creature's last-known
// counter intervening-if predicates to their lowered form. The negative form
// ("if it had no +1/+1 counters on it") is the Undying/Persist gate; the
// positive form ("if it had a +1/+1 counter on it") is its affirmative
// counterpart. Both read the permanent's last-known information (CR 603.10).
func eventSubjectCounterInterveningCondition(condition *compiler.CompiledCondition) (enterInterveningCondition, bool) {
	had := condition.Predicate == compiler.ConditionPredicateEventSubjectHadCounter
	if condition.Predicate != compiler.ConditionPredicateEventSubjectHadNoCounter && !had {
		return enterInterveningCondition{}, false
	}
	switch condition.Counter {
	case compiler.ConditionCounterPlusOnePlusOne:
		if had {
			return enterInterveningCondition{hadCounterKind: opt.Val(counter.PlusOnePlusOne)}, true
		}
		return enterInterveningCondition{hadNoCounterKind: opt.Val(counter.PlusOnePlusOne)}, true
	case compiler.ConditionCounterMinusOneMinusOne:
		if had {
			return enterInterveningCondition{hadCounterKind: opt.Val(counter.MinusOneMinusOne)}, true
		}
		return enterInterveningCondition{hadNoCounterKind: opt.Val(counter.MinusOneMinusOne)}, true
	default:
		return enterInterveningCondition{}, false
	}
}

func bodyReferences(
	references []compiler.CompiledReference,
	excludedSpans ...shared.Span,
) []compiler.CompiledReference {
	var body []compiler.CompiledReference
	for _, reference := range references {
		if spanCovered(reference.Span, excludedSpans) {
			continue
		}
		body = append(body, reference)
	}
	return body
}

func interveningIfText(trigger *compiler.CompiledTrigger) string {
	if trigger == nil || trigger.Condition == nil {
		return ""
	}
	return trigger.Condition.Text
}

// preparedTriggerBody is the body of a supported triggered ability, ready for
// content lowering. body is the body as a spell-shaped CompiledAbility, syntax
// is the matching parser syntax, and optional is the rendered
// TriggeredAbility.Optional flag (false when only the body's first instruction
// is optional, as in a "you may X. If you do, Y" resolving sequence).
type preparedTriggerBody struct {
	body     compiler.CompiledAbility
	syntax   parser.Ability
	optional bool
}

func exactEventPlayerPaymentCondition(ability compiler.CompiledAbility) bool {
	if ability.Trigger == nil ||
		len(ability.Content.Effects) != 1 ||
		len(ability.Content.Conditions) != 1 ||
		len(ability.Content.References) != 1 {
		return false
	}
	effect := ability.Content.Effects[0]
	condition := ability.Content.Conditions[0]
	reference := ability.Content.References[0]
	payment := effect.Payment
	if payment.Payer != parser.EffectPaymentPayerEventPlayer ||
		len(payment.ManaCost) == 0 ||
		reference.Kind != compiler.ReferenceThatPlayer ||
		reference.Binding != compiler.ReferenceBindingEventPlayer ||
		payment.Span.Start.Offset > reference.Span.Start.Offset ||
		payment.Span.End.Offset < reference.Span.End.Offset {
		return false
	}
	switch payment.Form {
	case parser.EffectPaymentFormUnless:
		return ability.Optional &&
			effect.Optional &&
			ability.OptionalSpan.Start == effect.Span.Start &&
			condition.Kind == compiler.ConditionUnless &&
			condition.Predicate == compiler.ConditionPredicateEventPlayerDoesNotPay &&
			condition.Order.Contains(payment.Order)
	case parser.EffectPaymentFormMayPayThenIfDoesNot:
		return !ability.Optional &&
			!effect.Optional &&
			condition.Kind == compiler.ConditionIf &&
			condition.Predicate == compiler.ConditionPredicateEventPlayerDoesNotPay &&
			condition.NodeID == payment.FailureConditionNodeID &&
			payment.Span.End.Offset < condition.Span.Start.Offset
	default:
		return false
	}
}

func triggerBodySpan(
	effects []compiler.CompiledEffect,
	conditions []compiler.CompiledCondition,
	eventPlayerPaymentCondition bool,
	resolutionCondition bool,
) shared.Span {
	bodySpan := shared.Span{
		Start: effects[0].Span.Start,
		End:   effects[len(effects)-1].Span.End,
	}
	if eventPlayerPaymentCondition {
		paymentStart := effects[0].Payment.Span.Start
		if paymentStart.Offset < bodySpan.Start.Offset {
			bodySpan.Start = paymentStart
		}
	}
	if resolutionCondition {
		// The resolution condition clause may sit before the first effect
		// ("if CONDITION, EFFECT") or after the last ("EFFECT if CONDITION"),
		// so widen the body span to cover every body condition clause.
		for _, condition := range conditions {
			if condition.Span.Start.Offset < bodySpan.Start.Offset {
				bodySpan.Start = condition.Span.Start
			}
			if condition.Span.End.Offset > bodySpan.End.Offset {
				bodySpan.End = condition.Span.End
			}
		}
	}
	// A folded "That token gains <keyword>." copy-token rider sits in a sentence
	// after the create effect, so widen the body span to cover its rider span;
	// otherwise the granted keyword falls outside the body and the keyword-span
	// reconciliation rejects the trigger body.
	for i := range effects {
		if len(effects[i].TokenCopyGrantKeywords) != 0 &&
			effects[i].TokenCopyGrantRiderSpan.End.Offset > bodySpan.End.Offset {
			bodySpan.End = effects[i].TokenCopyGrantRiderSpan.End
		}
	}
	// A folded "choose <keyword> or <keyword> at random." prelude sits in a
	// sentence before the "<source> gains that ability until end of turn." grant,
	// so widen the body span to cover its prelude span; otherwise the listed
	// keywords fall outside the body and the keyword-span reconciliation rejects
	// the trigger body.
	for i := range effects {
		if effects[i].KeywordGrantChoiceAtRandom &&
			effects[i].KeywordChoiceAtRandomPreludeSpan.Start.Offset < bodySpan.Start.Offset {
			bodySpan.Start = effects[i].KeywordChoiceAtRandomPreludeSpan.Start
		}
	}
	return bodySpan
}

// hasMandatoryIfYouDoCondition reports whether the ability carries a residual
// "if you do" gate (prior-instruction-accepted) condition that is distinct from
// the intervening trigger condition. It selects the mandatory result-gate body
// shape ("if CONDITION, exile it. If you do, create a token.") where the leading
// effect is not optional but its success gates the trailing effect.
func hasMandatoryIfYouDoCondition(conditions []compiler.CompiledCondition, interveningSpan shared.Span) bool {
	for ci := range conditions {
		condition := conditions[ci]
		if condition.Predicate == compiler.ConditionPredicatePriorInstructionAccepted &&
			condition.Span != interveningSpan {
			return true
		}
	}
	return false
}

// prepareTriggerBody builds the body CompiledAbility and syntax for a
// supported triggered ability. It handles condition consistency, effect
// filtering for intervening conditions, body span/text construction, reference
// exclusion, and optional "you may" stripping. Callers must have already
// verified that ability.Trigger is non-nil.
func prepareTriggerBody(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (preparedTriggerBody, bool) {
	if ability.Trigger == nil {
		return preparedTriggerBody{}, false
	}
	hasInterveningCondition := ability.Trigger.Condition != nil
	// An optional resolving sequence ("you may X. If you do, Y") carries its
	// "if you do" gate as a body condition. The shared content lowering consumes
	// that gate via the ordered-effect-sequence path and fails closed if it is
	// not exactly the optional-flow gate, so such body conditions may pass
	// through here rather than being rejected outright. The optional marker may
	// land on the whole ability ("Whenever X, you may Y. If you do, Z." parsed
	// with ability.Optional) or on the leading effect itself (the parser instead
	// flags the first effect Optional); either way the body is the same
	// optional-flow sequence, so detection keys on an optional resolving effect
	// rather than on the ability-level flag.
	optionalSequence := !hasInterveningCondition &&
		len(ability.Content.Effects) > 1 && hasOptionalResolvingEffect(ability.Content.Effects)
	// The same optional-flow sequence may also appear behind an intervening "if"
	// condition ("Whenever X, if CONDITION, you may Y. If you do, Z."). The
	// intervening condition is removed from the body (it gates the trigger, not
	// the resolution), and the remaining "if you do" gate stays in the body for
	// the shared ordered-effect-sequence path to consume. The shared lowering
	// fails closed if the residual body conditions are not exactly that gate.
	// The optional flow also covers the single-effect payment shape ("if
	// CONDITION, you may pay COST. If you do, EFFECT."), where paying the cost
	// is not itself a resolving effect: the lone effect carries the optional
	// payment and the residual "if you do" gate, exactly as the non-intervening
	// payment form does, so the same body composition applies.
	interveningOptionalSequence := hasInterveningCondition &&
		((len(ability.Content.Effects) > 1 && hasOptionalResolvingEffect(ability.Content.Effects)) ||
			hasOptionalPaymentResolvingEffect(ability.Content.Effects))
	// A mandatory "if you do" result gate ("Whenever X, if CONDITION, exile it.
	// If you do, create a token.") sits behind the intervening condition exactly
	// like the optional sequence: the intervening condition gates the trigger and
	// is removed from the body, while the residual "if you do" gate stays in the
	// body for the shared ordered-effect-sequence path to consume. The leading
	// effect is mandatory (not "you may"), so it is detected by the residual
	// prior-instruction-accepted gate rather than by an optional resolving effect.
	interveningResultSequence := hasInterveningCondition &&
		len(ability.Content.Effects) > 1 &&
		hasMandatoryIfYouDoCondition(ability.Content.Conditions, ability.Trigger.Condition.Span)
	keepResidualBodyConditions := interveningOptionalSequence || interveningResultSequence
	// A resolution condition ("Whenever X, EFFECT if CONDITION." or "Whenever
	// X, if CONDITION, EFFECT.") is a body condition checked only when the
	// ability resolves, not an intervening "if" re-checked at trigger time. It
	// is not the trigger's own condition, so it stays in the body and routes
	// through the shared content lowering exactly as the same condition does on
	// a spell. The shared lowering fails closed if it cannot lower the
	// condition, so passing it through here is safe. Bodies whose condition is
	// instead an optional-result "if you do" gate are excluded: those route
	// through the optional-sequence path, which composes any non-optional
	// leading effects ahead of the optional resolving pair.
	resolutionCondition := !hasInterveningCondition && !ability.Optional &&
		len(ability.Content.Conditions) != 0 &&
		!hasOptionalResolvingEffect(ability.Content.Effects)
	eventPlayerPaymentCondition := exactEventPlayerPaymentCondition(ability)
	if !optionalSequence && !resolutionCondition && !keepResidualBodyConditions && !eventPlayerPaymentCondition {
		if (len(ability.Content.Conditions) != 0 && !hasInterveningCondition) ||
			(hasInterveningCondition && (len(ability.Content.Conditions) != 1 ||
				ability.Content.Conditions[0].Span != ability.Trigger.Condition.Span)) {
			return preparedTriggerBody{}, false
		}
	}
	resolvingEffects := ability.Content.Effects
	if hasInterveningCondition {
		conditionSpan := []shared.Span{ability.Trigger.Condition.Span}
		resolvingEffects = slices.DeleteFunc(
			append([]compiler.CompiledEffect(nil), ability.Content.Effects...),
			func(effect compiler.CompiledEffect) bool {
				return spanCovered(effect.VerbSpan, conditionSpan)
			},
		)
	}
	if len(resolvingEffects) == 0 {
		return preparedTriggerBody{}, false
	}
	body := ability
	triggerOptional := ability.Optional
	body.Content.Effects = resolvingEffects
	body.Kind = compiler.AbilitySpell
	body.Span = triggerBodySpan(
		resolvingEffects,
		ability.Content.Conditions,
		eventPlayerPaymentCondition,
		resolutionCondition,
	)
	body.Text = titleFirst(
		ability.Text[body.Span.Start.Offset-ability.Span.Start.Offset : body.Span.End.Offset-ability.Span.Start.Offset],
	)
	body.Trigger = nil
	body.Optional = false
	body.OptionalSpan = shared.Span{}
	excludedReferenceSpans := []shared.Span{ability.Trigger.Span}
	if hasInterveningCondition {
		excludedReferenceSpans = append(excludedReferenceSpans, ability.Trigger.Condition.Span)
		if keepResidualBodyConditions {
			// Keep body conditions other than the intervening one (the residual
			// "if you do" optional-flow gate) so the shared ordered-effect
			// sequence can consume it; drop only the intervening condition.
			interveningSpan := []shared.Span{ability.Trigger.Condition.Span}
			body.Content.Conditions = slices.DeleteFunc(
				append([]compiler.CompiledCondition(nil), ability.Content.Conditions...),
				func(condition compiler.CompiledCondition) bool {
					return spanCovered(condition.Span, interveningSpan)
				},
			)
		} else {
			body.Content.Conditions = nil
		}
		bodyStart := slices.IndexFunc(syntax.Tokens, func(token shared.Token) bool {
			return token.Kind != shared.Comma &&
				token.Span.Start.Offset >= ability.Trigger.Condition.Span.End.Offset
		})
		if bodyStart < 0 {
			return preparedTriggerBody{}, false
		}
		effect := body.Content.Effects[0]
		originalSpan := effect.Span
		effect.Span.Start = syntax.Tokens[bodyStart].Span.Start
		effect.Text = ability.Text[effect.Span.Start.Offset-ability.Span.Start.Offset : effect.Span.End.Offset-ability.Span.Start.Offset]
		body.Content.Effects[0] = effect
		// A search/tutor group ("search ..., reveal ..., put ..., then shuffle")
		// models every grouped effect with the same single-sentence span. Moving
		// only the leading effect's start past the intervening condition would
		// desync the group, so move every effect that shared the leading effect's
		// original span to keep the group's same-span invariant intact.
		for i := 1; i < len(body.Content.Effects); i++ {
			if body.Content.Effects[i].Span != originalSpan {
				continue
			}
			grouped := body.Content.Effects[i]
			grouped.Span.Start = effect.Span.Start
			grouped.Text = effect.Text
			body.Content.Effects[i] = grouped
		}
		body.Span.Start = effect.Span.Start
		body.Text = titleFirst(
			ability.Text[body.Span.Start.Offset-ability.Span.Start.Offset : body.Span.End.Offset-ability.Span.Start.Offset],
		)
	}
	body.Content.References = bodyReferences(ability.Content.References, excludedReferenceSpans...)
	bodyTokens := parser.TokensFrom(syntax.Tokens, body.Span.Start.Offset)
	if len(bodyTokens) == 0 {
		return preparedTriggerBody{}, false
	}
	bodySyntax := *syntax
	bodySyntax.Kind = parser.AbilitySpell
	bodySyntax.Tokens = bodyTokens
	if ability.Optional {
		if len(ability.Content.Effects) != 1 {
			if !optionalUntapRemoveFromCombatBody(ability.Content) {
				// A multi-effect optional body ("you may X. If you do, Y") keeps
				// its resolving optionality inside the body so the shared content
				// lowering wires the optional first instruction and result gate.
				// The trigger fires unconditionally; only its first instruction
				// is optional. Intervening-condition bodies are not composed this
				// way.
				if hasInterveningCondition {
					return preparedTriggerBody{}, false
				}
				triggerOptional = false
			}
			// Otherwise the single "you may" governs both coordinated effects:
			// keep the trigger optional and lower two mandatory instructions.
		} else {
			effect := body.Content.Effects[0]
			switch {
			case hasInterveningCondition:
				body.Optional = true
				body.OptionalSpan = ability.OptionalSpan
			case eventPlayerPaymentCondition:
				triggerOptional = false
			case ability.OptionalSpan.Start != effect.Span.Start:
				return preparedTriggerBody{}, false
			default:
				effect.Text = effect.Text[effect.VerbSpan.Start.Offset-effect.Span.Start.Offset:]
				effect.Span.Start = effect.VerbSpan.Start
				effect.Optional = false
				effect.OptionalSpan = shared.Span{}
				body.Content.Effects = []compiler.CompiledEffect{effect}
				body.Span.Start = effect.Span.Start
				body.Text = titleFirst(
					ability.Text[body.Span.Start.Offset-ability.Span.Start.Offset : body.Span.End.Offset-ability.Span.Start.Offset],
				)
				bodyTokens = parser.TokensFrom(bodySyntax.Tokens, effect.VerbSpan.Start.Offset)
				if len(bodyTokens) == 0 {
					return preparedTriggerBody{}, false
				}
				bodySyntax.Tokens = bodyTokens
			}
		}
	}
	body.Content.Keywords = keywordsWithinSpan(ability.Content.Keywords, body.Span)
	if len(body.Content.Keywords) != len(ability.Content.Keywords) {
		return preparedTriggerBody{}, false
	}
	bodySyntax.Span = body.Span
	bodySyntax.Text = body.Text
	return preparedTriggerBody{body: body, syntax: bodySyntax, optional: triggerOptional}, true
}

func lowerPermanentZoneChangeTrigger(
	cardName string,
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.TriggeredAbility, *shared.Diagnostic) {
	const summary = "unsupported permanent zone-change trigger"
	const effectSummary = "unsupported permanent zone-change trigger effect"
	if ability.Trigger == nil {
		return game.TriggeredAbility{}, executableDiagnostic(ability, summary,
			"the executable source backend requires a semantic permanent zone-change trigger")
	}
	pattern, ok := lowerTriggerPattern(&ability.Trigger.Pattern)
	if !ok ||
		(pattern.Event != game.EventPermanentEnteredBattlefield &&
			pattern.Event != game.EventPermanentDied &&
			pattern.Event != game.EventZoneChanged) {
		return game.TriggeredAbility{}, executableDiagnostic(ability, summary,
			"the executable source backend does not support this semantic permanent zone-change trigger pattern")
	}
	triggerType, ok := lowerTriggerKind(ability.Trigger.Pattern.Kind)
	if !ok || (triggerType != game.TriggerWhen && triggerType != game.TriggerWhenever) {
		return game.TriggeredAbility{}, executableDiagnostic(ability, summary,
			"permanent zone-change triggers require When or Whenever")
	}
	intervening, ok := lowerPermanentZoneChangeInterveningCondition(&pattern, ability.Trigger)
	if !ok {
		return game.TriggeredAbility{}, executableDiagnostic(ability, summary,
			"the executable source backend does not support this semantic permanent zone-change trigger condition")
	}
	// Enter, dies, and zone-change bodies all lower through the same shared
	// content path, so they are gated identically here. The ability-word label
	// (e.g. "Chainsword —" or "Landfall —") is purely cosmetic (CR 207.2c) and
	// is excluded from the body span by lowerTriggeredAbilityKind, so any label
	// — whitelisted or not — passes through without affecting the lowered body.
	// A modal "choose one —" body routes through the shared modal-content
	// lowering, exactly like spell-cast triggers; non-modal mode lists and empty
	// effect lists remain unsupported and fail closed.
	if modalTriggerBody(ability) {
		content, diagnostic := lowerModalTriggerBody(cardName, ability, syntax, pattern.Event)
		if diagnostic != nil {
			return game.TriggeredAbility{}, diagnostic
		}
		return permanentZoneChangeTriggeredAbility(ability, ability.Optional, triggerType, &pattern, &intervening, content), nil
	}
	if len(ability.Content.Effects) == 0 || len(ability.Content.Modes) != 0 {
		return game.TriggeredAbility{}, executableDiagnostic(ability, effectSummary,
			"the executable source backend does not support this permanent zone-change trigger body")
	}
	prepared, ok := prepareTriggerBody(ability, syntax)
	if !ok {
		return game.TriggeredAbility{}, executableDiagnostic(ability, effectSummary,
			"the executable source backend does not support this permanent zone-change trigger body")
	}
	body, bodySyntax, triggerOptional := prepared.body, prepared.syntax, prepared.optional
	content, diagnostic := lowerTriggerBodyContent(cardName, body.Content, body.Optional, &bodySyntax, pattern)
	if diagnostic != nil {
		return game.TriggeredAbility{}, diagnostic
	}
	return permanentZoneChangeTriggeredAbility(ability, triggerOptional, triggerType, &pattern, &intervening, content), nil
}

func lowerPermanentZoneChangeInterveningCondition(
	pattern *game.TriggerPattern,
	trigger *compiler.CompiledTrigger,
) (enterInterveningCondition, bool) {
	if pattern.Source == game.TriggerSourceSelf {
		return lowerSelfInterveningCondition(pattern.Event, trigger)
	}
	if trigger != nil && trigger.Condition != nil {
		switch trigger.Condition.Predicate {
		case compiler.ConditionPredicateObjectMatches,
			compiler.ConditionPredicateObjectExists,
			compiler.ConditionPredicateEventSubjectNameUnique:
			if condition, ok := lowerCondition(*trigger.Condition, conditionContextInterveningTrigger); ok {
				return enterInterveningCondition{condition: opt.Val(condition)}, true
			}
		default:
		}
		if trigger.Condition.Predicate == compiler.ConditionPredicateEventSubjectHadCounters {
			if trigger.Condition.ObjectBinding != compiler.ReferenceBindingEventPermanent {
				return enterInterveningCondition{}, false
			}
			return enterInterveningCondition{hadCounters: true}, true
		}
		if pattern.Event == game.EventPermanentDied {
			if intervening, ok := eventSubjectCounterInterveningCondition(trigger.Condition); ok {
				return intervening, true
			}
		}
	}
	if pattern.Event == game.EventPermanentEnteredBattlefield {
		intervening, ok := lowerEnterInterveningCondition(trigger)
		if !ok ||
			(trigger.Condition != nil &&
				trigger.Condition.Predicate == compiler.ConditionPredicateEventSubjectWasCastByController) {
			return enterInterveningCondition{}, false
		}
		return intervening, true
	}
	return enterInterveningCondition{}, trigger == nil || trigger.Condition == nil
}

func permanentZoneChangeTriggeredAbility(
	ability compiler.CompiledAbility,
	triggerOptional bool,
	triggerType game.TriggerType,
	pattern *game.TriggerPattern,
	intervening *enterInterveningCondition,
	content game.AbilityContent,
) game.TriggeredAbility {
	return game.TriggeredAbility{
		Text: ability.Text,
		Trigger: game.TriggerCondition{
			Type:                                   triggerType,
			Pattern:                                *pattern,
			InterveningIf:                          interveningIfText(ability.Trigger),
			InterveningCondition:                   intervening.condition,
			InterveningIfEventPermanentHadCounters: intervening.hadCounters,
			InterveningIfEventPermanentHadNoCounterKind:                     intervening.hadNoCounterKind,
			InterveningIfEventPermanentHadCounterKind:                       intervening.hadCounterKind,
			InterveningIfEventPermanentWasKicked:                            intervening.wasKicked,
			InterveningIfEventPermanentWasCast:                              intervening.wasCast,
			InterveningIfEventPermanentWasCastByController:                  intervening.wasCastByController,
			InterveningIfEventPermanentWasCastFromControllerHand:            intervening.wasCastFromCtrlHand,
			InterveningIfEventPermanentEnteredOrCastFromGraveyard:           intervening.enteredOrCastFromGY,
			InterveningIfEventPermanentEnteredOrCastFromControllerGraveyard: intervening.enteredOrCastFromCGY,
		},
		Optional: triggerOptional,
		Content:  content,
	}
}

// lowerCastTrigger lowers a recognized semantic spell-cast trigger into a
// game.TriggeredAbility with EventSpellCast.
func lowerCastTrigger(
	cardName string,
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.TriggeredAbility, *shared.Diagnostic) {
	if ability.Trigger == nil {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability",
			"the executable source backend requires a semantic whenever spell-cast trigger")
	}
	selfCast := ability.Trigger.Pattern.SelfWasCast
	triggerType := game.TriggerWhenever
	switch ability.Trigger.Pattern.Kind {
	case compiler.TriggerWhenever:
		if selfCast {
			return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability",
				"the executable source backend requires a when self-cast spell trigger")
		}
	case compiler.TriggerWhen:
		// "When you cast this spell" is the only spell-cast trigger introduced by
		// "When"; every other spell-cast trigger uses "Whenever".
		if !selfCast {
			return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability",
				"the executable source backend requires a semantic whenever spell-cast trigger")
		}
		triggerType = game.TriggerWhen
	default:
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability",
			"the executable source backend requires a semantic whenever spell-cast trigger")
	}

	pattern, ok := lowerTriggerPattern(&ability.Trigger.Pattern)
	if !ok || pattern.Event != game.EventSpellCast {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability",
			"the executable source backend does not support this semantic spell-cast trigger pattern")
	}
	intervening, ok := lowerAtInterveningCondition(ability.Trigger)
	if !ok {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability",
			"the executable source backend does not support this semantic spell-cast trigger condition")
	}
	if modalTriggerBody(ability) {
		content, diagnostic := lowerModalTriggerBody(cardName, ability, syntax, pattern.Event)
		if diagnostic != nil {
			return game.TriggeredAbility{}, diagnostic
		}
		return game.TriggeredAbility{
			Text: ability.Text,
			Trigger: game.TriggerCondition{
				Type:                 triggerType,
				Pattern:              pattern,
				InterveningIf:        interveningIfText(ability.Trigger),
				InterveningCondition: intervening,
			},
			Optional: ability.Optional,
			Content:  content,
		}, nil
	}
	if triggerContentUnsupported(ability) {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability effect",
			"the executable source backend does not support this spell-cast trigger body")
	}

	prepared, ok := prepareTriggerBody(ability, syntax)
	if !ok {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability effect",
			"the executable source backend does not support this spell-cast trigger body")
	}
	body, bodySyntax, triggerOptional := prepared.body, prepared.syntax, prepared.optional
	content, diagnostic := lowerTriggerBodyContent(
		cardName,
		body.Content,
		body.Optional,
		&bodySyntax,
		pattern,
	)
	if diagnostic != nil {
		return game.TriggeredAbility{}, diagnostic
	}

	return game.TriggeredAbility{
		Text: ability.Text,
		Trigger: game.TriggerCondition{
			Type:                 triggerType,
			Pattern:              pattern,
			InterveningIf:        interveningIfText(ability.Trigger),
			InterveningCondition: intervening,
		},
		Optional: triggerOptional,
		Content:  content,
	}, nil
}

func spanCovered(span shared.Span, covering []shared.Span) bool {
	for _, candidate := range covering {
		if candidate.Start.Offset <= span.Start.Offset &&
			candidate.End.Offset >= span.End.Offset {
			return true
		}
	}
	return false
}
