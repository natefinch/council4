package payment

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
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

func applyCostModifiers(s State, ctx costModificationContext) spellCostOption {
	modifiers := s.CostModifiersForSpell(ctx.player, ctx.card, ctx.cardID, ctx.sourceZone, ctx.targets, ctx.bargained, ctx.bestowed)
	ctx.option.manaCost = applyGenericCostModifiers(ctx.option.manaCost, modifiers)
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

func applyGenericCostModifiers(manaCost *cost.Mana, modifiers []game.CostModifier) *cost.Mana {
	if len(modifiers) == 0 {
		return manaCost
	}
	generic := genericCostAmount(manaCost)
	minimum := 0
	taxInstances := 0
	var coloredIncrease []cost.Symbol
	set := (*int)(nil)
	for _, modifier := range modifiers {
		if modifier.SetGeneric.Exists {
			set = &modifier.SetGeneric.Val
		}
		generic += modifier.GenericIncrease
		generic -= modifier.GenericReduction
		taxInstances += modifier.LifePayableTaxInstances
		for _, c := range modifier.ColoredIncrease {
			coloredIncrease = append(coloredIncrease, cost.Symbol{Kind: cost.ColoredSymbol, Color: c})
		}
		if modifier.MinimumGeneric > minimum {
			minimum = modifier.MinimumGeneric
		}
	}
	if set != nil {
		generic = *set
	}
	if generic < minimum {
		generic = minimum
	}
	if generic < 0 {
		generic = 0
	}
	return costWithGenericAmount(manaCost, generic, taxInstances, coloredIncrease)
}

func genericCostAmount(manaCost *cost.Mana) int {
	if manaCost == nil {
		return 0
	}
	total := 0
	for _, symbol := range *manaCost {
		if symbol.Kind == cost.GenericSymbol {
			total += symbol.Generic
		}
	}
	return total
}

func costWithGenericAmount(manaCost *cost.Mana, generic, taxInstances int, coloredIncrease []cost.Symbol) *cost.Mana {
	var modified cost.Mana
	if generic > 0 {
		modified = append(modified, cost.O(generic))
	}
	for range taxInstances {
		modified = append(modified, cost.PhyrexianGeneric(2))
	}
	modified = append(modified, coloredIncrease...)
	if manaCost != nil {
		for _, symbol := range *manaCost {
			if symbol.Kind != cost.GenericSymbol {
				modified = append(modified, symbol)
			}
		}
	}
	return &modified
}
