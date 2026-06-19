package id

import "testing"

func TestRestoreCopiesCounterAndIsIndependent(t *testing.T) {
	var gen Generator
	for range 5 {
		gen.Next()
	}

	var clone Generator
	clone.Restore(gen.Current())
	if got := clone.Current(); got != 5 {
		t.Fatalf("clone counter = %d, want 5", got)
	}

	// Both produce the same next ID because they share the same counter value.
	if got, want := gen.Next(), clone.Next(); got != want {
		t.Fatalf("next IDs differ after clone: %d != %d", got, want)
	}

	// Advancing the clone must not advance the original.
	clone.Next()
	clone.Next()
	if got := gen.Current(); got != 6 {
		t.Fatalf("original counter = %d after advancing clone, want 6", got)
	}
	if got := clone.Current(); got != 8 {
		t.Fatalf("clone counter = %d, want 8", got)
	}
}

func TestRestoreToZero(t *testing.T) {
	var gen Generator
	var clone Generator
	clone.Restore(gen.Current())
	if got := clone.Current(); got != 0 {
		t.Fatalf("fresh clone counter = %d, want 0", got)
	}
	if got := clone.Next(); got != 1 {
		t.Fatalf("first ID from fresh clone = %d, want 1", got)
	}
}
