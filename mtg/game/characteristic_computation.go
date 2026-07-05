package game

import "github.com/natefinch/council4/mtg/game/id"

// BeginCharacteristicComputation marks objID's effective characteristics as being
// computed and reports whether they were ALREADY being computed higher in the
// current call stack — a characteristic-dependency loop (CR 613.8), where a
// characteristic-defining effect depends on the very characteristic it defines
// (for example a creature whose power is "the greatest power among creatures you
// control", a group that includes itself). The rules layer uses a true result to
// break the loop with base values rather than recursing without bound.
//
// Every call that returns false must be paired with EndCharacteristicComputation,
// typically via defer. The set is transient engine-computation state, not game
// state: it is empty except while a computation is in progress, mirrors
// staticFrame, and is never cloned.
func (g *Game) BeginCharacteristicComputation(objID id.ID) (alreadyComputing bool) {
	if g.computingCharacteristics[objID] {
		return true
	}
	if g.computingCharacteristics == nil {
		g.computingCharacteristics = make(map[id.ID]bool)
	}
	g.computingCharacteristics[objID] = true
	return false
}

// EndCharacteristicComputation clears the in-progress mark for objID set by a
// BeginCharacteristicComputation call that returned false.
func (g *Game) EndCharacteristicComputation(objID id.ID) {
	delete(g.computingCharacteristics, objID)
}
