package cardgen

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// triggerContentUnsupported reports whether a triggered ability's top-level
// content shape cannot route through the shared trigger-body lowering. The
// ability-word label is intentionally
// not gated: an ability word printed before a When/Whenever/At trigger is
// always rules-free flavor (rule 207.2c). Keyword abilities that carry rules
// meaning (Boast, Exhaust, Cohort, Renew, ...) are printed before an activation
// cost, never before a trigger word, so any label preceding a trigger is safe
// to ignore. The ability-word source span is still covered for completeness by
// lowerTriggeredAbilityKind, which spans the label-to-trigger region.
func triggerContentUnsupported(ability compiler.CompiledAbility) bool {
	return len(ability.Content.Modes) != 0 && !exactConnectionModeLabels(ability.Content.Modes)
}

// modalTriggerBody reports whether a triggered ability's resolving body is a
// generic modal "choose one/up to one —" choice that routes through the shared
// modal-content lowering rather than the single-effect trigger-body path. The
// exact labeled "first main phase" modal (Black Market Connections) keeps its
// dedicated handling, so it is excluded here.
func modalTriggerBody(ability compiler.CompiledAbility) bool {
	return len(ability.Content.Modes) != 0 && !exactConnectionModeLabels(ability.Content.Modes)
}

// lowerModalTriggerBody lowers a triggered ability whose resolving body is a
// modal choice. Each mode lowers as an independent already-supported effect; the
// runtime presents the modes when the ability resolves. The shared modal-content
// lowering fails closed on intervening/resolution conditions or effects shared
// across modes, so only a bare modal trigger body is accepted.
func lowerModalTriggerBody(
	cardName string,
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
	pattern game.TriggerPattern,
) (game.AbilityContent, *shared.Diagnostic) {
	ctx := contentCtx{
		text:                    ability.Text,
		span:                    ability.Span,
		content:                 ability.Content,
		enclosingKind:           compiler.AbilityTriggered,
		triggerCardCountEvent:   pattern.Event,
		triggerEvent:            pattern.Event,
		triggerSubjectSelection: pattern.SubjectSelection,
	}
	return lowerModalContent(cardName, ctx, syntax)
}

