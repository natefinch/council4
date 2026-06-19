package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestRegisterPrimitiveHandlerPanicsOnDuplicate(t *testing.T) {
	reg := &primitiveRegistry{}
	registerPrimitiveHandler(reg, handleDraw)

	defer func() {
		if recover() == nil {
			t.Fatal("registerPrimitiveHandler did not panic on duplicate registration")
		}
	}()

	registerPrimitiveHandler(reg, handleDraw)
}

func TestPrimitiveRegistryDispatchPanicsWithUnsupportedError(t *testing.T) {
	reg := &primitiveRegistry{}

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("dispatch did not panic on unregistered kind")
		}
		u, ok := r.(UnsupportedError)
		if !ok {
			t.Fatalf("dispatch panicked with %T, want rules.UnsupportedError", r)
		}
		if u.Kind != game.PrimitiveDamage {
			t.Errorf("UnsupportedError.Kind = %d, want %d", u.Kind, game.PrimitiveDamage)
		}
		if u.Error() == "" {
			t.Error("UnsupportedError.Error() is empty")
		}
	}()

	_ = reg.dispatch(game.PrimitiveDamage)
}

func TestPrimitiveRegistryDispatchPanicsOnOutOfRangeKind(t *testing.T) {
	reg := &primitiveRegistry{}

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("dispatch did not panic on out-of-range kind")
		}
		if _, ok := r.(UnsupportedError); !ok {
			t.Fatalf("dispatch panicked with %T, want rules.UnsupportedError", r)
		}
	}()

	_ = reg.dispatch(game.PrimitiveKind(game.PrimitiveKindCount + 100))
}
