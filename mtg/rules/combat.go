package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
)

func (e *Engine) runCombatPhase(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	g.Turn.Phase = game.PhaseCombat
	g.Combat = &game.CombatState{}
	defer func() {
		g.Combat = nil
	}()

	if !e.runCombatPriorityStep(g, agents, log, game.StepBeginningOfCombat) {
		return
	}

	g.Turn.Step = game.StepDeclareAttackers
	e.declareAttackers(g, agents, log)
	if !e.runCombatPriority(g, agents, log) {
		return
	}
	emptyManaPools(g)

	g.Turn.Step = game.StepDeclareBlockers
	e.declareBlockers(g, agents, log)
	if !e.runCombatPriority(g, agents, log) {
		return
	}
	emptyManaPools(g)

	if combatHasFirstStrikeDamage(g) {
		g.Turn.Step = game.StepFirstStrikeDamage
		e.resolveCombatDamagePass(g, firstStrikeCombatDamage, log)
		e.applyStateBasedActionsWithLog(g, log)
		if g.IsGameOver() {
			return
		}
		if !e.runCombatPriority(g, agents, log) {
			return
		}
		emptyManaPools(g)
	}

	g.Turn.Step = game.StepCombatDamage
	e.resolveCombatDamagePass(g, normalCombatDamage, log)
	e.applyStateBasedActionsWithLog(g, log)
	if g.IsGameOver() {
		return
	}
	if !e.runCombatPriority(g, agents, log) {
		return
	}
	emptyManaPools(g)

	e.runCombatPriorityStep(g, agents, log, game.StepEndOfCombat)
}

func (e *Engine) runCombatPriorityStep(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog, step game.Step) bool {
	g.Turn.Step = step
	if !e.runCombatPriority(g, agents, log) {
		return false
	}
	emptyManaPools(g)
	return true
}

func (e *Engine) runCombatPriority(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	g.Turn.PriorityPlayer = g.Turn.ActivePlayer
	e.runPriorityLoop(g, agents, log)
	return !g.IsGameOver()
}

func (e *Engine) declareAttackers(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	playerID := g.Turn.ActivePlayer
	legal := legalDeclareAttackersActions(g, playerID)
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

	if log != nil {
		log.Actions = append(log.Actions, ActionLog{
			Player: playerID,
			Action: chosen,
		})
	}

	if !e.applyDeclareAttackers(g, playerID, chosen.DeclareAttackers) {
		panic("applyDeclareAttackers failed for validated action")
	}
}

func (e *Engine) declareBlockers(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	for _, playerID := range defendingPlayersInOrder(g) {
		legal := legalDeclareBlockersActions(g, playerID)
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

		if log != nil {
			log.Actions = append(log.Actions, ActionLog{
				Player: playerID,
				Action: chosen,
			})
		}

		if !e.applyDeclareBlockers(g, playerID, chosen.DeclareBlockers) {
			panic("applyDeclareBlockers failed for validated action")
		}
	}
}

func (e *Engine) resolveCombatDamage(g *game.Game, log *TurnLog) {
	e.resolveCombatDamagePass(g, normalCombatDamage, log)
}

type combatDamagePass int

const (
	firstStrikeCombatDamage combatDamagePass = iota
	normalCombatDamage
)

func (e *Engine) resolveCombatDamagePass(g *game.Game, pass combatDamagePass, log *TurnLog) {
	if g == nil || g.Combat == nil {
		return
	}
	// Combat damage is simultaneous; state-based eliminations happen after all attackers deal damage.
	blockersByAttacker := blockersByAttacker(g)
	for _, declaration := range g.Combat.Attackers {
		attacker := permanentByObjectID(g, declaration.Attacker)
		if attacker == nil {
			continue
		}
		blockers := blockersByAttacker[declaration.Attacker]
		if attackerWasBlocked(g, declaration.Attacker) {
			resolveBlockedCombatDamage(g, attacker, blockers, declaration.Target.Player, pass, log)
			continue
		}
		if !declaration.Target.IsPlayerAttack() || !isPlayerAlive(g, declaration.Target.Player) {
			continue
		}
		resolveUnblockedCombatDamage(g, attacker, declaration.Target.Player, pass, log)
	}
}

