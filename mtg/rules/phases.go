package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

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
	expireGoadForActivePlayer(g)
	for _, permanent := range g.Battlefield {
		if effectiveController(g, permanent) != g.Turn.ActivePlayer {
			continue
		}
		if permanent.PhasedOut {
			permanent.PhasedOut = false
			if permanent.Exerted {
				permanent.Exerted = false
				permanent.SummoningSick = false
			}
			continue
		}
		if permanent.Exerted {
			permanent.Exerted = false
			permanent.SummoningSick = false
			continue
		}
		if ruleEffectPreventsUntap(g, permanent) {
			permanent.SummoningSick = false
			continue
		}
		// A stun counter replaces the permanent's untapping: instead of
		// untapping, remove one stun counter from it (CR 122.6f). The counter is
		// only consumed when the permanent would actually untap, so an already
		// untapped permanent keeps its stun counters.
		if permanent.Tapped && permanent.Counters.Has(counter.Stun) {
			permanent.Counters.Remove(counter.Stun, 1)
			permanent.SummoningSick = false
			continue
		}
		setPermanentTapped(g, permanent, false)
		permanent.SummoningSick = false
	}

	g.Turn.Step = game.StepUpkeep
	// Beginning-of-step triggers fire at the start of the upkeep and are put on
	// the stack before the game advances to draw (CR 603.6c, CR 117.3b).
	emitBeginningOfStepEvent(g, game.StepUpkeep)
	putBeginningOfNextUpkeepDelayedTriggersOnStack(g)
	e.processSuspendUpkeep(g, g.Turn.ActivePlayer)
	g.Turn.PriorityPlayer = g.Turn.ActivePlayer
	e.runPriorityLoop(g, agents, log)
	if g.IsGameOver() {
		return
	}

	if !consumeSkipStep(g, g.Turn.ActivePlayer, game.StepDraw) {
		g.Turn.Step = game.StepDraw
		emitBeginningOfStepEvent(g, game.StepDraw)
		cardID, ok := e.drawCard(g, g.Turn.ActivePlayer)
		log.addDraw(DrawLog{
			Player: g.Turn.ActivePlayer,
			CardID: cardID,
			Failed: !ok,
		})
		advanceSagas(g, g.Turn.ActivePlayer)
		g.Turn.PriorityPlayer = g.Turn.ActivePlayer
		e.runPriorityLoop(g, agents, log)
		if g.IsGameOver() {
			return
		}
	}
	e.applyStateBasedActionsWithLog(g, log)
	emptyManaPools(g)
}

func (e *Engine) runMainPhase(g *game.Game, agents [game.NumPlayers]PlayerAgent, phase game.Phase, log *TurnLog) {
	g.Turn.Phase = phase
	step := game.StepNone
	switch phase {
	case game.PhasePrecombatMain:
		step = game.StepPrecombatMain
	case game.PhasePostcombatMain:
		step = game.StepPostcombatMain
	default:
	}
	if step != game.StepNone {
		g.Turn.Step = step
		emitBeginningOfStepEvent(g, step)
	}
	// Main phases have no steps. Reset the synthetic trigger boundary before
	// priority so sorcery-speed actions remain legal.
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = g.Turn.ActivePlayer
	e.runPriorityLoop(g, agents, log)
	emptyManaPools(g)
}

func (e *Engine) runEndingPhase(g *game.Game, agents [game.NumPlayers]PlayerAgent) {
	g.Turn.Phase = game.PhaseEnding
	g.Turn.Step = game.StepEnd
	// "At the beginning of the end step" triggers use the same event as delayed
	// next-end-step triggers before the end-step priority window (CR 603.6c,
	// CR 603.7b).
	emitBeginningOfStepEvent(g, game.StepEnd)
	putBeginningOfEndStepDelayedTriggersOnStack(g)
	if e.putTriggeredAbilitiesOnStackWithChoices(g, agents, nil) || !g.Stack.IsEmpty() {
		g.Turn.PriorityPlayer = g.Turn.ActivePlayer
		e.runPriorityLoop(g, agents, nil)
		if g.IsGameOver() {
			return
		}
	}

	g.Turn.Step = game.StepCleanup
	discardToMaximumHandSize(g, g.Turn.ActivePlayer)
	for _, permanent := range g.Battlefield {
		permanent.MarkedDamage = 0
		permanent.MarkedDeathtouchDamage = false
		permanent.TemporaryPowerModifier = 0
		permanent.TemporaryToughnessModifier = 0
		permanent.RegenerationShields = 0
	}
	expireCleanupDurations(g)
	expirePreventionShields(g)
	expireReplacementEffects(g)
	expireRuleEffects(g)
	e.applyStateBasedActions(g)
	emptyManaPools(g)
	g.Combat = nil
}

func emitBeginningOfStepEvent(g *game.Game, step game.Step) {
	// Triggered abilities with "At the beginning of [step]" look for this
	// turn-based event (CR 603.6c).
	emitEvent(g, game.Event{
		Kind:       game.EventBeginningOfStep,
		Controller: g.Turn.ActivePlayer,
		Player:     g.Turn.ActivePlayer,
		Step:       step,
	})
}

func discardToMaximumHandSize(g *game.Game, playerID game.PlayerID) {
	player, ok := playerByID(g, playerID)
	if !ok || player.Eliminated || player.Hand.Size() <= maximumHandSize {
		return
	}
	if playerHasNoMaximumHandSize(g, playerID) {
		return
	}
	cards := player.Hand.All()
	for i := len(cards) - 1; i >= maximumHandSize; i-- {
		discardCardFromHand(g, playerID, cards[i])
	}
}

func (*Engine) advanceToNextTurn(g *game.Game) {
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
	g.TriggeredAbilitiesThisTurn = make(map[game.TriggeredAbilityUse]int)
	g.Combat = nil
	markCurrentTurnEventStart(g)
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
		player.ManaPool.Empty()
	}
}
