package rules

import (
	"github.com/natefinch/council4/mtg/game"
)

func untilEndOfTurnContinuousEffect(g *game.Game, obj *game.StackObject, permanent *game.Permanent, effect game.Effect) game.ContinuousEffect {
	sourceID, sourceObjectID := damageSourceIDs(g, obj)
	effectID := g.IDGen.Next()
	return game.ContinuousEffect{
		ID:               effectID,
		SourceCardID:     sourceID,
		SourceObjectID:   sourceObjectID,
		Controller:       obj.Controller,
		Timestamp:        int64(effectID),
		Duration:         game.DurationUntilEndOfTurn,
		CreatedTurn:      g.Turn.TurnNumber,
		AffectedObjectID: permanent.ObjectID,
		Layer:            game.LayerPowerToughnessModify,
		PowerDelta:       effect.PowerDelta,
		ToughnessDelta:   effect.ToughnessDelta,
	}
}

func effectDurationOrDefault(duration game.EffectDuration, fallback game.EffectDuration) game.EffectDuration {
	if duration == game.DurationPermanent {
		return fallback
	}
	return duration
}

func expireTurnStartDurations(g *game.Game) {
	if g == nil {
		return
	}
	g.ContinuousEffects = filterContinuousEffects(g.ContinuousEffects, func(effect game.ContinuousEffect) bool {
		return effect.Duration == game.DurationUntilYourNextTurn &&
			effect.ExpiresFor == g.Turn.ActivePlayer &&
			effect.CreatedTurn < g.Turn.TurnNumber
	})
}

func expireCleanupDurations(g *game.Game) {
	if g == nil {
		return
	}
	g.ContinuousEffects = filterContinuousEffects(g.ContinuousEffects, func(effect game.ContinuousEffect) bool {
		return effect.Duration == game.DurationUntilEndOfTurn || effect.Duration == game.DurationThisTurn
	})
}

func filterContinuousEffects(effects []game.ContinuousEffect, expired func(game.ContinuousEffect) bool) []game.ContinuousEffect {
	if len(effects) == 0 {
		return effects
	}
	kept := effects[:0]
	for _, effect := range effects {
		if expired(effect) {
			continue
		}
		kept = append(kept, effect)
	}
	return kept
}

func scheduleDelayedTrigger(g *game.Game, obj *game.StackObject, def *game.DelayedTriggerDef) bool {
	if g == nil || obj == nil || def == nil || def.Timing == 0 {
		return false
	}
	sourceID, sourceObjectID := damageSourceIDs(g, obj)
	ability := game.AbilityDef{
		Kind:     game.TriggeredAbility,
		Optional: def.Optional,
		Effects:  append([]game.Effect(nil), def.Effects...),
		Targets:  append([]game.TargetSpec(nil), def.Targets...),
	}
	g.DelayedTriggers = append(g.DelayedTriggers, game.DelayedTrigger{
		ID:             g.IDGen.Next(),
		SourceID:       sourceID,
		SourceObjectID: sourceObjectID,
		SourceTokenDef: obj.SourceTokenDef,
		Controller:     obj.Controller,
		CreatedTurn:    g.Turn.TurnNumber,
		Timing:         def.Timing,
		Ability:        ability,
	})
	return true
}

func putBeginningOfEndStepDelayedTriggersOnStack(g *game.Game) {
	if g == nil || len(g.DelayedTriggers) == 0 {
		return
	}
	remaining := g.DelayedTriggers[:0]
	var ready []game.DelayedTrigger
	for _, trigger := range g.DelayedTriggers {
		if trigger.Timing != game.DelayedAtBeginningOfNextEndStep {
			remaining = append(remaining, trigger)
			continue
		}
		ready = append(ready, trigger)
	}
	for _, trigger := range orderDelayedTriggersAPNAP(g, ready) {
		ability := trigger.Ability
		g.Stack.Push(&game.StackObject{
			ID:             g.IDGen.Next(),
			Kind:           game.StackTriggeredAbility,
			SourceID:       trigger.SourceObjectID,
			SourceCardID:   trigger.SourceID,
			SourceTokenDef: trigger.SourceTokenDef,
			Controller:     trigger.Controller,
			InlineAbility:  &ability,
		})
	}
	g.DelayedTriggers = remaining
}

func orderDelayedTriggersAPNAP(g *game.Game, triggers []game.DelayedTrigger) []game.DelayedTrigger {
	if len(triggers) <= 1 || g == nil {
		return triggers
	}
	ordered := make([]game.DelayedTrigger, 0, len(triggers))
	used := make([]bool, len(triggers))
	for _, playerID := range triggerAPNAPPlayers(g) {
		for i, trigger := range triggers {
			if trigger.Controller != playerID {
				continue
			}
			ordered = append(ordered, trigger)
			used[i] = true
		}
	}
	for i, trigger := range triggers {
		if !used[i] {
			ordered = append(ordered, trigger)
		}
	}
	return ordered
}
