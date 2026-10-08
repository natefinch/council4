package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

func libraryCardPublication(instr *game.Instruction) (game.LinkedKey, bool) {
	if instr.Primitive == nil {
		return "", false
	}
	switch instr.Primitive.Kind() {
	case game.PrimitiveLookAtLibraryTop:
		look, ok := instr.Primitive.(game.LookAtLibraryTop)
		if !ok {
			panic("LookAtLibraryTop kind has an incompatible primitive")
		}
		return look.PublishLinked, false
	case game.PrimitiveReveal:
		primitive, ok := instr.Primitive.(game.Reveal)
		if !ok {
			panic("Reveal kind has an incompatible primitive")
		}
		return primitive.PublishLinked, primitive.Card.Kind == game.CardReferenceLinked &&
			primitive.Card.LinkID == string(primitive.PublishLinked)
	default:
		return "", false
	}
}

func (r *effectResolver) clearLibraryCardPublication(key game.LinkedKey) {
	if key != "" {
		clearLinkedObjects(r.game, linkedObjectSourceKey(r.game, r.obj, string(key)))
	}
}

func (r *effectResolver) publishObservedCard(key game.LinkedKey, cardID id.ID) bool {
	card, ok := r.game.GetCardInstance(cardID)
	if !ok || card.Def == nil {
		return false
	}
	if key != "" {
		rememberLinkedObject(r.game, linkedObjectSourceKey(r.game, r.obj, string(key)), game.LinkedObjectRef{
			CardID: cardID, CardZoneVersion: card.ZoneVersion, CardZoneVersionSet: true,
		})
	}
	return true
}
