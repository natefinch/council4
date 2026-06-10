package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

// pushSpellToStack pushes obj onto the stack and emits the standard spell
// zone-change and spell-cast events derived from castEvent.
//
// castEvent must have SourceID, StackObjectID (obj.ID), Controller, CardID,
// Face, CardTypes, Colors, FromZone, and ToZone populated by the caller. The helper
// sets event Kind for both the zone-change and spell-cast emissions.
// emitTargetEvents fires before the cast event so that became-target events
// precede EventSpellCast in the event stream (CR 601.2h).
//
// Callers that produce storm copies must compute the copy count before calling
// this helper, because EventSpellCast is emitted inside.
func pushSpellToStack(g *game.Game, obj *game.StackObject, castEvent game.Event) {
	g.Stack.Push(obj)
	emitTargetEvents(g, obj)
	castEvent = emitZoneChangeEvent(g, castEvent)
	castEvent.Kind = game.EventSpellCast
	emitEvent(g, castEvent)
}

// pushAbilityToStack pushes obj onto the stack and emits target events for any
// targets recorded on the stack object. Used for activated and triggered
// abilities that may have player or permanent targets.
func pushAbilityToStack(g *game.Game, obj *game.StackObject) {
	g.Stack.Push(obj)
	emitTargetEvents(g, obj)
}

// cardInstanceFaceDef looks up a CardInstance and its face definition in one
// call. Returns (nil, nil, false) when the card is absent or has no such face.
func cardInstanceFaceDef(g *game.Game, cardID id.ID, face game.FaceIndex) (*game.CardInstance, *game.CardDef, bool) {
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return nil, nil, false
	}
	def, ok := cardFaceDef(card, face)
	return card, def, ok
}
