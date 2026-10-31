package game

import "github.com/natefinch/council4/mtg/game/id"

// BeginCharacteristicComputation marks objID's effective characteristics as being
// computed and reports whether they were ALREADY being computed higher in the
// current call stack — a characteristic-dependency loop (CR 613.8), where a
// characteristic-defining effect depends on the very characteristic it defines
// (for example a creature whose power is "the greatest power among creatures you
// control", a group that includes itself). The rules layer uses a true result to
// break the loop with the values the in-progress computation has already
// determined (see CharacteristicComputationLayer) rather than recursing without
// bound.
//
// Every call that returns false must be paired with EndCharacteristicComputation,
// typically via defer. The set is transient engine-computation state, not game
// state: it is empty except while a computation is in progress, mirrors
// staticFrame, and is never cloned.
func (g *Game) BeginCharacteristicComputation(objID id.ID) (alreadyComputing bool) {
	if _, computing := g.computingCharacteristics[objID]; computing {
		return true
	}
	if g.computingCharacteristics == nil {
		g.computingCharacteristics = make(map[id.ID]ContinuousLayer)
	}
	g.computingCharacteristics[objID] = 0
	return false
}

// EndCharacteristicComputation clears the in-progress mark for objID set by a
// BeginCharacteristicComputation call that returned false.
func (g *Game) EndCharacteristicComputation(objID id.ID) {
	delete(g.computingCharacteristics, objID)
}

// CharacteristicComputationLayer reports the layer an in-progress computation of
// objID's characteristics is applying. Zero means no layer has begun: only the
// pre-layer values are determined. ok is false when objID is not being computed.
func (g *Game) CharacteristicComputationLayer(objID id.ID) (layer ContinuousLayer, ok bool) {
	layer, ok = g.computingCharacteristics[objID]
	return layer, ok
}

// SetCharacteristicComputationLayer records that an in-progress computation of
// objID's characteristics is applying layer and returns the previously recorded
// layer. It does nothing and reports false when objID is not being computed.
func (g *Game) SetCharacteristicComputationLayer(objID id.ID, layer ContinuousLayer) (previous ContinuousLayer, ok bool) {
	previous, ok = g.computingCharacteristics[objID]
	if ok {
		g.computingCharacteristics[objID] = layer
	}
	return previous, ok
}
