package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/opt"
)

// Availability is separate from comparison truth: negation must not turn an
// unknown characteristic into a successful numeric condition.
func objectConditionInformationAvailable(g *game.Game, ctx conditionContext, cond *game.Condition) bool {
	if cond.TargetCardResultKey != "" && (!cond.Object.Exists ||
		cond.Object.Val.Kind() != game.ObjectReferenceTargetCard || !cond.ObjectMatches.Exists) {
		return false
	}
	if cond.Object.Exists && cond.Object.Val.Kind() == game.ObjectReferenceTargetCard &&
		(ctx.obj == nil || !conditionTargetCardAvailable(g, ctx.obj, cond)) {
		return false
	}
	if cond.UseCounteredSpellManaValue {
		if _, ok := counteredSpellConditionComparison(cond); !ok {
			return false
		}
	}
	if !cond.ObjectMatches.Exists {
		return true
	}
	original := cond.ObjectMatches.Val
	selection := game.Selection{}
	if original.ManaValue.Exists {
		selection.ManaValue = opt.Val(compare.Int{Op: compare.Any})
	}
	if original.Power.Exists {
		selection.Power = opt.Val(compare.Int{Op: compare.Any})
	}
	if original.Toughness.Exists {
		selection.Toughness = opt.Val(compare.Int{Op: compare.Any})
	}
	if selection.Empty() {
		return true
	}
	available := game.Condition{
		Object: cond.Object, ObjectMatches: opt.Val(selection),
		UseCounteredSpellManaValue: cond.UseCounteredSpellManaValue,
		TargetCardResultKey:        cond.TargetCardResultKey,
	}
	return conditionObjectMatches(g, ctx, &available)
}

func counteredSpellConditionTargetMatchesCapture(g *game.Game, obj *game.StackObject, cond *game.Condition) bool {
	if _, ok := counteredSpellConditionComparison(cond); !ok {
		return false
	}
	index := cond.Object.Val.TargetIndex()
	slot := remapTargetSlot(g, obj, index)
	if slot < 0 || slot >= len(obj.Targets) || obj.Targets[slot].Kind != game.TargetStackObject ||
		obj.Targets[slot].StackObjectID == 0 {
		return false
	}
	capturedID, captured := obj.TargetManaValueLKIObjectIDs[index]
	return !captured || capturedID == obj.Targets[slot].StackObjectID
}

func counteredSpellMatchesConditionSelection(g *game.Game, obj *game.StackObject, cond *game.Condition) bool {
	numeric, ok := counteredSpellConditionComparison(cond)
	if !ok {
		return false
	}
	index := cond.Object.Val.TargetIndex()
	slot := remapTargetSlot(g, obj, index)
	if slot < 0 || slot >= len(obj.Targets) || obj.Targets[slot].Kind != game.TargetStackObject {
		return false
	}
	targetID := obj.Targets[slot].StackObjectID
	if targetID == 0 || obj.TargetManaValueLKIObjectIDs[index] != targetID {
		return false
	}
	if _, live := stackObjectByID(g, targetID); live {
		return false
	}
	value, known := obj.TargetManaValueLKI[index]
	return known && numeric.Matches(value)
}

func counteredSpellConditionComparison(cond *game.Condition) (compare.Int, bool) {
	if !cond.Object.Exists || cond.Object.Val.Kind() != game.ObjectReferenceTargetStackObject ||
		!cond.ObjectMatches.Exists || len(cond.Types) != 0 {
		return compare.Int{}, false
	}
	selection := cond.ObjectMatches.Val
	numeric := selection.ManaValue
	selection.ManaValue = opt.V[compare.Int]{}
	if !numeric.Exists || !selection.Empty() {
		return compare.Int{}, false
	}
	return numeric.Val, true
}

func eventSpellMatchesConditionSelection(g *game.Game, obj *game.StackObject, ctx conditionContext, selection *game.Selection) bool {
	if !obj.HasTriggerEvent || obj.TriggerEvent.StackObjectID == 0 ||
		(obj.TriggerEvent.Kind != game.EventSpellCast && obj.TriggerEvent.Kind != game.EventSpellCopied) {
		return false
	}
	if stack, live := stackObjectByID(g, obj.TriggerEvent.StackObjectID); live {
		return resolvedObjectMatchesConditionSelection(g, ctx, &resolvedObjectReference{stack: stack}, selection)
	}
	subject := selectionSubject{
		kind: subjectCastSpell, g: g, event: obj.TriggerEvent,
		cardTypes: obj.TriggerEvent.CardTypes, controller: obj.TriggerEvent.Controller,
		viewer: ctx.controller,
	}
	return matchSelection(&subject, selection)
}
