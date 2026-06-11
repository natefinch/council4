package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

func emitEvent(g *game.Game, event game.Event) {
	g.AppendEvent(event)
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
		if effect.Kind == game.RuleEffectCastFromZone &&
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
	emitEvent(g, game.Event{
		Kind:        game.EventPermanentTapped,
		Controller:  effectiveController(g, permanent),
		Player:      permanent.Owner,
		CardID:      permanent.CardInstanceID,
		PermanentID: permanent.ObjectID,
		TokenName:   permanentTokenName(permanent),
		TokenDef:    permanent.TokenDef,
	})
}

func emitPermanentUntappedEvent(g *game.Game, permanent *game.Permanent) {
	emitEvent(g, game.Event{
		Kind:        game.EventPermanentUntapped,
		Controller:  effectiveController(g, permanent),
		Player:      permanent.Owner,
		CardID:      permanent.CardInstanceID,
		PermanentID: permanent.ObjectID,
		TokenName:   permanentTokenName(permanent),
		TokenDef:    permanent.TokenDef,
	})
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
