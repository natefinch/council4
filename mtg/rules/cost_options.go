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

func isFlashbackAlternative(alternative cost.Alternative) bool {
	return strings.EqualFold(strings.TrimSpace(alternative.Label), flashbackAlternativeLabel)
}

func spellAdditionalCosts(card *game.CardDef) []cost.Additional {
	if card == nil {
		return nil
	}
	return append([]cost.Additional(nil), card.AdditionalCosts...)
}

func abilityAdditionalCosts(additionalCosts []cost.Additional) []cost.Additional {
	return append([]cost.Additional(nil), additionalCosts...)
}

func sacrificeCostMatcher(sacCost string) (func(*game.CardDef) bool, bool) {
	typed, ok := sacrificeAdditionalCost(sacCost)
	if !ok {
		return nil, false
	}
	return func(card *game.CardDef) bool {
		return localAdditionalCostMatchesCard(card, cost.Additional{
			MatchCardType: typed.MatchPermanentType,
			CardType:      typed.PermanentType,
		})
	}, true
}

func sacrificeAdditionalCost(sacCost string) (cost.Additional, bool) {
	normalized := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(sacCost)), ".")
	switch normalized {
	case "sacrifice a creature":
		return cost.Additional{Kind: cost.AdditionalSacrifice, Text: sacCost, Amount: 1, MatchPermanentType: true, PermanentType: types.Creature}, true
	case "sacrifice an artifact":
		return cost.Additional{Kind: cost.AdditionalSacrifice, Text: sacCost, Amount: 1, MatchPermanentType: true, PermanentType: types.Artifact}, true
	case "sacrifice an enchantment":
		return cost.Additional{Kind: cost.AdditionalSacrifice, Text: sacCost, Amount: 1, MatchPermanentType: true, PermanentType: types.Enchantment}, true
	case "sacrifice a land":
		return cost.Additional{Kind: cost.AdditionalSacrifice, Text: sacCost, Amount: 1, MatchPermanentType: true, PermanentType: types.Land}, true
	case "sacrifice a permanent":
		return cost.Additional{Kind: cost.AdditionalSacrifice, Text: sacCost, Amount: 1}, true
	default:
		return cost.Additional{}, false
	}
}
