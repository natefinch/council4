package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func stormCopyCount(g *game.Game, spell *game.CardDef) int {
	if spell == nil || !spell.HasKeyword(game.Storm) {
		return 0
	}
	count := 0
	for _, event := range g.EventsThisTurn() {
		if event.Kind == game.EventSpellCast {
			count++
		}
	}
	return count
}

func createStormCopies(g *game.Game, original *game.StackObject, spell *game.CardDef, count int) {
	for range count {
		copyObj := &game.StackObject{
			ID:                  g.IDGen.Next(),
			Kind:                game.StackSpell,
			SourceID:            original.SourceID,
			Face:                original.Face,
			SourceCardID:        original.SourceCardID,
			SourceTokenDef:      original.SourceTokenDef,
			Controller:          original.Controller,
			Targets:             append([]game.Target(nil), original.Targets...),
			TargetCounts:        append([]int(nil), original.TargetCounts...),
			ChosenModes:         append([]int(nil), original.ChosenModes...),
			XValue:              original.XValue,
			KickerPaid:          original.KickerPaid,
			Overloaded:          original.Overloaded,
			Flashback:           original.Flashback,
			Suspend:             original.Suspend,
			Mutate:              original.Mutate,
			MutateTargetID:      original.MutateTargetID,
			Copy:                true,
			SourceZone:          original.SourceZone,
			AdditionalCostsPaid: append([]string(nil), original.AdditionalCostsPaid...),
		}
		g.Stack.Push(copyObj)
		emitSpellCopiedEvent(g, copyObj, spell)
	}
}

// emitSpellCopiedEvent records that a spell copy was created on the stack
// (CR 707). It mirrors the spell characteristics an EventSpellCast carries so
// "cast or copy" (magecraft) triggers can match copies, but uses a distinct
// event kind so cast-only triggers and cast counts ignore it.
func emitSpellCopiedEvent(g *game.Game, copyObj *game.StackObject, spell *game.CardDef) {
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
		event.CardTypes = cardTypes(spell)
		event.CardSupertypes = cardSupertypes(spell)
		event.CardSubtypes = cardSubtypes(spell)
		event.Colors = spellColors(spell)
		event.ManaValue = opt.Val(stackManaValue(spell, copyObj.XValue))
	}
	emitEvent(g, event)
}
