package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/rules/payment"
)

// combatEngine concentrates combat-phase behaviour for Engine. It is created
// once per combat phase and owns combat orchestration, declaration
// validation/application, legal-action enumeration, damage resolution, and
// attack-tax integration. Priority-loop management is delegated back to the
// outer Engine.
//
// Keeping the type in-package lets later work extract it to a subpackage once
// the surface area and dependencies are stable. The extraction decision
// criteria are documented in README.md.
type combatEngine struct {
	// e is required for phase orchestration and declaration application. Legal
	// action enumeration helpers used by tests may use a zero combatEngine.
	e *Engine
}

// runPhase drives the full combat phase: beginning-of-combat priority,
// attacker declaration, blocker declaration, optional first-strike damage,
// normal-damage, and end-of-combat priority. It initialises and clears
// g.Combat; callers must not touch g.Combat while this is running.
func (ce combatEngine) runPhase(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	g.Turn.Phase = game.PhaseCombat
	g.Combat = &game.CombatState{}
	defer func() {
		g.Combat = nil
	}()

	if !ce.runPriorityStep(g, agents, log, game.StepBeginningOfCombat) {
		return
	}

	g.Turn.Step = game.StepDeclareAttackers
	ce.declareAttackers(g, agents, log)
	if !ce.runPriority(g, agents, log) {
		return
	}
	emptyManaPools(g)

	g.Turn.Step = game.StepDeclareBlockers
	ce.declareBlockers(g, agents, log)
	if !ce.runPriority(g, agents, log) {
		return
	}
	emptyManaPools(g)

	if combatHasFirstStrikeDamage(g) {
		g.Turn.Step = game.StepFirstStrikeDamage
		ce.resolveDamagePass(g, firstStrikeCombatDamage, log)
		ce.e.applyStateBasedActionsWithLog(g, log)
		if g.IsGameOver() {
			return
		}
		if !ce.runPriority(g, agents, log) {
			return
		}
		emptyManaPools(g)
	}

	g.Turn.Step = game.StepCombatDamage
	ce.resolveDamagePass(g, normalCombatDamage, log)
	ce.e.applyStateBasedActionsWithLog(g, log)
	if g.IsGameOver() {
		return
	}
	if !ce.runPriority(g, agents, log) {
		return
	}
	emptyManaPools(g)

	ce.runPriorityStep(g, agents, log, game.StepEndOfCombat)
}

// runPriorityStep sets the current step, emits the beginning-of-step event,
// runs the priority loop, and empties mana pools. It returns false if the game
// ended during the priority window.
func (ce combatEngine) runPriorityStep(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog, step game.Step) bool {
	g.Turn.Step = step
	emitBeginningOfStepEvent(g, step)
	if !ce.runPriority(g, agents, log) {
		return false
	}
	emptyManaPools(g)
	return true
}

// runPriority gives priority to the active player and runs the priority loop.
// It returns false if the game ended during the window.
func (ce combatEngine) runPriority(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	g.Turn.PriorityPlayer = g.Turn.ActivePlayer
	ce.e.runPriorityLoop(g, agents, log)
	return !g.IsGameOver()
}

// declareAttackers runs the declare-attackers turn-based action: it enumerates
// legal attacker choices, asks the active player to pick one, logs it, and
// applies it.
func (ce combatEngine) declareAttackers(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	playerID := g.Turn.ActivePlayer
	legal := ce.legalAttackers(g, playerID)
	if len(legal) == 0 {
		return
	}

	chosen := legal[len(legal)-1]
	if agent := agentFor(agents, playerID); agent != nil {
		chosen = agent.ChooseAction(observe(g, playerID), legal)
	}
	if !containsAction(legal, chosen) {
		chosen = legal[len(legal)-1]
	}

	actionLog := combatActionLog(g, playerID, chosen)
	log.addAction(&actionLog)

	attackers, ok := chosen.DeclareAttackersPayload()
	if !ok || !ce.applyAttackers(g, playerID, attackers) {
		panic("applyAttackers failed for validated action")
	}
	ce.e.notifyActionObservers(g, agents, playerID, chosen)
}

