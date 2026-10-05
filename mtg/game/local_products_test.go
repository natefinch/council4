package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/id"
)

func TestLocalProductDeclarationRequiresExactPublisher(t *testing.T) {
	for _, tc := range []struct {
		name        string
		instruction Instruction
		valid       bool
	}{
		{"owned result", Instruction{Primitive: GainLife{Player: ControllerReference(), Amount: Fixed(1)},
			PublishResult: "owned", LocalProducts: LocalProducts{Results: []ResultKey{"owned"}}}, true},
		{"wrong result", Instruction{Primitive: GainLife{Player: ControllerReference(), Amount: Fixed(1)},
			PublishResult: "other", LocalProducts: LocalProducts{Results: []ResultKey{"owned"}}}, false},
		{"owned link", Instruction{Primitive: LookAtLibraryTop{Player: ControllerReference(), PublishLinked: "owned"},
			LocalProducts: LocalProducts{Links: []LinkedKey{"owned"}}}, true},
		{"wrong link", Instruction{Primitive: LookAtLibraryTop{Player: ControllerReference(), PublishLinked: "other"},
			LocalProducts: LocalProducts{Links: []LinkedKey{"owned"}}}, false},
		{"empty result", Instruction{Primitive: GainLife{Player: ControllerReference(), Amount: Fixed(1)},
			LocalProducts: LocalProducts{Results: []ResultKey{""}}}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := ValidateInstructionSequence([]Instruction{tc.instruction}); (err == nil) != tc.valid {
				t.Fatalf("validation=%v, want valid=%t", err, tc.valid)
			}
		})
	}
}

func TestLocalProductAddressesAreNotAliasedByStackCopy(t *testing.T) {
	original := &StackObject{ID: id.ID(1), LocalLinkedProducts: map[LinkedKey]LinkedObjectKey{
		"owned": {SourceID: id.ID(2), LinkID: "owned", ResolutionScope: id.ID(2)},
	}}
	copied := NewStackObjectCopy(original, id.ID(3))
	delete(copied.LocalLinkedProducts, "owned")
	if len(original.LocalLinkedProducts) != 1 {
		t.Fatal("stack copy shared mutable frame addresses")
	}
}