func lowerAtTrigger(
	cardName string,
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.TriggeredAbility, *shared.Diagnostic) {
	const summary = "unsupported phase/step trigger phrase"
	if ability.Trigger == nil {
		return game.TriggeredAbility{}, executableDiagnostic(
			ability,
			summary,
			"the executable source backend requires a semantic step trigger pattern",
		)
	}
	pattern, ok := lowerTriggerPattern(&ability.Trigger.Pattern)
	if !ok || pattern.Event != game.EventBeginningOfStep {
		_, detail := triggerPatternCapabilityDiagnostic(ability.Trigger)
		return game.TriggeredAbility{}, executableDiagnostic(
			ability,
			summary,
			detail,
		)
	}
	intervening, ok := lowerAtInterveningCondition(ability.Trigger)
	if !ok {
		return game.TriggeredAbility{}, executableDiagnostic(
			ability,
			summary,
			"the executable source backend does not support this intervening-if condition",
		)
	}
	if ability.KeywordShare != nil {
		return lowerKeywordShareTrigger(ability, pattern)
	}
	if ability.ExactSequence == compiler.ExactSequenceChosenTypeLibraryTopToHand {
		return lowerChosenTypeLibraryTopTrigger(ability, &pattern, intervening)
	}
	if ability.ExactSequence == compiler.ExactSequencePayHandSizeOrCantAttack {
		return lowerPayHandSizeOrCantAttackTrigger(ability, &pattern, intervening)
	}
	if ability.ExactSequence == compiler.ExactSequenceExtraDrawThenPayLifeOrTop {
		return lowerExtraDrawPayLifeOrTopTrigger(ability, &pattern, intervening)
	}
	if ability.ExactSequence == compiler.ExactSequenceSharedTypeSacrificePunisher {
		return lowerSharedTypeSacrificePunisherTrigger(ability, &pattern, intervening)
	}
	if triggerContentUnsupported(ability) {
		return game.TriggeredAbility{}, executableDiagnostic(
			ability,
			"unsupported phase/step trigger phrase effect",
			"modes and ability words are not supported in phase/step triggers",
		)
	}
	if len(ability.Content.Modes) != 0 {
		if !reflect.DeepEqual(pattern, game.TriggerPattern{
			Event:      game.EventBeginningOfStep,
			Controller: game.TriggerControllerYou,
			Step:       game.StepPrecombatMain,
		}) {
			return game.TriggeredAbility{}, executableDiagnostic(
				ability,
				"unsupported phase/step trigger phrase effect",
				"the exact labeled modal body is supported only at the beginning of its controller's first main phase",
			)
		}
		if ability.Optional || ability.Trigger.Condition != nil ||
			len(ability.Content.Effects) != 0 ||
			len(ability.Content.Targets) != 0 ||
			len(ability.Content.Keywords) != 0 ||
			len(ability.Content.Conditions) != 0 ||
			len(ability.Content.References) != 0 {
			return game.TriggeredAbility{}, executableDiagnostic(
				ability,
				"unsupported phase/step trigger phrase effect",
				"the executable source backend supports only the exact labeled modal trigger body",
			)
		}
		content, diagnostic := lowerAbilityContent(cardName, compiler.AbilityTriggered, ability.Content, false, syntax)
		if diagnostic != nil {
			return game.TriggeredAbility{}, diagnostic
		}
		return game.TriggeredAbility{
			Text: ability.Text,
			Trigger: game.TriggerCondition{
				Type:                 game.TriggerAt,
				Pattern:              pattern,
				InterveningIf:        interveningIfText(ability.Trigger),
				InterveningCondition: intervening,
			},
			Content: content,
		}, nil
	}
	prepared, ok := prepareTriggerBody(ability, syntax)
	if !ok {
		return game.TriggeredAbility{}, executableDiagnostic(
			ability,
			"unsupported phase/step trigger phrase effect",
			"the executable source backend does not support this phase/step trigger body",
		)
	}
	body, bodySyntax, triggerOptional := prepared.body, prepared.syntax, prepared.optional
	content, diagnostic := lowerAbilityContent(cardName, compiler.AbilityTriggered, body.Content, body.Optional, &bodySyntax)
	if diagnostic != nil {
		return game.TriggeredAbility{}, diagnostic
	}
	return game.TriggeredAbility{
		Text: ability.Text,
		Trigger: game.TriggerCondition{
			Type:                 game.TriggerAt,
			Pattern:              pattern,
			InterveningIf:        interveningIfText(ability.Trigger),
			InterveningCondition: intervening,
		},
		Optional: triggerOptional,
		Content:  content,
	}, nil
}

func lowerAtInterveningCondition(trigger *compiler.CompiledTrigger) (opt.V[game.Condition], bool) {
	if trigger == nil || trigger.Condition == nil {
		return opt.V[game.Condition]{}, true
	}
	condition := *trigger.Condition
	if lowered, ok := lowerCondition(condition, conditionContextInterveningTrigger); ok {
		return opt.Val(lowered), true
	}
	return opt.V[game.Condition]{}, false
}

