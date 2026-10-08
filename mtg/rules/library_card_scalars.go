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
	case game.PrimitiveMoveCard:
		return primitive.(game.MoveCard).PublishDepartureCharacteristics
	default:
		return game.LibraryCardCharacteristics{}
	}
}

func (r *effectResolver) clearLibraryCardScalars(outputs game.LibraryCardCharacteristics) {
	if r.obj == nil {
		return
	}
	for _, key := range []game.ResultKey{outputs.Power, outputs.Toughness, outputs.ManaValue} {
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
	if outputs.Power == "" && outputs.Toughness == "" && outputs.ManaValue == "" {
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
	r.publishCardScalars(card, outputs, false)
}

func (r *effectResolver) publishCardScalars(card *game.CardInstance, outputs game.LibraryCardCharacteristics, departure bool) {
	power, powerKnown := cardFacePT(card, func(face game.CardFace) opt.V[game.PT] { return face.Power })
	toughness, toughnessKnown := cardFacePT(card, func(face game.CardFace) opt.V[game.PT] { return face.Toughness })
	if departure && card != nil && card.Def != nil {
		face := card.Def.DefaultFace()
		powerKnown = powerKnown || !face.Power.Exists
		toughnessKnown = toughnessKnown || !face.Toughness.Exists
	}
	publish := func(key game.ResultKey, value int, known bool) {
		if key != "" && known {
			// Reading a known characteristic succeeds even when its value is zero.
			recordResultKey(r.obj, key, effectResolved{accepted: true, succeeded: true, amount: value})
		}
	}
	publish(outputs.Power, power, powerKnown)
	publish(outputs.Toughness, toughness, toughnessKnown)
	subject := selectionSubject{kind: subjectCard, card: card}
	manaValue, manaValueKnown := subject.manaValue()
	publish(outputs.ManaValue, manaValue, manaValueKnown)
}
