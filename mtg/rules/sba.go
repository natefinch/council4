package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

const maxStateBasedActionPasses = 1000

func (e *Engine) applyStateBasedActions(g *game.Game) []LossLog {
	losses, _ := e.applyStateBasedActionsWithDeaths(g)
	return losses
}

func (e *Engine) applyStateBasedActionsWithLog(g *game.Game, log *TurnLog) []LossLog {
	losses, deaths := e.applyStateBasedActionsWithDeaths(g)
	if log != nil {
		log.Losses = append(log.Losses, losses...)
		log.Deaths = append(log.Deaths, deaths...)
	}
	return losses
}

func (e *Engine) applyStateBasedActionsWithDeaths(g *game.Game) ([]LossLog, []PermanentDeathLog) {
	var losses []LossLog
	var deaths []PermanentDeathLog
	for i := 0; i < maxStateBasedActionPasses; i++ {
		changed, passLosses := e.checkStateBasedActions(g)
		permanentsChanged, passDeaths := e.checkPermanentStateBasedActions(g)
		losses = append(losses, passLosses...)
		deaths = append(deaths, passDeaths...)
		if !changed && !permanentsChanged {
			return losses, deaths
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

func (e *Engine) checkPermanentStateBasedActions(g *game.Game) (bool, []PermanentDeathLog) {
	if g == nil {
		return false, nil
	}

	type pendingDeath struct {
		objectID id.ID
		reason   PermanentDeathReason
	}
	var pending []pendingDeath
	for _, permanent := range g.Battlefield {
		reason, ok := permanentDeathReason(g, permanent)
		if ok {
			pending = append(pending, pendingDeath{
				objectID: permanent.ObjectID,
				reason:   reason,
			})
		}
	}
	if len(pending) == 0 {
		return false, nil
	}

	var deaths []PermanentDeathLog
	for _, death := range pending {
		permanent, ok := destroyPermanent(g, death.objectID)
		if !ok {
			continue
		}
		deaths = append(deaths, PermanentDeathLog{
			Permanent:  permanent.ObjectID,
			SourceID:   permanent.CardInstanceID,
			Owner:      permanent.Owner,
			Controller: permanent.Controller,
			Reason:     death.reason,
		})
	}
	return len(deaths) > 0, deaths
}

func permanentDeathReason(g *game.Game, permanent *game.Permanent) (PermanentDeathReason, bool) {
	card := permanentCardDef(g, permanent)
	if card == nil || !card.HasType(game.TypeCreature) {
		return "", false
	}
	toughness, ok := creatureToughness(card)
	if !ok {
		return "", false
	}
	if toughness <= 0 {
		return PermanentDeathReasonZeroToughness, true
	}
	if permanent.MarkedDamage >= toughness {
		return PermanentDeathReasonLethalDamage, true
	}
	return "", false
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