// declareBlockers runs the declare-blockers turn-based action for each
// defending player in priority order.
func (ce combatEngine) declareBlockers(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	for _, playerID := range defendingPlayersInOrder(g) {
		legal := ce.legalBlockers(g, playerID)
		if len(legal) == 0 {
			continue
		}

		chosen := legal[len(legal)-1]
		if agent := agentFor(agents, playerID); agent != nil {
			chosen = agent.ChooseAction(observe(g, playerID), legal)
		}
		if !containsAction(legal, chosen) {
			chosen = legal[len(legal)-1]
		}

		actionLog := combatActionLog(g, playerID, chosen)
		log.addAction(&actionLog)

		blockers, ok := chosen.DeclareBlockersPayload()
		if !ok || !ce.applyBlockers(g, playerID, blockers) {
			panic("applyBlockers failed for validated action")
		}
		ce.e.notifyActionObservers(g, agents, playerID, chosen)
	}
}

// legalAttackers returns the legal declare-attackers actions for playerID.
// The last element is always the no-attack action (or the only action when
// goad forces that result).
func (ce combatEngine) legalAttackers(g *game.Game, playerID game.PlayerID) []action.Action {
	if !canDeclareAttackers(g, playerID) {
		return nil
	}

	attackers := eligibleAttackers(g, playerID)
	targets := legalAttackTargets(g, playerID)
	actions := make([]action.Action, 0, len(targets)+1)
	eligibleByID := permanentMapByObjectID(attackers)
	if len(attackers) > 0 {
		for _, target := range targets {
			declarations := make([]game.AttackDeclaration, 0, len(attackers))
			for _, attacker := range attackers {
				single := []game.AttackDeclaration{{
					Attacker: attacker.ObjectID,
					Target:   target,
				}}
				if !canAttackTarget(g, attacker, target) {
					continue
				}
				if len(attackers) > 1 && ce.declareAttackersSatisfiesRequirements(g, playerID, single, eligibleByID) {
					act := actionBuild.declareAttackers(single)
					if !containsAction(actions, act) && ce.canPayAttackTax(g, playerID, single) {
						actions = append(actions, act)
					}
				}
				declarations = append(declarations, single[0])
			}
			if ce.declareAttackersSatisfiesRequirements(g, playerID, declarations, eligibleByID) {
				act := actionBuild.declareAttackers(declarations)
				if !containsAction(actions, act) && ce.canPayAttackTax(g, playerID, declarations) {
					actions = append(actions, act)
				}
			}
		}
	}
	if ce.declareAttackersSatisfiesRequirements(g, playerID, nil, eligibleByID) {
		actions = append(actions, actionBuild.declareAttackers(nil))
	} else if len(actions) == 0 {
		if declarations := preferredRequiredAttackDeclarations(g, playerID, attackers); len(declarations) > 0 {
			if ce.canPayAttackTax(g, playerID, declarations) {
				actions = append(actions, actionBuild.declareAttackers(declarations))
			}
		}
	}
	return actions
}

// legalBlockers returns the legal declare-blockers actions for playerID.
// The last element is always the no-block action.
func (combatEngine) legalBlockers(g *game.Game, playerID game.PlayerID) []action.Action {
	if !canDeclareBlockers(g, playerID) {
		return nil
	}

	attackers := attacksAgainstPlayer(g, playerID)
	blockers := eligibleBlockers(g, playerID)
	required := satisfiableMustBlockAttackers(g, playerID, attackers, blockers)
	actions := make([]action.Action, 0, len(attackers)*len(blockers)+1)
	for _, attacker := range attackers {
		var allBlockers []game.BlockDeclaration
		attackingPermanent, ok := permanentByObjectID(g, attacker.Attacker)
		if !ok {
			continue
		}
		for _, blocker := range blockers {
			if !canBlockAttacker(g, blocker, attackingPermanent) {
				continue
			}
			block := game.BlockDeclaration{
				Blocker:  blocker.ObjectID,
				Blocking: attacker.Attacker,
			}
			allBlockers = append(allBlockers, block)
			if !attackerRequiresMultipleBlockers(g, attackingPermanent) {
				declaration := []game.BlockDeclaration{block}
				if blockDeclarationsSatisfyMustBlockRequirements(required, declaration) {
					actions = append(actions, actionBuild.declareBlockers(declaration))
				}
			}
		}
		if len(allBlockers) > 1 {
			if blockDeclarationsSatisfyMustBlockRequirements(required, allBlockers) {
				actions = append(actions, actionBuild.declareBlockers(allBlockers))
			}
		}
	}
	if len(required) == 0 {
		actions = append(actions, actionBuild.declareBlockers(nil))
	}
	return actions
}