func lowerTriggeredAbility(
	cardName string,
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.TriggeredAbility, *shared.Diagnostic) {
	if ability.Trigger == nil {
		return game.TriggeredAbility{}, executableDiagnostic(
			ability,
			"unsupported triggered ability",
			"the executable source backend requires a semantic trigger pattern",
		)
	}
	pattern := ability.Trigger.Pattern
	if ability.ExactSequence == compiler.ExactSequenceConditionalLookAtTopReveal {
		return lowerConditionalLookAtTopTrigger(ability)
	}
	if ability.ExactSequence == compiler.ExactSequenceConditionalLookAtTopBattlefield {
		return lowerConditionalLookAtTopBattlefieldTrigger(ability)
	}
	if ability.ExactSequence == compiler.ExactSequenceDevotionLookWin {
		return lowerDevotionLookWinTrigger(ability)
	}
	if ability.ExactSequence == compiler.ExactSequenceDrawPutLandSubtypeLife {
		return lowerDrawPutLandSubtypeLifeTrigger(ability)
	}
	if pattern.Kind == compiler.TriggerAt {
		return lowerAtTrigger(cardName, ability, syntax)
	}
	if pattern.Kind == compiler.TriggerState {
		return lowerStateTrigger(cardName, ability, syntax)
	}
	switch pattern.Event {
	case compiler.TriggerEventCardDrawn, compiler.TriggerEventCardDiscarded, compiler.TriggerEventCycled:
		return lowerDrawDiscardTrigger(cardName, ability, syntax)
	case compiler.TriggerEventLifeGained, compiler.TriggerEventLifeLost, compiler.TriggerEventDamageDealt:
		return lowerLifeDamageTrigger(cardName, ability, syntax)
	case compiler.TriggerEventPermanentEnteredBattlefield,
		compiler.TriggerEventPermanentDied,
		compiler.TriggerEventZoneChanged,
		compiler.TriggerEventDoorUnlocked:
		return lowerPermanentZoneChangeTrigger(cardName, ability, syntax)
	case compiler.TriggerEventSpellCast:
		return lowerCastTrigger(cardName, ability, syntax)
	default:
		if pattern.Source == compiler.TriggerSourceSelf {
			return lowerEnterTrigger(cardName, ability, syntax)
		}
		return lowerGenericPatternTrigger(cardName, ability, syntax)
	}
}

