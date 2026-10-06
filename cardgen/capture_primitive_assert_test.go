package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func captureTestPrimitive[T game.Primitive](t *testing.T, primitive game.Primitive) T {
	t.Helper()
	value, err := assertPrimitive[T](primitive)
	if err != nil {
		t.Fatal(err)
	}
	return value
}
