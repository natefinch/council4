package payment

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
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

func applyGenericCostModifiers(cost *mana.Cost, modifiers []game.CostModifier) *mana.Cost {
	if len(modifiers) == 0 {
		return cost
	}
	generic := genericCostAmount(cost)
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
	return costWithGenericAmount(cost, generic)
}

func genericCostAmount(cost *mana.Cost) int {
	if cost == nil {
		return 0
	}
	total := 0
	for _, symbol := range *cost {
		if symbol.Kind == mana.GenericSymbol {
			total += symbol.Generic
		}
	}
	return total
}

func costWithGenericAmount(cost *mana.Cost, generic int) *mana.Cost {
	var modified mana.Cost
	if generic > 0 {
		modified = append(modified, mana.GenericMana(generic))
	}
	if cost != nil {
		for _, symbol := range *cost {
			if symbol.Kind != mana.GenericSymbol {
				modified = append(modified, symbol)
			}
		}
	}
	return &modified
}
