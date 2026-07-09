package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/opt"
)

func combinationAddMana(colors ...mana.Color) AddMana {
	return AddMana{Amount: Fixed(len(colors)), CombinationColors: colors}
}

// TestCombinationManaValidationAcceptsColorSet confirms a well-formed combination
// AddMana (two or more distinct offered colors, no conflicting color mechanism)
// validates.
func TestCombinationManaValidationAcceptsColorSet(t *testing.T) {
	t.Parallel()
	valid := []AddMana{
		combinationAddMana(mana.R, mana.G),
		combinationAddMana(mana.U, mana.B, mana.R),
		combinationAddMana(mana.W, mana.U, mana.B, mana.R, mana.G),
	}
	for _, prim := range valid {
		if err := ValidateInstructionSequence([]Instruction{{Primitive: prim}}); err != nil {
			t.Fatalf("ValidateInstructionSequence(%v) = %v, want nil", prim.CombinationColors, err)
		}
	}
}

// TestCombinationManaValidationRejectsMalformed confirms the invariants fail
// closed: fewer than two colors, duplicate colors, and mixing the combination
// set with another color-selection mechanism are each rejected rather than
// silently mislowered.
func TestCombinationManaValidationRejectsMalformed(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		prim AddMana
	}{
		{"single color", AddMana{Amount: Fixed(1), CombinationColors: []mana.Color{mana.R}}},
		{"duplicate color", AddMana{Amount: Fixed(2), CombinationColors: []mana.Color{mana.R, mana.R}}},
		{"with fixed color", AddMana{Amount: Fixed(2), ManaColor: mana.G, CombinationColors: []mana.Color{mana.R, mana.G}}},
		{"with entry choice", AddMana{Amount: Fixed(2), EntryChoiceFrom: "k", CombinationColors: []mana.Color{mana.R, mana.G}}},
		{"with spend rider", AddMana{
			Amount:            Fixed(2),
			CombinationColors: []mana.Color{mana.R, mana.G},
			SpendRider: opt.Val(ManaSpendRider{
				Condition: ManaSpendCastCommanderCreatureType,
				Effect:    Mode{Sequence: []Instruction{{Primitive: Scry{Amount: Fixed(1), Player: ControllerReference()}}}},
			}),
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := ValidateInstructionSequence([]Instruction{{Primitive: tc.prim}}); err == nil {
				t.Fatalf("ValidateInstructionSequence(%s) = nil, want error", tc.name)
			}
		})
	}
}
