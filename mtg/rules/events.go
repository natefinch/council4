package rules

import "github.com/natefinch/council4/mtg/game"

func emitEvent(g *game.Game, event game.GameEvent) {
	if g == nil || event.Kind == game.EventUnknown {
		return
	}
	g.Events = append(g.Events, event)
}

func emitZoneChangeEvent(g *game.Game, event game.GameEvent) {
	event.Kind = game.EventZoneChanged
	emitEvent(g, event)
}
