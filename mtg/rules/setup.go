package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

const openingHandSize = 7

func (e *Engine) drawOpeningHands(g *game.Game) {
	for _, player := range g.Players {
		for range openingHandSize {
			e.drawCard(g, player.ID)
		}
	}
}

func (*Engine) drawCard(g *game.Game, playerID game.PlayerID) (id.ID, bool) {
	if playerID < 0 || int(playerID) >= len(g.Players) {
		return 0, false
	}
	player := g.Players[playerID]

	cardID, ok := player.Library.Top()
	if !ok {
		g.FailedDraws[playerID] = true
		return 0, false
	}

	player.Library.Remove(cardID)
	player.Hand.Add(cardID)
	event := game.GameEvent{
		Player:   playerID,
		CardID:   cardID,
		FromZone: zone.Library,
		ToZone:   zone.Hand,
		Amount:   1,
	}
	emitZoneChangeEvent(g, event)
	event.Kind = game.EventCardDrawn
	emitEvent(g, event)
	return cardID, true
}
