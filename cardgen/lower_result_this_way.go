package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

func planOptionalFlow(content compiler.AbilityContent) (optionalFlowPlan, bool) {
	plan, ok := planOptionalFlowBase(content)
	if !ok || !plan.enabled || plan.gateCondition < 0 {
		return plan, ok
	}
	condition := content.Conditions[plan.gateCondition]
	if condition.Predicate != compiler.ConditionPredicateResultThisWay {
		return plan, true
	}
	if condition.ThisWaySelection == nil {
		return optionalFlowPlan{}, false
	}
	selector := *condition.ThisWaySelection
	// The publisher already supplies the object universe. A bare card noun
	// needs no card-type atom, just like a bare permanent in the shared projector.
	if selector.Kind == compiler.SelectorCard {
		selector.Kind = compiler.SelectorPermanent
	}
	selection, ok := SelectionForSelector(selector)
	selection.Colored = selector.Colored
	if !ok || len(game.ValidateResultObjectSelection(selection)) != 0 {
		return optionalFlowPlan{}, false
	}
	plan.resultSelection = opt.Val(selection)
	plan.resultCardNoun = condition.ThisWayCardNoun
	return plan, true
}

func (p optionalFlowPlan) resultGate(succeeded game.TriState) game.InstructionResultGate {
	gate := game.InstructionResultGate{
		Key: optionalIfYouDoResultKey, Succeeded: succeeded, ObjectSelection: p.resultSelection,
		CardOnly: p.resultCardNoun,
	}
	if p.resultSelection.Exists && succeeded == game.TriFalse {
		if p.elseGateCondition >= 0 {
			gate.ObjectSelection = opt.V[game.Selection]{}
			gate.CardOnly = false
		} else {
			gate.Succeeded = game.TriTrue
			gate.Negate = true
		}
	}
	return gate
}
