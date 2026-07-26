package cardgen

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// moveTopOfLibraryTo returns primitive as a game.MoveTopOfLibrary when it moves
// the top of a library to destination.
//
// Asserting the primitive type alone used to be enough to identify a family
// member, because each destination had its own type (game.Mill for the
// graveyard, game.ExileTopOfLibrary for exile). Now that the destination is a
// field, every site that recognizes "a mill" or "an exile the top" must check
// where the cards go, or a mill-shaped matcher would silently accept an
// exile-top and widen its behavior.
func moveTopOfLibraryTo(primitive game.Primitive, destination zone.Type) (game.MoveTopOfLibrary, bool) {
	move, ok := primitive.(game.MoveTopOfLibrary)
	if !ok || move.Destination != destination {
		return game.MoveTopOfLibrary{}, false
	}
	return move, true
}

// isMoveTopOfLibraryTo reports whether primitive moves the top of a library to
// destination.
func isMoveTopOfLibraryTo(primitive game.Primitive, destination zone.Type) bool {
	_, ok := moveTopOfLibraryTo(primitive, destination)
	return ok
}
