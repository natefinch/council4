package rules

import "github.com/natefinch/council4/mtg/game"

const maxStateBasedActionPasses = 1000

func (e *Engine) applyStateBasedActions(g *game.Game) {
	for i := 0; i < maxStateBasedActionPasses; i++ {
		if !e.checkStateBasedActions(g) {
			return
		}
	}
	panic("state-based actions did not converge")
}

func (e *Engine) checkStateBasedActions(g *game.Game) bool {
	if g == nil {
		return false
	}

	changed := false
	for _, player := range g.Players {
		if player == nil {
			continue
		}
		if player.Eliminated {
			delete(g.FailedDraws, player.ID)
			continue
		}
		if player.Life <= 0 ||
			player.HasLethalPoison() ||
			player.HasLethalCommanderDamage() ||
			g.FailedDraws[player.ID] {
			if e.eliminatePlayer(g, player.ID) {
				changed = true
			}
			delete(g.FailedDraws, player.ID)
		}
	}
	return changed
}

func (e *Engine) eliminatePlayer(g *game.Game, playerID game.PlayerID) bool {
	if g == nil || playerID < 0 || int(playerID) >= len(g.Players) {
		return false
	}

	player := g.Players[playerID]
	if player == nil {
		return false
	}

	if player.Eliminated && g.TurnOrder.IsEliminated(playerID) {
		return false
	}

	player.Eliminated = true
	g.TurnOrder.Eliminate(playerID)
	return true
}
