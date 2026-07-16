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
// runPhase runs the combat phase and its five steps in order (CR 506.1):
// beginning of combat (CR 507), declare attackers (CR 508), declare blockers
// (CR 509), combat damage (CR 510), and end of combat (CR 511). A first or double
// strike creature adds a second combat damage step (CR 510.4): the first pass
// deals first/double strike damage, the second deals the rest.
//
// CR 508.8: if no creature is attacking after the declare attackers step (none
// were declared and none were put onto the battlefield attacking), the declare
// blockers and combat damage steps are skipped and the phase proceeds directly to
// end of combat.
func (ce combatEngine) runPhase(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	g.Turn.Phase = game.PhaseCombat
	log.addPhase(game.PhaseCombat)
	g.Turn.CombatPhasesThisTurn++
	g.Combat = &game.CombatState{}
	defer func() {
		expireEndOfCombatRuleEffects(g)
		g.Combat = nil
	}()

	if !ce.runPriorityStep(g, agents, log, game.StepBeginningOfCombat) {
		return
	}

	g.Turn.Step = game.StepDeclareAttackers
	log.addStep(game.StepDeclareAttackers)
	ce.declareAttackers(g, agents, log)
	ce.resolveCombatAfterAttackers(g, agents, log)
}

// resolveCombatAfterAttackers runs the combat phase from the post-declare-
// attackers priority window through end of combat: the attack-declaration
// priority window, then (when a creature attacked, CR 508.8) the declare-blockers
// step, the first-strike and normal combat-damage steps with their priority
// windows, and finally the end-of-combat step. It assumes attackers have already
// been declared for the current combat and g is in the declare-attackers step.
//
// It is the resumable tail of runPhase, split out so a search agent's Simulator
// can play a candidate attack out to its combat result (see
// Simulator.ResolveCombatWithAttackers) using the same authoritative sequence
// real games use. runPhase is its only in-engine caller, so extracting it does
// not change real-game combat.
func (ce combatEngine) resolveCombatAfterAttackers(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	if !ce.runPriority(g, agents, log) {
		return
	}
	emptyManaPools(g)

	// CR 508.8: skip the declare blockers and combat damage steps if no creature
	// is attacking (none declared and none put onto the battlefield attacking).
	// This is a historical check (g.Combat.AttackersDeclared), not the current
	// attacker list, which can empty out as attackers leave combat.
	if g.Combat.AttackersDeclared {
		g.Turn.Step = game.StepDeclareBlockers
		log.addStep(game.StepDeclareBlockers)
		ce.declareBlockers(g, agents, log)
		if !ce.runPriority(g, agents, log) {
			return
		}
		emptyManaPools(g)

		if combatHasFirstStrikeDamage(g) {
			g.Turn.Step = game.StepFirstStrikeDamage
			log.addStep(game.StepFirstStrikeDamage)
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
		log.addStep(game.StepCombatDamage)
		ce.resolveDamagePass(g, normalCombatDamage, log)
		ce.e.applyStateBasedActionsWithLog(g, log)
		if g.IsGameOver() {
			return
		}
		if !ce.runPriority(g, agents, log) {
			return
		}
		emptyManaPools(g)
	}

	ce.runPriorityStep(g, agents, log, game.StepEndOfCombat)
}

// runPriorityStep sets the current step, emits the beginning-of-step event,
// runs the priority loop, and empties mana pools. It returns false if the game
// ended during the priority window.
func (ce combatEngine) runPriorityStep(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog, step game.Step) bool {
	g.Turn.Step = step
	log.addStep(step)
	emitBeginningOfStepEvent(g, step)
	if !ce.runPriority(g, agents, log) {
		return false
	}
	emptyManaPools(g)
	return true
}

// runPriority gives priority to the active player and runs the priority loop.
// It returns false if the game ended during the window. After every combat
// turn-based action (declaring attackers/blockers, dealing combat damage) it is
// the active player who receives priority (CR 508.2, CR 509.2, CR 510.3,
// CR 511.1).
func (ce combatEngine) runPriority(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	g.Turn.PriorityPlayer = g.Turn.ActivePlayer
	ce.e.runPriorityLoop(g, agents, log)
	return !g.IsGameOver()
}

// declareAttackers runs the declare-attackers turn-based action (CR 508.1: the
// active player declares attackers; this does not use the stack): it enumerates
// legal attacker choices, asks the active player to pick one, logs it, and
// applies it. Legality enforces CR 508.1a (chosen creatures are untapped and have
// haste or were controlled since the turn began) and the attack restrictions and
// requirements (CR 508.1b-d).
func (ce combatEngine) declareAttackers(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	playerID := g.Turn.ActivePlayer
	legal := ce.legalAttackers(g, playerID)
	if len(legal) == 0 {
		return
	}

	chosen := legal[len(legal)-1]
	if agent := agentFor(agents, playerID); agent != nil {
		chosen = ce.e.decideAction(g, agent, playerID, legal)
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

// declareBlockers runs the declare-blockers turn-based action for each defending
// player in turn order (CR 509.1: each defending player declares blockers; this
// does not use the stack). Legality enforces CR 509.1a (chosen blockers are
// untapped and assigned to an attacker that is attacking that player or a
// planeswalker/battle they control or protect) and the block restrictions and
// requirements (CR 509.1b-c). After all blockers are declared, attackers with no
// blockers become unblocked (CR 509.1h).
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
	emitUnblockedAttackerEvents(g)
}

// attackerDefendingPlayer returns the defending player an attacker was declared
// against (AttackTarget.Player holds it even when a planeswalker or battle is the
// direct target). A blocked attacker is always a current declared attacker, so
// the lookup succeeds; it falls back to the active player only if no declaration
// matches.
func attackerDefendingPlayer(g *game.Game, attackerObjectID id.ID) game.PlayerID {
	if g.Combat != nil {
		for _, declaration := range g.Combat.Attackers {
			if declaration.Attacker == attackerObjectID {
				return declaration.Target.Player
			}
		}
	}
	return g.Turn.ActivePlayer
}

// emitUnblockedAttackerEvents fires EventAttackerBecameUnblocked once for each
// attacker that no creature blocked, after every defending player has finished
// declaring blockers (CR 509.1h). The events share a simultaneous batch so
// "whenever this creature attacks and isn't blocked" triggers all see a single
// declare-blockers boundary.
func emitUnblockedAttackerEvents(g *game.Game) {
	if g.Combat == nil || len(g.Combat.Attackers) == 0 {
		return
	}
	batchID := g.IDGen.Next()
	for _, declaration := range g.Combat.Attackers {
		if g.Combat.BlockedAttackers[declaration.Attacker] {
			continue
		}
		attacker, ok := permanentByObjectID(g, declaration.Attacker)
		if !ok {
			continue
		}
		emitEvent(g, game.Event{
			Kind:           game.EventAttackerBecameUnblocked,
			SourceID:       attacker.CardInstanceID,
			SourceObjectID: attacker.ObjectID,
			Controller:     effectiveController(g, attacker),
			Player:         declaration.Target.Player,
			PermanentID:    attacker.ObjectID,
			AttackTarget:   declaration.Target,
			SimultaneousID: batchID,
		})
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
				if len(attackers) > 1 && ce.declareAttackersSatisfiesRequirements(g, playerID, single, eligibleByID) &&
					attackDeclarationsSatisfyAloneRestriction(g, single, eligibleByID) {
					act := actionBuild.declareAttackers(single)
					if !containsAction(actions, act) && ce.canPayAttackTax(g, playerID, single) {
						actions = append(actions, act)
					}
				}
				declarations = append(declarations, single[0])
			}
			if ce.declareAttackersSatisfiesRequirements(g, playerID, declarations, eligibleByID) &&
				attackDeclarationsSatisfyAloneRestriction(g, declarations, eligibleByID) {
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
		if declarations := preferredRequiredAttackDeclarations(g, playerID, attackers); len(declarations) > 0 &&
			attackDeclarationsSatisfyAloneRestriction(g, declarations, eligibleByID) {
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
	required := mustBlockRequirements(g, attackers)
	maxRequired, preferredRequired := maximumSatisfiedMustBlockRequirements(g, required, blockers)
	lures := trueLureAttackers(g, attackers, blockers)
	lureForcesBlock := lureForcesAnyBlocker(g, lures, blockers)
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
				if blockDeclarationsSatisfyMustBlockRequirements(required, maxRequired, declaration) &&
					blockDeclarationsSatisfyLures(g, lures, blockers, declaration) &&
					blockDeclarationsSatisfyAloneRestriction(g, appendCombatBlockers(g, declaration)) {
					actions = append(actions, actionBuild.declareBlockers(declaration))
				}
			}
		}
		if len(allBlockers) > 1 && !ruleEffectLimitsBlockersToOne(g, attackingPermanent) {
			if blockDeclarationsSatisfyMustBlockRequirements(required, maxRequired, allBlockers) &&
				blockDeclarationsSatisfyLures(g, lures, blockers, allBlockers) &&
				blockDeclarationsSatisfyAloneRestriction(g, appendCombatBlockers(g, allBlockers)) {
				actions = append(actions, actionBuild.declareBlockers(allBlockers))
			}
		}
	}
	actions = append(actions, multiAttackerBlockActions(g, attackers, blockers, required, maxRequired, lures)...)
	requiredDeclarations := [][]game.BlockDeclaration{preferredRequired}
	if unitMustBlockRequirements(required) {
		requiredDeclarations = maximumUnitMustBlockDeclarations(g, required, blockers, maxRequired)
	}
	for _, declaration := range requiredDeclarations {
		if len(declaration) > 1 &&
			blockDeclarationsSatisfyLures(g, lures, blockers, declaration) &&
			blockDeclarationsSatisfyAloneRestriction(g, appendCombatBlockers(g, declaration)) {
			actions = append(actions, actionBuild.declareBlockers(declaration))
		}
	}
	if maxRequired == 0 && !lureForcesBlock {
		actions = append(actions, actionBuild.declareBlockers(nil))
	}
	return actions
}

// multiAttackerBlockActions enumerates declare-blockers actions in which a single
// creature that can block more than one attacker ("can block an additional
// creature", blockerBlockLimit > 1) blocks a subset of the attackers it can
// legally block. It offers one such declaration per capable blocker per legal
// attacker subset (size two up to the blocker's limit), so the added capability
// is reachable through the action interface. Menace attackers (which need at
// least two blockers) are excluded because a lone blocker can't satisfy them, and
// every declaration is validated against the same must-block, lure, and alone
// restrictions as the single-attacker declarations.
func multiAttackerBlockActions(g *game.Game, attackers []game.AttackDeclaration, blockers []*game.Permanent, required map[id.ID]int, maxRequired int, lures map[id.ID]bool) []action.Action {
	var actions []action.Action
	for _, blocker := range blockers {
		limit := blockerBlockLimit(g, blocker)
		if limit < 2 {
			continue
		}
		var blockable []game.AttackDeclaration
		for _, attacker := range attackers {
			attackingPermanent, ok := permanentByObjectID(g, attacker.Attacker)
			if !ok || attackerRequiresMultipleBlockers(g, attackingPermanent) {
				continue
			}
			if canBlockAttacker(g, blocker, attackingPermanent) {
				blockable = append(blockable, attacker)
			}
		}
		if len(blockable) < 2 {
			continue
		}
		for _, subset := range attackerSubsets(blockable, 2, min(limit, len(blockable))) {
			declaration := make([]game.BlockDeclaration, 0, len(subset))
			for _, attacker := range subset {
				declaration = append(declaration, game.BlockDeclaration{
					Blocker:  blocker.ObjectID,
					Blocking: attacker.Attacker,
				})
			}
			if blockDeclarationsSatisfyMustBlockRequirements(required, maxRequired, declaration) &&
				blockDeclarationsSatisfyLures(g, lures, blockers, declaration) &&
				blockDeclarationsSatisfyAloneRestriction(g, appendCombatBlockers(g, declaration)) {
				actions = append(actions, actionBuild.declareBlockers(declaration))
			}
		}
	}
	return actions
}

// attackerSubsets returns every subset of items whose size is between minSize and
// maxSize inclusive, preserving the input (attacker declaration) order within each
// subset so the generated block declarations are deterministic.
func attackerSubsets(items []game.AttackDeclaration, minSize, maxSize int) [][]game.AttackDeclaration {
	var subsets [][]game.AttackDeclaration
	var build func(start int, current []game.AttackDeclaration)
	build = func(start int, current []game.AttackDeclaration) {
		if len(current) >= minSize {
			subset := make([]game.AttackDeclaration, len(current))
			copy(subset, current)
			subsets = append(subsets, subset)
		}
		if len(current) == maxSize {
			return
		}
		for i := start; i < len(items); i++ {
			next := make([]game.AttackDeclaration, len(current)+1)
			copy(next, current)
			next[len(current)] = items[i]
			build(i+1, next)
		}
	}
	build(0, nil)
	return subsets
}

// trueLureAttackers returns the set of attacker object IDs whose true-lure
// requirement is active and satisfiable (every creature able to block them must
// do so, CR 509.1c). A lure whose block-count constraints make the
// every-able-blocker requirement impossible — a menaced lure with fewer than two
// legal blockers, or a lure that can be blocked by at most one creature while
// more than one is able — is excluded so it fails open as an ordinary attacker
// rather than forcing blocks or invalidating other attackers' legal blocks.
func trueLureAttackers(g *game.Game, attackers []game.AttackDeclaration, blockers []*game.Permanent) map[id.ID]bool {
	var lures map[id.ID]bool
	for _, attack := range attackers {
		attacker, ok := permanentByObjectID(g, attack.Attacker)
		if !ok || !ruleEffectRequiresBeingBlockedByAllAble(g, attacker) {
			continue
		}
		if !lureRequirementSatisfiable(g, attacker, blockers) {
			continue
		}
		if lures == nil {
			lures = make(map[id.ID]bool)
		}
		lures[attack.Attacker] = true
	}
	return lures
}

// lureRequirementSatisfiable reports whether every creature able to block the
// lure attacker can legally do so simultaneously, given the attacker's
// block-count constraints. It mirrors satisfiableMustBlockAttackers' menace
// guard and additionally rejects the limit-to-one conflict.
func lureRequirementSatisfiable(g *game.Game, attacker *game.Permanent, blockers []*game.Permanent) bool {
	legalBlockerCount := 0
	for _, blocker := range blockers {
		if canBlockAttacker(g, blocker, attacker) {
			legalBlockerCount++
		}
	}
	if legalBlockerCount == 0 {
		return false
	}
	if attackerRequiresMultipleBlockers(g, attacker) && legalBlockerCount < 2 {
		return false
	}
	if ruleEffectLimitsBlockersToOne(g, attacker) && legalBlockerCount > 1 {
		return false
	}
	return true
}

// lureForcesAnyBlocker reports whether at least one eligible blocker is able to
// block a true-lure attacker, so the defending player can't decline to block.
func lureForcesAnyBlocker(g *game.Game, lures map[id.ID]bool, blockers []*game.Permanent) bool {
	if len(lures) == 0 {
		return false
	}
	for _, blocker := range blockers {
		if blockerCanBlockAnyLure(g, lures, blocker) {
			return true
		}
	}
	return false
}

// blockerCanBlockAnyLure reports whether blocker is able to block any of the
// true-lure attackers.
func blockerCanBlockAnyLure(g *game.Game, lures map[id.ID]bool, blocker *game.Permanent) bool {
	for lureID := range lures {
		lure, ok := permanentByObjectID(g, lureID)
		if ok && canBlockAttacker(g, blocker, lure) {
			return true
		}
	}
	return false
}

// blockDeclarationsSatisfyLures reports whether the block declarations honor every
// true-lure requirement: each eligible blocker able to block a lure attacker is
// declared blocking one of the lure attackers it can block (CR 509.1c).
func blockDeclarationsSatisfyLures(g *game.Game, lures map[id.ID]bool, blockers []*game.Permanent, declarations []game.BlockDeclaration) bool {
	if len(lures) == 0 {
		return true
	}
	declaredBlocking := make(map[id.ID]id.ID, len(declarations))
	for _, declaration := range declarations {
		declaredBlocking[declaration.Blocker] = declaration.Blocking
	}
	for _, blocker := range blockers {
		if !blockerCanBlockAnyLure(g, lures, blocker) {
			continue
		}
		blocking, ok := declaredBlocking[blocker.ObjectID]
		if !ok || !lures[blocking] {
			return false
		}
	}
	return true
}

// mustBlockRequirements returns each attacker's minimum legal blocker count for
// satisfying its "must be blocked if able" requirement. An internally
// contradictory requirement (menace plus at-most-one) is omitted because no
// legal declaration can satisfy it.
func mustBlockRequirements(g *game.Game, attackers []game.AttackDeclaration) map[id.ID]int {
	required := make(map[id.ID]int)
	for _, attack := range attackers {
		attacker, ok := permanentByObjectID(g, attack.Attacker)
		if !ok || !ruleEffectRequiresBeingBlocked(g, attacker) {
			continue
		}
		minimum := 1
		if attackerRequiresMultipleBlockers(g, attacker) {
			minimum = 2
		}
		if minimum > 1 && ruleEffectLimitsBlockersToOne(g, attacker) {
			continue
		}
		required[attack.Attacker] = minimum
	}
	return required
}

// blockDeclarationsSatisfyMustBlockRequirements enforces CR 509.1c by requiring
// a declaration to satisfy the maximum number of simultaneously satisfiable
// must-block requirements.
func blockDeclarationsSatisfyMustBlockRequirements(required map[id.ID]int, maximum int, declarations []game.BlockDeclaration) bool {
	if len(required) == 0 {
		return true
	}
	counts := make(map[id.ID]int)
	for _, declaration := range declarations {
		counts[declaration.Blocking]++
	}
	satisfied := 0
	for attackerID, minimum := range required {
		if counts[attackerID] >= minimum {
			satisfied++
		}
	}
	return satisfied == maximum
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
	if !attackDeclarationsSatisfyAloneRestriction(g, declare.Attackers, eligibleByID) {
		return false
	}
	if tax, ok := ce.attackTaxCost(g, declare.Attackers); ok {
		if !ce.payAttackTax(g, playerID, declare.Attackers, tax) {
			return false
		}
	}

	g.Combat.Attackers = append([]game.AttackDeclaration(nil), declare.Attackers...)
	if len(declare.Attackers) > 0 {
		// CR 508.8: record that creatures attacked so the declare blockers and
		// combat damage steps run even if those attackers later leave combat.
		g.Combat.AttackersDeclared = true
	}
	if g.Combat.PlayersAttacked == nil {
		g.Combat.PlayersAttacked = make(map[game.PlayerID]bool)
	}
	for _, declaration := range declare.Attackers {
		g.Combat.PlayersAttacked[declaration.Target.Player] = true
	}
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
			Kind:               game.EventAttackerDeclared,
			SourceID:           attacker.CardInstanceID,
			SourceObjectID:     attacker.ObjectID,
			Controller:         effectiveController(g, attacker),
			Player:             declaration.Target.Player,
			PermanentID:        attacker.ObjectID,
			SubjectGoaded:      isGoadedNow(g, attacker),
			SubjectGoadedKnown: true,
			AttackTarget:       declaration.Target,
			SimultaneousID:     simultaneousID,
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
	alreadyBlocking := make(map[id.ID]int)
	blockerCounts := make(map[id.ID]int)
	for _, block := range g.Combat.Blockers {
		alreadyBlocking[block.Blocker]++
		blockerCounts[block.Blocking]++
	}

	seenBlockers := make(map[id.ID]int)
	seenPairs := make(map[[2]id.ID]bool)
	for _, block := range g.Combat.Blockers {
		seenPairs[[2]id.ID{block.Blocker, block.Blocking}] = true
	}
	for _, block := range declare.Blockers {
		pair := [2]id.ID{block.Blocker, block.Blocking}
		if seenPairs[pair] {
			return false
		}
		seenPairs[pair] = true
		seenBlockers[block.Blocker]++
		if eligibleByID[block.Blocker] == nil {
			return false
		}
		if seenBlockers[block.Blocker]+alreadyBlocking[block.Blocker] > blockerBlockLimit(g, eligibleByID[block.Blocker]) {
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
		if !ok || count == 0 {
			continue
		}
		if count < 2 && attackerRequiresMultipleBlockers(g, attacker) {
			return false
		}
		if count > 1 && ruleEffectLimitsBlockersToOne(g, attacker) {
			return false
		}
	}
	allBlockers := append([]game.BlockDeclaration(nil), g.Combat.Blockers...)
	allBlockers = append(allBlockers, declare.Blockers...)
	attackers := attacksAgainstPlayer(g, playerID)
	eligible := eligibleBlockers(g, playerID)
	if !blockDeclarationsSatisfyAloneRestriction(g, allBlockers) {
		return false
	}
	required := mustBlockRequirements(g, attackers)
	maxRequired, _ := maximumSatisfiedMustBlockRequirements(g, required, eligible)
	if !blockDeclarationsSatisfyMustBlockRequirements(required, maxRequired, allBlockers) {
		return false
	}
	if !blockDeclarationsSatisfyLures(g, trueLureAttackers(g, attackers, eligible), eligible, allBlockers) {
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
				attackTarget, _ := attackTargetForAttacker(g, attacker.ObjectID)
				emitEvent(g, game.Event{
					Kind:               game.EventAttackerBecameBlocked,
					SourceID:           attacker.CardInstanceID,
					SourceObjectID:     attacker.ObjectID,
					Controller:         effectiveController(g, attacker),
					Player:             attackerDefendingPlayer(g, attacker.ObjectID),
					PermanentID:        attacker.ObjectID,
					RelatedPermanentID: block.Blocker,
					AttackTarget:       attackTarget,
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
			SubjectGoaded:      isGoadedNow(g, blocker),
			SubjectGoadedKnown: true,
			RelatedPermanentID: block.Blocking,
			BlockedAttackerID:  block.Blocking,
			SimultaneousID:     g.Combat.BlockDeclarationBatchID,
		})
	}
	return true
}

// resolveDamagePass assigns and marks combat damage for all attackers in the
// given damage pass (first-strike or normal), implementing the combat damage step
// (CR 510). Combat damage is assigned per CR 510.1 (unblocked creatures to the
// player/planeswalker/battle they attack, blocked creatures to their blockers)
// and then all of it is dealt simultaneously (CR 510.2); the events are batched
// so they share one simultaneous timestamp and trigger together.
func (combatEngine) resolveDamagePass(g *game.Game, pass combatDamagePass, log *TurnLog) {
	if g.Combat == nil {
		return
	}
	eventStart := len(g.Events)
	// Combat damage in this pass is dealt simultaneously (CR 510.2). Reset the
	// per-step Phantom counter-removal latch so a gang-blocked Phantom loses at
	// most one +1/+1 counter this pass (each pass is a separate step, so a
	// double-strike attacker can still cost two counters across both passes).
	for _, permanent := range g.Battlefield {
		permanent.DamagePreventionCounterRemovedThisStep = false
	}
	blockerMap := blockersByAttacker(g)
	blockerDamage := blockerCombatDamageAssignments(g, blockerMap, pass)
	for _, declaration := range g.Combat.Attackers {
		attacker, ok := permanentByObjectID(g, declaration.Attacker)
		if !ok || attacker.PhasedOut {
			continue
		}
		blockers := blockerMap[declaration.Attacker]
		if attackerWasBlocked(g, declaration.Attacker) {
			resolveBlockedCombatDamage(g, attacker, blockers, declaration.Target, pass, blockerDamageForAttacker(blockerDamage, declaration.Attacker), log)
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
	effects := activeRuleEffects(g)
	for _, declaration := range declarations {
		for i := range effects {
			if ruleEffectAttackTaxApplies(&effects[i], declaration) {
				total += effects[i].AttackTaxGeneric
			}
			if ruleEffectPerCreatureAttackTaxApplies(&effects[i], declaration) {
				total += perCreatureAttackTaxAmount(g, &effects[i])
			}
		}
	}
	if total <= 0 {
		return nil, false
	}
	manaCost := cost.Mana{cost.O(total)}
	return &manaCost, true
}

func ruleEffectAttackTaxApplies(effect *game.RuleEffect, declaration game.AttackDeclaration) bool {
	return effect != nil &&
		effect.Kind == game.RuleEffectAttackTax &&
		effect.AttackTaxGeneric > 0 &&
		declaration.Target.IsPlayerAttack() &&
		playerRelationMatches(effect.Controller, declaration.Target.Player, effect.AffectedPlayer)
}

// ruleEffectPerCreatureAttackTaxApplies reports whether a per-creature attack
// tax (Baird, Archon of Absolution, Sphere of Safety, Collective Restraint)
// charges the given attacker. The tax protects the effect controller; when
// AttackTaxIncludesPlaneswalkers is set it also covers any planeswalker that
// controller controls, so it taxes both direct attacks on that player and
// attacks on their planeswalkers, but never attacks on a battle nor an attacker
// whose target already left combat. Without that flag it taxes only direct
// attacks on the controller.
func ruleEffectPerCreatureAttackTaxApplies(effect *game.RuleEffect, declaration game.AttackDeclaration) bool {
	if effect == nil || effect.Kind != game.RuleEffectAttackTaxPerCreature {
		return false
	}
	if !playerRelationMatches(effect.Controller, declaration.Target.Player, effect.AffectedPlayer) {
		return false
	}
	if effect.AttackTaxIncludesPlaneswalkers {
		return !declaration.Target.NoTarget && declaration.Target.BattleID == 0
	}
	return declaration.Target.IsPlayerAttack()
}

// perCreatureAttackTaxAmount evaluates the per-attacker generic mana a
// per-creature attack tax charges, from its single configured amount source: a
// CardSelection permanent count ("for each of those creatures, where X is the
// number of enchantments you control", Sphere of Safety), a board-derived
// aggregate ("where X is the number of basic land types among lands you
// control", Collective Restraint), or a fixed generic value ("pays {1} for each
// of those creatures", Baird, Archon of Absolution).
func perCreatureAttackTaxAmount(g *game.Game, effect *game.RuleEffect) int {
	if !effect.CardSelection.Empty() {
		return countPermanentsMatchingGroup(g, nil, effect.Controller, game.BattlefieldGroup(effect.CardSelection))
	}
	if effect.AttackTaxScaledAmount != game.AggregateNone {
		value, ok := aggregateValue(g, conditionContext{controller: effect.Controller}, effect.AttackTaxScaledAmount)
		if !ok {
			return 0
		}
		return value
	}
	return effect.AttackTaxGeneric
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
