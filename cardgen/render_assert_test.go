package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestAssertPrimitiveNarrowsConcreteType proves assertPrimitive returns the
// concrete primitive value when the dynamic type matches the requested type, the
// success path every renderer dispatch arm relies on after switching on
// primitive.Kind().
func TestAssertPrimitiveNarrowsConcreteType(t *testing.T) {
	t.Parallel()
	value, err := assertPrimitive[game.Draw](game.Draw{Player: game.TargetPlayerReference(3)})
	if err != nil {
		t.Fatalf("assertPrimitive returned error for matching type: %v", err)
	}
	if value.Player != game.TargetPlayerReference(3) {
		t.Fatalf("narrowed Draw = %+v, want Player TargetPlayerReference(3)", value)
	}
}

// TestAssertPrimitiveRejectsMismatchedType proves assertPrimitive fails closed
// when the concrete type does not match the requested type, reproducing the
// unreachable internal-error guard each dispatch arm previously inlined rather
// than silently returning a zero value.
func TestAssertPrimitiveRejectsMismatchedType(t *testing.T) {
	t.Parallel()
	value, err := assertPrimitive[game.Draw](game.Discard{})
	if err == nil {
		t.Fatal("assertPrimitive accepted a mismatched concrete type")
	}
	if value.Player.Kind() != game.PlayerReferenceNone {
		t.Fatalf("mismatched assertion returned non-zero value %+v", value)
	}
}
