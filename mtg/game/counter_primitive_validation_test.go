package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/counter"
)

func TestCounterPlacementPrimitiveValidation(t *testing.T) {
	t.Parallel()
	permanentTargets := []TargetSpec{{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      TargetAllowPermanent,
	}}
	playerTargets := []TargetSpec{{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      TargetAllowPlayer,
	}}
	legacyPermanentTargets := []TargetSpec{{
		MinTargets: 1,
		MaxTargets: 1,
		Constraint: "target creature",
	}}
	legacyPlayerTargets := []TargetSpec{{
		MinTargets: 1,
		MaxTargets: 1,
		Constraint: "target opponent",
	}}
	valid := []struct {
		name      string
		primitive Primitive
		targets   []TargetSpec
	}{
		{
			"permanent",
			AddCounter{Amount: Fixed(1), Object: TargetPermanentReference(0), CounterKind: counter.Charge},
			permanentTargets,
		},
		{
			"player",
			AddPlayerCounter{Amount: Fixed(1), Player: TargetPlayerReference(0), CounterKind: counter.Poison},
			playerTargets,
		},
		{
			"dynamic source",
			AddPlayerCounter{
				Amount: Dynamic(DynamicAmount{
					Kind:   DynamicAmountObjectPower,
					Object: SourcePermanentReference(),
				}),
				Player:      TargetPlayerReference(0),
				CounterKind: counter.Energy,
			},
			playerTargets,
		},
		{
			"legacy permanent",
			AddCounter{Amount: Fixed(1), Object: TargetPermanentReference(0), CounterKind: counter.Charge},
			legacyPermanentTargets,
		},
		{
			"legacy player",
			AddPlayerCounter{Amount: Fixed(1), Player: TargetPlayerReference(0), CounterKind: counter.Poison},
			legacyPlayerTargets,
		},
	}
	for _, test := range valid {
		if err := ValidateInstructionSequence(
			[]Instruction{{Primitive: test.primitive}},
			test.targets,
		); err != nil {
			t.Fatalf("%s: ValidateInstructionSequence() = %v", test.name, err)
		}
	}

	invalid := []struct {
		name      string
		primitive Primitive
		targets   []TargetSpec
	}{
		{
			"player kind on permanent",
			AddCounter{Amount: Fixed(1), Object: TargetPermanentReference(0), CounterKind: counter.Energy},
			permanentTargets,
		},
		{
			"permanent kind on player",
			AddPlayerCounter{Amount: Fixed(1), Player: TargetPlayerReference(0), CounterKind: counter.Charge},
			playerTargets,
		},
		{
			"unknown kind",
			AddCounter{Amount: Fixed(1), Object: TargetPermanentReference(0), CounterKind: counter.Kind(1000)},
			permanentTargets,
		},
		{
			"zero",
			AddCounter{Amount: Fixed(0), Object: TargetPermanentReference(0), CounterKind: counter.Charge},
			permanentTargets,
		},
		{
			"negative",
			AddPlayerCounter{Amount: Fixed(-1), Player: TargetPlayerReference(0), CounterKind: counter.Poison},
			playerTargets,
		},
		{
			"target out of range",
			AddPlayerCounter{Amount: Fixed(1), Player: TargetPlayerReference(1), CounterKind: counter.Poison},
			playerTargets,
		},
		{
			"player reference to permanent target",
			AddPlayerCounter{Amount: Fixed(1), Player: TargetPlayerReference(0), CounterKind: counter.Poison},
			permanentTargets,
		},
		{
			"permanent reference to player target",
			AddCounter{Amount: Fixed(1), Object: TargetPermanentReference(0), CounterKind: counter.Charge},
			playerTargets,
		},
		{
			"permanent reference to legacy player target",
			AddCounter{Amount: Fixed(1), Object: TargetPermanentReference(0), CounterKind: counter.Charge},
			legacyPlayerTargets,
		},
		{
			"player reference to legacy permanent target",
			AddPlayerCounter{Amount: Fixed(1), Player: TargetPlayerReference(0), CounterKind: counter.Poison},
			legacyPermanentTargets,
		},
		{
			"permanent reference to mixed target",
			AddCounter{Amount: Fixed(1), Object: TargetPermanentReference(0), CounterKind: counter.Charge},
			[]TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      TargetAllowPermanent | TargetAllowPlayer,
			}},
		},
		{
			"player reference to mixed target",
			AddPlayerCounter{Amount: Fixed(1), Player: TargetPlayerReference(0), CounterKind: counter.Poison},
			[]TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      TargetAllowPermanent | TargetAllowPlayer,
			}},
		},
		{
			"permanent reference to unknown target domain",
			AddCounter{Amount: Fixed(1), Object: TargetPermanentReference(0), CounterKind: counter.Charge},
			[]TargetSpec{{MinTargets: 1, MaxTargets: 1}},
		},
	}
	for _, test := range invalid {
		if err := ValidateInstructionSequence(
			[]Instruction{{Primitive: test.primitive}},
			test.targets,
		); err == nil {
			t.Fatalf("%s: ValidateInstructionSequence() = nil", test.name)
		}
	}
}

func TestAddPlayerCounterInstructionReferences(t *testing.T) {
	t.Parallel()
	err := ValidateInstructionSequence([]Instruction{{
		Primitive: AddPlayerCounter{
			Amount: Dynamic(DynamicAmount{
				Kind:      DynamicAmountPreviousEffectResult,
				ResultKey: ResultKey("missing"),
			}),
			Player:      ControllerReference(),
			CounterKind: counter.Energy,
		},
	}}, nil)
	if err == nil {
		t.Fatal("missing previous result reference was accepted")
	}
}

func TestCounterObjectPrimitiveValidation(t *testing.T) {
	t.Parallel()
	counter := CounterObject{Object: TargetStackObjectReference(0)}
	stackTarget := []TargetSpec{{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      TargetAllowStackObject,
		Predicate:  TargetPredicate{StackObjectKinds: []StackObjectKind{StackSpell}},
	}}
	if err := ValidateInstructionSequence([]Instruction{{Primitive: counter}}, stackTarget); err != nil {
		t.Fatalf("stack target: ValidateInstructionSequence() = %v", err)
	}

	for _, targets := range [][]TargetSpec{
		{{MinTargets: 1, MaxTargets: 1, Allow: TargetAllowPermanent}},
		{{MinTargets: 1, MaxTargets: 1, Allow: TargetAllowPermanent | TargetAllowStackObject}},
	} {
		if err := ValidateInstructionSequence([]Instruction{{Primitive: counter}}, targets); err == nil {
			t.Fatalf("incompatible target %+v: ValidateInstructionSequence() = nil", targets[0])
		}
	}
}
