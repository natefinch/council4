package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/color"
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

// TestApplyContinuousAllTargetPermanentsValidation covers "one or more target
// creatures become white": ApplyContinuous accepts the whole-target-spec
// reference, bounded by the declared specs, but not with a single linked
// publish or a resolution-time group choice.
func TestApplyContinuousAllTargetPermanentsValidation(t *testing.T) {
	t.Parallel()
	specs := []TargetSpec{{MinTargets: 1, MaxTargets: 99, Allow: TargetAllowPermanent}}
	white := []ContinuousEffect{{Layer: LayerColor, SetColors: []color.Color{color.White}}}
	for _, tc := range []struct {
		name    string
		prim    ApplyContinuous
		wantErr bool
	}{
		{"valid", ApplyContinuous{Object: opt.Val(AllTargetPermanentsReference(0)), ContinuousEffects: white}, false},
		{"spec out of bounds", ApplyContinuous{Object: opt.Val(AllTargetPermanentsReference(1)), ContinuousEffects: white}, true},
		{"publish linked", ApplyContinuous{
			Object: opt.Val(AllTargetPermanentsReference(0)), ContinuousEffects: white, PublishLinked: "chosen",
		}, true},
		{"choose from", ApplyContinuous{
			Object: opt.Val(AllTargetPermanentsReference(0)), ContinuousEffects: white,
			ChooseFrom: BattlefieldGroup(Selection{}), ChooseUpTo: Fixed(1),
		}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateInstructionSequence([]Instruction{{Primitive: tc.prim}}, specs)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ValidateInstructionSequence() = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
