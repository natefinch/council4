package rules

import "github.com/natefinch/council4/mtg/game"

func staticAbilityCardHasLayer(card *game.CardDef, onBattlefield bool, layer game.ContinuousLayer) bool {
	if card == nil {
		return false
	}
	for i := range card.StaticAbilities {
		body := &card.StaticAbilities[i]
		if !staticAbilityFunctionsInZone(body, onBattlefield) {
			continue
		}
		if staticAbilityHasEffectForLayer(body, layer) {
			return true
		}
	}
	return false
}

func staticAbilityCardHasContinuousEffects(card *game.CardDef, onBattlefield bool) bool {
	if card == nil {
		return false
	}
	for i := range card.StaticAbilities {
		body := &card.StaticAbilities[i]
		if staticAbilityFunctionsInZone(body, onBattlefield) && len(body.ContinuousEffects) > 0 {
			return true
		}
	}
	return false
}