func resolveUnblockedCombatDamage(g *game.Game, attacker *game.Permanent, defendingPlayer game.PlayerID, pass combatDamagePass, log *TurnLog) {
	if !dealsCombatDamageInPass(g, attacker, pass) {
		return
	}
	damage := effectivePower(g, attacker)
	if damage <= 0 {
		return
	}
	markPlayerCombatDamage(g, attacker, defendingPlayer, damage, log)
}

func resolveBlockedCombatDamage(g *game.Game, attacker *game.Permanent, blockers []*game.Permanent, defendingPlayer game.PlayerID, pass combatDamagePass, log *TurnLog) {
	if len(blockers) == 0 && (!dealsCombatDamageInPass(g, attacker, pass) || !hasKeyword(g, attacker, game.Trample)) {
		return
	}
	if dealsCombatDamageInPass(g, attacker, pass) {
		assignments, tramplingDamage := assignAttackerCombatDamage(g, attacker, blockers)
		for _, assignment := range assignments {
			markCreatureCombatDamage(g, attacker, assignment.permanent, assignment.damage, log)
		}
		markPlayerCombatDamage(g, attacker, defendingPlayer, tramplingDamage, log)
	}
	for _, blocker := range blockers {
		if dealsCombatDamageInPass(g, blocker, pass) {
			markCreatureCombatDamage(g, blocker, attacker, effectivePower(g, blocker), log)
		}
	}
}

func markCreatureCombatDamage(g *game.Game, source *game.Permanent, damaged *game.Permanent, damage int, log *TurnLog) {
	if source == nil || damaged == nil || damage <= 0 {
		return
	}
	damaged.MarkedDamage += damage
	if hasKeyword(g, source, game.Deathtouch) {
		damaged.MarkedDeathtouchDamage = true
	}
	applyLifelink(g, source, damage)
	if log != nil {
		log.CreatureDamage = append(log.CreatureDamage, CreatureDamageLog{
			SourcePermanent:   source.ObjectID,
			SourceID:          source.CardInstanceID,
			Controller:        source.Controller,
			DamagedPermanent:  damaged.ObjectID,
			DamagedSourceID:   damaged.CardInstanceID,
			DamagedController: damaged.Controller,
			Damage:            damage,
		})
	}
}

func markPlayerCombatDamage(g *game.Game, source *game.Permanent, defendingPlayer game.PlayerID, damage int, log *TurnLog) {
	if source == nil || damage <= 0 || !isPlayerAlive(g, defendingPlayer) {
		return
	}
	defender := g.Players[defendingPlayer]
	defender.Life -= damage
	if sourceIsControllerCommander(g, source) {
		defender.CommanderDamage[source.CardInstanceID] += damage
	}
	applyLifelink(g, source, damage)
	if log != nil {
		log.CombatDamage = append(log.CombatDamage, CombatDamageLog{
			Attacker:        source.ObjectID,
			SourceID:        source.CardInstanceID,
			Controller:      source.Controller,
			DefendingPlayer: defendingPlayer,
			Damage:          damage,
		})
	}
}

func applyLifelink(g *game.Game, source *game.Permanent, damage int) {
	if source == nil || damage <= 0 || !hasKeyword(g, source, game.Lifelink) {
		return
	}
	if source.Controller < 0 || int(source.Controller) >= len(g.Players) {
		return
	}
	controller := g.Players[source.Controller]
	if controller == nil {
		return
	}
	controller.Life += damage
}

func sourceIsControllerCommander(g *game.Game, source *game.Permanent) bool {
	if g == nil || source == nil || source.CardInstanceID == 0 {
		return false
	}
	if source.Controller < 0 || int(source.Controller) >= len(g.Players) {
		return false
	}
	controller := g.Players[source.Controller]
	return controller != nil && controller.CommanderInstanceID == source.CardInstanceID
}

func combatHasFirstStrikeDamage(g *game.Game) bool {
	if g == nil || g.Combat == nil {
		return false
	}
	for _, attack := range g.Combat.Attackers {
		if hasFirstOrDoubleStrike(g, permanentByObjectID(g, attack.Attacker)) {
			return true
		}
	}
	for _, block := range g.Combat.Blockers {
		if hasFirstOrDoubleStrike(g, permanentByObjectID(g, block.Blocker)) {
			return true
		}
	}
	return false
}

