package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

const maximumHandSize = 7

func (e *Engine) runTurn(g *game.Game, agents [game.NumPlayers]PlayerAgent) (log TurnLog) {
	g.BeginPlayerControlTurn()
	log = TurnLog{
		TurnNumber:   g.Turn.TurnNumber,
		ActivePlayer: g.Turn.ActivePlayer,
	}
	// Count this turn toward the active player's own turn tally (CR 500.1). It
	// runs once per whole turn, including the game's first turn and any extra
	// turn the active player takes, so a per-player turn-ordinal condition
	// ("your first, second, or third turn of the game", Starting Town) reads the
	// controller's true turn count rather than the global turn number.
	if active, ok := playerByID(g, g.Turn.ActivePlayer); ok {
		active.TurnsTaken++
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
	e.runTurnFromPhase(g, agents, &log, game.PhasePrecombatMain)

	return log
}

// runTurnFromPhase plays the active player's turn from phase `from` onward and
// then advances to the next turn, skipping any earlier phase (whose priority the
// caller has already resolved). runTurn plays a whole turn; a search rollout
// resumes a turn it interrupted at a main phase (see Simulator.PlayForward). The
// beginning phase, which has steps a caller cannot cleanly resume past (untap,
// upkeep, draw), is only ever run by runTurn itself, so `from` is always a main
// phase or later.
func (e *Engine) runTurnFromPhase(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog, from game.Phase) {
	if from > game.PhasePrecombatMain {
		g.Turn.MainPhasesThisTurn = max(1, g.Turn.MainPhasesThisTurn)
	}
	g.Turn.MainPhaseContinuation = opt.Val(from)
	// Extra phases queued by the phase immediately before `from` run before the
	// next regular phase (CR 500.8). For a whole turn that preceding phase is the
	// beginning phase; for a resumed search rollout it is the main phase whose
	// priority the caller already resolved.
	e.runExtraPhases(g, agents, log)
	if g.IsGameOver() {
		return
	}
	if from <= game.PhasePrecombatMain {
		e.runMainPhase(g, agents, game.PhasePrecombatMain, log)
		recordManaDevelopment(g, log)
		if g.IsGameOver() {
			return
		}
		e.runExtraPhases(g, agents, log)
		if g.IsGameOver() {
			return
		}
	}
	if from <= game.PhaseCombat {
		e.runCombatPhase(g, agents, log)
		if g.IsGameOver() {
			return
		}
		g.Turn.MainPhaseContinuation = opt.Val(game.PhasePostcombatMain)
		e.runExtraPhases(g, agents, log)
		if g.IsGameOver() {
			return
		}
	}
	if from <= game.PhasePostcombatMain {
		e.runMainPhase(g, agents, game.PhasePostcombatMain, log)
		if g.IsGameOver() {
			return
		}
		e.runExtraPhases(g, agents, log)
		if g.IsGameOver() {
			return
		}
	}
	e.runEndingPhaseWithLog(g, agents, log)
	if g.IsGameOver() {
		return
	}
	e.advanceToNextTurn(g)
}

// finishCurrentTurn plays the active player's turn to its end from the phase
// after the current one — the caller having already resolved the current phase's
// priority — and advances to the next turn. It is used by a search rollout that
// interrupted a turn at a main phase to evaluate a candidate action; the agents
// drive any choices the tail of the turn requires (blockers, discards, triggers).
func (e *Engine) finishCurrentTurn(g *game.Game, agents [game.NumPlayers]PlayerAgent) {
	if !g.Turn.IsMainPhase() || g.Turn.Step != game.StepNone {
		panic("cannot resume a turn outside a resolved main-phase priority window")
	}
	if g.IsGameOver() {
		return
	}
	log := TurnLog{TurnNumber: g.Turn.TurnNumber, ActivePlayer: g.Turn.ActivePlayer}
	e.setReplacementChoiceContext(g, agents, &log)
	defer g.ClearChoiceContext()
	completeTurnBoundary(g)
	// The current phase's priority and exit obligations have completed.
	next := g.Turn.Phase + 1
	if g.Turn.MainPhaseContinuation.Exists {
		next = g.Turn.MainPhaseContinuation.Val
	}
	e.runTurnFromPhase(g, agents, &log, next)
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
			next := g.Turn.Phase + 1
			if g.Turn.MainPhaseContinuation.Exists {
				next = g.Turn.MainPhaseContinuation.Val
			}
			e.runMainPhaseWithContinuation(g, agents, phase, next, log)
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
		setPermanentTapped(g, permanent, false)
		permanent.SummoningSick = false
	}
	untapDuringOtherPlayersUntapStep(g)
	completeTurnBoundary(g)

	e.runUpkeepStep(g, agents, log)
	if g.IsGameOver() {
		return
	}
	// Drain any additional upkeep steps queued during this upkeep (Paradox Haze:
	// "that player gets an additional upkeep step after this step."). Each extra
	// upkeep step runs immediately after the current one and before the draw step
	// (CR 505.5b), and re-runs every "at the beginning of upkeep" trigger and
	// turn-based upkeep action; because runUpkeepStep increments
	// UpkeepStepsThisTurn, the "first upkeep each turn" trigger does not re-fire,
	// so the loop terminates. A drained step may itself queue more (another
	// source's additional upkeep), which the loop then drains in turn.
	for g.Turn.ExtraUpkeepSteps > 0 {
		g.Turn.ExtraUpkeepSteps--
		e.runUpkeepStep(g, agents, log)
		if g.IsGameOver() {
			return
		}
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
		g.Turn.PriorityPlayer = g.Turn.ActivePlayer
		e.runPriorityLoop(g, agents, log)
		if g.IsGameOver() {
			return
		}
		completeTurnBoundary(g)
	}
	e.applyStateBasedActionsWithLog(g, log)
	completeTurnBoundary(g)
}

