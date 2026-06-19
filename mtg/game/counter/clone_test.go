package counter

import "testing"

func TestSetCloneDeepCopiesCounts(t *testing.T) {
	original := NewSet()
	original.Add(PlusOnePlusOne, 3)
	original.Add(Loyalty, 4)

	clone := original.Clone()
	if got, want := clone.Get(PlusOnePlusOne), 3; got != want {
		t.Fatalf("clone +1/+1 = %d, want %d", got, want)
	}
	if got, want := clone.Get(Loyalty), 4; got != want {
		t.Fatalf("clone loyalty = %d, want %d", got, want)
	}

	// Mutating the clone must not affect the original.
	clone.Add(PlusOnePlusOne, 10)
	clone.Remove(Loyalty, 4)
	if got, want := original.Get(PlusOnePlusOne), 3; got != want {
		t.Fatalf("original +1/+1 = %d after mutating clone, want %d", got, want)
	}
	if got, want := original.Get(Loyalty), 4; got != want {
		t.Fatalf("original loyalty = %d after mutating clone, want %d", got, want)
	}

	// Mutating the original must not affect the clone.
	original.Add(Charge, 1)
	if got := clone.Get(Charge); got != 0 {
		t.Fatalf("clone charge = %d after mutating original, want 0", got)
	}
}

func TestSetCloneEmpty(t *testing.T) {
	var original Set
	clone := original.Clone()
	clone.Add(PlusOnePlusOne, 1)
	if !original.IsEmpty() {
		t.Fatal("original set gained counters from clone mutation")
	}
}
