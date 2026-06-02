package payment

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
)

// costModificationContext carries the context needed to apply cost modifiers.
type costModificationContext struct {
	player     game.PlayerID
	card       *game.CardDef
	cardID     id.ID
	sourceZone game.ZoneType
	option     spellCostOption
}

func applyCostModifiers(s State, ctx costModificationContext) spellCostOption {
	ctx.option.manaCost = applyGenericCostModifiers(ctx.option.manaCost, s.CostModifiersForSpell(ctx.player, ctx.card, ctx.cardID, ctx.sourceZone))
	return ctx.option
}

func applyGenericCostModifiers(manaCost *cost.Mana, modifiers []game.CostModifier) *cost.Mana {
	if len(modifiers) == 0 {
		return manaCost
	}
	generic := genericCostAmount(manaCost)
	minimum := 0
	set := (*int)(nil)
	for _, modifier := range modifiers {
		if modifier.SetGeneric.Exists {
			set = &modifier.SetGeneric.Val
		}
		generic += modifier.GenericIncrease
		generic -= modifier.GenericReduction
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
	return costWithGenericAmount(manaCost, generic)
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

func costWithGenericAmount(manaCost *cost.Mana, generic int) *cost.Mana {
	var modified cost.Mana
	if generic > 0 {
		modified = append(modified, cost.O(generic))
	}
	if manaCost != nil {
		for _, symbol := range *manaCost {
			if symbol.Kind != cost.GenericSymbol {
				modified = append(modified, symbol)
			}
		}
	}
	return &modified
}