func satisfiableMustBlockAttackers(g *game.Game, playerID game.PlayerID, attackers []game.AttackDeclaration, blockers []*game.Permanent) map[id.ID]bool {
	required := make(map[id.ID]bool)
	for _, attack := range attackers {
		attacker, ok := permanentByObjectID(g, attack.Attacker)
		if !ok || !ruleEffectRequiresBeingBlocked(g, attacker) {
			continue
		}
		legalBlockerCount := 0
		for _, blocker := range blockers {
			if canBlockAttacker(g, blocker, attacker) {
				legalBlockerCount++
			}
		}
		if legalBlockerCount == 0 {
			continue
		}
		if attackerRequiresMultipleBlockers(g, attacker) && legalBlockerCount < 2 {
			continue
		}
		required[attack.Attacker] = true
	}
	return required
}

func blockDeclarationsSatisfyMustBlockRequirements(required map[id.ID]bool, declarations []game.BlockDeclaration) bool {
	if len(required) == 0 {
		return true
	}
	// The current blocker enumerator models one attacker at a time. Satisfying at
	// least one satisfiable requirement is exact for single-requirement effects
	// like Neyith; multiple simultaneous requirements need broader enumeration.
	for _, declaration := range declarations {
		if required[declaration.Blocking] {
			return true
		}
	}
	return false
}

// applyAttackers validates and applies the declare-attackers action for
// playerID. It pays any attack tax and emits attacker-declared events.
func (ce combatEngine) applyAttackers(g *game.Game, playerID game.PlayerID, declare action.DeclareAttackersAction) bool {
	if !canDeclareAttackers(g, playerID) {
		return false
	}

	eligibleByID := make(map[id.ID]*game.Permanent)
	for _, attacker := range eligibleAttackers(g, playerID) {
		eligibleByID[attacker.ObjectID] = attacker
	}

	seen := make(map[id.ID]bool)
	for _, declaration := range declare.Attackers {
		if seen[declaration.Attacker] {
			return false
		}
		seen[declaration.Attacker] = true

		if eligibleByID[declaration.Attacker] == nil {
			return false
		}
		if !isLegalAttackTarget(g, playerID, declaration.Target) {
			return false
		}
		if !canAttackTarget(g, eligibleByID[declaration.Attacker], declaration.Target) {
			return false
		}
	}
	if !ce.declareAttackersSatisfiesRequirements(g, playerID, declare.Attackers, eligibleByID) {
		return false
	}
	if tax, ok := ce.attackTaxCost(g, declare.Attackers); ok {
		if !ce.payAttackTax(g, playerID, declare.Attackers, tax) {
			return false
		}
	}

	g.Combat.Attackers = append([]game.AttackDeclaration(nil), declare.Attackers...)
	simultaneousID := id.ID(0)
	if len(declare.Attackers) > 1 {
		simultaneousID = g.IDGen.Next()
	}
	for _, declaration := range declare.Attackers {
		attacker := eligibleByID[declaration.Attacker]
		if !hasKeyword(g, attacker, game.Vigilance) {
			setPermanentTapped(g, attacker, true)
		}

		emitEvent(g, game.Event{
			Kind:           game.EventAttackerDeclared,
			SourceID:       attacker.CardInstanceID,
			SourceObjectID: attacker.ObjectID,
			Controller:     effectiveController(g, attacker),
			Player:         declaration.Target.Player,
			PermanentID:    attacker.ObjectID,
			AttackTarget:   declaration.Target,
			SimultaneousID: simultaneousID,
		})
	}
	return true
}

