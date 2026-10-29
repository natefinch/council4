package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// stormTriggeredAbility is the "When you cast this spell" triggered ability a
// storm instance creates (CR 702.40a). It carries the Storm keyword so
// resolution dispatches to the shared storm operation. It has no targets, so it
// can be countered (Stifle) or copied like any triggered ability.
func stormTriggeredAbility() *game.TriggeredAbility {
	return &game.TriggeredAbility{
		Text:             "Storm",
		KeywordAbilities: game.SimpleKeywords(game.Storm),
	}
}

// spellsCastEarlierThisTurn counts the spells any player has cast this turn
// before the spell-cast event being prepared now. Copies put directly onto the
// stack are not cast and emit no EventSpellCast.
func spellsCastEarlierThisTurn(g *game.Game) int {
	return eventsThisTurnWindow(g).count(func(event game.Event) bool {
		return event.Kind == game.EventSpellCast
	})
}

// stormCastTriggers returns one pending storm trigger for each storm instance a
// genuinely cast spell had as it became cast (CR 702.40a/b). Only
// EventSpellCast reaches here, so non-cast stack copies never storm (CR 707.10),
// while a cast copy of a card does (CR 707.12). The trigger keeps the cast
// event, whose SpellsCastBeforeThisTurn fixes the number of copies. Emitted
// spell-cast events carry the instances recorded as the spell became cast; a raw
// event that bypassed emission is evaluated now.
func stormCastTriggers(g *game.Game, source *game.Permanent, event game.Event) []pendingTriggeredAbility {
	if event.Kind != game.EventSpellCast {
		return nil
	}
	count := event.StormInstances.Val
	if !event.StormInstances.Exists {
		count = spellKeywordInstancesAtCast(g, event, game.Storm)
	}
	if count == 0 {
		return nil
	}
	pending := make([]pendingTriggeredAbility, 0, count)
	for range count {
		pending = append(pending, pendingTriggeredAbility{
			controller:      event.Controller,
			sourceID:        source.ObjectID,
			sourceCardID:    source.CardInstanceID,
			sourceToken:     source.TokenDef,
			face:            event.Face,
			inline:          stormTriggeredAbility(),
			event:           event,
			hasEvent:        true,
			ordinaryTrigger: true,
		})
	}
	return pending
}

// resolveStormTriggeredAbility copies the spell that triggered it once for each
// other spell cast before it this turn (CR 702.40a). The spell is copied as it
// currently exists on the stack, or from its last known information if it has
// already left the stack (CR 608.2h), so countering the spell in response does
// not stop the copies; countering this trigger does. The trigger's controller
// puts each copy onto the stack and, when the spell has targets, may choose new
// targets for each copy independently (CR 707.10c). A copy is put onto the
// stack before the next copy's targets are chosen, so later copies can target
// earlier ones.
func (e *Engine) resolveStormTriggeredAbility(g *game.Game, obj *game.StackObject, agents [game.NumPlayers]PlayerAgent, log *TurnLog) string {
	if !obj.HasTriggerEvent || !obj.TriggerEvent.SpellsCastBeforeThisTurn.Exists {
		return "missing source"
	}
	original, ok := stormSpell(g, obj)
	if !ok {
		return "missing source"
	}
	count := obj.TriggerEvent.SpellsCastBeforeThisTurn.Val
	if count > 0 {
		addend, _ := additionalSpellCopies(g, obj.Controller)
		count += addend
	}
	for range count {
		copyObj := game.NewStackObjectCopy(original, g.IDGen.Next())
		copyDef, copyOK := prepareStackSpellCopy(g, original, copyObj, nil)
		if !copyOK {
			continue
		}
		copyObj.Controller = obj.Controller
		copyObj.CopyOwner = opt.Val(obj.Controller)
		e.mayRetargetStackObjectChoice(g, obj.Controller, copyObj, agents, log)
		g.Stack.Push(copyObj)
		emitSpellCopiedEvent(g, copyObj, copyDef)
	}
	return "resolved"
}

// stormSpell returns the spell a storm trigger copies: the spell itself while it
// is still on the stack, otherwise its last known information.
func stormSpell(g *game.Game, obj *game.StackObject) (*game.StackObject, bool) {
	if spell, ok := stackObjectByID(g, obj.TriggerEvent.StackObjectID); ok && spell.Kind == game.StackSpell {
		return spell, true
	}
	if obj.TriggerSpellLKI != nil && obj.TriggerSpellLKI.ID == obj.TriggerEvent.StackObjectID {
		return obj.TriggerSpellLKI, true
	}
	return nil, false
}

// emitSpellCopiedEvent records that a spell copy was created on the stack
// (CR 707). It mirrors the spell characteristics an EventSpellCast carries so
// "cast or copy" (magecraft) triggers can match copies, but uses a distinct
// event kind so cast-only triggers and cast counts ignore it.
func emitSpellCopiedEvent(g *game.Game, copyObj *game.StackObject, spell *game.CardDef) {
	emitTargetEvents(g, copyObj)
	event := game.Event{
		Kind:          game.EventSpellCopied,
		SourceID:      copyObj.SourceID,
		StackObjectID: copyObj.ID,
		Controller:    copyObj.Controller,
		CardID:        copyObj.SourceCardID,
		Face:          copyObj.Face,
		ToZone:        zone.Stack,
	}
	if spell != nil {
		typeLine := stackSpellTypeLine(g, copyObj, spell)
		event.CardTypes = typeLine.types
		event.CardSupertypes = typeLine.supertypes
		event.CardSubtypes = typeLine.subtypes
		event.Colors = spellColors(spell)
		event.CardCharacteristicsKnown = true
		event.ManaValue = stackObjectKnownManaValue(g, copyObj)
	}
	emitEvent(g, event)
}
