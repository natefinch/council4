package cardgen

import (
	"fmt"

	"github.com/natefinch/council4/mtg/game"
)

// assertPrimitive narrows a game.Primitive to its concrete type T. The renderer
// dispatches on primitive.Kind() and then needs the concrete value; every
// dispatch arm previously inlined the same
// `value, ok := primitive.(game.T); if !ok { return "", errors.New(...) }`
// block. assertPrimitive centralizes that paired assertion so the explicit
// per-kind dispatch stays visible at each call site while the boilerplate and
// its unreachable internal-error message live in one place. The error fires only
// if a Kind/type pairing is ever inconsistent, an invariant violation the sealed
// primitive model otherwise prevents.
func assertPrimitive[T game.Primitive](primitive game.Primitive) (T, error) {
	value, ok := primitive.(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("render: internal error: primitive kind %d has unexpected concrete type %T (want %T)", primitive.Kind(), primitive, zero)
	}
	return value, nil
}
