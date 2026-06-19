package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/opt"
)

// validScryRider is the Path of Ancestry spend rider used as a valid baseline in
// these validation tests: scry 1 when the tagged mana is spent to cast a
// creature spell sharing a creature type with the commander.
func validScryRider() ManaSpendRider {
	return ManaSpendRider{
		Condition: ManaSpendCastCommanderCreatureType,
		Effect: Mode{Sequence: []Instruction{
			{Primitive: Scry{Amount: Fixed(1), Player: ControllerReference()}},
		}},
	}
}

func addManaWithRider(rider ManaSpendRider) AddMana {
	return AddMana{Amount: Fixed(1), ManaColor: mana.G, SpendRider: opt.Val(rider)}
}

// TestAddManaSpendRiderValidationAcceptsModeledRider confirms a fully modeled
// rider (recognized condition, non-empty untargeted effect) validates.
func TestAddManaSpendRiderValidationAcceptsModeledRider(t *testing.T) {
	t.Parallel()
	if err := ValidateInstructionSequence([]Instruction{{Primitive: addManaWithRider(validScryRider())}}); err != nil {
		t.Fatalf("ValidateInstructionSequence() = %v, want nil", err)
	}
}

// TestAddManaSpendRiderValidationRejectsUnknownCondition confirms the unknown
// condition value is rejected rather than treated as a no-op rider.
func TestAddManaSpendRiderValidationRejectsUnknownCondition(t *testing.T) {
	t.Parallel()
	rider := validScryRider()
	rider.Condition = ManaSpendConditionUnknown
	if err := ValidateInstructionSequence([]Instruction{{Primitive: addManaWithRider(rider)}}); err == nil {
		t.Fatal("ValidateInstructionSequence() = nil, want error for unknown condition")
	}
}

// TestAddManaSpendRiderValidationRejectsOutOfRangeCondition confirms the
// exhaustive enum switch rejects any value outside the modeled conditions, not
// just the zero unknown value.
func TestAddManaSpendRiderValidationRejectsOutOfRangeCondition(t *testing.T) {
	t.Parallel()
	rider := validScryRider()
	rider.Condition = ManaSpendCastCommanderCreatureType + 1
	if err := ValidateInstructionSequence([]Instruction{{Primitive: addManaWithRider(rider)}}); err == nil {
		t.Fatal("ValidateInstructionSequence() = nil, want error for out-of-range condition")
	}
}

// TestAddManaSpendRiderValidationRejectsEmptyEffect confirms a rider with no
// effect instructions is rejected.
func TestAddManaSpendRiderValidationRejectsEmptyEffect(t *testing.T) {
	t.Parallel()
	rider := validScryRider()
	rider.Effect = Mode{}
	if err := ValidateInstructionSequence([]Instruction{{Primitive: addManaWithRider(rider)}}); err == nil {
		t.Fatal("ValidateInstructionSequence() = nil, want error for empty rider effect")
	}
}

// TestAddManaSpendRiderValidationRejectsDeclaredTargets confirms a rider that
// declares target specs is rejected: a fired rider is put on the stack with no
// targets of its own, so it could never choose a legal target.
func TestAddManaSpendRiderValidationRejectsDeclaredTargets(t *testing.T) {
	t.Parallel()
	rider := validScryRider()
	rider.Effect.Targets = []TargetSpec{{MinTargets: 1, MaxTargets: 1}}
	if err := ValidateInstructionSequence([]Instruction{{Primitive: addManaWithRider(rider)}}); err == nil {
		t.Fatal("ValidateInstructionSequence() = nil, want error for declared rider targets")
	}
}

// TestAddManaSpendRiderValidationRejectsTargetedInstruction confirms a rider
// whose effect references a target is rejected even when it declares no target
// specs, because the sequence is validated against an empty target set.
func TestAddManaSpendRiderValidationRejectsTargetedInstruction(t *testing.T) {
	t.Parallel()
	rider := validScryRider()
	rider.Effect = Mode{Sequence: []Instruction{
		{Primitive: Destroy{Object: TargetPermanentReference(0)}},
	}}
	if err := ValidateInstructionSequence([]Instruction{{Primitive: addManaWithRider(rider)}}); err == nil {
		t.Fatal("ValidateInstructionSequence() = nil, want error for targeted rider instruction")
	}
}
