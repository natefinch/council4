package mana

import "testing"

func TestPoolCloneDeepCopiesUnits(t *testing.T) {
	original := NewPool()
	original.Add(G, 2)
	original.AddSnow(U, 1)

	clone := original.Clone()
	if got, want := clone.Amount(G), 2; got != want {
		t.Fatalf("clone green = %d, want %d", got, want)
	}
	if got, want := clone.SnowAmount(), 1; got != want {
		t.Fatalf("clone snow = %d, want %d", got, want)
	}

	// Mutating the clone must not affect the original.
	clone.Add(G, 5)
	clone.Spend(U, 1)
	if got, want := original.Amount(G), 2; got != want {
		t.Fatalf("original green = %d after mutating clone, want %d", got, want)
	}
	if got, want := original.SnowAmount(), 1; got != want {
		t.Fatalf("original snow = %d after mutating clone, want %d", got, want)
	}

	// Mutating the original must not affect the clone.
	original.Add(R, 3)
	if got := clone.Amount(R); got != 0 {
		t.Fatalf("clone red = %d after mutating original, want 0", got)
	}
}

func TestPoolCloneEmpty(t *testing.T) {
	var original Pool
	clone := original.Clone()
	clone.Add(W, 1)
	if !original.IsEmpty() {
		t.Fatal("original pool gained mana from clone mutation")
	}
}