// runUpkeepStep runs a single upkeep step of the active player's beginning phase:
// the turn's normal upkeep and every additional upkeep step drained from
// TurnState.ExtraUpkeepSteps share this body. It counts the step toward
// UpkeepStepsThisTurn (backing the "first upkeep each turn" gate), emits the
// beginning-of-step event so "at the beginning of upkeep" triggers fire each
// step, performs the upkeep turn-based actions (initiative venture, suspend and
// rebound processing), and resolves the upkeep priority window.
func (e *Engine) runUpkeepStep(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	g.Turn.UpkeepStepsThisTurn++
	g.Turn.Step = game.StepUpkeep
	log.addStep(game.StepUpkeep)
	// Beginning-of-step triggers fire at the start of the upkeep and are put on
	// the stack before the game advances to draw (CR 603.6c, CR 117.3b).
	emitBeginningOfStepEvent(g, game.StepUpkeep)
	// The initiative-holder ventures into Undercity at the beginning of their
	// upkeep (CR 720). Upkeeps occur only on the active player's turn, so this is
	// the initiative-holder's upkeep exactly when the active player has the
	// initiative. The venture is queued and resolved by the upcoming trigger pass,
	// where player choices are available.
	if holder, ok := playerByID(g, g.Turn.ActivePlayer); ok && holder.HasInitiative {
		queueInitiativeVenture(g, g.Turn.ActivePlayer)
	}
	e.processSuspendUpkeep(g, g.Turn.ActivePlayer)
	e.processReboundUpkeep(g, g.Turn.ActivePlayer, agents, log)
	g.Turn.PriorityPlayer = g.Turn.ActivePlayer
	e.runPriorityLoop(g, agents, log)
	completeTurnBoundary(g)
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
	seen := make(map[*game.Permanent]bool)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind != game.RuleEffectUntapDuringOtherPlayersUntapStep ||
			effect.Controller == g.Turn.ActivePlayer {
			continue
		}
		if effect.AffectedSource {
			if permanent, ok := permanentByObjectID(g, effect.AffectedObjectID); ok && !seen[permanent] {
				seen[permanent] = true
				untapForOtherPlayersStep(g, permanent)
			}
			continue
		}
		for _, permanent := range g.Battlefield {
			if !activeBattlefieldPermanent(permanent) ||
				!ruleEffectMatchesPermanent(g, effect, permanent) || seen[permanent] {
				continue
			}
			seen[permanent] = true
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
	setPermanentTapped(g, permanent, false)
}

func (e *Engine) runMainPhase(g *game.Game, agents [game.NumPlayers]PlayerAgent, phase game.Phase, log *TurnLog) {
	e.runMainPhaseWithContinuation(g, agents, phase, phase+1, log)
}

func (e *Engine) runMainPhaseWithContinuation(g *game.Game, agents [game.NumPlayers]PlayerAgent, phase, next game.Phase, log *TurnLog) {
	if phase != game.PhasePrecombatMain && phase != game.PhasePostcombatMain {
		panic("main-phase entry requested for a non-main phase")
	}
	g.Turn.MainPhaseContinuation = opt.Val(next)
	phase = game.PhasePostcombatMain
	if g.Turn.MainPhasesThisTurn == 0 {
		phase = game.PhasePrecombatMain
	}
	g.Turn.MainPhasesThisTurn++
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
	if phase == game.PhasePrecombatMain {
		advanceSagas(g, g.Turn.ActivePlayer)
	}
	// Main phases have no steps. Reset the synthetic trigger boundary before
	// priority so sorcery-speed actions remain legal.
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = g.Turn.ActivePlayer
	e.runPriorityLoop(g, agents, log)
	completeTurnBoundary(g)
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
	e.putTriggeredAbilitiesOnStackWithChoices(g, agents, log)
	g.Turn.PriorityPlayer = g.Turn.ActivePlayer
	e.runPriorityLoop(g, agents, log)
	if g.IsGameOver() {
		return
	}
	completeTurnBoundary(g)

	for {
		applied := e.runCleanupStep(g, agents, log)
		triggered := e.putTriggeredAbilitiesOnStackWithChoices(g, agents, log)
		if g.IsGameOver() || (!applied && !triggered) {
			completeTurnBoundary(g)
			return
		}
		// CR 514.3a: cleanup SBAs or triggers open priority; after everyone
		// passes on an empty stack, begin another cleanup step. Turn control
		// remains active through every such step until the next turn begins.
		g.Turn.PriorityPlayer = g.Turn.ActivePlayer
		e.runPriorityLoop(g, agents, log)
		if g.IsGameOver() {
			return
		}
		completeTurnBoundary(g)
	}
}

func (e *Engine) runCleanupStep(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	g.Turn.Step = game.StepCleanup
	log.addStep(game.StepCleanup)
	e.discardToMaximumHandSizeWithChoices(g, g.Turn.ActivePlayer, agents, log)
	for _, permanent := range g.Battlefield {
		if activeBattlefieldPermanent(permanent) {
			permanent.MarkedDamage = 0
			permanent.MarkedDeathtouchDamage = false
		}
		permanent.TemporaryPowerModifier = 0
		permanent.TemporaryToughnessModifier = 0
		permanent.RegenerationShields = 0
		permanent.Saddled = false
		permanent.SaddleContributorIDs = nil
	}
	expireCleanupDurations(g)
	_, applied := e.applyStateBasedActionsWithLogResult(g, log)
	clearPersistentMana(g)
	g.Combat = nil
	return applied
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

func (e *Engine) discardToMaximumHandSizeWithChoices(g *game.Game, playerID game.PlayerID, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	player, ok := playerByID(g, playerID)
	if !ok || player.Eliminated || player.Hand.Size() <= maximumHandSize || playerHasNoMaximumHandSize(g, playerID) {
		return
	}
	cards := player.Hand.All()
	count := len(cards) - maximumHandSize
	options := make([]game.ChoiceOption, len(cards))
	for i, card := range cards {
		options[i] = game.ChoiceOption{Index: i, Card: cardChoiceInfo(g, card)}
	}
	defaults := make([]int, count)
	for i := range defaults {
		defaults[i] = len(cards) - 1 - i
	}
	selected := e.chooseChoice(g, agents, game.ChoiceRequest{
		Kind: game.ChoiceZoneSelection, Player: playerID, Prompt: "Discard to maximum hand size.",
		Options: options, MinChoices: count, MaxChoices: count, DefaultSelection: defaults,
	}, log)
	for _, index := range selected {
		discardCardFromHand(g, playerID, cards[index])
	}
}

func (*Engine) advanceToNextTurn(g *game.Game) {
	next := nextActualTurn(g)

	g.Turn.TurnNumber++
	g.Turn.ActivePlayer = next
	g.Turn.PriorityPlayer = next
	g.Turn.Phase = game.PhaseBeginning
	g.Turn.Step = game.StepUntap
	g.Turn.LandsPlayedThisTurn = 0
	g.Turn.LandsAllowedThisTurn = 1
	g.Turn.MainPhasesThisTurn = 0
	g.Turn.MainPhaseContinuation = opt.V[game.Phase]{}
	g.Turn.ExtraPhases = nil
	g.Turn.CombatPhasesThisTurn = 0
	g.Turn.UpkeepStepsThisTurn = 0
	g.Turn.ExtraUpkeepSteps = 0
	g.Turn.MonarchAtTurnStart = currentMonarch(g)
	for i := range g.Players {
		g.Players[i].CantBecomeMonarchThisTurn = false
	}
	g.ActivatedAbilitiesThisTurn = make(map[game.ActivatedAbilityUse]bool)
	g.AbilityActivationsThisTurn = make(map[game.ActivatedAbilityUse]int)
	g.ExilePlayPermissionUsedThisTurn = make(map[game.ObjectID]bool)
	g.PermanentTypePlayUsesThisTurn = make(map[game.PlayPermissionUse]bool)
	g.TriggeredAbilitiesThisTurn = make(map[game.TriggeredAbilityUse]int)
	g.ResolvedTriggeredAbilitiesThisTurn = make(map[game.TriggeredAbilityUse]int)
	g.ResolvedActivatedAbilitiesThisTurn = make(map[game.ActivatedAbilityResolutionUse]int)
	g.ChosenModesThisTurn = make(map[game.TriggeredAbilityUse]uint64)
	g.Combat = nil
	g.BeginPlayerControlTurn()
	markCurrentTurnEventStart(g)
}

func nextActualTurn(g *game.Game) game.PlayerID {
	regular := g.Turn.ActivePlayer
	for {
		next, extra := popExtraTurn(g)
		if !extra {
			next = nextRegularTurn(g, regular)
			regular = next
		}
		if g.SkippedTurns[next] > 0 {
			g.SkippedTurns[next]--
			continue
		}
		return next
	}
}

func popExtraTurn(g *game.Game) (game.PlayerID, bool) {
	for len(g.Turn.ExtraTurns) > 0 {
		last := len(g.Turn.ExtraTurns) - 1
		next := g.Turn.ExtraTurns[last]
		g.Turn.ExtraTurns = g.Turn.ExtraTurns[:last]
		if departedPlayer(g, next) {
			expireDepartedNextTurnDurations(g, next)
		}
		if !g.TurnOrder.IsEliminated(next) && !playerSkipsExtraTurns(g, next) {
			return next, true
		}
	}
	return 0, false
}

func nextRegularTurn(g *game.Game, current game.PlayerID) game.PlayerID {
	index := slices.Index(g.TurnOrder.Order[:], current)
	for offset := 1; offset <= game.NumPlayers; offset++ {
		next := g.TurnOrder.Order[(index+offset)%game.NumPlayers]
		if departedPlayer(g, next) {
			expireDepartedNextTurnDurations(g, next)
			continue
		}
		return next
	}
	return current
}

func playerSkipsExtraTurns(g *game.Game, player game.PlayerID) bool {
	for _, effect := range activeRuleEffects(g) {
		if effect.Kind == game.RuleEffectSkipExtraTurns &&
			playerRelationMatches(effect.Controller, player, effect.AffectedPlayer) {
			return true
		}
	}
	return false
}

// completeTurnBoundary finishes a step or phase without advancing or repeating
// its entry actions. Live execution and main-phase Simulator resumes share it.
func completeTurnBoundary(g *game.Game) {
	emptyManaPools(g)
}

func emptyManaPools(g *game.Game) {
	for _, player := range g.Players {
		player.ManaPool.Empty()
		if len(player.ManaRiders) == 0 {
			continue
		}
		remaining := player.ManaPool.Units()
		kept := player.ManaRiders[:0]
		for _, rider := range player.ManaRiders {
			if remaining[rider.Unit] > 0 {
				kept = append(kept, rider)
				remaining[rider.Unit]--
			}
		}
		player.ManaRiders = kept
	}
}

// clearPersistentMana releases every player's until-end-of-turn mana reservation
// (Pool.AddPersistent) at end-of-turn cleanup, so the reserved mana empties like
// any other mana as the following step or phase ends (CR 500.4 exception ending,
// CR 514.2). It is called during the cleanup turn-based action; remaining mana
// is spendable in exceptional cleanup priority and empties when the step ends.
func clearPersistentMana(g *game.Game) {
	for _, player := range g.Players {
		player.ManaPool.ClearPersistent()
	}
}
