package game

import (
	"slices"

	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

// UntilSourceLeavesExile records the objects one resolution of a CR 610.3
// "exile ... until <this source> leaves the battlefield" one-shot effect
// actually exiled. It is game-owned so the duration survives cloning and is
// never stored on an immutable card program.
//
// SourceObjectID is the battlefield object identity (CR 400.7) of the source
// permanent when the exile happened. A permanent that leaves and returns is a
// new object with a new ObjectID, so an absent SourceObjectID means the source
// has left the battlefield; a phased-out permanent keeps its ObjectID and has
// not left. Immediately after that happens the Engine returns every recorded
// card still in exile as the same incarnation to its previous zone, ReturnZone
// (CR 610.3): to the battlefield under its owner's control (CR 610.3c) or to
// its owner's hand, as one simultaneous event for every source that left
// together (CR 610.3d). This return is a one-shot effect, not a
// triggered ability: it never uses the stack and no player can respond to it.
type UntilSourceLeavesExile struct {
	SourceObjectID id.ID
	// Objects are the exact exiled card incarnations, captured after the move
	// with CardZoneVersionSet, so a card that has since left exile (or a token,
	// which ceased to exist) is never returned.
	Objects []LinkedObjectRef
	// ReturnZone is the zone the objects were exiled from and return to: zone.Hand
	// for hand cards, or zone.None (the zero value) or zone.Battlefield for
	// exiled permanents.
	ReturnZone zone.Type
}

func cloneUntilSourceLeavesExile(record UntilSourceLeavesExile) UntilSourceLeavesExile {
	record.Objects = slices.Clone(record.Objects)
	return record
}