func dealsCombatDamageInPass(g *game.Game, permanent *game.Permanent, pass combatDamagePass) bool {
	card := permanentCardDef(g, permanent)
	if card == nil {
		return false
	}
	hasFirst := hasKeyword(g, permanent, game.FirstStrike)
	hasDouble := hasKeyword(g, permanent, game.DoubleStrike)
	switch pass {
	case firstStrikeCombatDamage:
		return hasFirst || hasDouble
	case normalCombatDamage:
		return !hasFirst || hasDouble
	default:
		return false
	}
}

func hasFirstOrDoubleStrike(g *game.Game, permanent *game.Permanent) bool {
	return hasKeyword(g, permanent, game.FirstStrike) || hasKeyword(g, permanent, game.DoubleStrike)
}

func hasKeyword(g *game.Game, permanent *game.Permanent, keyword game.Keyword) bool {
	card := permanentCardDef(g, permanent)
	if card != nil && card.HasKeyword(keyword) {
		return true
	}
	counterKind, ok := keywordCounterKind(keyword)
	return ok && permanent != nil && permanent.Counters.Get(counterKind) > 0
}

func keywordCounterKind(keyword game.Keyword) (counter.Kind, bool) {
	switch keyword {
	case game.Deathtouch:
		return counter.Deathtouch, true
	case game.FirstStrike:
		return counter.FirstStrike, true
	case game.Flying:
		return counter.Flying, true
	case game.Hexproof:
		return counter.Hexproof, true
	case game.Indestructible:
		return counter.Indestructible, true
	case game.Lifelink:
		return counter.Lifelink, true
	case game.Menace:
		return counter.Menace, true
	case game.Reach:
		return counter.Reach, true
	case game.Trample:
		return counter.Trample, true
	case game.Vigilance:
		return counter.Vigilance, true
	default:
		return 0, false
	}
}

func attackerWasBlocked(g *game.Game, attackerID id.ID) bool {
	if g == nil || g.Combat == nil {
		return false
	}
	for _, block := range g.Combat.Blockers {
		if block.Blocking == attackerID {
			return true
		}
	}
	return false
}

type creatureDamageAssignment struct {
	permanent *game.Permanent
	damage    int
}

func assignAttackerCombatDamage(g *game.Game, attacker *game.Permanent, blockers []*game.Permanent) ([]creatureDamageAssignment, int) {
	damageRemaining := effectivePower(g, attacker)
	if damageRemaining <= 0 {
		return nil, 0
	}
	assignments := make([]creatureDamageAssignment, 0, len(blockers))
	hasTrample := hasKeyword(g, attacker, game.Trample)
	for i, blocker := range blockers {
		if blocker == nil || damageRemaining <= 0 {
			continue
		}
		damage := damageRemaining
		if hasTrample || i < len(blockers)-1 {
			damage = min(damageRemaining, lethalDamageRemainingFromSource(g, attacker, blocker))
		}
		if damage <= 0 {
			continue
		}
		assignments = append(assignments, creatureDamageAssignment{
			permanent: blocker,
			damage:    damage,
		})
		damageRemaining -= damage
	}
	if hasTrample {
		return assignments, damageRemaining
	}
	return assignments, 0
}

func lethalDamageRemainingFromSource(g *game.Game, source *game.Permanent, permanent *game.Permanent) int {
	if hasKeyword(g, source, game.Deathtouch) {
		if permanent.MarkedDeathtouchDamage {
			return 0
		}
		return 1
	}
	return lethalDamageRemaining(g, permanent)
}

func lethalDamageRemaining(g *game.Game, permanent *game.Permanent) int {
	lethal, ok := lethalDamageNeeded(g, permanent)
	if !ok {
		return 0
	}
	return max(0, lethal-permanent.MarkedDamage)
}

func blockersByAttacker(g *game.Game) map[id.ID][]*game.Permanent {
	blockers := make(map[id.ID][]*game.Permanent)
	if g == nil || g.Combat == nil {
		return blockers
	}
	blockerByID := make(map[id.ID]*game.Permanent)
	for _, block := range g.Combat.Blockers {
		blocker := permanentByObjectID(g, block.Blocker)
		if blocker == nil {
			continue
		}
		blockerByID[block.Blocker] = blocker
		if len(g.Combat.BlockerOrder[block.Blocking]) == 0 {
			blockers[block.Blocking] = append(blockers[block.Blocking], blocker)
		}
	}
	for attackerID, order := range g.Combat.BlockerOrder {
		for _, blockerID := range order {
			blocker := blockerByID[blockerID]
			if blocker == nil {
				continue
			}
			blockers[attackerID] = append(blockers[attackerID], blocker)
		}
	}
	return blockers
}

