package cardgen

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// Quantities can read targets through objects, players, groups, or operands.
func transformQuantity(amount game.Quantity, transform targetIndexTransform) (game.Quantity, bool) {
	dynamic := amount.DynamicAmount()
	if !dynamic.Exists {
		return amount, true
	}
	value := dynamic.Val
	if value.Kind == game.DynamicAmountTargetPower || value.Kind == game.DynamicAmountTargetToughness ||
		value.Kind == game.DynamicAmountTargetManaValue || value.Kind == game.DynamicAmountTargetCounters {
		return game.Quantity{}, false
	}
	if wholeTargetObjectReference(value.Object) {
		return game.Quantity{}, false
	}
	if objectReferenceCarriesTargetIndex(value.Object) {
		object, ok := transformObjectReference(value.Object, transform)
		if !ok {
			return game.Quantity{}, false
		}
		value.Object = object
	}
	if value.Player != nil {
		player, ok := transformPlayerReference(*value.Player, transform)
		if !ok {
			return game.Quantity{}, false
		}
		value.Player = &player
	}
	group, ok := transformQuantityGroup(value.Group, transform)
	if !ok {
		return game.Quantity{}, false
	}
	value.Group = group
	if value.Selection != nil {
		selection, ok := transformSelectionTargetIndices(*value.Selection, transform)
		if !ok {
			return game.Quantity{}, false
		}
		value.Selection = &selection
	}
	value.Operands = slices.Clone(value.Operands)
	for i := range value.Operands {
		operand, ok := transformQuantity(game.Dynamic(value.Operands[i]), transform)
		if !ok {
			return game.Quantity{}, false
		}
		value.Operands[i] = operand.DynamicAmount().Val
	}
	return game.Dynamic(value), true
}

func transformPlayerQuantityTargets(player *game.PlayerReference, amount *game.Quantity, transform targetIndexTransform) bool {
	if player.Kind() != game.PlayerReferenceNone {
		projected, ok := transformPlayerReference(*player, transform)
		if !ok {
			return false
		}
		*player = projected
	}
	projected, ok := transformQuantity(*amount, transform)
	if ok {
		*amount = projected
	}
	return ok
}

func transformRepeatEnvelopes(sequence []game.Instruction, transform targetIndexTransform) bool {
	for i := range sequence {
		instruction := &sequence[i]
		if instruction.Condition.Exists {
			condition, ok := transformEffectConditionTargetIndices(instruction.Condition.Val, transform)
			if !ok {
				return false
			}
			instruction.Condition = opt.Val(condition)
		}
		if instruction.CardCondition.Exists {
			condition := instruction.CardCondition.Val
			card, ok := transformCardReference(condition.Card, transform)
			if !ok {
				return false
			}
			condition.Card = card
			selection, ok := transformSelectionTargetIndices(condition.Selection, transform)
			if !ok {
				return false
			}
			condition.Selection = selection
			instruction.CardCondition = opt.Val(condition)
		}
		if instruction.ResultGate.Exists && instruction.ResultGate.Val.ObjectSelection.Exists {
			gate := instruction.ResultGate.Val
			selection, ok := transformSelectionTargetIndices(gate.ObjectSelection.Val, transform)
			if !ok {
				return false
			}
			gate.ObjectSelection = opt.Val(selection)
			instruction.ResultGate = opt.Val(gate)
		}
		if instruction.OptionalActor.Exists {
			player, ok := transformPlayerReference(instruction.OptionalActor.Val, transform)
			if !ok {
				return false
			}
			instruction.OptionalActor = opt.Val(player)
		}
		// Whole-target player groups cannot be narrowed to one clause's targets.
		for _, group := range []opt.V[game.PlayerGroupReference]{instruction.OptionalActorGroup, instruction.ForEachPlayerGroup} {
			if group.Exists && group.Val.Kind == game.PlayerGroupReferenceTargetedPlayers {
				return false
			}
		}
		if len(instruction.TemptingOfferBody) != 0 {
			return false
		}
	}
	return true
}

