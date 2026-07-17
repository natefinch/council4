package game

import (
	"testing"

	"github.com/natefinch/council4/opt"
)

func TestDelayedTriggerEventPlayerValidatesTargetReference(t *testing.T) {
	primitive := CreateDelayedTrigger{Trigger: DelayedTriggerDef{
		EventPattern: opt.Val(TriggerPattern{Event: EventDamageDealt}),
		EventPlayer:  opt.Val(TargetPlayerReference(0)),
		Window:       DelayedWindowThisTurn,
		Content:      Mode{Sequence: []Instruction{{Primitive: Draw{Amount: Fixed(1), Player: ControllerReference()}}}}.Ability(),
	}}
	playerTargets := []TargetSpec{{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      TargetAllowPlayer,
	}}

	if err := primitive.validatePrimitive(playerTargets, true); err != nil {
		t.Fatalf("valid player-bound delayed trigger: %v", err)
	}
	if err := primitive.validatePrimitive(nil, true); err == nil {
		t.Fatal("player-bound delayed trigger accepted a missing target specification")
	}
	permanentTargets := []TargetSpec{{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      TargetAllowPermanent,
	}}
	if err := primitive.validatePrimitive(permanentTargets, true); err == nil {
		t.Fatal("player-bound delayed trigger accepted a permanent target specification")
	}

	primitive.Trigger.EventPattern = opt.V[TriggerPattern]{}
	primitive.Trigger.Timing = DelayedAtBeginningOfNextEndStep
	primitive.Trigger.Window = DelayedWindowNone
	if err := primitive.validatePrimitive(playerTargets, true); err == nil {
		t.Fatal("fixed-phase delayed trigger accepted EventPlayer")
	}
}
