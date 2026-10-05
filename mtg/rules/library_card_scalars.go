package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

func libraryCardCharacteristicOutputs(primitive game.Primitive) game.LibraryCardCharacteristics {
	if primitive == nil {
		return game.LibraryCardCharacteristics{}
	}
	switch primitive.Kind() {
	case game.PrimitiveLookAtLibraryTop:
		return primitive.(game.LookAtLibraryTop).PublishCharacteristics
	case game.PrimitiveReveal:
		return primitive.(game.Reveal).PublishCharacteristics
	default:
		return game.LibraryCardCharacteristics{}
	}
}

func (r *effectResolver) clearLibraryCardScalars(outputs game.LibraryCardCharacteristics) {
	if r.obj == nil {
		return
	}
	for _, key := range []game.ResultKey{outputs.Power, outputs.Toughness} {
		if key == "" {
			continue
		}
		delete(r.obj.ResolutionResults, string(key))
		delete(r.obj.ResolutionResultObjects, string(key))
		delete(r.obj.ResolvedAmounts, string(key))
		delete(r.obj.ResolvedExcessDamage, string(key))
	}
}

func (r *effectResolver) publishLibraryCardScalars(link game.LinkedKey, outputs game.LibraryCardCharacteristics) {
	if outputs.Power == "" && outputs.Toughness == "" {
		return
	}
	observed := linkedObjects(r.game, linkedObjectSourceKey(r.game, r.obj, string(link)))
	if len(observed) != 1 || observed[0].ObjectID != 0 || !observed[0].CardZoneVersionSet {
		return
	}
	card, available := linkedCardInstance(r.game, observed[0])
	if !available {
		return
	}
	power, powerKnown := cardFacePT(card, func(face game.CardFace) opt.V[game.PT] { return face.Power })
	toughness, toughnessKnown := cardFacePT(card, func(face game.CardFace) opt.V[game.PT] { return face.Toughness })
	publish := func(key game.ResultKey, value int, known bool) {
		if key != "" && known {
			// Reading a known characteristic succeeds even when its value is zero.
			recordResultKey(r.obj, key, effectResolved{accepted: true, succeeded: true, amount: value})
		}
	}
	publish(outputs.Power, power, powerKnown)
	publish(outputs.Toughness, toughness, toughnessKnown)
}
