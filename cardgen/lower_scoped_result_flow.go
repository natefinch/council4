package cardgen

import (
	"fmt"
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

type scopedResultFlow struct {
	publishers   map[int]game.ResultKey
	gates        map[int]game.InstructionResultGate
	optional     map[int]bool
	otherwise    map[int]bool
	clearNegated map[int]bool
	conditions   map[int]bool
	actions      optionalActionGroups
}

func (p optionalFlowPlan) singleOptionalTail(effectCount int) bool {
	if p.scoped == nil {
		return true
	}
	if len(p.scoped.publishers) != 1 || len(p.scoped.conditions) != 1 ||
		p.optionalIndex != 0 || p.gateIndex != 1 ||
		len(p.scoped.gates) != effectCount-1 {
		return false
	}
	for ei := 1; ei < effectCount; ei++ {
		if p.scoped.gates[ei].Key == "" || p.scoped.otherwise[ei] {
			return false
		}
	}
	return true
}

func (p optionalFlowPlan) consumesCondition(ci int) bool {
	if p.scoped != nil {
		return p.scoped.conditions[ci]
	}
	return p.enabled && (ci == p.gateCondition || ci == p.elseGateCondition)
}

func resultConditionGate(condition compiler.CompiledCondition) (game.InstructionResultGate, bool) {
	gate := game.InstructionResultGate{Succeeded: game.TriTrue}
	if condition.Predicate == compiler.ConditionPredicatePriorInstructionNotAccepted {
		gate.Succeeded = game.TriFalse
		return gate, true
	}
	if condition.Predicate != compiler.ConditionPredicateResultThisWay {
		return gate, true
	}
	if condition.ThisWaySelection == nil {
		return game.InstructionResultGate{}, false
	}
	selector := *condition.ThisWaySelection
	if selector.Kind == compiler.SelectorCard {
		selector.Kind = compiler.SelectorPermanent
	}
	selection, ok := SelectionForSelector(selector)
	selection.Colored = selector.Colored
	if !ok || len(game.ValidateResultObjectSelection(selection)) != 0 {
		return game.InstructionResultGate{}, false
	}
	gate.ObjectSelection = opt.Val(selection)
	gate.CardOnly = condition.ThisWayCardNoun
	if condition.ThisWayCount > 0 {
		count := game.IntRange{Min: condition.ThisWayCount, Max: condition.ThisWayCount}
		if condition.ThisWayCountComparison == parser.ConditionComparisonAtLeast {
			count.Max = 0
		} else if condition.ThisWayCountComparison != parser.ConditionComparisonNone {
			return game.InstructionResultGate{}, false
		}
		gate.ObjectCountRange = opt.Val(count)
	}
	return gate, true
}

func planScopedResultFlow(content compiler.AbilityContent) (result optionalFlowPlan, valid, handled bool) {
	flow := &scopedResultFlow{
		publishers: make(map[int]game.ResultKey), gates: make(map[int]game.InstructionResultGate),
		optional: make(map[int]bool), otherwise: make(map[int]bool),
		clearNegated: make(map[int]bool), conditions: make(map[int]bool),
	}
	plan := optionalFlowPlan{
		scoped: flow, bareIndex: -1, elseIndex: -1, elseGateCondition: -1,
		optionalIndex: -1, gateIndex: -1, gateCondition: -1, extraOptionalIndex: -1,
		failureCategory: "structural — actual-result condition has no unique earlier typed producer",
	}
	indices := make(map[int]int)
	for ei, effect := range content.Effects {
		if effect.ClauseID <= 0 {
			continue
		}
		if _, exists := indices[effect.ClauseID]; exists {
			return plan, false, true
		}
		indices[effect.ClauseID] = ei
		flow.optional[ei] = effect.Optional && !effect.DelayedSubject.OptionalAtDelayedTime
	}
	var actionReason string
	flow.actions, actionReason = planOptionalActionGroups(content.Effects, indices)
	if actionReason != "" {
		plan.failureCategory = actionReason
		return plan, false, true
	}
	for ci, condition := range content.Conditions {
		if !isResolvingSuccessGate(condition.Predicate) &&
			condition.Predicate != compiler.ConditionPredicatePriorInstructionNotAccepted {
			continue
		}
		flow.conditions[ci] = true
		owners, reason := conditionClauseIndices(content.Effects, condition)
		producer, exists := indices[condition.Ownership.ResultProducerClauseID]
		if reason != "" || !exists || producer >= owners[0] ||
			condition.Kind != compiler.ConditionIf || condition.Negated || condition.Intervening ||
			!resultThisWayMatchesEffect(condition, content.Effects[producer]) {
			return plan, false, true
		}
		effect := content.Effects[producer]
		if condition.Predicate != compiler.ConditionPredicateResultThisWay && flow.actions.isCompound(producer) {
			plan.failureCategory = "structural — whole optional action outcome not modeled"
			return plan, false, true
		}
		if condition.Predicate == compiler.ConditionPredicateCounterSucceeded {
			if owned, ok := counterOutcomeProducer(content, condition); !ok || owned != producer {
				plan.failureCategory = counterOwnershipCategory
				return plan, false, true
			}
		}
		if effect.Negated || effect.DelayedTiming != 0 {
			plan.failureCategory = "structural — actual-result producer is negated or delayed"
			return plan, false, true
		}
		if condition.ThisWayController && !resultProducerActsForController(effect) {
			plan.failureCategory = "structural — actual-result actor ownership not modeled"
			return plan, false, true
		}
		flow.publishers[producer] = game.ResultKey(fmt.Sprintf("result-clause-%d", effect.ClauseID))
		gate, ok := resultConditionGate(condition)
		if !ok {
			plan.failureCategory = "structural — actual-result selection or count not modeled"
			return plan, false, true
		}
		gate.Key = flow.publishers[producer]
		for _, owner := range owners {
			if _, exists := flow.gates[owner]; exists {
				plan.failureCategory = "structural — overlapping actual-result predicates not modeled"
				return plan, false, true
			}
			flow.gates[owner] = gate
			if condition.Predicate == compiler.ConditionPredicatePriorInstructionNotAccepted {
				flow.clearNegated[owner] = true
				flow.otherwise[owner] = true
				plan.elseGateCondition = ci
			}
		}
		for ei, effect := range content.Effects {
			if effect.ResultElseOfClauseID != content.Effects[owners[0]].ClauseID {
				continue
			}
			if ei <= owners[len(owners)-1] {
				return plan, false, true
			}
			complement := gate
			if complement.ObjectSelection.Exists {
				complement.Negate = true
			} else {
				complement.Succeeded = game.TriFalse
			}
			if _, exists := flow.gates[ei]; exists {
				return plan, false, true
			}
			flow.gates[ei] = complement
			flow.otherwise[ei] = true
		}
		if plan.gateCondition < 0 && condition.Predicate != compiler.ConditionPredicatePriorInstructionNotAccepted {
			plan.gateCondition, plan.gateIndex = ci, owners[0]
			plan.optionalIndex = producer
			plan.publishWithoutOptional = !effect.Optional
			plan.resultSelection, plan.resultCardNoun = gate.ObjectSelection, gate.CardOnly
		}
	}
	if contentHasReflexiveGate(content) && (len(flow.conditions) != 1 || len(flow.publishers) != 1) {
		plan.failureCategory = "structural — mixed reflexive actual-result flow not modeled"
		return plan, false, true
	}
	if len(flow.conditions) == 0 {
		hasOptional := false
		for _, optional := range flow.optional {
			hasOptional = hasOptional || optional
		}
		hasDelayedOptional := slices.ContainsFunc(content.Effects, func(effect compiler.CompiledEffect) bool {
			return effect.Optional && effect.DelayedSubject.OptionalAtDelayedTime && fixedPhaseSubjectEffectModeled(effect)
		})
		if !hasOptional && !hasDelayedOptional {
			return optionalFlowPlan{}, false, false
		}
	}
	plan.enabled = true
	if len(flow.publishers) == 1 {
		// Preserve existing single-producer definitions; only independent
		// publications need new scoped keys.
		for producer := range flow.publishers {
			flow.publishers[producer] = optionalIfYouDoResultKey
			plan.optionalIndex = producer
			plan.publishWithoutOptional = !content.Effects[producer].Optional
		}
		for owner, gate := range flow.gates {
			gate.Key = optionalIfYouDoResultKey
			flow.gates[owner] = gate
		}
	} else {
		// Specialized single-tail consumers cannot consume this plan.
		plan.optionalIndex = -1
	}
	for ei, effect := range content.Effects {
		if effect.Negated && !flow.clearNegated[ei] ||
			effect.Optional && effect.DelayedTiming != 0 && !fixedPhaseSubjectEffectModeled(effect) {
			return plan, false, true
		}
		if flow.gates[ei].Key == "" && !flow.actions.ownsOptionalAntecedents(content, ei) && optionalAntecedentUnmodeled(content, ei) {
			plan.failureCategory = "structural — optional published-subject antecedent not modeled"
			return plan, false, true
		}
		if flow.otherwise[ei] && plan.elseIndex < 0 {
			plan.elseIndex = ei
		}
		if effect.Optional && ei != plan.optionalIndex && flow.gates[ei].Key != "" {
			if plan.extraOptionalIndex >= 0 {
				plan.optionalIndex = -1
			}
			plan.extraOptionalIndex = ei
		}
	}
	return plan, true, true
}

func optionalAntecedentUnmodeled(content compiler.AbilityContent, ei int) bool {
	if fixedPhaseSubjectEffectModeled(content.Effects[ei]) &&
		content.Effects[ei].DelayedSubject.Kind == parser.DelayedSubjectProduct {
		return false
	}
	amount := content.Effects[ei].Amount
	if ei > 0 && content.Effects[ei-1].Optional &&
		(amount.DynamicKind == compiler.DynamicAmountSourceManaValue ||
			amount.DynamicKind == compiler.DynamicAmountSourcePower ||
			amount.DynamicKind == compiler.DynamicAmountSourceToughness) {
		return true
	}
	for _, reference := range content.Effects[ei].References {
		if reference.Binding == compiler.ReferenceBindingPriorInstructionResult &&
			reference.PriorInstruction >= 0 && reference.PriorInstruction < ei &&
			content.Effects[reference.PriorInstruction].Optional {
			return true
		}
	}
	return ei > 0 && content.Effects[ei-1].Optional &&
		(content.Effects[ei].Context == parser.EffectContextReferencedObject ||
			content.Effects[ei].Context == parser.EffectContextPriorSubject)
}
func resultProducerActsForController(effect compiler.CompiledEffect) bool {
	if effect.Context == parser.EffectContextController {
		return true
	}
	if (effect.Kind != compiler.EffectExile && effect.Kind != compiler.EffectDestroy) ||
		effect.Context != parser.EffectContextTarget || len(effect.Targets) == 0 {
		return false
	}
	for _, target := range effect.Targets {
		kind := target.Selector.Kind
		if kind == compiler.SelectorPlayer || kind == compiler.SelectorOpponent ||
			kind == compiler.SelectorUnknown || kind == compiler.SelectorAny {
			return false
		}
	}
	return true
}

func (flow *scopedResultFlow) apply(ei int, sequence []game.Instruction) (string, bool) {
	key, publishes := flow.publishers[ei]
	if publishes {
		index := 0
		counterIndex, isCounter := counterActionInstruction(sequence)
		if !flow.optional[ei] && isCounter {
			index = counterIndex
		} else if len(sequence) != 1 {
			return "structural — result producer or optional effect requires one instruction", false
		}
		if sequence[index].Optional || sequence[index].PublishResult != "" {
			return "structural — result producer or optional effect requires one instruction", false
		}
		if sequence[index].ResultGate.Exists && !isCounter {
			return "structural — result publication conflicts with clause result wiring", false
		}
		sequence[index].PublishResult = key
	}
	if gate, exists := flow.gates[ei]; exists {
		gated, ok := appendResultGatedBranch(nil, gate, sequence)
		if !ok || len(gated) == 0 {
			return "structural — overlapping actual-result gates not modeled", false
		}
		copy(sequence, gated)
	}
	if reason := flow.actions.apply(ei, sequence); reason != "" {
		return reason, false
	}
	return "", true
}
