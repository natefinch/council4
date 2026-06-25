package game

import (
	"strings"
	"testing"
)

// TestValidateCapturedTargetControllerPlayerAndQuantity proves the consolidated
// reference-then-quantity validator preserves every prior accept/reject branch:
// it accepts a valid captured-target-controller player paired with a fixed
// amount, rejects when the player reference is out of range (the reference
// branch), and rejects when the amount's captured stack-object reference is out
// of range (the quantity branch). The 13 single-player primitives that previously
// inlined this body now route through this helper, so these branches lock their
// shared behavior.
func TestValidateCapturedTargetControllerPlayerAndQuantity(t *testing.T) {
	t.Parallel()
	stackTargets := []TargetSpec{{MinTargets: 1, MaxTargets: 1, Allow: TargetAllowStackObject}}

	if err := validateCapturedTargetControllerPlayerAndQuantity(
		CapturedTargetControllerReference(0), Fixed(1), stackTargets, true,
	); err != nil {
		t.Fatalf("valid player+quantity rejected: %v", err)
	}

	err := validateCapturedTargetControllerPlayerAndQuantity(
		CapturedTargetControllerReference(1), Fixed(1), stackTargets, true,
	)
	if err == nil || !strings.Contains(err.Error(), "target index 1 has no matching target specification") {
		t.Fatalf("out-of-range player reference: err = %v, want bounds failure", err)
	}

	dynamicOutOfRange := Dynamic(DynamicAmount{
		Kind:   DynamicAmountCapturedTargetManaValue,
		Object: CapturedTargetStackObjectReference(1),
	})
	err = validateCapturedTargetControllerPlayerAndQuantity(
		TargetPlayerReference(0), dynamicOutOfRange, stackTargets, true,
	)
	if err == nil || !strings.Contains(err.Error(), "target index 1 has no matching target specification") {
		t.Fatalf("out-of-range quantity reference: err = %v, want bounds failure", err)
	}

	if err := validateCapturedTargetControllerPlayerAndQuantity(
		CapturedTargetControllerReference(5), Fixed(1), stackTargets, false,
	); err != nil {
		t.Fatalf("checkTargets=false should skip bounds: %v", err)
	}
}
