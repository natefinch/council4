package rules

import (
	"maps"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
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
			if !onBattlefield || !activeBattlefieldPermanent(affected) {
				return true
			}
			return !permanentIsEnchanted(g, affected)
		}
		return false
	}
	changed := false
	for {
		before := len(g.ContinuousEffects)
		g.ContinuousEffects = filterContinuousEffects(g.ContinuousEffects, expired)
		if len(g.ContinuousEffects) == before {
			return changed
		}
		changed = true
	}
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
	if obj == nil || def == nil || (def.Timing == 0 && !def.EventPattern.Exists) {
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
		EventPattern:                def.EventPattern,
		OneShot:                     def.OneShot,
		Window:                      def.Window,
		CapturedTargetControllerLKI: clonePlayerIDMap(obj.TargetControllerLKI),
		CapturedTargetManaValueLKI:  cloneIntMap(obj.TargetManaValueLKI),
		BoundDamageSourceObjectID:   capturedDamageSourceObjectID(g, obj, def),
	})
	return true
}

// capturedDamageSourceObjectID resolves the permanent a combat-damage delayed
// trigger binds to from the creating ability's DamageSourceObject reference, so
// the scheduled trigger fires only on combat damage dealt by that specific
// object ("... target creature ... Whenever that creature deals combat damage to
// a player this turn, ..."). It returns zero when the definition carries no such
// reference or the captured permanent is already gone, in which case the trigger
// never fires.
func capturedDamageSourceObjectID(g *game.Game, obj *game.StackObject, def *game.DelayedTriggerDef) id.ID {
	if !def.DamageSourceObject.Exists {
		return 0
	}
	reference := def.DamageSourceObject.Val
	if reference.Kind() != game.ObjectReferenceLinkedObject {
		return 0
	}
	for _, linked := range linkedObjects(g, linkedObjectSourceKey(g, obj, reference.LinkID())) {
		if linked.ObjectID != 0 {
			return linked.ObjectID
		}
	}
	return 0
}

// drainReadyEventDelayedTriggers fires event-based delayed triggers whose stored
// event pattern matches one of the freshly emitted events, reusing the ordinary
// triggered-ability matcher bound to the trigger's stored controller. A one-shot
// trigger ("the next time you cast ...") is removed once it fires; a repeating
// trigger ("whenever you cast ... this turn") stays until its window ends. Each
// trigger fires at most once per drain even if several matching events occurred,
// retaining the first matching event as its triggering event.
func drainReadyEventDelayedTriggers(g *game.Game, events []game.Event) []pendingTriggeredAbility {
	if len(g.DelayedTriggers) == 0 || len(events) == 0 {
		return nil
	}
	remaining := g.DelayedTriggers[:0]
	var pending []pendingTriggeredAbility
	for i := range g.DelayedTriggers {
		trigger := &g.DelayedTriggers[i]
		if !trigger.EventPattern.Exists {
			remaining = append(remaining, *trigger)
			continue
		}
		matched, matchEvent := matchEventDelayedTrigger(g, trigger, events)
		if !matched {
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
			event:                       matchEvent,
			hasEvent:                    true,
			capturedTargetControllerLKI: clonePlayerIDMap(trigger.CapturedTargetControllerLKI),
			capturedTargetManaValueLKI:  cloneIntMap(trigger.CapturedTargetManaValueLKI),
		})
		if !trigger.OneShot {
			remaining = append(remaining, *trigger)
		}
	}
	g.DelayedTriggers = remaining
	return pending
}

// matchEventDelayedTrigger reports whether any freshly emitted event satisfies an
// event-based delayed trigger's pattern, returning the first match. The trigger's
// stored controller drives controller-relative pattern filters so "you cast a
// spell" stays bound to the trigger's controller regardless of the creating
// permanent's current state. The source permanent, when still present, supplies
// object identity for self-referential filters.
func matchEventDelayedTrigger(g *game.Game, trigger *game.DelayedTrigger, events []game.Event) (bool, game.Event) {
	pattern := trigger.EventPattern.Val
	source, _ := permanentByObjectID(g, trigger.SourceObjectID)
	for i := range events {
		if pattern.DamageSourceCaptured &&
			(trigger.BoundDamageSourceObjectID == 0 ||
				events[i].SourceObjectID != trigger.BoundDamageSourceObjectID) {
			continue
		}
		if triggerMatchesEventForController(g, source, trigger.Controller, &pattern, events[i]) {
			return true, events[i]
		}
	}
	return false, game.Event{}
}

// expireEventDelayedTriggers removes event-based delayed triggers whose
// this-turn window has ended. It runs during the cleanup step, so a "whenever
// you cast a spell this turn" rider stops firing once its turn is over.
func expireEventDelayedTriggers(g *game.Game) {
	if len(g.DelayedTriggers) == 0 {
		return
	}
	kept := g.DelayedTriggers[:0]
	for i := range g.DelayedTriggers {
		trigger := &g.DelayedTriggers[i]
		if trigger.EventPattern.Exists && trigger.Window == game.DelayedWindowThisTurn {
			continue
		}
		kept = append(kept, *trigger)
	}
	g.DelayedTriggers = kept
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
		case game.StepEndOfCombat:
			return game.DelayedAtEndOfCombat
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