// applyBlockers validates and applies the declare-blockers action for
// playerID. It emits blocker-declared events.
func (combatEngine) applyBlockers(g *game.Game, playerID game.PlayerID, declare action.DeclareBlockersAction) bool {
	if !canDeclareBlockers(g, playerID) {
		return false
	}

	eligibleByID := make(map[id.ID]*game.Permanent)
	for _, blocker := range eligibleBlockers(g, playerID) {
		eligibleByID[blocker.ObjectID] = blocker
	}
	attackersByID := make(map[id.ID]bool)
	for _, attack := range attacksAgainstPlayer(g, playerID) {
		attackersByID[attack.Attacker] = true
	}
	alreadyBlocking := make(map[id.ID]bool)
	blockerCounts := make(map[id.ID]int)
	for _, block := range g.Combat.Blockers {
		alreadyBlocking[block.Blocker] = true
		blockerCounts[block.Blocking]++
	}

	seenBlockers := make(map[id.ID]bool)
	for _, block := range declare.Blockers {
		if seenBlockers[block.Blocker] || alreadyBlocking[block.Blocker] {
			return false
		}
		seenBlockers[block.Blocker] = true
		if eligibleByID[block.Blocker] == nil {
			return false
		}
		if !attackersByID[block.Blocking] {
			return false
		}
		attacker, ok := permanentByObjectID(g, block.Blocking)
		if !ok || !canBlockAttacker(g, eligibleByID[block.Blocker], attacker) {
			return false
		}
		blockerCounts[block.Blocking]++
	}
	for attackerID, count := range blockerCounts {
		attacker, ok := permanentByObjectID(g, attackerID)
		if ok && count > 0 && count < 2 && attackerRequiresMultipleBlockers(g, attacker) {
			return false
		}
	}
	allBlockers := append([]game.BlockDeclaration(nil), g.Combat.Blockers...)
	allBlockers = append(allBlockers, declare.Blockers...)
	if !blockDeclarationsSatisfyMustBlockRequirements(satisfiableMustBlockAttackers(g, playerID, attacksAgainstPlayer(g, playerID), eligibleBlockers(g, playerID)), allBlockers) {
		return false
	}

	g.Combat.Blockers = append(g.Combat.Blockers, declare.Blockers...)
	if len(declare.Blockers) > 0 && g.Combat.BlockedAttackers == nil {
		g.Combat.BlockedAttackers = make(map[id.ID]bool)
	}
	if len(declare.Blockers) > 0 && g.Combat.BlockerOrder == nil {
		g.Combat.BlockerOrder = make(map[id.ID][]id.ID)
	}
	if len(declare.Blockers) > 0 && g.Combat.BlockDeclarationBatchID == 0 {
		g.Combat.BlockDeclarationBatchID = g.IDGen.Next()
	}
	for _, block := range declare.Blockers {
		attackerBecameBlocked := !g.Combat.BlockedAttackers[block.Blocking]
		g.Combat.BlockedAttackers[block.Blocking] = true
		g.Combat.BlockerOrder[block.Blocking] = append(g.Combat.BlockerOrder[block.Blocking], block.Blocker)
		if attackerBecameBlocked {
			if attacker, ok := permanentByObjectID(g, block.Blocking); ok {
				emitEvent(g, game.Event{
					Kind:               game.EventAttackerBecameBlocked,
					SourceID:           attacker.CardInstanceID,
					SourceObjectID:     attacker.ObjectID,
					Controller:         effectiveController(g, attacker),
					PermanentID:        attacker.ObjectID,
					RelatedPermanentID: block.Blocker,
					SimultaneousID:     g.Combat.BlockDeclarationBatchID,
				})
			}
		}
		blocker := eligibleByID[block.Blocker]
		if blocker == nil {
			continue
		}
		emitEvent(g, game.Event{
			Kind:               game.EventBlockerDeclared,
			SourceID:           blocker.CardInstanceID,
			SourceObjectID:     blocker.ObjectID,
			Controller:         effectiveController(g, blocker),
			PermanentID:        blocker.ObjectID,
			RelatedPermanentID: block.Blocking,
			BlockedAttackerID:  block.Blocking,
			SimultaneousID:     g.Combat.BlockDeclarationBatchID,
		})
	}
	return true
}

