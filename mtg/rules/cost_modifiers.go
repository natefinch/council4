package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
)

func applyCostModifiers(g *game.Game, context costModificationContext) spellCostOption {
	context.option.manaCost = applyGenericCostModifiers(context.option.manaCost, costModifiersForContext(g, context))
	return context.option
}

func costModifiersForContext(g *game.Game, context costModificationContext) []game.CostModifier {
	var modifiers []game.CostModifier
	for _, modifier := range g.CostModifiers {
		if modifier.Kind != game.CostModifierSpell {
			continue
		}
		if modifier.MatchCardType && (context.card == nil || !context.card.HasType(modifier.CardType)) {
			continue
		}
		modifiers = append(modifiers, modifier)
	}
	if context.sourceZone == game.ZoneCommand && context.cardID != 0 {
		player, ok := playerByID(g, context.player)
		if ok && player.CommanderInstanceID == context.cardID && player.CommanderTax() > 0 {
			modifiers = append(modifiers, game.CostModifier{
				Kind:            game.CostModifierSpell,
				GenericIncrease: player.CommanderTax(),
			})
		}
	}
	modifiers = append(modifiers, staticCostModifiersForContext(g, context)...)
	return modifiers
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
