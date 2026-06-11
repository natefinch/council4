package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
)

func TestSacrificePermanentsPrimitiveValidationRejectsContradictorySelection(t *testing.T) {
	t.Parallel()
	err := ValidateInstructionSequence([]Instruction{{
		Primitive: SacrificePermanents{
			Player: ControllerReference(),
			Amount: Fixed(1),
			Selection: Selection{
				RequiredTypes: []types.Card{types.Creature},
				ExcludedTypes: []types.Card{types.Creature},
			},
		},
	}}, nil)
	if err == nil {
		t.Fatal("ValidateInstructionSequence() = nil")
	}
}