// resolveDamagePass assigns and marks combat damage for all attackers in the
// given damage pass (first-strike or normal).
func (combatEngine) resolveDamagePass(g *game.Game, pass combatDamagePass, log *TurnLog) {
	if g.Combat == nil {
		return
	}
	eventStart := len(g.Events)
	blockerMap := blockersByAttacker(g)
	for _, declaration := range g.Combat.Attackers {
		attacker, ok := permanentByObjectID(g, declaration.Attacker)
		if !ok || attacker.PhasedOut {
			continue
		}
		blockers := blockerMap[declaration.Attacker]
		if attackerWasBlocked(g, declaration.Attacker) {
			resolveBlockedCombatDamage(g, attacker, blockers, declaration.Target, pass, log)
			continue
		}
		if !isLegalAttackTarget(g, effectiveController(g, attacker), declaration.Target) {
			continue
		}
		resolveUnblockedCombatDamage(g, attacker, declaration.Target, pass, log)
	}
	batchCombatDamageEvents(g, eventStart)
}

func batchCombatDamageEvents(g *game.Game, eventStart int) {
	damageEvents := 0
	for i := eventStart; i < len(g.Events); i++ {
		if g.Events[i].Kind == game.EventDamageDealt && g.Events[i].CombatDamage {
			damageEvents++
		}
	}
	if damageEvents < 2 {
		return
	}
	simultaneousID := g.IDGen.Next()
	for i := eventStart; i < len(g.Events); i++ {
		if g.Events[i].Kind == game.EventDamageDealt && g.Events[i].CombatDamage {
			g.Events[i].SimultaneousID = simultaneousID
		}
	}
}

// canPayAttackTax reports whether playerID can currently pay the attack tax
// for the given attack declarations.
func (ce combatEngine) canPayAttackTax(g *game.Game, playerID game.PlayerID, declarations []game.AttackDeclaration) bool {
	manaCost, ok := ce.attackTaxCost(g, declarations)
	if !ok {
		return true
	}
	return paymentOrch.canPayGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: manaCost, Exclude: ce.attackingPermanentExclusions(declarations)})
}

// payAttackTax pays the attack tax for the given attack declarations.
func (ce combatEngine) payAttackTax(g *game.Game, playerID game.PlayerID, declarations []game.AttackDeclaration, manaCost *cost.Mana) bool {
	return paymentOrch.payGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: manaCost, Exclude: ce.attackingPermanentExclusions(declarations)})
}

// attackTaxCost computes the total attack-tax cost for the given declarations.
// It returns (nil, false) when no tax applies.
func (combatEngine) attackTaxCost(g *game.Game, declarations []game.AttackDeclaration) (*cost.Mana, bool) {
	total := 0
	for _, declaration := range declarations {
		for _, tax := range g.AttackTaxes {
			if tax.DefendingPlayer == declaration.Target.Player && tax.Amount > 0 {
				total += tax.Amount
			}
		}
	}
	if total <= 0 {
		return nil, false
	}
	manaCost := cost.Mana{cost.O(total)}
	return &manaCost, true
}

// attackingPermanentExclusions returns a set of permanent object IDs that are
// excluded from mana-payment plans because they are declared attackers.
func (combatEngine) attackingPermanentExclusions(declarations []game.AttackDeclaration) map[id.ID]bool {
	excluded := make(map[id.ID]bool, len(declarations))
	for _, declaration := range declarations {
		excluded[declaration.Attacker] = true
	}
	return excluded
}
