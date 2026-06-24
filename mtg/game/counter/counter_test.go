package counter

import "testing"

func TestAgeCounterKind(t *testing.T) {
	t.Parallel()
	if !Age.Valid() {
		t.Fatal("Age.Valid() = false; want true")
	}
	if Age.PlayerOnly() {
		t.Fatal("Age.PlayerOnly() = true; want false")
	}
	if got, want := Age.String(), "age"; got != want {
		t.Fatalf("Age.String() = %q; want %q", got, want)
	}
}

func TestLevelCounterKind(t *testing.T) {
	t.Parallel()
	if !Level.Valid() {
		t.Fatal("Level.Valid() = false; want true")
	}
	if Level.PlayerOnly() {
		t.Fatal("Level.PlayerOnly() = true; want false")
	}
	if got, want := Level.String(), "level"; got != want {
		t.Fatalf("Level.String() = %q; want %q", got, want)
	}
}
