package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

func emitEvent(g *game.Game, event game.Event) {
	if event.Kind == game.EventSpellCast && event.PlayerEventOrdinalThisTurn == 0 {
		event.PlayerEventOrdinalThisTurn = nextSpellCastOrdinalThisTurn(g, event.Controller)
	}
	if event.Kind == game.EventCardDrawn || event.Kind == game.EventBeginningOfStep {
		event.TriggeredAbilities = captureEventTriggeredAbilities(g, event)
		event.TriggeredAbilitiesCaptured = true
	}
	if !event.TriggeredAbilitiesCaptured {
		if doublers := captureChosenTypeTriggerDoublers(g); len(doublers) > 0 {
			event.ChosenTypeTriggerDoublers = &game.ChosenTypeTriggerDoublerSnapshot{Doublers: doublers}
		}
	}
	g.AppendEvent(event)
}

func nextPlayerEventOrdinalThisTurn(g *game.Game, kind game.EventKind, playerID game.PlayerID) int {
	ordinal := 1
	start := 0
	index := g.Turn.TurnNumber - 1
	if index >= 0 && index < len(g.EventTurnStarts) {
		start = g.EventTurnStarts[index]
	}
	for _, event := range g.Events[start:] {
		if event.Kind == kind && event.Player == playerID {
			ordinal++
		}
	}
	return ordinal
}

// nextSpellCastOrdinalThisTurn reports the per-turn ordinal position of the
// spell about to be cast by controller, counting only prior EventSpellCast
// events this turn (CR 700.6, "Nth spell each turn"). Spell copies emit
// EventSpellCopied and are deliberately excluded so copies do not advance the
// count.
func nextSpellCastOrdinalThisTurn(g *game.Game, controller game.PlayerID) int {
	ordinal := 1
	start := 0
	index := g.Turn.TurnNumber - 1
	if index >= 0 && index < len(g.EventTurnStarts) {
		start = g.EventTurnStarts[index]
	}
	for _, event := range g.Events[start:] {
		if event.Kind == game.EventSpellCast && event.Controller == controller {
			ordinal++
		}
	}
	return ordinal
}

func emitZoneChangeEvent(g *game.Game, event game.Event) game.Event {
	if event.CardID != 0 && event.CardZoneVersion == 0 {
		if card, ok := g.GetCardInstance(event.CardID); ok {
			card.ZoneVersion++
			event.CardZoneVersion = card.ZoneVersion
		}
	}
	if event.FromZone == zone.Exile && event.ToZone != zone.Exile {
		delete(g.AdventureCards, event.CardID)
		delete(g.SuspendedCards, event.CardID)
	}
	if event.CardID != 0 && event.FromZone != event.ToZone {
		clearCardCastPermissions(g, event.CardID, event.FromZone)
	}
	event.Kind = game.EventZoneChanged
	emitEvent(g, event)
	return event
}

func clearCardCastPermissions(g *game.Game, cardID game.ObjectID, fromZone zone.Type) {
	kept := g.RuleEffects[:0]
	for i := range g.RuleEffects {
		effect := &g.RuleEffects[i]
		if (effect.Kind == game.RuleEffectCastFromZone || effect.Kind == game.RuleEffectPlayFromZone) &&
			effect.AffectedCardID == cardID &&
			effect.CastFromZone == fromZone {
			continue
		}
		kept = append(kept, *effect)
	}
	g.RuleEffects = kept
}

func markCurrentTurnEventStart(g *game.Game) {
	index := g.Turn.TurnNumber - 1
	for len(g.EventTurnStarts) <= index {
		g.EventTurnStarts = append(g.EventTurnStarts, len(g.Events))
	}
	g.EventTurnStarts[index] = len(g.Events)
	g.TriggerEventCursor = len(g.Events)
}

func emitPermanentTappedEvent(g *game.Game, permanent *game.Permanent) {
	emitEvent(g, permanentTappedEvent(g, permanent, true))
}

func permanentTappedEvent(g *game.Game, permanent *game.Permanent, tapped bool) game.Event {
	kind := game.EventPermanentUntapped
	if tapped {
		kind = game.EventPermanentTapped
	}
	return game.Event{
		Kind:        kind,
		Controller:  effectiveController(g, permanent),
		Player:      permanent.Owner,
		CardID:      permanent.CardInstanceID,
		PermanentID: permanent.ObjectID,
		TokenName:   permanentTokenName(permanent),
		TokenDef:    permanent.TokenDef,
	}
}

func emitPermanentUntappedEvent(g *game.Game, permanent *game.Permanent) {
	emitEvent(g, permanentTappedEvent(g, permanent, false))
}

func setPermanentsTappedSimultaneously(g *game.Game, permanents []*game.Permanent, tapped bool) bool {
	var changed []*game.Permanent
	for _, permanent := range permanents {
		if permanent != nil && permanent.Tapped != tapped {
			changed = append(changed, permanent)
			permanent.Tapped = tapped
		}
	}
	if len(changed) == 0 {
		return false
	}
	simultaneousID := g.IDGen.Next()
	for _, permanent := range changed {
		event := permanentTappedEvent(g, permanent, tapped)
		event.SimultaneousID = simultaneousID
		emitEvent(g, event)
	}
	return true
}