func legalDeclareBlockersActions(g *game.Game, playerID game.PlayerID) []action.Action {
	if !canDeclareBlockers(g, playerID) {
		return nil
	}

	attackers := attacksAgainstPlayer(g, playerID)
	blockers := eligibleBlockers(g, playerID)
	actions := make([]action.Action, 0, len(attackers)*len(blockers)+1)
	for _, attacker := range attackers {
		var allBlockers []game.BlockDeclaration
		attackingPermanent := permanentByObjectID(g, attacker.Attacker)
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
				actions = append(actions, action.DeclareBlockers([]game.BlockDeclaration{
					block,
				}))
			}
		}
		if len(allBlockers) > 1 {
			actions = append(actions, action.DeclareBlockers(allBlockers))
		}
	}
	actions = append(actions, action.DeclareBlockers(nil))
	return actions
}

func eligibleBlockers(g *game.Game, playerID game.PlayerID) []*game.Permanent {
	if !isPlayerAlive(g, playerID) {
		return nil
	}
	var blockers []*game.Permanent
	for _, permanent := range g.Battlefield {
		if !canBlockWith(g, permanent, playerID) {
			continue
		}
		blockers = append(blockers, permanent)
	}
	return blockers
}

func canBlockWith(g *game.Game, permanent *game.Permanent, playerID game.PlayerID) bool {
	if permanent == nil || permanent.Controller != playerID || permanent.Tapped {
		return false
	}
	card := permanentCardDef(g, permanent)
	return card != nil && card.HasType(game.TypeCreature)
}

func canBlockAttacker(g *game.Game, blocker *game.Permanent, attacker *game.Permanent) bool {
	if blocker == nil || attacker == nil {
		return false
	}
	if hasKeyword(g, attacker, game.Flying) && !hasKeyword(g, blocker, game.Flying) && !hasKeyword(g, blocker, game.Reach) {
		return false
	}
	return true
}

func attackerRequiresMultipleBlockers(g *game.Game, attacker *game.Permanent) bool {
	return hasKeyword(g, attacker, game.Menace)
}

func (e *Engine) applyDeclareBlockers(g *game.Game, playerID game.PlayerID, declare action.DeclareBlockersAction) bool {
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
		if !canBlockAttacker(g, eligibleByID[block.Blocker], permanentByObjectID(g, block.Blocking)) {
			return false
		}
		blockerCounts[block.Blocking]++
	}
	for attackerID, count := range blockerCounts {
		if count > 0 && count < 2 && attackerRequiresMultipleBlockers(g, permanentByObjectID(g, attackerID)) {
			return false
		}
	}

	g.Combat.Blockers = append(g.Combat.Blockers, declare.Blockers...)
	if len(declare.Blockers) > 0 && g.Combat.BlockerOrder == nil {
		g.Combat.BlockerOrder = make(map[id.ID][]id.ID)
	}
	for _, block := range declare.Blockers {
		g.Combat.BlockerOrder[block.Blocking] = append(g.Combat.BlockerOrder[block.Blocking], block.Blocker)
	}
	return true
}

func canDeclareBlockers(g *game.Game, playerID game.PlayerID) bool {
	return g != nil &&
		g.Combat != nil &&
		g.Turn.Phase == game.PhaseCombat &&
		g.Turn.Step == game.StepDeclareBlockers &&
		isPlayerAlive(g, playerID)
}

func attacksAgainstPlayer(g *game.Game, playerID game.PlayerID) []game.AttackDeclaration {
	if g == nil || g.Combat == nil {
		return nil
	}
	var attacks []game.AttackDeclaration
	for _, attack := range g.Combat.Attackers {
		if attack.Target.IsPlayerAttack() && attack.Target.Player == playerID {
			attacks = append(attacks, attack)
		}
	}
	return attacks
}

func defendingPlayersInOrder(g *game.Game) []game.PlayerID {
	if g == nil || g.Combat == nil {
		return nil
	}
	var defenders []game.PlayerID
	seen := make(map[game.PlayerID]bool)
	current := g.Turn.ActivePlayer
	for range game.NumPlayers - 1 {
		current = g.TurnOrder.NextPriority(current)
		if current == g.Turn.ActivePlayer || seen[current] {
			break
		}
		seen[current] = true
		if len(attacksAgainstPlayer(g, current)) > 0 {
			defenders = append(defenders, current)
		}
	}
	return defenders
}

