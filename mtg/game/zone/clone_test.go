package zone

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/id"
)

func TestCloneDeepCopiesCardsAndFaceDown(t *testing.T) {
	original := New(Exile)
	original.AddToBottom(1)
	original.AddToBottom(2)
	original.AddToBottom(3)
	original.SetFaceDown(2, true)

	clone := original.Clone()

	if !slices.Equal(clone.All(), original.All()) {
		t.Fatalf("clone cards = %v, want %v", clone.All(), original.All())
	}
	if !clone.IsFaceDown(2) {
		t.Fatal("clone lost face-down state for card 2")
	}
	if clone.Type != original.Type {
		t.Fatalf("clone type = %v, want %v", clone.Type, original.Type)
	}
}

func TestCloneIsIndependentOfOriginal(t *testing.T) {
	original := New(Library)
	original.AddToBottom(1)
	original.AddToBottom(2)
	original.SetFaceDown(1, true)

	clone := original.Clone()

	// Mutate the clone; the original must be unchanged.
	clone.AddToBottom(99)
	clone.SetFaceDown(2, true)
	clone.SetFaceDown(1, false)

	if original.Size() != 2 {
		t.Fatalf("original size = %d after mutating clone, want 2", original.Size())
	}
	if !original.IsFaceDown(1) {
		t.Fatal("original lost face-down state after mutating clone")
	}
	if original.IsFaceDown(2) {
		t.Fatal("original gained face-down state from clone mutation")
	}

	// Mutate the original; the clone must be unchanged.
	original.AddToBottom(50)
	if clone.Contains(50) {
		t.Fatal("clone observed card added to original")
	}
}

func TestCloneEmptyZone(t *testing.T) {
	original := New(Hand)
	clone := original.Clone()
	clone.AddToBottom(id.ID(7))
	if original.Size() != 0 {
		t.Fatalf("original size = %d, want 0", original.Size())
	}
}
