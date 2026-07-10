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
	// Make the player agents reachable from the replacement-selection chokepoint
	// (CR 616.1) for the duration of the turn; see replacement_choice.go.
	e.setReplacementChoiceContext(g, agents, &log)
	defer func() {
		g.ClearChoiceContext()
		log.ManaSpent = g.Players[activePlayer].ManaPool.Spent() - manaSpentBefore
	}()

	e.runBeginningPhase(g, agents, &log)
	if g.IsGameOver() {
		return log
	}
	e.runExtraPhases(g, agents, &log)
	if g.IsGameOver() {
		return log
	}
	e.runMainPhase(g, agents, game.PhasePrecombatMain, &log)
	recordManaDevelopment(g, &log)
	if g.IsGameOver() {
		return log
	}
	e.runExtraPhases(g, agents, &log)
	if g.IsGameOver() {
		return log
	}
	e.runCombatPhase(g, agents, &log)
	if g.IsGameOver() {
		return log
	}
	e.runExtraPhases(g, agents, &log)
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
	e.runEndingPhaseWithLog(g, agents, &log)
	if g.IsGameOver() {
		return log
	}
	e.advanceToNextTurn(g)

	return log
}

// runExtraPhases drains the additional phases queued onto the turn by
// extra-phase effects ("After this main phase, there is an additional combat
// phase followed by an additional main phase." — Aggravated Assault; "there is
// an additional beginning phase after this phase." — Sphinx of the Second Sun,
// Cyclonus, Cybertronian Fighter). It is called after each base phase, so a
// queued phase runs immediately after the phase during which it was queued
// (CR 500.7): an "after this phase" effect that resolves during combat (Éomer,
// Marshal of Rohan; Cyclonus) runs before the postcombat main phase, while an
// effect that resolves during the postcombat main phase (Sphinx, Aggravated
// Assault) runs after it. Because each call empties the queue, calls after
// phases that queued nothing are no-ops. A queued main phase that re-activates
// the source re-queues more phases, so the loop continues until the queue empties
// (the extra-combat combo, CR 505.5 / 506.2).
func (e *Engine) runExtraPhases(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	for len(g.Turn.ExtraPhases) > 0 {
		phase := g.Turn.ExtraPhases[0]
		g.Turn.ExtraPhases = g.Turn.ExtraPhases[1:]
		switch phase {
		case game.PhaseBeginning:
			e.runBeginningPhase(g, agents, log)
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
	log.addPhase(game.PhaseBeginning)

	g.Turn.Step = game.StepUntap
	log.addStep(game.StepUntap)
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
	log.addStep(game.StepUpkeep)
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

	// Consume any scheduled one-shot draw-step skip unconditionally: a queued
	// "skip your next draw step" applies to this turn's draw step even when a
	// static "Skip your draw step." effect (Necropotence, Yawgmoth's Bargain)
	// also skips it, so it must not survive to skip a later, unintended draw
	// step. Either source skips the draw step entirely (CR 500.8).
	scheduledSkip := consumeSkipStep(g, g.Turn.ActivePlayer, game.StepDraw)
	if !playerSkipsDrawStep(g, g.Turn.ActivePlayer) && !scheduledSkip {
		g.Turn.Step = game.StepDraw
		log.addStep(game.StepDraw)
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
	log.addPhase(phase)
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
	e.runEndingPhaseWithLog(g, agents, nil)
}

func (e *Engine) runEndingPhaseWithLog(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	g.Turn.Phase = game.PhaseEnding
	log.addPhase(game.PhaseEnding)
	g.Turn.Step = game.StepEnd
	log.addStep(game.StepEnd)
	// "At the beginning of the end step" triggers use the same event as delayed
	// next-end-step triggers before the end-step priority window (CR 603.6c,
	// CR 603.7b).
	emitBeginningOfStepEvent(g, game.StepEnd)
	// The monarch draws a card at the beginning of their end step (CR 720.5).
	// End steps occur only on the active player's turn, so this is the monarch's
	// end step exactly when the active player is the monarch.
	if monarch, ok := playerByID(g, g.Turn.ActivePlayer); ok && monarch.IsMonarch {
		e.drawCards(g, g.Turn.ActivePlayer, 1, agents, log)
	}
	if e.putTriggeredAbilitiesOnStackWithChoices(g, agents, log) || !g.Stack.IsEmpty() {
		g.Turn.PriorityPlayer = g.Turn.ActivePlayer
		e.runPriorityLoop(g, agents, log)
		if g.IsGameOver() {
			return
		}
	}

	g.Turn.Step = game.StepCleanup
	log.addStep(game.StepCleanup)
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
	e.applyStateBasedActionsWithLog(g, log)
	clearPersistentMana(g)
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
	g.Turn.MonarchAtTurnStart = currentMonarch(g)
	for i := range g.Players {
		g.Players[i].CantBecomeMonarchThisTurn = false
	}
	g.ActivatedAbilitiesThisTurn = make(map[game.ActivatedAbilityUse]bool)
	g.AbilityActivationsThisTurn = make(map[game.ActivatedAbilityUse]int)
	g.ExilePlayPermissionUsedThisTurn = make(map[game.ObjectID]bool)
	g.TriggeredAbilitiesThisTurn = make(map[game.TriggeredAbilityUse]int)
	g.ResolvedTriggeredAbilitiesThisTurn = make(map[game.TriggeredAbilityUse]int)
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

// clearPersistentMana releases every player's until-end-of-turn mana reservation
// (Pool.AddPersistent) at end-of-turn cleanup, so the reserved mana empties like
// any other mana as the following step or phase ends (CR 500.4 exception ending,
// CR 514.2). It is called during the cleanup step immediately before
// emptyManaPools; for pools that never received persistent mana it is a no-op.
func clearPersistentMana(g *game.Game) {
	for _, player := range g.Players {
		player.ManaPool.ClearPersistent()
	}
}