func lowerDrawDiscardTrigger(
	cardName string,
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.TriggeredAbility, *shared.Diagnostic) {
	const summary = "unsupported draw/discard trigger"
	const effectSummary = "unsupported draw/discard trigger effect"
	if ability.Trigger == nil {
		return game.TriggeredAbility{}, executableDiagnostic(ability, summary,
			"the executable source backend requires a semantic trigger pattern")
	}
	// Draw and discard triggers are recurring ("Whenever you ..."). The
	// self-source cycle trigger "When you cycle this card" fires once from the
	// cycled card itself (CR 702.29e) and is the only one introduced by "When".
	kind := ability.Trigger.Pattern.Kind
	selfCycle := ability.Trigger.Pattern.Source == compiler.TriggerSourceSelf &&
		ability.Trigger.Pattern.Event == compiler.TriggerEventCycled
	if kind != compiler.TriggerWhenever && (kind != compiler.TriggerWhen || !selfCycle) {
		return game.TriggeredAbility{}, executableDiagnostic(ability, summary,
			"the executable source backend supports only TriggerWhenever draw and discard triggers")
	}
	triggerType, ok := lowerTriggerKind(kind)
	if !ok {
		return game.TriggeredAbility{}, executableDiagnostic(ability, summary,
			"the executable source backend does not support this semantic trigger kind")
	}
	pattern, ok := lowerTriggerPattern(&ability.Trigger.Pattern)
	if !ok ||
		(pattern.Event != game.EventCardDrawn &&
			pattern.Event != game.EventCardDiscarded &&
			pattern.Event != game.EventCycled) {
		return game.TriggeredAbility{}, executableDiagnostic(ability, summary,
			"unrecognized semantic draw, discard, or cycling trigger pattern")
	}
	intervening, ok := lowerAtInterveningCondition(ability.Trigger)
	if !ok {
		return game.TriggeredAbility{}, executableDiagnostic(ability, summary,
			"the executable source backend does not support this semantic draw/discard trigger condition")
	}
	if triggerContentUnsupported(ability) {
		return game.TriggeredAbility{}, executableDiagnostic(ability, effectSummary,
			"the executable source backend does not support this draw/discard trigger body")
	}
	prepared, ok := prepareTriggerBody(ability, syntax)
	if !ok {
		return game.TriggeredAbility{}, executableDiagnostic(ability, effectSummary,
			"the executable source backend does not support this draw/discard trigger body")
	}
	body, bodySyntax, triggerOptional := prepared.body, prepared.syntax, prepared.optional
	content, diagnostic := lowerTriggerBodyContent(cardName, body.Content, body.Optional, &bodySyntax, pattern)
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

func lowerStateTrigger(
	cardName string,
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.TriggeredAbility, *shared.Diagnostic) {
	const summary = "unsupported state-triggered ability"
	if ability.Trigger == nil || ability.Trigger.Pattern.StateCondition == nil {
		return game.TriggeredAbility{}, executableDiagnostic(ability, summary,
			"the executable source backend requires a semantic state-trigger condition")
	}
	// A state trigger carries no event pattern, only a board-state condition. It
	// takes no intervening "if" and no modal, per-turn, or optional body: the
	// only supported shape is a bare "When <state>, <effect>." clause.
	if ability.Trigger.Condition != nil ||
		ability.Trigger.MaxTriggersPerTurn != 0 ||
		modalTriggerBody(ability) {
		return game.TriggeredAbility{}, executableDiagnostic(ability, summary,
			"the executable source backend supports only a bare state trigger with no intervening, modal, or per-turn qualifier")
	}
	condition, ok := lowerCondition(*ability.Trigger.Pattern.StateCondition, conditionContextStateTrigger)
	if !ok {
		return game.TriggeredAbility{}, executableDiagnostic(ability, summary,
			"the executable source backend does not support this state-trigger condition")
	}
	if triggerContentUnsupported(ability) {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported state-triggered ability effect",
			"the executable source backend does not support this state-trigger body")
	}
	prepared, ok := prepareTriggerBody(ability, syntax)
	if !ok {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported state-triggered ability effect",
			"the executable source backend does not support this state-trigger body")
	}
	body, bodySyntax, triggerOptional := prepared.body, prepared.syntax, prepared.optional
	content, diagnostic := lowerAbilityContent(cardName, compiler.AbilityTriggered, body.Content, body.Optional, &bodySyntax)
	if diagnostic != nil {
		return game.TriggeredAbility{}, diagnostic
	}
	return game.TriggeredAbility{
		Text: ability.Text,
		Trigger: game.TriggerCondition{
			Type:  game.TriggerState,
			State: opt.Val(game.StateTriggerCondition{Condition: opt.Val(condition)}),
		},
		Optional: triggerOptional,
		Content:  content,
	}, nil
}

func lowerGenericPatternTrigger(
	cardName string,
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.TriggeredAbility, *shared.Diagnostic) {
	if ability.Trigger == nil {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability",
			"the executable source backend requires a semantic trigger pattern")
	}
	pattern, ok := lowerTriggerPattern(&ability.Trigger.Pattern)
	if !ok {
		if ability.Trigger.Pattern.OneOrMore {
			if diagnostic := triggerBodyDiagnostic(cardName, ability, syntax); diagnostic != nil {
				return game.TriggeredAbility{}, diagnostic
			}
		}
		summary, detail := triggerPatternCapabilityDiagnostic(ability.Trigger)
		return game.TriggeredAbility{}, executableDiagnostic(ability, summary, detail)
	}
	triggerType, ok := lowerTriggerKind(ability.Trigger.Pattern.Kind)
	if !ok || triggerType == game.TriggerAt {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability",
			"the executable source backend does not support this semantic trigger kind")
	}

	intervening, ok := lowerAtInterveningCondition(ability.Trigger)
	if !ok {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability",
			"the executable source backend does not support this semantic trigger condition")
	}
	if modalTriggerBody(ability) {
		content, diagnostic := lowerModalTriggerBody(cardName, ability, syntax, pattern)
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
			"the executable source backend does not support this trigger body")
	}
	prepared, ok := prepareTriggerBody(ability, syntax)
	if !ok {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability effect",
			"the executable source backend does not support this trigger body")
	}
	body, bodySyntax, triggerOptional := prepared.body, prepared.syntax, prepared.optional
	var content game.AbilityContent
	var diagnostic *shared.Diagnostic
	if groupContent, ok := lowerAttackGroupKeywordGrant(body.Content, pattern, body.Optional); ok {
		content = groupContent
	} else if pattern.Event == game.EventCountersAdded || attackDefendingPlayerEvent(pattern.Event) ||
		pattern.RequireProducedManaColorFromEntryChoice {
		content, diagnostic = lowerTriggerBodyContent(cardName, body.Content, body.Optional, &bodySyntax, pattern)
	} else {
		content, diagnostic = lowerAbilityContent(cardName, compiler.AbilityTriggered, body.Content, body.Optional, &bodySyntax)
	}
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