func eligibleAttackers(g *game.Game, playerID game.PlayerID) []*game.Permanent {
	if !isPlayerAlive(g, playerID) {
		return nil
	}

	var eligible []*game.Permanent
	for _, permanent := range g.Battlefield {
		if !canAttackWith(g, permanent, playerID) {
			continue
		}
		eligible = append(eligible, permanent)
	}
	return eligible
}

func canAttackWith(g *game.Game, permanent *game.Permanent, playerID game.PlayerID) bool {
	if permanent == nil || permanent.Controller != playerID || permanent.Tapped {
		return false
	}
	card := permanentCardDef(g, permanent)
	if card == nil || !card.HasType(game.TypeCreature) || hasKeyword(g, permanent, game.Defender) {
		return false
	}
	return !permanent.SummoningSick || hasKeyword(g, permanent, game.Haste)
}

func legalDeclareAttackersActions(g *game.Game, playerID game.PlayerID) []action.Action {
	if !canDeclareAttackers(g, playerID) {
		return nil
	}

	attackers := eligibleAttackers(g, playerID)
	opponents := aliveOpponents(g, playerID)
	actions := make([]action.Action, 0, len(opponents)+1)
	eligibleByID := permanentMapByObjectID(attackers)
	if len(attackers) > 0 {
		for _, opponent := range opponents {
			declarations := make([]game.AttackDeclaration, 0, len(attackers))
			for _, attacker := range attackers {
				declarations = append(declarations, game.AttackDeclaration{
					Attacker: attacker.ObjectID,
					Target: game.AttackTarget{
						Player: opponent,
					},
				})
			}
			if declareAttackersSatisfiesGoad(g, playerID, declarations, eligibleByID) {
				actions = append(actions, action.DeclareAttackers(declarations))
			}
		}
	}
	if !hasGoadedEligibleAttacker(attackers) {
		actions = append(actions, action.DeclareAttackers(nil))
	} else if len(actions) == 0 {
		if declarations := preferredGoadAttackDeclarations(g, playerID, attackers); len(declarations) > 0 {
			actions = append(actions, action.DeclareAttackers(declarations))
		}
	}
	return actions
}

func (e *Engine) applyDeclareAttackers(g *game.Game, playerID game.PlayerID, declare action.DeclareAttackersAction) bool {
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
	}
	if !declareAttackersSatisfiesGoad(g, playerID, declare.Attackers, eligibleByID) {
		return false
	}

	g.Combat.Attackers = append([]game.AttackDeclaration(nil), declare.Attackers...)
	for _, declaration := range declare.Attackers {
		attacker := eligibleByID[declaration.Attacker]
		if !hasKeyword(g, attacker, game.Vigilance) {
			attacker.Tapped = true
		}
	}
	return true
}

func canDeclareAttackers(g *game.Game, playerID game.PlayerID) bool {
	return g != nil &&
		g.Combat != nil &&
		g.Turn.Phase == game.PhaseCombat &&
		g.Turn.Step == game.StepDeclareAttackers &&
		playerID == g.Turn.ActivePlayer &&
		isPlayerAlive(g, playerID)
}

func isLegalAttackTarget(g *game.Game, attackerController game.PlayerID, target game.AttackTarget) bool {
	return target.IsPlayerAttack() &&
		target.Player != attackerController &&
		isPlayerAlive(g, target.Player)
}

func declareAttackersSatisfiesGoad(g *game.Game, playerID game.PlayerID, declarations []game.AttackDeclaration, eligibleByID map[id.ID]*game.Permanent) bool {
	declared := make(map[id.ID]game.AttackTarget, len(declarations))
	for _, declaration := range declarations {
		declared[declaration.Attacker] = declaration.Target
	}
	for _, attacker := range eligibleByID {
		target, isAttacking := declared[attacker.ObjectID]
		if !isGoaded(attacker) {
			continue
		}
		if !isAttacking {
			return false
		}
		if !goadAllowsAttackTarget(g, playerID, attacker, target) {
			return false
		}
	}
	return true
}

func hasGoadedEligibleAttacker(attackers []*game.Permanent) bool {
	for _, attacker := range attackers {
		if isGoaded(attacker) {
			return true
		}
	}
	return false
}

