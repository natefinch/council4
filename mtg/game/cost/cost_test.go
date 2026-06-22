package cost

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/mana"
)

func TestManaMultiplyPreservesExactRequirements(t *testing.T) {
	t.Parallel()
	base := Mana{O(1), U, HybridMana(mana.W, mana.B), C}
	got := base.Multiply(2)
	want := Mana{O(2), U, HybridMana(mana.W, mana.B), C, U, HybridMana(mana.W, mana.B), C}
	if !slices.Equal(got, want) {
		t.Fatalf("%s.Multiply(2) = %s; want %s", base, got, want)
	}
	if !slices.Equal(base, Mana{O(1), U, HybridMana(mana.W, mana.B), C}) {
		t.Fatalf("Multiply mutated base cost: %s", base)
	}
}

func TestManaMultiplyBoundaryCounts(t *testing.T) {
	t.Parallel()
	if got, want := (Mana{O(2)}).Multiply(3), (Mana{O(6)}); !slices.Equal(got, want) {
		t.Fatalf("{2}.Multiply(3) = %s; want %s", got, want)
	}
	if got, want := (Mana{O(0)}).Multiply(3), (Mana{O(0)}); !slices.Equal(got, want) {
		t.Fatalf("{0}.Multiply(3) = %#v; want %s", got, want)
	}
	for _, count := range []int{-1, 0} {
		if got, want := (Mana{U}).Multiply(count), (Mana{O(0)}); !slices.Equal(got, want) {
			t.Fatalf("{U}.Multiply(%d) = %s; want %s", count, got, want)
		}
	}
}

func TestPhyrexianGenericSymbol(t *testing.T) {
	symbol := PhyrexianGeneric(2)
	if symbol.Kind != PhyrexianGenericSymbol {
		t.Fatalf("kind = %v, want PhyrexianGenericSymbol", symbol.Kind)
	}
	if symbol.Generic != 2 {
		t.Fatalf("generic = %d, want 2", symbol.Generic)
	}
	if got := symbol.String(); got != "{2/P}" {
		t.Fatalf("String() = %q, want {2/P}", got)
	}
	manaCost := Mana{PhyrexianGeneric(2)}
	if got := manaCost.ManaValue(); got != 2 {
		t.Fatalf("ManaValue() = %d, want 2", got)
	}
}
