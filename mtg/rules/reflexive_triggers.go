package rules

import (
	"github.com/natefinch/council4/mtg/game"
)

// queueReflexiveTrigger enqueues a reflexive triggered ability (CR 603.11) from
// the resolving object. It is called by the CreateReflexiveTrigger handler,
// whose instruction is gated on the enabling action's published result, so the
// trigger is queued only when the enabling action was performed. The queued
// trigger is put on the stack by drainReadyReflexiveTriggers the next time
// triggered abilities are gathered (immediately after the enabling ability
// finishes resolving), with its targets chosen then.
func queueReflexiveTrigger(g *game.Game, obj *game.StackObject, def *game.ReflexiveTriggerDef) bool {
	if obj == nil || def == nil || len(def.Content.Modes) == 0 {
		return false
	}
	sourceID, sourceObjectID := damageSourceIDs(g, obj)
	g.PendingReflexiveTriggers = append(g.PendingReflexiveTriggers, game.ReflexiveTrigger{
		SourceID:                    sourceID,
		SourceObjectID:              sourceObjectID,
		SourceTokenDef:              obj.SourceTokenDef,
		Controller:                  obj.Controller,
		Ability:                     game.TriggeredAbility{Content: def.Content},
		CapturedTargetControllerLKI: clonePlayerIDMap(obj.TargetControllerLKI),
		CapturedTargetManaValueLKI:  cloneIntMap(obj.TargetManaValueLKI),
		TriggerEvent:                obj.TriggerEvent,
		HasTriggerEvent:             obj.HasTriggerEvent,
	})
	return true
}

// drainReadyReflexiveTriggers removes every queued reflexive trigger and returns
// it as a pending triggered ability ready to be put on the stack. Reflexive
// triggers have no timing or event window (unlike delayed triggers): each is
// always ready the first time triggered abilities are gathered after it is
// queued. It is invoked from putTriggeredAbilitiesOnStackWithChoices alongside
// the other trigger drains so the reflexive ability is ordered in APNAP order
// with any ordinary triggers from the same resolution and its targets are chosen
// as it is put on the stack.
func drainReadyReflexiveTriggers(g *game.Game) []pendingTriggeredAbility {
	if len(g.PendingReflexiveTriggers) == 0 {
		return nil
	}
	pending := make([]pendingTriggeredAbility, 0, len(g.PendingReflexiveTriggers))
	for i := range g.PendingReflexiveTriggers {
		trigger := &g.PendingReflexiveTriggers[i]
		ability := trigger.Ability
		pending = append(pending, pendingTriggeredAbility{
			controller:                  trigger.Controller,
			sourceID:                    trigger.SourceObjectID,
			sourceCardID:                trigger.SourceID,
			sourceToken:                 trigger.SourceTokenDef,
			inline:                      &ability,
			event:                       trigger.TriggerEvent,
			hasEvent:                    trigger.HasTriggerEvent,
			capturedTargetControllerLKI: clonePlayerIDMap(trigger.CapturedTargetControllerLKI),
			capturedTargetManaValueLKI:  cloneIntMap(trigger.CapturedTargetManaValueLKI),
			capturedObjectID:            trigger.CapturedObjectID,
		})
	}
	g.PendingReflexiveTriggers = nil
	return pending
}
