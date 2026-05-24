package rules

import "github.com/natefinch/council4/mtg/game"

const maximumHandSize = 7

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
	expireTurnStartDurations(g)
	for _, permanent := range g.Battlefield {
		if permanent == nil {
			continue
		}
		if effectiveController(g, permanent) == g.Turn.ActivePlayer {
			if permanent.PhasedOut {
				permanent.PhasedOut = false
				continue
			}
			permanent.Tapped = false
			permanent.SummoningSick = false
		}
	}

	g.Turn.Step = game.StepUpkeep

	g.Turn.Step = game.StepDraw
	if !consumeSkipStep(g, g.Turn.ActivePlayer, game.StepDraw) {
		cardID, ok := e.drawCard(g, g.Turn.ActivePlayer)
		if log != nil {
			log.Draws = append(log.Draws, DrawLog{
				Player: g.Turn.ActivePlayer,
				CardID: cardID,
				Failed: !ok,
			})
		}
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
	putBeginningOfEndStepDelayedTriggersOnStack(g)
	if !g.Stack.IsEmpty() {
		g.Turn.PriorityPlayer = g.Turn.ActivePlayer
		e.runPriorityLoop(g, agents, nil)
		if g.IsGameOver() {
			return
		}
	}

	g.Turn.Step = game.StepCleanup
	discardToMaximumHandSize(g, g.Turn.ActivePlayer)
	for _, permanent := range g.Battlefield {
		if permanent == nil {
			continue
		}
		permanent.MarkedDamage = 0
		permanent.MarkedDeathtouchDamage = false
		permanent.TemporaryPowerModifier = 0
		permanent.TemporaryToughnessModifier = 0
		permanent.RegenerationShields = 0
	}
	expireCleanupDurations(g)
	expirePreventionShields(g)
	e.applyStateBasedActions(g)
	emptyManaPools(g)
	g.Combat = nil
}

func discardToMaximumHandSize(g *game.Game, playerID game.PlayerID) {
	player := playerByID(g, playerID)
	if player == nil || player.Eliminated || player.Hand.Size() <= maximumHandSize {
		return
	}
	cards := player.Hand.All()
	for i := len(cards) - 1; i >= maximumHandSize; i-- {
		discardCardFromHand(g, playerID, cards[i])
	}
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
	g.ActivatedAbilitiesThisTurn = make(map[game.ActivatedAbilityUse]bool)
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
