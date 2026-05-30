package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/opt"
	"strings"
)

var paymentColors = []mana.Color{
	mana.White,
	mana.Blue,
	mana.Black,
	mana.Red,
	mana.Green,
	mana.Colorless,
}

const flashbackAlternativeLabel = "Flashback"

func manaCostPtr(cost opt.V[mana.Cost]) *mana.Cost {
	if !cost.Exists {
		return nil
	}
	return &cost.Val
}

type spellCostOption struct {
	index           int
	label           string
	card            *game.CardDef
	manaCost        *mana.Cost
	additionalCosts []game.AdditionalCost
}

func spellCostOptionsForZoneAndKicker(card *game.CardDef, sourceZone game.ZoneType, kickerPaid bool) []spellCostOption {
	if card == nil {
		return nil
	}
	ability, ok := firstSpellAbility(card)
	if !ok {
		return []spellCostOption{{index: 0, label: "Normal cost", card: card, manaCost: manaCostPtr(card.ManaCost)}}
	}
	requiredAdditional := abilityAdditionalCosts(ability)
	options := []spellCostOption{
		{
			index:           0,
			label:           "Normal cost",
			card:            card,
			manaCost:        spellManaCostWithKicker(manaCostPtr(card.ManaCost), ability, kickerPaid),
			additionalCosts: append([]game.AdditionalCost(nil), requiredAdditional...),
		},
	}
	for i, alternative := range ability.AlternativeCosts {
		if isFlashbackAlternative(alternative) && sourceZone != game.ZoneGraveyard {
			continue
		}
		if sourceZone == game.ZoneGraveyard && !isFlashbackAlternative(alternative) {
			continue
		}
		additional := append([]game.AdditionalCost(nil), requiredAdditional...)
		additional = append(additional, alternative.AdditionalCosts...)
		label := alternative.Label
		if label == "" {
			label = "Alternative cost"
		}
		options = append(options, spellCostOption{
			index:           i + 1,
			label:           label,
			card:            card,
			manaCost:        spellManaCostWithKicker(manaCostPtr(alternative.ManaCost), ability, kickerPaid),
			additionalCosts: additional,
		})
	}
	if sourceZone == game.ZoneGraveyard {
		return options[1:]
	}
	return options
}

func isFlashbackAlternative(alternative game.AlternativeCost) bool {
	return strings.EqualFold(strings.TrimSpace(alternative.Label), flashbackAlternativeLabel)
}

func spellManaCostWithKicker(base *mana.Cost, ability *game.AbilityDef, kickerPaid bool) *mana.Cost {
	if !kickerPaid || ability == nil || !ability.KickerCost.Exists {
		return base
	}
	combined := mana.Cost{}
	if base != nil {
		combined = append(combined, (*base)...)
	}
	combined = append(combined, ability.KickerCost.Val...)
	return &combined
}

func spellAdditionalCosts(card *game.CardDef) []game.AdditionalCost {
	if card == nil {
		return nil
	}
	for _, ability := range card.Abilities {
		if ability.Kind == game.SpellAbility {
			return abilityAdditionalCosts(&ability)
		}
	}
	return nil
}

func abilityAdditionalCosts(ability *game.AbilityDef) []game.AdditionalCost {
	if ability == nil {
		return nil
	}
	return append([]game.AdditionalCost(nil), ability.AdditionalCosts...)
}

func sacrificeCostMatcher(cost string) (func(*game.CardDef) bool, bool) {
	typed, ok := sacrificeAdditionalCost(cost)
	if !ok {
		return nil, false
	}
	return additionalCostCardMatcher(typed), true
}

func sacrificeAdditionalCost(cost string) (game.AdditionalCost, bool) {
	normalized := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(cost)), ".")
	switch normalized {
	case "sacrifice a creature":
		return game.AdditionalCost{Kind: game.AdditionalCostSacrifice, Text: cost, Amount: 1, MatchPermanentType: true, PermanentType: game.TypeCreature}, true
	case "sacrifice an artifact":
		return game.AdditionalCost{Kind: game.AdditionalCostSacrifice, Text: cost, Amount: 1, MatchPermanentType: true, PermanentType: game.TypeArtifact}, true
	case "sacrifice an enchantment":
		return game.AdditionalCost{Kind: game.AdditionalCostSacrifice, Text: cost, Amount: 1, MatchPermanentType: true, PermanentType: game.TypeEnchantment}, true
	case "sacrifice a land":
		return game.AdditionalCost{Kind: game.AdditionalCostSacrifice, Text: cost, Amount: 1, MatchPermanentType: true, PermanentType: game.TypeLand}, true
	case "sacrifice a permanent":
		return game.AdditionalCost{Kind: game.AdditionalCostSacrifice, Text: cost, Amount: 1}, true
	default:
		return game.AdditionalCost{}, false
	}
}
