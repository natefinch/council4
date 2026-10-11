package rules

import "github.com/natefinch/council4/mtg/game"

func setEffectiveTriggerOrigin(values *permanentEffectiveValues, origin game.TriggerOrigin) {
	index := len(values.abilities) - 1
	body, ok := values.abilities[index].(*game.TriggeredAbility)
	if !ok || body.MaxTriggersPerTurn == 0 {
		return
	}
	if values.triggerOrigins == nil {
		values.triggerOrigins = make(map[int]game.TriggerOrigin)
	}
	values.triggerOrigins[index] = origin
}
