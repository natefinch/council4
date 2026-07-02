package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// emitStateTriggerClauses recognizes state-trigger event clauses (CR 603.8) for
// triggers whose event text matched no event-based clause family. It runs after
// emitTriggerEventClauses so an event-based reading always wins; only a residual
// "When you control no <selection>" clause becomes a state trigger.
func emitStateTriggerClauses(abilities []Ability) {
	for i := range abilities {
		trigger := abilities[i].Trigger
		if trigger == nil ||
			trigger.PhaseStep != nil ||
			trigger.PlayerEvent != nil ||
			trigger.TriggerEvent != nil ||
			trigger.State != nil {
			continue
		}
		if trigger.Introduction.Kind != TriggerIntroductionWhen {
			continue
		}
		clause, ok := recognizeControllerControlsNoStateCondition(trigger.eventTokens, abilities[i].Atoms)
		if !ok {
			continue
		}
		trigger.State = &StateTriggerClause{Condition: clause}
	}
}

// recognizeControllerControlsNoStateCondition matches the "you control no
// <selection>" state-trigger idiom (CR 603.8), e.g. "When you control no
// Islands, sacrifice this creature." It reuses the typed controls-condition
// recognizer and accepts only the controller-scoped "no" (at most zero) shape so
// event-based and numeric-comparison wordings do not become state triggers. Any
// other scope, comparison, or unrepresentable selection fails closed.
func recognizeControllerControlsNoStateCondition(tokens []shared.Token, atoms Atoms) (ConditionClause, bool) {
	clause, ok := recognizeControlsCondition(tokens, atoms)
	if !ok {
		return ConditionClause{}, false
	}
	if clause.Predicate != ConditionPredicateControls ||
		clause.Scope != ConditionControlScopeController ||
		clause.Comparison != ConditionComparisonAtMost ||
		clause.CompareValue != 0 {
		return ConditionClause{}, false
	}
	return clause, true
}
