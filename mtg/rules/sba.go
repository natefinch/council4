package rules

import "github.com/natefinch/council4/mtg/game"

const maxStateBasedActionPasses = 1000

func (e *Engine) applyStateBasedActions(g *game.Game) []LossLog {
	var losses []LossLog
	for i := 0; i < maxStateBasedActionPasses; i++ {
		changed, passLosses := e.checkStateBasedActions(g)
		losses = append(losses, passLosses...)
		if !changed {
			return losses
		}
	}
	panic("state-based actions did not converge")
}

func (e *Engine) checkStateBasedActions(g *game.Game) (bool, []LossLog) {
	if g == nil {
		return false, nil
	}

	changed := false
	var losses []LossLog
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
			reason := lossReason(g, player)
			if e.eliminatePlayer(g, player.ID) {
				changed = true
				losses = append(losses, LossLog{
					Player: player.ID,
					Reason: reason,
				})
			}
			delete(g.FailedDraws, player.ID)
		}
	}
	return changed, losses
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

func lossReason(g *game.Game, player *game.Player) LossReason {
	if g.FailedDraws[player.ID] {
		return LossReasonEmptyLibraryDraw
	}
	if player.Life <= 0 {
		return LossReasonZeroLife
	}
	if player.HasLethalPoison() {
		return LossReasonPoisonCounters
	}
	if player.HasLethalCommanderDamage() {
		return LossReasonCommanderDamage
	}
	return LossReasonStateBasedEliminate
}