func triggerBodyDiagnostic(cardName string, ability compiler.CompiledAbility, syntax *parser.Ability) *shared.Diagnostic {
	if triggerContentUnsupported(ability) {
		return nil
	}
	prepared, ok := prepareTriggerBody(ability, syntax)
	if !ok {
		return nil
	}
	body, bodySyntax := prepared.body, prepared.syntax
	_, diagnostic := lowerAbilityContent(cardName, compiler.AbilityTriggered, body.Content, body.Optional, &bodySyntax)
	return diagnostic
}

// zoneChangeTriggerPatternDetail is the diagnostic detail shared by the
// unsupported permanent zone-change trigger patterns. It is named so the
// summary selection in triggerPatternCapabilityDiagnostic does not duplicate
// the literal.
const zoneChangeTriggerPatternDetail = "the executable source backend does not support this semantic permanent zone-change trigger pattern"

// triggerPatternCapabilityDiagnostic returns the (summary, detail) describing
// why a semantic Trigger Pattern has no runtime lowering adapter. It runs only
// after typed lowering (lowerTriggerPattern) has already failed closed, so the
// card is unsupported regardless of this message and no supported/unsupported
// outcome or generated behavior depends on it. The retained read of the raw
// event-clause text inside triggerPatternCapabilityDetail is therefore a
// diagnostic-only refinement that gives more specific feedback for event
// families the compiler does not yet recognize into typed data; it never gates
// behavior. The summary is derived here rather than by string-comparing the
// detail in lowering callers.
func triggerPatternCapabilityDiagnostic(trigger *compiler.CompiledTrigger) (summary, detail string) {
	detail = triggerPatternCapabilityDetail(trigger)
	if detail == zoneChangeTriggerPatternDetail {
		return "unsupported permanent zone-change trigger", detail
	}
	return "unsupported triggered ability", detail
}

