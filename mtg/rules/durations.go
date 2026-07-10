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
		case game.DurationForAsLongAsPlayerIsMonarch:
			// Expire when the player whose monarchy created the effect is no
			// longer the monarch (a different player took the crown, or no
			// player is the monarch). The bound player rides ExpiresFor.
			monarch := currentMonarch(g)
			return !monarch.Exists || monarch.Val != effect.ExpiresFor
		default:
			// Other durations are not conditional-control durations.
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
	// Carry an intervening-if condition onto the scheduled ability's trigger so
	// the ordinary intervening-if machinery re-checks it both when the captured
	// event fires (drainReadyEventDelayedTriggers) and when the ability resolves
	// (resolveTriggeredAbilityBodyWithChoices), evaluated against the delayed
	// trigger's controller ("... if you control your commander, ...").
	if def.InterveningCondition.Exists {
		ability.Trigger.InterveningCondition = def.InterveningCondition
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
		BoundAttackerObjectID:       capturedAttackerObjectID(g, obj, def),
		BoundDyingObjectID:          capturedDyingObjectID(g, obj, def),
		CapturedObjectID:            capturedObjectID(g, obj, def),
	})
	return true
}

// capturedObjectID freezes the permanent a fixed-phase delayed trigger binds to
// from the creating ability's CapturedObject reference, resolving it against the
// creating ability's context at schedule time. It backs delayed "at end of
// combat" disposal of the creature involved in combat ("destroy that creature at
// end of combat"), where the original combat event is gone once the trigger
// fires, so the blocked or damaged creature must be captured now, and delayed
// disposal of a permanent an earlier clause in the same resolution published
// under a linked key ("Create a token ... Sacrifice it at the beginning of the
// next end step.", Feldon of the Third Path), where the reference is that linked
// object so each activation captures the permanent it created. It returns zero
// when the definition carries no such reference or the permanent cannot be
// identified, in which case the trigger's content finds nothing and does nothing.
func capturedObjectID(g *game.Game, obj *game.StackObject, def *game.DelayedTriggerDef) id.ID {
	if !def.CapturedObject.Exists {
		return 0
	}
	objectID, _ := newReferenceResolver(g, obj).objectIdentityID(def.CapturedObject.Val)
	return objectID
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

// capturedAttackerObjectID resolves the permanent an attacker-declared delayed
// trigger binds to from the creating ability's CapturedAttackerObject reference,
// so the scheduled trigger fires only when that specific object is declared as
// an attacker ("... target creature ... Whenever that creature attacks the
// monarch this turn, ..."). It mirrors capturedDamageSourceObjectID and returns
// zero when the definition carries no such reference or the captured permanent is
// already gone, in which case the trigger never fires.
func capturedAttackerObjectID(g *game.Game, obj *game.StackObject, def *game.DelayedTriggerDef) id.ID {
	if !def.CapturedAttackerObject.Exists {
		return 0
	}
	reference := def.CapturedAttackerObject.Val
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

// capturedDyingObjectID resolves the permanent a permanent-died delayed trigger
// binds to from the creating ability's CapturedDyingObject reference, so the
// scheduled trigger fires only when that specific object dies ("... target
// creature an opponent controls ... When the creature an opponent controls dies
// this turn, ..."). Unlike capturedDamageSourceObjectID and
// capturedAttackerObjectID, the reference is a target permanent rather than a
// linked object, so it resolves through objectIdentityID (the same path
// capturedObjectID uses). It returns zero when the definition carries no such
// reference or the captured permanent cannot be identified, in which case the
// trigger never fires.
func capturedDyingObjectID(g *game.Game, obj *game.StackObject, def *game.DelayedTriggerDef) id.ID {
	if !def.CapturedDyingObject.Exists {
		return 0
	}
	objectID, _ := newReferenceResolver(g, obj).objectIdentityID(def.CapturedDyingObject.Val)
	return objectID
}

// drainReadyEventDelayedTriggers fires event-based delayed triggers whose stored
// event pattern matches the freshly emitted events, reusing the ordinary
// triggered-ability matcher bound to the trigger's stored controller. Like an
// ordinary triggered ability (CR 603.2, one instance per matching event), a
// per-object repeating trigger ("whenever a creature dies this turn") fires once
// per matching event, so several creatures dying simultaneously fire it once
// each; the trigger stays until its window ends. A "one or more ... this turn"
// (OneOrMore) trigger instead fires once per simultaneous batch (CR 603.3e),
// matching how ordinary triggers coalesce OneOrMore pendings. A one-shot trigger
// ("the next time you cast ...") fires at most once and is then removed, even if
// several events match in the same drain; if none match it is retained to catch a
// later event.
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
		matchEvents := matchEventDelayedTriggerEvents(g, trigger, events)
		if len(matchEvents) == 0 {
			remaining = append(remaining, *trigger)
			continue
		}
		pattern := trigger.EventPattern.Val
		fired := false
		seenOneOrMoreBatch := make(map[triggerBatchKey]bool)
		for j := range matchEvents {
			matchEvent := matchEvents[j]
			// A "one or more ... this turn" delayed trigger fires once per
			// simultaneous batch, not once per event (CR 603.3e), mirroring how
			// ordinary triggered abilities coalesce OneOrMore pendings sharing a
			// SimultaneousID (coalescePendingTriggeredAbilities). Per-object
			// patterns keep the per-event fan-out below; SimultaneousID 0 and
			// distinct SimultaneousIDs remain separate firings.
			if pattern.OneOrMore && matchEvent.SimultaneousID != 0 {
				key := triggerBatchKey{
					sourceID:     trigger.SourceObjectID,
					controller:   trigger.Controller,
					event:        matchEvent.Kind,
					simultaneous: matchEvent.SimultaneousID,
				}
				if pattern.OneOrMorePerAttackTarget {
					key.attackTarget = matchEvent.AttackTarget
				}
				if seenOneOrMoreBatch[key] {
					continue
				}
				seenOneOrMoreBatch[key] = true
			}
			// An intervening-if condition is re-checked as each captured event
			// fires (CR 603.4). When it fails the ability does not trigger for
			// that event ("... if you control your commander, ...").
			if trigger.Ability.Trigger.InterveningCondition.Exists {
				source, _ := permanentByObjectID(g, trigger.SourceObjectID)
				if !triggerInterveningIf(g, source, trigger.Controller, &trigger.Ability.Trigger, &matchEvent) {
					continue
				}
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
			fired = true
			// A one-shot delayed trigger fires only once even when several
			// matching events occur simultaneously.
			if trigger.OneShot {
				break
			}
		}
		// A repeating trigger stays until its window ends; a one-shot trigger is
		// retained only if it did not fire, so it can still catch a later event.
		if !trigger.OneShot || !fired {
			remaining = append(remaining, *trigger)
		}
	}
	g.DelayedTriggers = remaining
	return pending
}

// matchEventDelayedTriggerEvents returns every freshly emitted event that
// satisfies an event-based delayed trigger's pattern, in emission order. The
// trigger's stored controller drives controller-relative pattern filters so "you
// cast a spell" stays bound to the trigger's controller regardless of the
// creating permanent's current state. The source permanent, when still present,
// supplies object identity for self-referential filters. Returning all matches
// lets the caller fire the trigger once per matching event, as an ordinary
// triggered ability would.
func matchEventDelayedTriggerEvents(g *game.Game, trigger *game.DelayedTrigger, events []game.Event) []game.Event {
	pattern := trigger.EventPattern.Val
	source, _ := permanentByObjectID(g, trigger.SourceObjectID)
	var matched []game.Event
	for i := range events {
		if pattern.DamageSourceCaptured &&
			(trigger.BoundDamageSourceObjectID == 0 ||
				events[i].SourceObjectID != trigger.BoundDamageSourceObjectID) {
			continue
		}
		if pattern.AttackerCaptured &&
			(trigger.BoundAttackerObjectID == 0 ||
				events[i].SourceObjectID != trigger.BoundAttackerObjectID) {
			continue
		}
		if pattern.DyingObjectCaptured &&
			(trigger.BoundDyingObjectID == 0 ||
				events[i].PermanentID != trigger.BoundDyingObjectID) {
			continue
		}
		if triggerMatchesEventForController(g, source, trigger.Controller, &pattern, events[i]) {
			matched = append(matched, events[i])
		}
	}
	return matched
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
			capturedObjectID:            trigger.CapturedObjectID,
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
		default:
			// Other steps do not schedule a delayed trigger; keep scanning.
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
