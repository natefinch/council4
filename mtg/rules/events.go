package rules

import "github.com/natefinch/council4/mtg/game"

func emitEvent(g *game.Game, event game.Event) {
	g.AppendEvent(event)
}

func emitZoneChangeEvent(g *game.Game, event game.Event) {
	event.Kind = game.EventZoneChanged
	emitEvent(g, event)
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