func triggerPatternCapabilityDetail(trigger *compiler.CompiledTrigger) string {
	if trigger == nil {
		return "the trigger shell is missing a semantic Trigger Pattern"
	}
	if trigger.Pattern.Event == compiler.TriggerEventAbilityActivated && !trigger.Pattern.ExcludeManaAbility {
		return "the runtime ability-activated event stream omits payment-time mana abilities, so unrestricted activated-ability triggers require a missing runtime capability"
	}
	if trigger.Pattern.Event != compiler.TriggerEventUnknown {
		return "the semantic Trigger Pattern contains a field with no runtime lowering adapter"
	}
	event := strings.ToLower(trigger.Event)
	for _, boundary := range []string{
		"declare attackers step",
		"declare blockers step",
		"first strike damage step",
		"combat damage step",
		"cleanup step",
	} {
		if strings.Contains(event, boundary) {
			return fmt.Sprintf("the runtime does not emit a beginning-of-%s event", boundary)
		}
	}
	if strings.Contains(event, " dies") && strings.Contains(event, "blocking this") {
		return zoneChangeTriggerPatternDetail
	}
	switch event {
	case "an enchanted creature dies",
		"an equipped creature you control dies":
		return zoneChangeTriggerPatternDetail
	case "a renowned creature you control deals combat damage to a player",
		"an enchanted creature you control deals combat damage to a player",
		"a goaded creature deals combat damage to one of your opponents",
		"a noncreature source you control deals damage":
		return "the executable source backend does not support this semantic life or damage trigger pattern"
	case "an enchanted creature attacks one of your opponents",
		"a goaded creature attacks",
		"one or more suspected creatures you control attack":
		return "the semantic Trigger Pattern contains a field with no runtime lowering adapter"
	}
	if strings.Contains(event, "attack") ||
		strings.Contains(event, "block") ||
		strings.Contains(event, "damage") ||
		strings.Contains(event, "combat") ||
		strings.Contains(event, "upkeep") ||
		strings.Contains(event, "draw step") ||
		strings.Contains(event, "end step") ||
		strings.Contains(event, "main phase") {
		return "the runtime event exists, but this combat, phase, or step relation requires a missing runtime capability"
	}
	if strings.Contains(event, " or ") {
		return "the runtime events exist, but this trigger requires a missing event-or-subject-union semantic slot"
	}
	if strings.Contains(event, "first time") ||
		strings.Contains(event, "second time") ||
		strings.Contains(event, "third time") ||
		strings.Contains(event, "during your turn") ||
		strings.Contains(event, "during their turn") ||
		strings.Contains(event, "once each turn") {
		return "the runtime event exists, but this trigger requires a missing ordinal, active-turn, or temporal semantic slot"
	}
	if strings.Contains(event, "target") {
		return "the object-became-target event exists, but this trigger requires a missing target-subject, targeting-cause, or source relation slot"
	}
	if unrestrictedAbilityActivatedEvent(event) {
		if trigger.Condition != nil && strings.Contains(strings.ToLower(trigger.Condition.Text), "mana ability") {
			return "the ability-activated event exists, but non-mana exclusion in an intervening condition requires a missing semantic condition slot"
		}
		if !strings.Contains(event, "isn't a mana ability") {
			return "the runtime ability-activated event stream omits payment-time mana abilities, so unrestricted activated-ability triggers require a missing runtime capability"
		}
		return "the ability-activated event exists, but this trigger requires a missing source, activation-cost, or ability-provenance semantic slot"
	}
	if strings.Contains(event, "ability") {
		return "the ability-activated event exists, but this trigger requires a missing source, activation-cost, or ability-provenance semantic slot"
	}
	if strings.Contains(event, "cast") || strings.Contains(event, "spell") || strings.Contains(event, "copied") {
		return "the spell event exists, but this trigger requires a missing spell-event relation, copy, or provenance semantic slot"
	}
	if strings.Contains(event, "sacrific") {
		return "the permanent-sacrificed event exists, but this trigger requires a missing subject, actor, or sacrifice-provenance semantic slot"
	}
	if strings.Contains(event, "scry") || strings.Contains(event, "surveil") {
		return "the player-action event exists, but this trigger requires a missing action amount, provenance, or temporal semantic slot"
	}
	if strings.Contains(event, "tap") || strings.Contains(event, "untap") {
		if strings.Contains(event, "for mana") {
			return "the permanent-tapped event exists, but the runtime event lacks tapped-for-mana provenance"
		}
		return "the permanent-state event exists, but this trigger requires a missing subject, source, or turn-provenance semantic slot"
	}
	if strings.Contains(event, "counter") {
		return "the counter event exists, but this trigger requires a missing counter-kind, subject, controller, or removal semantic slot"
	}
	if strings.Contains(event, "draw") || strings.Contains(event, "discard") || strings.Contains(event, "cycl") {
		return "the player-card event exists, but this trigger requires a missing count, card-selection, source, or turn-provenance semantic slot"
	}
	if strings.Contains(event, "turned face up") {
		return "the permanent-turned-face-up event exists, but this trigger requires a missing subject, source, or Selection semantic slot"
	}
	if strings.Contains(event, "turned face down") {
		return "the runtime does not emit an authoritative permanent-turned-face-down event"
	}
	if strings.Contains(event, " enters") ||
		strings.Contains(event, " dies") ||
		strings.Contains(event, " leaves") ||
		strings.Contains(event, "graveyard") ||
		strings.Contains(event, "exiled") {
		return "the zone-change event exists, but this trigger requires a missing subject, zone, source, or Selection semantic slot"
	}
	if strings.Contains(event, "token") {
		return "the token-created event exists, but this trigger requires a missing creator, subject, or Selection semantic slot"
	}
	if strings.Contains(event, "transform") ||
		strings.Contains(event, "investigate") ||
		strings.Contains(event, "proliferate") ||
		strings.Contains(event, "explore") ||
		strings.Contains(event, "monstrous") ||
		strings.Contains(event, "venture") ||
		strings.Contains(event, "roll") ||
		strings.Contains(event, "vote") ||
		strings.Contains(event, "clash") {
		return "the runtime does not emit an authoritative event for this game action"
	}
	return "the runtime does not emit an authoritative event for this trigger action"
}

