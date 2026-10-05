package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func captureTestPrimitive[T game.Primitive](t *testing.T, primitive game.Primitive) T {
	t.Helper()
	value, ok := primitive.(T)
	if !ok {
		t.Fatalf("primitive = %T, want %T", primitive, value)
	}
	return value
}
