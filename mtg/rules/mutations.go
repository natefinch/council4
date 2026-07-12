package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

// pushSpellToStack pushes obj onto the stack and emits the standard spell
// zone-change and spell-cast events derived from castEvent.
//
// castEvent must have SourceID, StackObjectID (obj.ID), Controller, CardID,
// Face, CardTypes, CardSupertypes, CardSubtypes, Colors, ManaValue, FromZone,
// and ToZone populated by the caller. The helper sets event Kind for both the
// zone-change and spell-cast emissions.
// emitTargetEvents fires before the cast event so that became-target events
// precede EventSpellCast in the event stream (CR 601.2h).
//
// Callers that produce storm copies must compute the copy count before calling
// this helper, because EventSpellCast is emitted inside.
func pushSpellToStack(g *game.Game, obj *game.StackObject, castEvent game.Event) {
	g.Stack.Push(obj)
	emitSpellCastEvents(g, obj, castEvent)
}

// emitSpellCastEvents emits the events that mark a spell as cast (CR 601.2i):
// it consumes "next spell can't be countered" effects, emits target events, the
// zone-change to the stack, and the spell-cast event (firing "when you cast"
// triggers). It is separated from the stack push so a cast can put the card on
// the stack first (CR 601.2a), pay its costs (CR 601.2f-h), and only then become
// cast; obj must already be on the stack when this is called.
func emitSpellCastEvents(g *game.Game, obj *game.StackObject, castEvent game.Event) {
	consumeNextSpellCantBeCounteredEffects(g, obj)
	consumeNextSpellAsThoughFlashEffects(g, obj)
	emitTargetEvents(g, obj)
	fromZone := castEvent.FromZone
	castCardID := castEvent.CardID
	castPlayer := castEvent.Controller
	castEvent = emitZoneChangeEvent(g, castEvent)
	castEvent.Kind = game.EventSpellCast
	castEvent.PlayerEventOrdinalThisTurn = nextSpellCastOrdinalThisTurn(g, castEvent.Controller)
	emitEvent(g, castEvent)
	emitCardPlayedFromExileEvent(g, castPlayer, castCardID, fromZone)
}

func consumeNextSpellAsThoughFlashEffects(g *game.Game, obj *game.StackObject) {
	if len(g.RuleEffects) == 0 {
		return
	}
	spellDef, ok := spellDefForStackObject(g, obj)
	if !ok {
		return
	}
	kept := g.RuleEffects[:0]
	for i := range g.RuleEffects {
		effect := &g.RuleEffects[i]
		if effect.Kind != game.RuleEffectCastSpellsAsThoughFlash ||
			!effect.AppliesToNextSpellOnly ||
			!playerRelationMatches(effect.Controller, obj.Controller, effect.AffectedPlayer) ||
			!castAsThoughFlashEffectMatchesCard(g, effect, spellDef) {
			kept = append(kept, *effect)
		}
	}
	g.RuleEffects = kept
}

// emitCardPlayedFromExileEvent emits EventCardPlayedFromExile when a card was
// played (cast or played as a land) from exile, so "whenever a player plays a
// card exiled with this" triggers (Prowl, Stoic Strategist) can fire. It is a
// no-op for plays from any other zone.
func emitCardPlayedFromExileEvent(g *game.Game, player game.PlayerID, cardID id.ID, fromZone zone.Type) {
	if fromZone != zone.Exile || cardID == 0 {
		return
	}
	emitEvent(g, game.Event{
		Kind:       game.EventCardPlayedFromExile,
		Controller: player,
		Player:     player,
		CardID:     cardID,
	})
}

// emitLandPlayedEvent emits EventLandPlayed when a player plays a land as the
// land-play special action (CR 305), so "whenever a player/an opponent/you
// play(s) a land" triggers (Burgeoning, Dirtcowl Wurm, Horn of Greed) fire. It
// fires for every played land regardless of source zone or which land-play
// permission allowed it, but not for lands an effect puts onto the battlefield
// without playing them.
func emitLandPlayedEvent(g *game.Game, player game.PlayerID, cardID id.ID) {
	emitEvent(g, game.Event{
		Kind:       game.EventLandPlayed,
		Controller: player,
		Player:     player,
		CardID:     cardID,
	})
}

// consumeNextSpellCantBeCounteredEffects applies and consumes any
// "The next spell you cast this turn can't be countered." rule effects (Mistrise
// Village) whose controller and spell-type filter match the spell just cast.
// Each matching effect is snapshotted onto the spell as an object-scoped
// cant-be-countered effect, so the spell stays uncounterable for the rest of its
// time on the stack, and is then removed from the game's rule effects so later
// spells are unaffected. It is a no-op when no such one-shot effect is active.
func consumeNextSpellCantBeCounteredEffects(g *game.Game, obj *game.StackObject) {
	if len(g.RuleEffects) == 0 {
		return
	}
	spellDef, ok := spellDefForStackObject(g, obj)
	if !ok {
		return
	}
	kept := g.RuleEffects[:0]
	for i := range g.RuleEffects {
		effect := &g.RuleEffects[i]
		if effect.Kind != game.RuleEffectCantBeCountered ||
			!effect.AppliesToNextSpellOnly ||
			!controllerRelationMatches(effect.Controller, obj.Controller, effect.AffectedController) ||
			!spellTypesMatch(spellDef, effect.SpellTypes) {
			kept = append(kept, *effect)
			continue
		}
		obj.RuleEffects = append(obj.RuleEffects, game.RuleEffect{
			Kind:             game.RuleEffectCantBeCountered,
			Controller:       effect.Controller,
			SourceObjectID:   effect.SourceObjectID,
			SourceCardID:     effect.SourceCardID,
			AffectedObjectID: obj.ID,
		})
	}
	g.RuleEffects = kept
}

// spellDefForStackObject resolves the card definition of a spell stack object,
// whether it is backed by a card instance or a token definition.
func spellDefForStackObject(g *game.Game, obj *game.StackObject) (*game.CardDef, bool) {
	if obj.SourceTokenDef != nil {
		return obj.SourceTokenDef.FaceDef(obj.Face)
	}
	_, spellDef, ok := cardInstanceFaceDef(g, obj.SourceID, obj.Face)
	return spellDef, ok
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