func transformEffectConditionTargetIndices(effect game.EffectCondition, transform targetIndexTransform) (game.EffectCondition, bool) {
	if wholeTargetObjectReference(effect.Object) ||
		effect.Condition.Exists && effect.Condition.Val.Object.Exists &&
			wholeTargetObjectReference(effect.Condition.Val.Object.Val) {
		return game.EffectCondition{}, false
	}
	if effect.Object.Kind() != game.ObjectReferenceNone {
		object, ok := transformObjectReference(effect.Object, transform)
		if !ok {
			return game.EffectCondition{}, false
		}
		effect.Object = object
	}
	if !effect.Condition.Exists {
		return effect, true
	}
	condition := effect.Condition.Val
	for _, selection := range []*opt.V[game.Selection]{
		&condition.ObjectMatches, &condition.ControlsGreatestManaValueInGroup,
	} {
		if selection.Exists {
			value, ok := transformSelectionTargetIndices(selection.Val, transform)
			if !ok {
				return game.EffectCondition{}, false
			}
			*selection = opt.Val(value)
		}
	}
	for _, count := range []*opt.V[game.SelectionCount]{
		&condition.ControlsMatching, &condition.AnyOpponentControls, &condition.OpponentsControl,
	} {
		if count.Exists {
			value := count.Val
			selection, ok := transformSelectionTargetIndices(value.Selection, transform)
			if !ok {
				return game.EffectCondition{}, false
			}
			value.Selection = selection
			*count = opt.Val(value)
		}
	}
	if condition.ControlComparison.Exists {
		comparison := condition.ControlComparison.Val
		selection, ok := transformSelectionTargetIndices(comparison.Selection, transform)
		if !ok {
			return game.EffectCondition{}, false
		}
		comparison.Selection = selection
		condition.ControlComparison = opt.Val(comparison)
	}
	if condition.Object.Exists {
		object, ok := transformObjectReference(condition.Object.Val, transform)
		if !ok {
			return game.EffectCondition{}, false
		}
		condition.Object = opt.Val(object)
	}
	if condition.TargetEnteredThisTurn.Exists {
		index, ok := transform(targetIndexObject, condition.TargetEnteredThisTurn.Val)
		if !ok {
			return game.EffectCondition{}, false
		}
		condition.TargetEnteredThisTurn = opt.Val(index)
	}
	condition.Aggregates = slices.Clone(condition.Aggregates)
	for i := range condition.Aggregates {
		comparison := &condition.Aggregates[i]
		if !comparison.ValueAmount.Exists {
			continue
		}
		amount, ok := transformQuantity(game.Dynamic(comparison.ValueAmount.Val), transform)
		if !ok {
			return game.EffectCondition{}, false
		}
		comparison.ValueAmount = amount.DynamicAmount()
	}
	effect.Condition = opt.Val(condition)
	return effect, true
}

func transformQuantityGroup(group game.GroupReference, transform targetIndexTransform) (game.GroupReference, bool) {
	if players, ok := group.PlayerGroup(); ok && players.Kind == game.PlayerGroupReferenceTargetedPlayers {
		return game.GroupReference{}, false
	}
	if anchor, ok := group.Anchor(); ok {
		if wholeTargetObjectReference(anchor) {
			return game.GroupReference{}, false
		}
		if objectReferenceCarriesTargetIndex(anchor) {
			return transformGroupReference(group, transform)
		}
	}
	if exclude, ok := group.Exclusion(); ok {
		if wholeTargetObjectReference(exclude) {
			return game.GroupReference{}, false
		}
		if objectReferenceCarriesTargetIndex(exclude) {
			return transformGroupReference(group, transform)
		}
	}
	if player, ok := group.PlayerAnchor(); ok {
		transformed, ok := transformPlayerReference(player, transform)
		if !ok {
			return game.GroupReference{}, false
		}
		if transformed != player {
			return transformGroupReference(group, transform)
		}
	}
	if selectionHasDynamicGroup(group.Selection()) {
		return transformGroupReference(group, transform)
	}
	return group, true
}

func wholeTargetObjectReference(object game.ObjectReference) bool {
	return object.Kind() == game.ObjectReferenceAllTargetPermanents ||
		object.Kind() == game.ObjectReferenceAllTargetStackObjects
}

func transformSelectionTargetIndices(selection game.Selection, transform targetIndexTransform) (game.Selection, bool) {
	selection.AnyOf = slices.Clone(selection.AnyOf)
	for i := range selection.AnyOf {
		alternative, ok := transformSelectionTargetIndices(selection.AnyOf[i], transform)
		if !ok {
			return game.Selection{}, false
		}
		selection.AnyOf[i] = alternative
	}
	if !selection.ManaValueDynamic.Exists || selection.ManaValueDynamic.Val.Group == nil {
		return selection, true
	}

	bound := selection.ManaValueDynamic.Val
	group, ok := transformQuantityGroup(*bound.Group, transform)
	if !ok {
		return game.Selection{}, false
	}
	bound.Group = &group
	selection.ManaValueDynamic = opt.Val(bound)
	return selection, true
}

func selectionHasDynamicGroup(selection game.Selection) bool {
	if selection.ManaValueDynamic.Exists && selection.ManaValueDynamic.Val.Group != nil {
		return true
	}
	return slices.ContainsFunc(selection.AnyOf, selectionHasDynamicGroup)
}
