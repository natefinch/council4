package game

import (
	"strings"
	"testing"
)

// TestValidateEachTargetDamage verifies that an EachTarget damage primitive is
// accepted when its recipient names an any-target spec that admits at least one
// target, mirroring the divided-damage validation the each-of-them Comet Storm
// lowering reuses.
func TestValidateEachTargetDamage(t *testing.T) {
	specs := []TargetSpec{{
		MinTargets:               1,
		MaxTargets:               21,
		Allow:                    TargetAllowPermanent | TargetAllowPlayer,
		CountEqualsKickerPlusOne: true,
	}}
	seq := []Instruction{{Primitive: Damage{
		Amount:     Dynamic(DynamicAmount{Kind: DynamicAmountX}),
		Recipient:  AnyTargetDamageRecipient(0),
		EachTarget: true,
	}}}
	if err := ValidateInstructionSequence(seq, specs); err != nil {
		t.Fatalf("ValidateInstructionSequence() = %v, want nil", err)
	}
}

// TestValidateEachTargetDamageRejectsDividedCombo verifies a Damage primitive
// cannot set both Divided and EachTarget, since the two split-versus-full
// resolutions are mutually exclusive.
func TestValidateEachTargetDamageRejectsDividedCombo(t *testing.T) {
	specs := []TargetSpec{{MinTargets: 1, MaxTargets: 2, Allow: TargetAllowPermanent | TargetAllowPlayer}}
	seq := []Instruction{{Primitive: Damage{
		Amount:     Fixed(2),
		Recipient:  AnyTargetDamageRecipient(0),
		Divided:    true,
		EachTarget: true,
	}}}
	err := ValidateInstructionSequence(seq, specs)
	if err == nil || !strings.Contains(err.Error(), "both divided and dealt to each target") {
		t.Fatalf("ValidateInstructionSequence() = %v, want a divided/each-target conflict error", err)
	}
}

// TestValidateEachTargetDamageRequiresAnyTargetRecipient verifies an EachTarget
// damage primitive is rejected when its recipient is not an any-target spec
// reference, since the resolution enumerates the chosen targets by spec index.
func TestValidateEachTargetDamageRequiresAnyTargetRecipient(t *testing.T) {
	specs := []TargetSpec{{MinTargets: 1, MaxTargets: 2, Allow: TargetAllowPermanent}}
	seq := []Instruction{{Primitive: Damage{
		Amount:     Fixed(2),
		Recipient:  PlayerDamageRecipient(ControllerReference()),
		EachTarget: true,
	}}}
	err := ValidateInstructionSequence(seq, specs)
	if err == nil || !strings.Contains(err.Error(), "each-target damage requires an any-target recipient") {
		t.Fatalf("ValidateInstructionSequence() = %v, want an any-target recipient error", err)
	}
}
