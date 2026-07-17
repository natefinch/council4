package game

import (
	"strings"
	"testing"
)

func TestValidateCopyStackObjectDynamicCount(t *testing.T) {
	t.Parallel()
	valid := CopyStackObject{
		Object: ResolvingStackObjectReference(),
		DynamicCount: Dynamic(DynamicAmount{
			Kind:       DynamicAmountCommanderCastCount,
			Multiplier: 1,
		}),
	}
	if err := ValidateInstructionSequence([]Instruction{{Primitive: valid}}); err != nil {
		t.Fatalf("dynamic copy count validation error = %v", err)
	}

	withFixed := valid
	withFixed.Count = 2
	if err := ValidateInstructionSequence([]Instruction{{Primitive: withFixed}}); err == nil ||
		!strings.Contains(err.Error(), "combine fixed and dynamic") {
		t.Fatalf("combined count error = %v", err)
	}

	fixedDynamicField := valid
	fixedDynamicField.DynamicCount = Fixed(2)
	if err := ValidateInstructionSequence([]Instruction{{Primitive: fixedDynamicField}}); err == nil ||
		!strings.Contains(err.Error(), "dynamic count must be dynamic") {
		t.Fatalf("fixed dynamic-count field error = %v", err)
	}
}

func TestCopyStackObjectDynamicCountValidatesDependencies(t *testing.T) {
	t.Parallel()
	copyEffect := CopyStackObject{
		Object: ResolvingStackObjectReference(),
		DynamicCount: Dynamic(DynamicAmount{
			Kind:      DynamicAmountPreviousEffectResult,
			ResultKey: "copies",
		}),
	}
	if err := ValidateInstructionSequence([]Instruction{{Primitive: copyEffect}}); err == nil ||
		!strings.Contains(err.Error(), "copies") {
		t.Fatalf("missing result-key validation error = %v", err)
	}
}
