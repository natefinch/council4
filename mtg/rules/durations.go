package rules

import (
	"github.com/natefinch/council4/mtg/game"
)

func untilEndOfTurnContinuousEffect(g *game.Game, obj *game.StackObject, permanent *game.Permanent, effect *game.Effect) game.ContinuousEffect {
	powerDelta := effect.PowerDelta
	if effect.PowerDeltaDynamic.Exists {
		powerDelta += dynamicAmountValue(g, obj, stackObjectController(obj), effect.PowerDeltaDynamic.Val)
	}
	toughnessDelta := effect.ToughnessDelta
	if effect.ToughnessDeltaDynamic.Exists {
		toughnessDelta += dynamicAmountValue(g, obj, stackObjectController(obj), effect.ToughnessDeltaDynamic.Val)
	}
	return untilEndOfTurnPTContinuousEffect(g, obj, permanent, powerDelta, toughnessDelta)
}

func untilEndOfTurnPTContinuousEffect(g *game.Game, obj *game.StackObject, permanent *game.Permanent, powerDelta, toughnessDelta int) game.ContinuousEffect {
	sourceID, sourceObjectID := damageSourceIDs(g, obj)
	effectID := g.IDGen.Next()
	return game.ContinuousEffect{
		ID:               effectID,
		SourceCardID:     sourceID,
		SourceObjectID:   sourceObjectID,
		Controller:       obj.Controller,
		Timestamp:        game.Timestamp(effectID),
		Duration:         game.DurationUntilEndOfTurn,
		CreatedTurn:      g.Turn.TurnNumber,
		AffectedObjectID: permanent.ObjectID,
		Layer:            game.LayerPowerToughnessModify,
		PowerDelta:       powerDelta,
		ToughnessDelta:   toughnessDelta,
	}
}

func effectDurationOrDefault(duration, fallback game.EffectDuration) game.EffectDuration {
	if duration == game.DurationPermanent {
		return fallback
	}
	return duration
}

func expireTurnStartDurations(g *game.Game) {
	g.ContinuousEffects = filterContinuousEffects(g.ContinuousEffects, func(effect *game.ContinuousEffect) bool {
		return effect.Duration == game.DurationUntilYourNextTurn &&
			effect.ExpiresFor == g.Turn.ActivePlayer &&
			effect.CreatedTurn < g.Turn.TurnNumber
	})
}

func expireCleanupDurations(g *game.Game) {
	g.ContinuousEffects = filterContinuousEffects(g.ContinuousEffects, func(effect *game.ContinuousEffect) bool {
		return effect.Duration == game.DurationUntilEndOfTurn || effect.Duration == game.DurationThisTurn
	})
}

func filterContinuousEffects(effects []game.ContinuousEffect, expired func(*game.ContinuousEffect) bool) []game.ContinuousEffect {
	if len(effects) == 0 {
		return effects
	}
	kept := effects[:0]
	for i := range effects {
		if expired(&effects[i]) {
			continue
		}
		kept = append(kept, effects[i])
	}
	return kept
}

func scheduleDelayedTrigger(g *game.Game, obj *game.StackObject, def *game.DelayedTriggerDef) bool {
	if obj == nil || def == nil || def.Timing == 0 {
		return false
	}
	sourceID, sourceObjectID := damageSourceIDs(g, obj)
	effects := append([]game.Effect(nil), def.Effects...)
	targets := append([]game.TargetSpec(nil), def.Targets...)
	ability := game.AbilityDef{
		Kind:     game.TriggeredAbility,
		Optional: def.Optional,
		Body: game.TriggeredAbilityBody{
			Optional: def.Optional,
			Content:  game.PlainAbilityContent{Targets: append([]game.TargetSpec(nil), targets...), LegacyEffects: append([]game.Effect(nil), effects...)},
		},
		Effects: effects,
		Targets: targets,
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
	if len(g.DelayedTriggers) == 0 {
		return
	}
	remaining := g.DelayedTriggers[:0]
	var ready []game.DelayedTrigger
	for i := range g.DelayedTriggers {
		trigger := &g.DelayedTriggers[i]
		if trigger.Timing != game.DelayedAtBeginningOfNextEndStep {
			remaining = append(remaining, *trigger)
			continue
		}
		ready = append(ready, *trigger)
	}
	ordered := orderDelayedTriggersAPNAP(g, ready)
	for i := range ordered {
		trigger := &ordered[i]
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
	if len(triggers) <= 1 {
		return triggers
	}
	ordered := make([]game.DelayedTrigger, 0, len(triggers))
	used := make([]bool, len(triggers))
	for _, playerID := range triggerAPNAPPlayers(g) {
		for i := range triggers {
			trigger := &triggers[i]
			if trigger.Controller != playerID {
				continue
			}
			ordered = append(ordered, *trigger)
			used[i] = true
		}
	}
	for i := range triggers {
		if !used[i] {
			ordered = append(ordered, triggers[i])
		}
	}
	return ordered
}
