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

func TestPrimitiveRegistryDispatchPanicsOnUnregisteredKind(t *testing.T) {
	reg := &primitiveRegistry{}

	defer func() {
		if recover() == nil {
			t.Fatal("dispatch did not panic on unregistered kind")
		}
	}()

	_ = reg.dispatch(game.PrimitiveDamage)
}
