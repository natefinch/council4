package rules

import (
	"maps"

	"github.com/natefinch/council4/mtg/game"
)

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
	if len(g.RuleEffects) == 0 {
		return
	}
	kept := g.RuleEffects[:0]
	for i := range g.RuleEffects {
		effect := &g.RuleEffects[i]
		if effect.Duration == game.DurationUntilYourNextTurn &&
			effect.ExpiresFor == g.Turn.ActivePlayer &&
			effect.CreatedTurn < g.Turn.TurnNumber {
			continue
		}
		kept = append(kept, *effect)
	}
	g.RuleEffects = kept
}

func expireCleanupDurations(g *game.Game) {
	g.ContinuousEffects = filterContinuousEffects(g.ContinuousEffects, func(effect *game.ContinuousEffect) bool {
		return effect.Duration == game.DurationUntilEndOfTurn || effect.Duration == game.DurationThisTurn
	})
}

// expireConditionalControlDurations removes continuous effects whose duration
// is tied to a runtime condition: the source permanent's presence on the
// battlefield, the effect controller's continued control of the source, or the
// affected creature remaining enchanted.  It is called at state-based-action
// cadence so that stale effects are removed before legality and selector
// decisions are made.  Returns true when at least one effect was removed.
func expireConditionalControlDurations(g *game.Game) bool {
	expired := func(effect *game.ContinuousEffect) bool {
		switch effect.Duration {
		case game.DurationForAsLongAsSourceOnBattlefield:
			// Phasing makes the source nonexistent, ending the duration even
			// though its identity remains stored on the battlefield.
			source, onBattlefield := permanentByObjectID(g, effect.SourceObjectID)
			return !onBattlefield || !activeBattlefieldPermanent(source)
		case game.DurationForAsLongAsYouControlSource:
			// Expire when the source is gone or no longer controlled by the
			// effect's controller.
			src, onBattlefield := permanentByObjectID(g, effect.SourceObjectID)
			if !onBattlefield || !activeBattlefieldPermanent(src) {
				return true
			}
			return effectiveController(g, src) != effect.Controller
		case game.DurationForAsLongAsControlledCreatureEnchanted:
			// Expire when the affected creature has left the battlefield or is
			// no longer enchanted (has no Aura attached).
			affected, onBattlefield := permanentByObjectID(g, effect.AffectedObjectID)
			if !onBattlefield {
				return true
			}
			return !permanentIsEnchanted(g, affected)
		}
		return false
	}
	before := len(g.ContinuousEffects)
	g.ContinuousEffects = filterContinuousEffects(g.ContinuousEffects, expired)
	return len(g.ContinuousEffects) < before
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
	ability := game.TriggeredAbility{
		Optional: def.Optional,
		Content:  def.Content,
	}
	g.DelayedTriggers = append(g.DelayedTriggers, game.DelayedTrigger{
		ID:                          g.IDGen.Next(),
		SourceID:                    sourceID,
		SourceObjectID:              sourceObjectID,
		SourceTokenDef:              obj.SourceTokenDef,
		Controller:                  obj.Controller,
		CreatedTurn:                 g.Turn.TurnNumber,
		Timing:                      def.Timing,
		Ability:                     ability,
		CapturedTargetControllerLKI: clonePlayerIDMap(obj.TargetControllerLKI),
		CapturedTargetManaValueLKI:  cloneIntMap(obj.TargetManaValueLKI),
	})
	return true
}

func drainReadyDelayedTriggers(g *game.Game, events []game.Event) []pendingTriggeredAbility {
	if len(g.DelayedTriggers) == 0 {
		return nil
	}
	timing := delayedTriggerTimingForStepBoundary(events)
	if timing == 0 {
		return nil
	}
	remaining := g.DelayedTriggers[:0]
	var pending []pendingTriggeredAbility
	for i := range g.DelayedTriggers {
		trigger := &g.DelayedTriggers[i]
		if trigger.Timing != timing ||
			timing == game.DelayedAtBeginningOfNextUpkeep &&
				trigger.CreatedTurn >= g.Turn.TurnNumber ||
			timing == game.DelayedAtBeginningOfNextMainPhase &&
				trigger.Controller != g.Turn.ActivePlayer {
			remaining = append(remaining, *trigger)
			continue
		}
		ability := trigger.Ability
		pending = append(pending, pendingTriggeredAbility{
			controller:                  trigger.Controller,
			sourceID:                    trigger.SourceObjectID,
			sourceCardID:                trigger.SourceID,
			sourceToken:                 trigger.SourceTokenDef,
			inline:                      &ability,
			capturedTargetControllerLKI: clonePlayerIDMap(trigger.CapturedTargetControllerLKI),
			capturedTargetManaValueLKI:  cloneIntMap(trigger.CapturedTargetManaValueLKI),
		})
	}
	g.DelayedTriggers = remaining
	return pending
}

func delayedTriggerTimingForStepBoundary(events []game.Event) game.DelayedTriggerTiming {
	for i := range events {
		event := &events[i]
		if event.Kind != game.EventBeginningOfStep {
			continue
		}
		switch event.Step {
		case game.StepUpkeep:
			return game.DelayedAtBeginningOfNextUpkeep
		case game.StepEnd:
			return game.DelayedAtBeginningOfNextEndStep
		case game.StepPrecombatMain, game.StepPostcombatMain:
			return game.DelayedAtBeginningOfNextMainPhase
		}
	}
	return 0
}

func clonePlayerIDMap(source map[int]game.PlayerID) map[int]game.PlayerID {
	if len(source) == 0 {
		return nil
	}
	clone := make(map[int]game.PlayerID, len(source))
	maps.Copy(clone, source)
	return clone
}

func cloneIntMap(source map[int]int) map[int]int {
	if len(source) == 0 {
		return nil
	}
	clone := make(map[int]int, len(source))
	maps.Copy(clone, source)
	return clone
}
