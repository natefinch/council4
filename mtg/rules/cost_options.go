package rules

import (
	"strings"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const flashbackAlternativeLabel = "Flashback"

func manaCostPtr(manaCost opt.V[cost.Mana]) *cost.Mana {
	if !manaCost.Exists {
		return nil
	}
	return &manaCost.Val
}

func isFlashbackAlternative(alternative game.AlternativeCost) bool {
	return strings.EqualFold(strings.TrimSpace(alternative.Label), flashbackAlternativeLabel)
}

func spellAdditionalCosts(card *game.CardDef) []game.AdditionalCost {
	if card == nil {
		return nil
	}
	abilities := card.AbilityDefs()
	for i := range abilities {
		ability := &abilities[i]
		if ability.IsSpell() {
			return abilityAdditionalCosts(ability)
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

func sacrificeCostMatcher(sacCost string) (func(*game.CardDef) bool, bool) {
	typed, ok := sacrificeAdditionalCost(sacCost)
	if !ok {
		return nil, false
	}
	return func(card *game.CardDef) bool {
		return localAdditionalCostMatchesCard(card, game.AdditionalCost{
			MatchCardType: typed.MatchPermanentType,
			CardType:      typed.PermanentType,
		})
	}, true
}

func sacrificeAdditionalCost(sacCost string) (game.AdditionalCost, bool) {
	normalized := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(sacCost)), ".")
	switch normalized {
	case "sacrifice a creature":
		return game.AdditionalCost{Kind: game.AdditionalCostSacrifice, Text: sacCost, Amount: 1, MatchPermanentType: true, PermanentType: types.Creature}, true
	case "sacrifice an artifact":
		return game.AdditionalCost{Kind: game.AdditionalCostSacrifice, Text: sacCost, Amount: 1, MatchPermanentType: true, PermanentType: types.Artifact}, true
	case "sacrifice an enchantment":
		return game.AdditionalCost{Kind: game.AdditionalCostSacrifice, Text: sacCost, Amount: 1, MatchPermanentType: true, PermanentType: types.Enchantment}, true
	case "sacrifice a land":
		return game.AdditionalCost{Kind: game.AdditionalCostSacrifice, Text: sacCost, Amount: 1, MatchPermanentType: true, PermanentType: types.Land}, true
	case "sacrifice a permanent":
		return game.AdditionalCost{Kind: game.AdditionalCostSacrifice, Text: sacCost, Amount: 1}, true
	default:
		return game.AdditionalCost{}, false
	}
}
