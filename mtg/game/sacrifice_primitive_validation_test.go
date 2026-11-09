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

func TestSacrificePermanentsExtremumValidation(t *testing.T) {
	t.Parallel()
	creature := Selection{RequiredTypes: []types.Card{types.Creature}}
	for _, test := range []struct {
		name  string
		prim  SacrificePermanents
		valid bool
	}{
		{name: "greatest power", valid: true, prim: SacrificePermanents{
			PlayerGroup: OpponentsReference(), Amount: Fixed(1), Selection: creature, Extremum: PermanentChoiceGreatestPower,
		}},
		{name: "unknown extremum", prim: SacrificePermanents{
			PlayerGroup: OpponentsReference(), Amount: Fixed(1), Selection: creature, Extremum: PermanentChoiceGreatestManaValue + 1,
		}},
		{name: "with all", prim: SacrificePermanents{
			PlayerGroup: OpponentsReference(), Selection: creature, All: true, Extremum: PermanentChoiceGreatestPower,
		}},
		{name: "with any number", prim: SacrificePermanents{
			PlayerGroup: OpponentsReference(), Selection: creature, AnyNumber: true, Extremum: PermanentChoiceGreatestPower,
		}},
		{name: "amount two", prim: SacrificePermanents{
			PlayerGroup: OpponentsReference(), Amount: Fixed(2), Selection: creature, Extremum: PermanentChoiceGreatestPower,
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateInstructionSequence([]Instruction{{Primitive: test.prim}}, nil)
			if (err == nil) != test.valid {
				t.Fatalf("ValidateInstructionSequence() = %v, want valid %v", err, test.valid)
			}
		})
	}
}
