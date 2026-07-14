package game

import "testing"

// TestIterativeLibraryProcessAllowAbsentNameRequiresChosenNameStop proves the
// absent-name sentinel is scoped to the chosen-name predicate: pairing it with
// the duplicate-name stop (which names no card) is rejected.
func TestIterativeLibraryProcessAllowAbsentNameRequiresChosenNameStop(t *testing.T) {
	t.Parallel()
	err := ValidateInstructionSequence([]Instruction{{
		Primitive: IterativeLibraryProcess{
			Player:          ControllerReference(),
			Stop:            IterativeLibraryStopDuplicateName,
			OptionalTake:    true,
			AllowAbsentName: true,
		},
	}}, nil)
	if err == nil {
		t.Fatal("ValidateInstructionSequence() = nil, want error for AllowAbsentName without chosen-name stop")
	}
}

// TestIterativeLibraryProcessChosenNameAllowsAbsentName proves the Demonic
// Consultation configuration — chosen-name stop with the absent-name sentinel —
// validates cleanly.
func TestIterativeLibraryProcessChosenNameAllowsAbsentName(t *testing.T) {
	t.Parallel()
	err := ValidateInstructionSequence([]Instruction{{
		Primitive: IterativeLibraryProcess{
			Player:          ControllerReference(),
			Stop:            IterativeLibraryStopChosenName,
			PreExile:        Fixed(6),
			ChooseName:      true,
			Reveal:          true,
			AllowAbsentName: true,
		},
	}}, nil)
	if err != nil {
		t.Fatalf("ValidateInstructionSequence() = %v, want nil", err)
	}
}
