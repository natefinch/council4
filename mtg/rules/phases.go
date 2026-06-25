package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

const maximumHandSize = 7

func (e *Engine) runTurn(g *game.Game, agents [game.NumPlayers]PlayerAgent) (log TurnLog) {
	log = TurnLog{
		TurnNumber:   g.Turn.TurnNumber,
		ActivePlayer: g.Turn.ActivePlayer,
	}
	for seat := range g.Players {
		log.LifeTotals[seat] = g.Players[seat].Life
	}

	activePlayer := log.ActivePlayer
	manaSpentBefore := g.Players[activePlayer].ManaPool.Spent()
	defer func() {
		log.ManaSpent = g.Players[activePlayer].ManaPool.Spent() - manaSpentBefore
	}()

	e.runBeginningPhase(g, agents, &log)
	if g.IsGameOver() {
		return log
	}
	e.runMainPhase(g, agents, game.PhasePrecombatMain, &log)
	recordManaDevelopment(g, &log)
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
	e.runExtraPhases(g, agents, &log)
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

// runExtraPhases drains the additional phases queued onto the turn by
// extra-phase effects ("After this main phase, there is an additional combat
// phase followed by an additional main phase." — Aggravated Assault). Queued
// phases run in order after the postcombat main phase; a queued main phase that
// re-activates the source re-queues more phases, so the loop continues until the
// queue empties (the extra-combat combo, CR 505.5 / 506.2).
func (e *Engine) runExtraPhases(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	for len(g.Turn.ExtraPhases) > 0 {
		phase := g.Turn.ExtraPhases[0]
		g.Turn.ExtraPhases = g.Turn.ExtraPhases[1:]
		switch phase {
		case game.PhaseCombat:
			e.runCombatPhase(g, agents, log)
		case game.PhasePrecombatMain, game.PhasePostcombatMain:
			e.runMainPhase(g, agents, phase, log)
		default:
		}
		if g.IsGameOver() {
			return
		}
	}
}

func (e *Engine) runBeginningPhase(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	g.Turn.Phase = game.PhaseBeginning

	g.Turn.Step = game.StepUntap
	expireTurnStartDurations(g)
	expireGoadForActivePlayer(g)
	for _, permanent := range g.Battlefield {
		if !permanent.PhasedOut {
			continue
		}
		phaseInFor := effectiveController(g, permanent)
		if permanent.PhaseInScheduled {
			phaseInFor = permanent.PhasedOutFor
		}
		if phaseInFor != g.Turn.ActivePlayer {
			continue
		}
		permanent.PhasedOut = false
		permanent.PhasedOutFor = 0
		permanent.PhaseInScheduled = false
		delete(g.LastKnownInformation, permanent.ObjectID)
		emitEvent(g, game.Event{
			Kind:        game.EventPermanentPhasedIn,
			Controller:  effectiveController(g, permanent),
			Player:      g.Turn.ActivePlayer,
			PermanentID: permanent.ObjectID,
			CardID:      permanent.CardInstanceID,
		})
	}
	for _, permanent := range g.Battlefield {
		if !activeBattlefieldPermanent(permanent) ||
			effectiveController(g, permanent) != g.Turn.ActivePlayer {
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
	untapDuringOtherPlayersUntapStep(g)

	g.Turn.Step = game.StepUpkeep
	// Beginning-of-step triggers fire at the start of the upkeep and are put on
	// the stack before the game advances to draw (CR 603.6c, CR 117.3b).
	emitBeginningOfStepEvent(g, game.StepUpkeep)
	e.processSuspendUpkeep(g, g.Turn.ActivePlayer)
	e.processReboundUpkeep(g, g.Turn.ActivePlayer, agents, log)
	g.Turn.PriorityPlayer = g.Turn.ActivePlayer
	e.runPriorityLoop(g, agents, log)
	if g.IsGameOver() {
		return
	}

	if !consumeSkipStep(g, g.Turn.ActivePlayer, game.StepDraw) {
		g.Turn.Step = game.StepDraw
		emitBeginningOfStepEvent(g, game.StepDraw)
		e.drawCardWithReplacements(g, g.Turn.ActivePlayer, agents, log, true)
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

// untapDuringOtherPlayersUntapStep grants the extra untap that Seedborn
// Muse-style static abilities give during every other player's untap step.
// Because the active player's permanents have already untapped, it untaps only
// the permanents of controllers other than the active player whose static
// RuleEffectUntapDuringOtherPlayersUntapStep applies. Summoning sickness is not
// cleared (that is tied to the controller's own turn) and exert is not consumed
// (it is spent only on the controller's own untap step).
func untapDuringOtherPlayersUntapStep(g *game.Game) {
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind != game.RuleEffectUntapDuringOtherPlayersUntapStep ||
			effect.Controller == g.Turn.ActivePlayer {
			continue
		}
		if effect.AffectedSource {
			if permanent, ok := permanentByObjectID(g, effect.AffectedObjectID); ok {
				untapForOtherPlayersStep(g, permanent)
			}
			continue
		}
		for _, permanent := range g.Battlefield {
			if !activeBattlefieldPermanent(permanent) ||
				!ruleEffectMatchesPermanent(g, effect, permanent) {
				continue
			}
			untapForOtherPlayersStep(g, permanent)
		}
	}
}

// untapForOtherPlayersStep untaps one permanent during another player's untap
// step, honoring the same untap prohibitions and stun-counter replacement the
// active player's own untap applies.
func untapForOtherPlayersStep(g *game.Game, permanent *game.Permanent) {
	if ruleEffectPreventsUntap(g, permanent) {
		return
	}
	if permanent.Tapped && permanent.Counters.Has(counter.Stun) {
		permanent.Counters.Remove(counter.Stun, 1)
		return
	}
	setPermanentTapped(g, permanent, false)
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
	// The monarch draws a card at the beginning of their end step (CR 720.5).
	// End steps occur only on the active player's turn, so this is the monarch's
	// end step exactly when the active player is the monarch.
	if monarch, ok := playerByID(g, g.Turn.ActivePlayer); ok && monarch.IsMonarch {
		e.drawCards(g, g.Turn.ActivePlayer, 1, agents, nil)
	}
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
		if activeBattlefieldPermanent(permanent) {
			permanent.MarkedDamage = 0
			permanent.MarkedDeathtouchDamage = false
		}
		permanent.TemporaryPowerModifier = 0
		permanent.TemporaryToughnessModifier = 0
		permanent.RegenerationShields = 0
		permanent.Saddled = false
	}
	expireCleanupDurations(g)
	expireEventDelayedTriggers(g)
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
	g.Turn.CombatPhasesThisTurn = 0
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

// emptyManaPools empties every player's mana pool. Each player's mana pool empties
// at the end of each step and phase, and any unspent mana is lost (CR 106.4); the
// engine calls this at those boundaries.
func emptyManaPools(g *game.Game) {
	for _, player := range g.Players {
		player.ManaPool.Empty()
		player.ManaRiders = nil
	}
}
