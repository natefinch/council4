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
