package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
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
	content, diagnostic := lowerTriggerBodyContent(cardName, body.Content, body.Optional, &bodySyntax, pattern.Event)
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
			InterveningIfEventPermanentHadNoCounterKind: intervening.hadNoCounterKind,
			InterveningIfEventPermanentWasKicked:        intervening.wasKicked,
			InterveningIfEventPermanentWasCast:          intervening.wasCast,
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
		pattern.Event,
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
		if (!effect.Exact && !referencesContainKind(ctx.content.References, compiler.ReferenceThatObject)) ||
			effect.ToZone != zone.Hand || ctx.optional {
			return game.AbilityContent{}, false
		}
	case compiler.EffectExile:
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
	case compiler.EffectExile:
		return game.Mode{Sequence: []game.Instruction{{
			Primitive: game.MoveCard{
				Card:        eventCard,
				FromZone:    zone.Graveyard,
				Destination: zone.Exile,
			},
		}}}.Ability(), true
	default:
		return game.AbilityContent{}, false
	}
}

type enterInterveningCondition struct {
	condition        opt.V[game.Condition]
	hadCounters      bool
	hadNoCounterKind opt.V[counter.Kind]
	wasKicked        bool
	wasCast          bool
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
		game.EventObjectBecameTarget:
		return kind == compiler.TriggerWhen || kind == compiler.TriggerWhenever
	case game.EventPermanentMutated,
		game.EventAttackerBecameBlocked,
		game.EventAttackerDeclared,
		game.EventBlockerDeclared,
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
		return enterInterveningCondition{}, false
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
	if condition.Predicate != compiler.ConditionPredicateEventSubjectHadNoCounter {
		return enterInterveningCondition{}, false
	}
	switch condition.Counter {
	case compiler.ConditionCounterPlusOnePlusOne:
		return enterInterveningCondition{hadNoCounterKind: opt.Val(counter.PlusOnePlusOne)}, true
	case compiler.ConditionCounterMinusOneMinusOne:
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
	return ability.Optional &&
		effect.Optional &&
		ability.OptionalSpan.Start == effect.Span.Start &&
		payment.Payer == parser.EffectPaymentPayerEventPlayer &&
		len(payment.ManaCost) != 0 &&
		condition.Kind == compiler.ConditionUnless &&
		condition.Predicate == compiler.ConditionPredicateEventPlayerDoesNotPay &&
		condition.Order.Contains(payment.Order) &&
		reference.Kind == compiler.ReferenceThatPlayer &&
		reference.Binding == compiler.ReferenceBindingEventPlayer &&
		payment.Span.Start.Offset <= reference.Span.Start.Offset &&
		payment.Span.End.Offset >= reference.Span.End.Offset
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
	interveningOptionalSequence := hasInterveningCondition &&
		len(ability.Content.Effects) > 1 && hasOptionalResolvingEffect(ability.Content.Effects)
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
	if !optionalSequence && !resolutionCondition && !interveningOptionalSequence && !eventPlayerPaymentCondition {
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
	bodySpanStart := resolvingEffects[0].Span.Start
	bodySpanEnd := resolvingEffects[len(resolvingEffects)-1].Span.End
	if resolutionCondition {
		// The resolution condition clause may sit before the first effect
		// ("if CONDITION, EFFECT") or after the last ("EFFECT if CONDITION"),
		// so widen the body span to cover every body condition clause.
		for _, condition := range ability.Content.Conditions {
			if condition.Span.Start.Offset < bodySpanStart.Offset {
				bodySpanStart = condition.Span.Start
			}
			if condition.Span.End.Offset > bodySpanEnd.Offset {
				bodySpanEnd = condition.Span.End
			}
		}
	}
	body.Span = shared.Span{Start: bodySpanStart, End: bodySpanEnd}
	body.Text = titleFirst(
		ability.Text[body.Span.Start.Offset-ability.Span.Start.Offset : body.Span.End.Offset-ability.Span.Start.Offset],
	)
	body.Trigger = nil
	body.Optional = false
	body.OptionalSpan = shared.Span{}
	excludedReferenceSpans := []shared.Span{ability.Trigger.Span}
	if hasInterveningCondition {
		excludedReferenceSpans = append(excludedReferenceSpans, ability.Trigger.Condition.Span)
		if interveningOptionalSequence {
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
		effect.Span.Start = syntax.Tokens[bodyStart].Span.Start
		effect.Text = ability.Text[effect.Span.Start.Offset-ability.Span.Start.Offset : effect.Span.End.Offset-ability.Span.Start.Offset]
		body.Content.Effects[0] = effect
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
			// A multi-effect optional body ("you may X. If you do, Y") keeps its
			// resolving optionality inside the body so the shared content
			// lowering wires the optional first instruction and its result gate.
			// The trigger fires unconditionally; only its first instruction is
			// optional. Intervening-condition bodies are not composed this way.
			if hasInterveningCondition {
				return preparedTriggerBody{}, false
			}
			triggerOptional = false
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
	// (e.g. "Chainsword —") is purely cosmetic (CR 207.2c) and is excluded from
	// the body span by lowerTriggeredAbilityKind, so any label — whitelisted or
	// not — passes through without affecting the lowered body. Modes and empty
	// effect lists remain unsupported and fail closed.
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
	content, diagnostic := lowerTriggerBodyContent(cardName, body.Content, body.Optional, &bodySyntax, pattern.Event)
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
		case compiler.ConditionPredicateObjectMatches, compiler.ConditionPredicateObjectExists:
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
			InterveningIfEventPermanentHadNoCounterKind: intervening.hadNoCounterKind,
			InterveningIfEventPermanentWasKicked:        intervening.wasKicked,
			InterveningIfEventPermanentWasCast:          intervening.wasCast,
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
	if ability.Trigger == nil ||
		ability.Trigger.Pattern.Kind != compiler.TriggerWhenever {
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
		pattern.Event,
	)
	if diagnostic != nil {
		return game.TriggeredAbility{}, diagnostic
	}

	return game.TriggeredAbility{
		Text: ability.Text,
		Trigger: game.TriggerCondition{
			Type:                 game.TriggerWhenever,
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