func preferredGoadAttackDeclarations(g *game.Game, playerID game.PlayerID, attackers []*game.Permanent) []game.AttackDeclaration {
	declarations := make([]game.AttackDeclaration, 0, len(attackers))
	for _, attacker := range attackers {
		target, ok := preferredGoadAttackTarget(g, playerID, attacker)
		if !ok {
			if isGoaded(attacker) {
				return nil
			}
			continue
		}
		declarations = append(declarations, game.AttackDeclaration{
			Attacker: attacker.ObjectID,
			Target: game.AttackTarget{
				Player: target,
			},
		})
	}
	return declarations
}

func preferredGoadAttackTarget(g *game.Game, playerID game.PlayerID, attacker *game.Permanent) (game.PlayerID, bool) {
	opponents := aliveOpponents(g, playerID)
	for _, opponent := range opponents {
		target := game.AttackTarget{Player: opponent}
		if goadAllowsAttackTarget(g, playerID, attacker, target) {
			return opponent, true
		}
	}
	if len(opponents) == 0 {
		return 0, false
	}
	return opponents[0], true
}

func goadAllowsAttackTarget(g *game.Game, playerID game.PlayerID, attacker *game.Permanent, target game.AttackTarget) bool {
	if !target.IsPlayerAttack() || !isLegalAttackTarget(g, playerID, target) {
		return false
	}
	if !isGoaded(attacker) || !hasNonGoadingOpponent(g, playerID, attacker) {
		return true
	}
	return !wasGoadedBy(attacker, target.Player)
}

func hasNonGoadingOpponent(g *game.Game, playerID game.PlayerID, attacker *game.Permanent) bool {
	for _, opponent := range aliveOpponents(g, playerID) {
		if !wasGoadedBy(attacker, opponent) {
			return true
		}
	}
	return false
}

func isGoaded(permanent *game.Permanent) bool {
	if permanent == nil {
		return false
	}
	for _, goaded := range permanent.Goaded {
		if goaded {
			return true
		}
	}
	return false
}

func wasGoadedBy(permanent *game.Permanent, player game.PlayerID) bool {
	return permanent != nil && permanent.Goaded[player]
}

func permanentMapByObjectID(permanents []*game.Permanent) map[id.ID]*game.Permanent {
	byID := make(map[id.ID]*game.Permanent, len(permanents))
	for _, permanent := range permanents {
		if permanent != nil {
			byID[permanent.ObjectID] = permanent
		}
	}
	return byID
}

func aliveOpponents(g *game.Game, playerID game.PlayerID) []game.PlayerID {
	var opponents []game.PlayerID
	for opponent := game.Player1; opponent < game.NumPlayers; opponent++ {
		if opponent != playerID && isPlayerAlive(g, opponent) {
			opponents = append(opponents, opponent)
		}
	}
	return opponents
}

func permanentCardDef(g *game.Game, permanent *game.Permanent) *game.CardDef {
	if g == nil || permanent == nil {
		return nil
	}
	if permanent.Token {
		return permanent.TokenDef
	}
	card := g.GetCardInstance(permanent.CardInstanceID)
	if card == nil {
		return nil
	}
	return card.Def
}

func permanentByObjectID(g *game.Game, objectID id.ID) *game.Permanent {
	if g == nil {
		return nil
	}
	for _, permanent := range g.Battlefield {
		if permanent != nil && permanent.ObjectID == objectID {
			return permanent
		}
	}
	return nil
}

func effectivePower(g *game.Game, permanent *game.Permanent) int {
	card := permanentCardDef(g, permanent)
	if card == nil || card.Power == nil || card.Power.IsStar {
		return 0
	}
	return max(0, card.Power.Value+powerToughnessCounterDelta(permanent))
}

func effectiveToughness(g *game.Game, permanent *game.Permanent) (int, bool) {
	card := permanentCardDef(g, permanent)
	if card == nil || card.Toughness == nil || card.Toughness.IsStar {
		return 0, false
	}
	return card.Toughness.Value + powerToughnessCounterDelta(permanent), true
}

func lethalDamageNeeded(g *game.Game, permanent *game.Permanent) (int, bool) {
	toughness, ok := effectiveToughness(g, permanent)
	if !ok || toughness <= 0 {
		return 0, ok
	}
	return toughness, true
}

func powerToughnessCounterDelta(permanent *game.Permanent) int {
	if permanent == nil {
		return 0
	}
	return permanent.Counters.Get(counter.PlusOnePlusOne) - permanent.Counters.Get(counter.MinusOneMinusOne)
}