func sacrificePermanent(g *game.Game, permanent *game.Permanent) bool {
	return sacrificePermanentsSimultaneously(g, []*game.Permanent{permanent})
}

func sacrificePermanentsSimultaneously(g *game.Game, permanents []*game.Permanent) bool {
	if len(permanents) == 0 {
		return false
	}
	simultaneousID := g.IDGen.Next()
	events := make([]game.Event, 0, len(permanents))
	for _, permanent := range permanents {
		if permanent == nil {
			continue
		}
		events = append(events, game.Event{
			Kind:           game.EventPermanentSacrificed,
			SimultaneousID: simultaneousID,
			Controller:     effectiveController(g, permanent),
			Player:         effectiveController(g, permanent),
			CardID:         permanent.CardInstanceID,
			PermanentID:    permanent.ObjectID,
			TokenName:      permanentTokenName(permanent),
			TokenDef:       permanent.TokenDef,
		})
	}
	if !movePermanentsToZoneSimultaneously(g, permanents, zone.Graveyard) {
		return false
	}
	succeeded := false
	for _, event := range events {
		if _, stillOnBattlefield := permanentByObjectID(g, event.PermanentID); stillOnBattlefield {
			continue
		}
		emitEvent(g, event)
		succeeded = true
	}
	return succeeded
}

func setPermanentTapped(g *game.Game, permanent *game.Permanent, tapped bool) {
	if permanent.Tapped == tapped {
		return
	}
	permanent.Tapped = tapped
	if tapped {
		emitPermanentTappedEvent(g, permanent)
		return
	}
	emitPermanentUntappedEvent(g, permanent)
}

// setPermanentTappedForMana taps a permanent and records tapped-for-mana
// provenance on the emitted event so "is tapped for mana" triggers (Wild Growth
// and the mana-additional aura family) can fire.
func setPermanentTappedForMana(g *game.Game, permanent *game.Permanent) {
	if permanent.Tapped {
		return
	}
	permanent.Tapped = true
	event := permanentTappedEvent(g, permanent, true)
	event.TappedForMana = true
	emitEvent(g, event)
}

// manaPoolColorSnapshot records, per color, how much mana the player currently
// holds. producedManaColorsSince diffs against it to learn which types a mana
// ability just added.
func manaPoolColorSnapshot(g *game.Game, playerID game.PlayerID) map[mana.Color]int {
	snapshot := make(map[mana.Color]int)
	player, ok := playerByID(g, playerID)
	if !ok {
		return snapshot
	}
	for unit, amount := range player.ManaPool.Units() {
		snapshot[unit.Color] += amount
	}
	return snapshot
}

// producedManaColorsSince returns the distinct mana types whose count in the
// player's pool grew relative to before, in WUBRG order with colorless last. It
// reports exactly the types a just-resolved mana ability added, backing the
// "one mana of any type that land produced" mana-doubler trigger.
func producedManaColorsSince(g *game.Game, playerID game.PlayerID, before map[mana.Color]int) []mana.Color {
	after := manaPoolColorSnapshot(g, playerID)
	var produced []mana.Color
	for _, c := range []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G, mana.C} {
		if after[c] > before[c] {
			produced = append(produced, c)
		}
	}
	return produced
}

// recordTappedForManaProduced annotates the most recent tapped-for-mana event
// for permanentID with the mana types its tap produced, so a "one mana of any
// type that land produced" trigger (Mirari's Wake) can mirror them at
// resolution. It is a no-op when no colors were produced or no matching event
// is found.
func recordTappedForManaProduced(g *game.Game, permanentID id.ID, colors []mana.Color) {
	if len(colors) == 0 {
		return
	}
	for i := len(g.Events) - 1; i >= 0; i-- {
		event := &g.Events[i]
		if event.Kind == game.EventPermanentTapped && event.TappedForMana && event.PermanentID == permanentID {
			event.ProducedManaColors = append(event.ProducedManaColors, colors...)
			return
		}
	}
}

func emitTargetEvents(g *game.Game, obj *game.StackObject) {
	for _, target := range obj.Targets {
		event := game.Event{
			Kind:          game.EventObjectBecameTarget,
			StackObjectID: obj.ID,
			Controller:    obj.Controller,
			Target:        target,
		}
		event.SourceID, event.SourceObjectID = damageSourceIDs(g, obj)
		switch target.Kind {
		case game.TargetPermanent:
			event.PermanentID = target.PermanentID
		case game.TargetPlayer:
			event.Player = target.PlayerID
		default:
		}
		emitEvent(g, event)
	}
}

func emitAbilityActivatedEvent(g *game.Game, obj *game.StackObject, permanentID game.ObjectID, manaAbility bool) {
	emitEvent(g, game.Event{
		Kind:           game.EventAbilityActivated,
		SourceID:       obj.SourceCardID,
		SourceObjectID: permanentID,
		StackObjectID:  obj.ID,
		AbilityIndex:   obj.AbilityIndex,
		ManaAbility:    manaAbility,
		Controller:     obj.Controller,
		Player:         obj.Controller,
		CardID:         obj.SourceCardID,
		PermanentID:    permanentID,
		TokenDef:       obj.SourceTokenDef,
	})
}
