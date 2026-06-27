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
	option     spellCostOption
}

func applyCostModifiers(s State, ctx costModificationContext) spellCostOption {
	ctx.option.manaCost = applyGenericCostModifiers(ctx.option.manaCost, s.CostModifiersForSpell(ctx.player, ctx.card, ctx.cardID, ctx.sourceZone, ctx.targets))
	return ctx.option
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
