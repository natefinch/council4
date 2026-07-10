package rules

import "github.com/natefinch/council4/mtg/game"

func staticAbilityCardHasLayer(face *game.CardFace, onBattlefield bool, layer game.ContinuousLayer) bool {
	if face == nil {
		return false
	}
	for i := range face.StaticAbilities {
		body := &face.StaticAbilities[i]
		if !staticAbilityFunctionsInZone(body, onBattlefield) {
			continue
		}
		if staticAbilityHasEffectForLayer(body, layer) {
			return true
		}
	}
	return false
}

func staticAbilityCardHasContinuousEffects(face *game.CardFace, onBattlefield bool) bool {
	if face == nil {
		return false
	}
	for i := range face.StaticAbilities {
		body := &face.StaticAbilities[i]
		if staticAbilityFunctionsInZone(body, onBattlefield) && len(body.ContinuousEffects) > 0 {
			return true
		}
	}
	return false
}