func unrestrictedAbilityActivatedEvent(event string) bool {
	for _, prefix := range []string{
		"you activate ",
		"an opponent activates ",
		"a player activates ",
	} {
		ability, ok := strings.CutPrefix(event, prefix)
		if !ok {
			continue
		}
		return ability == "an ability" || strings.HasPrefix(ability, "an ability of ")
	}
	return false
}

func lowerTriggeredAbilityKind(
	cardName string,
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (abilityLowering, *shared.Diagnostic) {
	triggeredAbility, diagnostic := lowerTriggeredAbility(cardName, ability, syntax)
	if diagnostic != nil {
		return abilityLowering{}, diagnostic
	}
	triggeredAbility.MaxTriggersPerTurn = ability.Trigger.MaxTriggersPerTurn
	spans := []shared.Span{ability.Trigger.Span}
	if len(ability.Content.Modes) != 0 {
		spans = append(spans, ability.Span)
	}
	if ability.Trigger.MaxTriggersPerTurn > 0 {
		spans = append(spans, ability.Trigger.MaxTriggersPerTurnSpan)
	}
	if syntax.AbilityWord != nil {
		spans = append(spans, shared.Span{
			Start: ability.Span.Start,
			End:   ability.Trigger.Span.Start,
		})
	}
	for i := range ability.Content.Effects {
		spans = append(spans, ability.Content.Effects[i].Span)
		if ability.Content.Effects[i].Payment.Span != (shared.Span{}) {
			spans = append(spans, ability.Content.Effects[i].Payment.Span)
		}
		if ability.Content.Effects[i].CopyMayChooseNewTargets {
			spans = append(spans, ability.Content.Effects[i].CopyChooseNewTargetsRiderSpan)
		}
		if len(ability.Content.Effects[i].TokenCopyGrantKeywords) != 0 {
			spans = append(spans, ability.Content.Effects[i].TokenCopyGrantRiderSpan)
		}
		if ability.Content.Effects[i].TokenGrantedAbilityRiderSpan != (shared.Span{}) {
			spans = append(spans, ability.Content.Effects[i].TokenGrantedAbilityRiderSpan)
		}
		if ability.Content.Effects[i].KeywordGrantChoiceAtRandom {
			spans = append(spans, ability.Content.Effects[i].KeywordChoiceAtRandomPreludeSpan)
		}
		if ability.Content.Effects[i].ReturnAsEnchantment {
			spans = append(spans, ability.Content.Effects[i].ReturnAsEnchantmentRiderSpan)
		}
		if ability.Content.Effects[i].Mana.PersistUntilEndOfTurn {
			spans = append(spans, ability.Content.Effects[i].PersistUntilEndOfTurnRiderSpan)
		}
		if ability.Content.Effects[i].PlayFromTopPayLife {
			spans = append(spans, ability.Content.Effects[i].PlayFromTopPayLifeRiderSpan)
		}
		if ability.Content.Effects[i].PileSplitSequence && ability.Content.Effects[i].PileSplitMiddleSpan != (shared.Span{}) {
			spans = append(spans, ability.Content.Effects[i].PileSplitMiddleSpan)
		}
	}
	for _, target := range ability.Content.Targets {
		spans = append(spans, target.Span)
		// A dedicated content recognizer that consumes a separate "choose ...
		// target" sentence also consumes that sentence's leading "choose" verb.
		// The generic trigger body does not yet cover such a ChoiceSpan, so this
		// is keyed on the exact Court of Vantress upkeep shape to keep every other
		// triggered ability byte-identical.
		if target.ChoiceSpan != (shared.Span{}) && courtOfVantressUpkeepContent(ability.Content) {
			spans = append(spans, target.ChoiceSpan)
		}
	}
	for _, condition := range ability.Content.Conditions {
		spans = append(spans, condition.Span)
	}
	for _, reference := range ability.Content.References {
		spans = append(spans, reference.Span)
	}
	spans = appendKeywordSpans(spans, ability.Content.Keywords)
	if ability.ExactSequence != compiler.ExactSequenceUnknown {
		spans = append(spans, ability.Content.Span)
		if ability.Optional {
			spans = append(spans, ability.OptionalSpan)
		}
	}
	for _, reminder := range syntax.Reminders {
		spans = append(spans, reminder.Span)
	}
	return abilityLowering{
		triggeredAbility: opt.Val(triggeredAbility),
		consumed: semanticConsumption{
			trigger:    true,
			optional:   ability.Optional,
			modes:      len(ability.Content.Modes),
			targets:    len(ability.Content.Targets),
			conditions: len(ability.Content.Conditions),
			effects:    len(ability.Content.Effects),
			keywords:   len(ability.Content.Keywords),
			references: len(ability.Content.References),
		},
		sourceSpans: spans,
	}, nil
}

func (lowering *abilityLowering) complete(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) bool {
	staticDeclarations := 0
	if ability.Static != nil {
		staticDeclarations = len(ability.Static.Declarations)
	}
	if lowering.consumed.cost != (ability.Cost != nil) ||
		lowering.consumed.alternativeCost != (ability.AlternativeCost != nil) ||
		lowering.consumed.trigger != (ability.Trigger != nil) ||
		lowering.consumed.optional != ability.Optional ||
		lowering.consumed.modes != len(ability.Content.Modes) ||
		lowering.consumed.targets != len(ability.Content.Targets) ||
		lowering.consumed.conditions != len(ability.Content.Conditions) ||
		lowering.consumed.effects != len(ability.Content.Effects) ||
		lowering.consumed.keywords != len(ability.Content.Keywords) ||
		lowering.consumed.references != len(ability.Content.References) ||
		lowering.consumed.declarations != staticDeclarations {
		return false
	}
	for _, span := range syntax.CoverageSpans() {
		if (syntax.AbilityWord != nil && rulesFreeAbilityWordLabel(ability.AbilityWord) &&
			(span == syntax.AbilityWord.SeparatorSpan ||
				spanCoveredByAbilityWord(span, syntax.AbilityWord))) ||
			spanCovered(span, lowering.sourceSpans) {
			continue
		}
		// A recognized coin flip re-parses each branch clause, so its compiled
		// branch effects carry spans in the branch sub-document rather than the
		// card source. The recognizer recorded the consumed flip and branch
		// sentence spans (card coordinates), which cover the whole coin-flip body.
		if syntax.CoinFlip != nil && spanCovered(span, syntax.CoinFlip.Spans) {
			continue
		}
		// A recognized keyword share (Odric, Lunarch Marshal) clears its body's
		// effect, keyword, and condition semantics in favor of the typed shared
		// keyword list, so the compiled content carries no spans for the body. The
		// recognizer recorded the consumed sentence spans, which cover the whole
		// construct body after the trigger clause.
		if syntax.KeywordShareGrant != nil && spanCovered(span, syntax.KeywordShareGrant.Spans) {
			continue
		}
		return false
	}
	return true
}
