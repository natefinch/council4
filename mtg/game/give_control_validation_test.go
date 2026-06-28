package game

import (
	"testing"

	"github.com/natefinch/council4/opt"
)

// giveControlInstruction builds a give-control ApplyContinuous instruction whose
// new controller is resolved from the supplied player reference, matching the
// shape lowerGiveControlSpell emits for "target player gains control of ...".
func giveControlInstruction(effect ContinuousEffect) []Instruction {
	return []Instruction{{
		Primitive: ApplyContinuous{
			Object:            opt.Val[ObjectReference](TargetPermanentReference(1)),
			ContinuousEffects: []ContinuousEffect{effect},
			Duration:          DurationPermanent,
		},
	}}
}

func TestGiveControlNewControllerRefValidation(t *testing.T) {
	t.Parallel()
	playerThenPermanent := []TargetSpec{
		{MinTargets: 1, MaxTargets: 1, Allow: TargetAllowPlayer},
		{MinTargets: 1, MaxTargets: 1, Allow: TargetAllowPermanent},
	}

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		err := ValidateInstructionSequence(giveControlInstruction(ContinuousEffect{
			Layer:            LayerControl,
			NewControllerRef: opt.Val(TargetPlayerReference(0)),
		}), playerThenPermanent)
		if err != nil {
			t.Fatalf("ValidateInstructionSequence() = %v, want nil", err)
		}
	})

	t.Run("out-of-bounds player target", func(t *testing.T) {
		t.Parallel()
		err := ValidateInstructionSequence(giveControlInstruction(ContinuousEffect{
			Layer:            LayerControl,
			NewControllerRef: opt.Val(TargetPlayerReference(5)),
		}), playerThenPermanent)
		if err == nil {
			t.Fatal("ValidateInstructionSequence() = nil, want out-of-bounds error")
		}
	})

	t.Run("wrong layer", func(t *testing.T) {
		t.Parallel()
		err := ValidateInstructionSequence(giveControlInstruction(ContinuousEffect{
			Layer:            LayerPowerToughnessModify,
			NewControllerRef: opt.Val(TargetPlayerReference(0)),
		}), playerThenPermanent)
		if err == nil {
			t.Fatal("ValidateInstructionSequence() = nil, want control-layer error")
		}
	})

	t.Run("mutually exclusive with NewController", func(t *testing.T) {
		t.Parallel()
		err := ValidateInstructionSequence(giveControlInstruction(ContinuousEffect{
			Layer:            LayerControl,
			NewController:    opt.Val(Player1),
			NewControllerRef: opt.Val(TargetPlayerReference(0)),
		}), playerThenPermanent)
		if err == nil {
			t.Fatal("ValidateInstructionSequence() = nil, want mutual-exclusion error")
		}
	})
}
