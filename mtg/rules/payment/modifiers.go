package payment

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

// costModificationContext carries the context needed to apply cost modifiers.
type costModificationContext struct {
	player     game.PlayerID
	card       *game.CardDef
	cardID     id.ID
	sourceZone zone.Type
	targets    []game.Target
	bargained  bool
	bestowed   bool
	option     spellCostOption
}

func resolveManaCostX(manaCost *cost.Mana, xValue int) (*cost.Mana, bool) {
	if xValue < 0 {
		return nil, false
	}
	if manaCost == nil {
		return nil, true
	}
	resolved := *manaCost
	variable := costHasVariableMana(manaCost)
	if variable {
		resolved = append(cost.Mana(nil), resolved...)
	}
	generic := 0
	maximum := int(^uint(0) >> 1)
	for index, symbol := range resolved {
		if symbol.Kind == cost.VariableSymbol {
			resolved[index] = cost.O(xValue)
			symbol = resolved[index]
		}
		if symbol.Kind == cost.GenericSymbol {
			if symbol.Generic < 0 || symbol.Generic > maximum-generic {
				return nil, false
			}
			generic += symbol.Generic
		}
	}
	if variable {
		return &resolved, true
	}
	return manaCost, true
}

func applyCostModifiers(s State, ctx costModificationContext) spellCostOption {
	modifiers := s.CostModifiersForSpell(ctx.player, ctx.card, ctx.cardID, ctx.sourceZone, ctx.targets, ctx.bargained, ctx.bestowed)
	ctx.option.manaCost = applyManaCostModifiers(ctx.option.manaCost, modifiers)
	ctx.option.additionalCosts = appendLifeCostModifiers(ctx.option.additionalCosts, modifiers)
	return ctx.option
}

// appendLifeCostModifiers appends the additional life a cast-cost tax imposes
// on the spell ("Spells your opponents cast that target this creature cost an
// additional 3 life to cast.", Terror of the Peaks). It sums the LifeIncrease of
// every applicable modifier into one pay-life additional cost so the payment
// planner requires the caster to have that much life. No matching modifier
// leaves the spell's additional costs unchanged.
func appendLifeCostModifiers(additionalCosts []cost.Additional, modifiers []game.CostModifier) []cost.Additional {
	life := 0
	for _, modifier := range modifiers {
		if modifier.LifeIncrease > 0 {
			life += modifier.LifeIncrease
		}
	}
	if life == 0 {
		return additionalCosts
	}
	return append(additionalCosts, cost.Additional{Kind: cost.AdditionalPayLife, Amount: life})
}

// applyManaCostModifiers determines the mana component of a spell's total cost
// (CR 601.2f): the chosen mana or alternative cost plus every increase, minus
// every reduction. A generic reduction lowers only the generic component (CR
// 118.7a). A colored reduction removes a symbol of its color from the cost
// including any colored increase, or, when no such symbol remains, lowers the
// generic component instead (CR 118.7b, 118.7c). The generic component never
// falls below zero.
func applyManaCostModifiers(manaCost *cost.Mana, modifiers []game.CostModifier) *cost.Mana {
	if len(modifiers) == 0 {
		return manaCost
	}
	generic := genericCostAmount(manaCost)
	minimum := 0
	taxInstances := 0
	var manaIncrease cost.Mana
	var coloredReduction []mana.Color
	set := (*int)(nil)
	for _, modifier := range modifiers {
		if modifier.SetGeneric.Exists {
			set = &modifier.SetGeneric.Val
		}
		generic += modifier.GenericIncrease
		generic += genericManaAmount(modifier.ManaIncrease)
		generic -= modifier.GenericReduction
		coloredReduction = append(coloredReduction, modifier.ColoredReduction...)
		taxInstances += modifier.LifePayableTaxInstances
		for _, c := range modifier.ColoredIncrease {
			manaIncrease = append(manaIncrease, cost.Symbol{Kind: cost.ColoredSymbol, Color: c})
		}
		for _, symbol := range modifier.ManaIncrease {
			if symbol.Kind != cost.GenericSymbol {
				manaIncrease = append(manaIncrease, symbol)
			}
		}
		if modifier.MinimumGeneric > minimum {
			minimum = modifier.MinimumGeneric
		}
	}
	var printed cost.Mana
	if manaCost != nil {
		for _, symbol := range *manaCost {
			if symbol.Kind != cost.GenericSymbol {
				printed = append(printed, symbol)
			}
		}
	}
	nonGeneric, excess := append(manaIncrease, printed...).ReduceColored(coloredReduction)
	generic -= excess
	if set != nil {
		generic = *set
	}
	if generic < minimum {
		generic = minimum
	}
	if generic < 0 {
		generic = 0
	}
	return costWithGenericAmount(generic, taxInstances, nonGeneric)
}

func genericCostAmount(manaCost *cost.Mana) int {
	if manaCost == nil {
		return 0
	}
	return genericManaAmount(*manaCost)
}

func genericManaAmount(manaCost cost.Mana) int {
	total := 0
	for _, symbol := range manaCost {
		if symbol.Kind == cost.GenericSymbol {
			total += symbol.Generic
		}
	}
	return total
}

// costWithGenericAmount builds a total mana cost from its generic amount, its
// life-payable commander tax instances and its remaining non-generic symbols.
func costWithGenericAmount(generic, taxInstances int, nonGeneric cost.Mana) *cost.Mana {
	var modified cost.Mana
	if generic > 0 {
		modified = append(modified, cost.O(generic))
	}
	for range taxInstances {
		modified = append(modified, cost.PhyrexianGeneric(2))
	}
	modified = append(modified, nonGeneric...)
	return &modified
}
