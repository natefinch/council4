package cardgen

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// movePermanentTo returns primitive as a game.MovePermanent when it moves a
// permanent to destination.
//
// Asserting the primitive type alone used to be enough to identify a family
// member, because each destination had its own type (game.MovePermanent, game.MovePermanent,
// game.MovePermanent). Now that the destination is a field, every site
// that recognizes "an exile" or "a bounce" must check where the permanent goes,
// or an exile-shaped matcher would silently accept a bounce.
func movePermanentTo(primitive game.Primitive, destination zone.Type) (game.MovePermanent, bool) {
	move, ok := primitive.(game.MovePermanent)
	if !ok || move.Destination != destination {
		return game.MovePermanent{}, false
	}
	return move, true
}

// isMovePermanentTo reports whether primitive moves a permanent to destination.
func isMovePermanentTo(primitive game.Primitive, destination zone.Type) bool {
	_, ok := movePermanentTo(primitive, destination)
	return ok
}
