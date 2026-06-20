package game

import (
	"strings"
	"testing"
)

func TestValidatePonderPrimitives(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name      string
		primitive Primitive
		wantError string
	}{
		{
			name: "reorder",
			primitive: ReorderLibraryTop{
				Player: ControllerReference(),
				Amount: Fixed(3),
			},
		},
		{
			name:      "shuffle",
			primitive: ShuffleLibrary{Player: ControllerReference()},
		},
		{
			name: "zero reorder",
			primitive: ReorderLibraryTop{
				Player: ControllerReference(),
				Amount: Fixed(0),
			},
			wantError: "positive number",
		},
		{
			name:      "missing reorder player",
			primitive: ReorderLibraryTop{Amount: Fixed(3)},
			wantError: "player reference",
		},
		{
			name:      "missing shuffle player",
			primitive: ShuffleLibrary{},
			wantError: "player reference",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateInstructionSequence([]Instruction{{Primitive: test.primitive}})
			if test.wantError == "" {
				if err != nil {
					t.Fatalf("ValidateInstructionSequence() error = %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("ValidateInstructionSequence() error = %v, want containing %q", err, test.wantError)
			}
		})
	}
}
