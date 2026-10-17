package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

const openingHandSize = 7

func (e *Engine) drawOpeningHands(g *game.Game) {
	for _, player := range g.Players {
		if player.Eliminated {
			continue
		}
		for range openingHandSize {
			e.drawCard(g, player.ID, false)
		}
	}
}

func (*Engine) drawCard(g *game.Game, playerID game.PlayerID, firstInDrawStep bool) (id.ID, bool) {
	if playerID < 0 || int(playerID) >= len(g.Players) {
		return 0, false
	}
	player := g.Players[playerID]

	cardID, ok := player.Library.Top()
	if !ok {
		g.FailedDraws[playerID] = true
		return 0, false
	}

	event := game.Event{
		Controller: playerID,
		Player:     playerID,
		CardID:     cardID,
		FromZone:   zone.Library,
		ToZone:     zone.Hand,
		Amount:     1,
	}
	destination, moved := moveCardBetweenZonesWithEvent(g, event, zoneMoveOptions{})
	if !moved || destination != zone.Hand {
		return cardID, false
	}
	card, _ := g.GetCardInstance(cardID)
	event.CardZoneVersion = card.ZoneVersion
	event.Kind = game.EventCardDrawn
	event.PlayerEventOrdinalThisTurn = nextPlayerEventOrdinalThisTurn(g, game.EventCardDrawn, playerID)
	event.FirstInDrawStep = firstInDrawStep
	emitEvent(g, event)
	return cardID, true
}
