package rules

import "github.com/natefinch/council4/mtg/game"

const openingHandSize = 7

func (e *Engine) drawOpeningHands(g *game.Game) {
	if g == nil {
		return
	}
	for _, player := range g.Players {
		if player == nil {
			continue
		}
		for range openingHandSize {
			e.drawCard(g, player.ID)
		}
	}
}

func (e *Engine) drawCard(g *game.Game, playerID game.PlayerID) bool {
	if g == nil || playerID < 0 || int(playerID) >= len(g.Players) {
		return false
	}
	player := g.Players[playerID]
	if player == nil {
		return false
	}

	cardID, ok := player.Library.Top()
	if !ok {
		g.FailedDraws[playerID] = true
		return false
	}

	player.Library.Remove(cardID)
	player.Hand.Add(cardID)
	return true
}
