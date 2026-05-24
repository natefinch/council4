package rules

import "github.com/natefinch/council4/mtg/game"

func (e *Engine) runTurn(g *game.Game, agents [game.NumPlayers]PlayerAgent) TurnLog {
	log := TurnLog{
		TurnNumber:   g.Turn.TurnNumber,
		ActivePlayer: g.Turn.ActivePlayer,
	}

	e.runBeginningPhase(g, agents, &log)
	if g.IsGameOver() {
		return log
	}
	e.runMainPhase(g, agents, game.PhasePrecombatMain, &log)
	if g.IsGameOver() {
		return log
	}
	e.runCombatPhase(g, agents, &log)
	if g.IsGameOver() {
		return log
	}
	e.runMainPhase(g, agents, game.PhasePostcombatMain, &log)
	if g.IsGameOver() {
		return log
	}
	e.runEndingPhase(g, agents)
	if g.IsGameOver() {
		return log
	}
	e.advanceToNextTurn(g)

	return log
}

func (e *Engine) runBeginningPhase(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	g.Turn.Phase = game.PhaseBeginning

	g.Turn.Step = game.StepUntap
	for _, permanent := range g.Battlefield {
		if permanent == nil {
			continue
		}
		if permanent.Controller == g.Turn.ActivePlayer {
			permanent.Tapped = false
			permanent.SummoningSick = false
		}
	}

	g.Turn.Step = game.StepUpkeep

	g.Turn.Step = game.StepDraw
	cardID, ok := e.drawCard(g, g.Turn.ActivePlayer)
	if log != nil {
		log.Draws = append(log.Draws, DrawLog{
			Player: g.Turn.ActivePlayer,
			CardID: cardID,
			Failed: !ok,
		})
	}
	e.applyStateBasedActionsWithLog(g, log)
	emptyManaPools(g)
}

func (e *Engine) runMainPhase(g *game.Game, agents [game.NumPlayers]PlayerAgent, phase game.Phase, log *TurnLog) {
	g.Turn.Phase = phase
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = g.Turn.ActivePlayer
	e.runPriorityLoop(g, agents, log)
	emptyManaPools(g)
}

func (e *Engine) runEndingPhase(g *game.Game, agents [game.NumPlayers]PlayerAgent) {
	g.Turn.Phase = game.PhaseEnding
	g.Turn.Step = game.StepEnd

	g.Turn.Step = game.StepCleanup
	for _, permanent := range g.Battlefield {
		if permanent == nil {
			continue
		}
		permanent.MarkedDamage = 0
		permanent.MarkedDeathtouchDamage = false
	}
	emptyManaPools(g)
	g.Combat = nil
	// TODO: remove "until end of turn" effects when continuous effects exist.
}

func (e *Engine) advanceToNextTurn(g *game.Game) {
	next, ok := popExtraTurn(&g.Turn.ExtraTurns, &g.TurnOrder)
	if !ok {
		next = g.TurnOrder.NextActivePlayer(g.Turn.ActivePlayer)
	}

	g.Turn.TurnNumber++
	g.Turn.ActivePlayer = next
	g.Turn.PriorityPlayer = next
	g.Turn.Phase = game.PhaseBeginning
	g.Turn.Step = game.StepUntap
	g.Turn.LandsPlayedThisTurn = 0
	g.Turn.LandsAllowedThisTurn = 1
	g.Combat = nil
}

func popExtraTurn(extraTurns *[]game.PlayerID, turnOrder *game.TurnOrder) (game.PlayerID, bool) {
	for len(*extraTurns) > 0 {
		last := len(*extraTurns) - 1
		next := (*extraTurns)[last]
		*extraTurns = (*extraTurns)[:last]
		if !turnOrder.IsEliminated(next) {
			return next, true
		}
	}
	return 0, false
}

func emptyManaPools(g *game.Game) {
	for _, player := range g.Players {
		if player == nil {
			continue
		}
		player.ManaPool.Empty()
	}
}
