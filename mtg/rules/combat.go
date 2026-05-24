package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
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
		if attacker == nil || attacker.PhasedOut {
			continue
		}
		blockers := blockersByAttacker[declaration.Attacker]
		if attackerWasBlocked(g, declaration.Attacker) {
			resolveBlockedCombatDamage(g, attacker, blockers, declaration.Target, pass, log)
			continue
		}
		if !isLegalAttackTarget(g, effectiveController(g, attacker), declaration.Target) {
			continue
		}
		resolveUnblockedCombatDamage(g, attacker, declaration.Target, pass, log)
	}
}

func resolveUnblockedCombatDamage(g *game.Game, attacker *game.Permanent, target game.AttackTarget, pass combatDamagePass, log *TurnLog) {
	if !dealsCombatDamageInPass(g, attacker, pass) {
		return
	}
	damage := effectivePower(g, attacker)
	if damage <= 0 {
		return
	}
	markAttackTargetCombatDamage(g, attacker, target, damage, log)
}

func resolveBlockedCombatDamage(g *game.Game, attacker *game.Permanent, blockers []*game.Permanent, target game.AttackTarget, pass combatDamagePass, log *TurnLog) {
	if len(blockers) == 0 && (!dealsCombatDamageInPass(g, attacker, pass) || !hasKeyword(g, attacker, game.Trample)) {
		return
	}
	if dealsCombatDamageInPass(g, attacker, pass) {
		assignments, tramplingDamage := assignAttackerCombatDamage(g, attacker, blockers)
		for _, assignment := range assignments {
			markCreatureCombatDamage(g, attacker, assignment.permanent, assignment.damage, log)
		}
		markAttackTargetCombatDamage(g, attacker, target, tramplingDamage, log)
	}
	for _, blocker := range blockers {
		if dealsCombatDamageInPass(g, blocker, pass) {
			markCreatureCombatDamage(g, blocker, attacker, effectivePower(g, blocker), log)
		}
	}
}

func markAttackTargetCombatDamage(g *game.Game, source *game.Permanent, target game.AttackTarget, damage int, log *TurnLog) {
	if target.IsPlayerAttack() {
		markPlayerCombatDamage(g, source, target.Player, damage, log)
		return
	}
	permanent := attackTargetPermanent(g, target)
	if permanent == nil || source == nil || damage <= 0 {
		return
	}
	sourceController := effectiveController(g, source)
	dealt := dealPermanentDamage(g, source.CardInstanceID, source.ObjectID, sourceController, permanent, damage, true)
	applyLifelink(g, source, dealt)
	if dealt <= 0 {
		return
	}
	if log != nil {
		log.CreatureDamage = append(log.CreatureDamage, CreatureDamageLog{
			SourcePermanent:   source.ObjectID,
			SourceID:          source.CardInstanceID,
			Controller:        sourceController,
			DamagedPermanent:  permanent.ObjectID,
			DamagedSourceID:   permanent.CardInstanceID,
			DamagedController: effectiveController(g, permanent),
			Damage:            dealt,
		})
	}
}

func markCreatureCombatDamage(g *game.Game, source *game.Permanent, damaged *game.Permanent, damage int, log *TurnLog) {
	if source == nil || damaged == nil || damage <= 0 {
		return
	}
	sourceController := effectiveController(g, source)
	dealt := dealPermanentDamage(g, source.CardInstanceID, source.ObjectID, sourceController, damaged, damage, true)
	if dealt > 0 && hasKeyword(g, source, game.Deathtouch) {
		damaged.MarkedDeathtouchDamage = true
	}
	applyLifelink(g, source, dealt)
	if dealt <= 0 {
		return
	}
	if log != nil {
		log.CreatureDamage = append(log.CreatureDamage, CreatureDamageLog{
			SourcePermanent:   source.ObjectID,
			SourceID:          source.CardInstanceID,
			Controller:        sourceController,
			DamagedPermanent:  damaged.ObjectID,
			DamagedSourceID:   damaged.CardInstanceID,
			DamagedController: effectiveController(g, damaged),
			Damage:            dealt,
		})
	}
}

func markPlayerCombatDamage(g *game.Game, source *game.Permanent, defendingPlayer game.PlayerID, damage int, log *TurnLog) {
	if source == nil || damage <= 0 || !isPlayerAlive(g, defendingPlayer) {
		return
	}
	defender := g.Players[defendingPlayer]
	sourceController := effectiveController(g, source)
	dealt := dealPlayerDamage(g, source.CardInstanceID, source.ObjectID, sourceController, defendingPlayer, damage, true)
	if sourceIsControllerCommander(g, source) {
		defender.CommanderDamage[source.CardInstanceID] += dealt
	}
	applyLifelink(g, source, dealt)
	if dealt <= 0 {
		return
	}
	if log != nil {
		log.CombatDamage = append(log.CombatDamage, CombatDamageLog{
			Attacker:        source.ObjectID,
			SourceID:        source.CardInstanceID,
			Controller:      sourceController,
			DefendingPlayer: defendingPlayer,
			Damage:          dealt,
		})
	}
}

func markPermanentDamage(g *game.Game, permanent *game.Permanent, damage int) {
	if permanent == nil || damage <= 0 {
		return
	}
	switch {
	case permanentHasType(g, permanent, game.TypePlaneswalker):
		permanent.Counters.Remove(counter.Loyalty, damage)
	case permanentHasType(g, permanent, game.TypeBattle):
		permanent.Counters.Remove(counter.Defense, damage)
	default:
		permanent.MarkedDamage += damage
	}
}

func dealPlayerDamage(g *game.Game, sourceID, sourceObjectID id.ID, controller, playerID game.PlayerID, damage int, combatDamage bool) int {
	if g == nil || damage <= 0 || !isPlayerAlive(g, playerID) {
		return 0
	}
	dealt := applyDamagePrevention(g, damageEvent{
		sourceID:       sourceID,
		sourceObjectID: sourceObjectID,
		controller:     controller,
		player:         playerID,
		amount:         damage,
		combatDamage:   combatDamage,
	})
	if dealt <= 0 {
		return 0
	}
	g.Players[playerID].Life -= dealt
	emitEvent(g, game.GameEvent{
		Kind:            game.EventDamageDealt,
		SourceID:        sourceID,
		SourceObjectID:  sourceObjectID,
		Controller:      controller,
		Player:          playerID,
		Amount:          dealt,
		DamageRecipient: game.DamageRecipientPlayer,
		CombatDamage:    combatDamage,
	})
	return dealt
}

func dealPermanentDamage(g *game.Game, sourceID, sourceObjectID id.ID, controller game.PlayerID, permanent *game.Permanent, damage int, combatDamage bool) int {
	if permanent == nil || damage <= 0 {
		return 0
	}
	dealt := applyDamagePrevention(g, damageEvent{
		sourceID:       sourceID,
		sourceObjectID: sourceObjectID,
		controller:     controller,
		permanent:      permanent,
		amount:         damage,
		combatDamage:   combatDamage,
	})
	if dealt <= 0 {
		return 0
	}
	markPermanentDamage(g, permanent, dealt)
	emitEvent(g, game.GameEvent{
		Kind:            game.EventDamageDealt,
		SourceID:        sourceID,
		SourceObjectID:  sourceObjectID,
		Controller:      controller,
		PermanentID:     permanent.ObjectID,
		CardID:          permanent.CardInstanceID,
		TokenName:       permanentTokenName(permanent),
		TokenDef:        permanent.TokenDef,
		Amount:          dealt,
		DamageRecipient: game.DamageRecipientPermanent,
		CombatDamage:    combatDamage,
	})
	return dealt
}

func applyLifelink(g *game.Game, source *game.Permanent, damage int) {
	if source == nil || damage <= 0 || !hasKeyword(g, source, game.Lifelink) {
		return
	}
	controllerID := effectiveController(g, source)
	if controllerID < 0 || int(controllerID) >= len(g.Players) {
		return
	}
	controller := g.Players[controllerID]
	if controller == nil {
		return
	}
	controller.Life += damage
}

func sourceIsControllerCommander(g *game.Game, source *game.Permanent) bool {
	if g == nil || source == nil || source.CardInstanceID == 0 {
		return false
	}
	controllerID := effectiveController(g, source)
	if controllerID < 0 || int(controllerID) >= len(g.Players) {
		return false
	}
	controller := g.Players[controllerID]
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
	if permanent == nil {
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

func removePermanentFromCombat(g *game.Game, permanentID id.ID) {
	if g == nil || g.Combat == nil || permanentID == 0 {
		return
	}
	var attackers []game.AttackDeclaration
	removedAttacker := false
	for _, attack := range g.Combat.Attackers {
		if attack.Attacker == permanentID {
			removedAttacker = true
			continue
		}
		attackers = append(attackers, attack)
	}
	g.Combat.Attackers = attackers

	var blockers []game.BlockDeclaration
	for _, block := range g.Combat.Blockers {
		if block.Blocker == permanentID || (removedAttacker && block.Blocking == permanentID) {
			continue
		}
		blockers = append(blockers, block)
	}
	g.Combat.Blockers = blockers
	for attackerID, order := range g.Combat.BlockerOrder {
		if attackerID == permanentID {
			delete(g.Combat.BlockerOrder, attackerID)
			continue
		}
		g.Combat.BlockerOrder[attackerID] = removePermanentID(order, permanentID)
	}
	delete(g.Combat.DamageAssignment, permanentID)
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
	if assignments, tramplingDamage, ok := attackerChosenDamageAssignments(g, attacker, blockers, damageRemaining); ok {
		return assignments, tramplingDamage
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

func attackerChosenDamageAssignments(g *game.Game, attacker *game.Permanent, blockers []*game.Permanent, damage int) ([]creatureDamageAssignment, int, bool) {
	if g == nil || g.Combat == nil || attacker == nil || len(g.Combat.DamageAssignment) == 0 {
		return nil, 0, false
	}
	var assignments []creatureDamageAssignment
	total := 0
	hasTrample := hasKeyword(g, attacker, game.Trample)
	for _, blocker := range blockers {
		if blocker == nil {
			continue
		}
		assigned := g.Combat.DamageAssignment[blocker.ObjectID]
		if assigned < 0 {
			return nil, 0, false
		}
		if hasTrample && assigned < lethalDamageRemainingFromSource(g, attacker, blocker) {
			return nil, 0, false
		}
		if assigned == 0 {
			continue
		}
		total += assigned
		assignments = append(assignments, creatureDamageAssignment{permanent: blocker, damage: assigned})
	}
	if !hasTrample && !damageAssignmentFollowsBlockerOrder(g, attacker, blockers) {
		return nil, 0, false
	}
	if total == 0 || total > damage {
		return nil, 0, false
	}
	if hasTrample {
		return assignments, damage - total, true
	}
	if total != damage {
		return nil, 0, false
	}
	return assignments, 0, true
}

func damageAssignmentFollowsBlockerOrder(g *game.Game, attacker *game.Permanent, blockers []*game.Permanent) bool {
	if g == nil || g.Combat == nil || attacker == nil {
		return false
	}
	mayAssignToLater := true
	for _, blocker := range blockers {
		if blocker == nil {
			continue
		}
		assigned := g.Combat.DamageAssignment[blocker.ObjectID]
		if assigned > 0 && !mayAssignToLater {
			return false
		}
		if assigned < lethalDamageRemainingFromSource(g, attacker, blocker) {
			mayAssignToLater = false
		}
	}
	return true
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
	if permanent == nil || effectiveController(g, permanent) != playerID || permanent.Tapped {
		return false
	}
	if permanent.PhasedOut {
		return false
	}
	return permanentHasType(g, permanent, game.TypeCreature)
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
		blocker := eligibleByID[block.Blocker]
		if blocker == nil {
			continue
		}
		emitEvent(g, game.GameEvent{
			Kind:              game.EventBlockerDeclared,
			SourceID:          blocker.CardInstanceID,
			SourceObjectID:    blocker.ObjectID,
			Controller:        effectiveController(g, blocker),
			PermanentID:       blocker.ObjectID,
			BlockedAttackerID: block.Blocking,
		})
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
		if attack.Target.Player == playerID {
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
	if permanent == nil || effectiveController(g, permanent) != playerID || permanent.Tapped || permanent.PhasedOut {
		return false
	}
	if !permanentHasType(g, permanent, game.TypeCreature) || hasKeyword(g, permanent, game.Defender) {
		return false
	}
	return !permanent.SummoningSick || hasKeyword(g, permanent, game.Haste)
}

func legalDeclareAttackersActions(g *game.Game, playerID game.PlayerID) []action.Action {
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
				if len(attackers) > 1 && declareAttackersSatisfiesGoad(g, playerID, single, eligibleByID) {
					action := action.DeclareAttackers(single)
					if !containsAction(actions, action) && canPayAttackTax(g, playerID, single) {
						actions = append(actions, action)
					}
				}
				declarations = append(declarations, single[0])
			}
			if declareAttackersSatisfiesGoad(g, playerID, declarations, eligibleByID) {
				action := action.DeclareAttackers(declarations)
				if !containsAction(actions, action) && canPayAttackTax(g, playerID, declarations) {
					actions = append(actions, action)
				}
			}
		}
	}
	if !hasGoadedEligibleAttacker(attackers) {
		actions = append(actions, action.DeclareAttackers(nil))
	} else if len(actions) == 0 {
		if declarations := preferredGoadAttackDeclarations(g, playerID, attackers); len(declarations) > 0 {
			if canPayAttackTax(g, playerID, declarations) {
				actions = append(actions, action.DeclareAttackers(declarations))
			}
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
	tax := attackTaxCost(g, declare.Attackers)
	if !payAttackTax(g, playerID, declare.Attackers, tax) {
		return false
	}

	g.Combat.Attackers = append([]game.AttackDeclaration(nil), declare.Attackers...)
	for _, declaration := range declare.Attackers {
		attacker := eligibleByID[declaration.Attacker]
		if !hasKeyword(g, attacker, game.Vigilance) {
			attacker.Tapped = true
		}

		emitEvent(g, game.GameEvent{
			Kind:           game.EventAttackerDeclared,
			SourceID:       attacker.CardInstanceID,
			SourceObjectID: attacker.ObjectID,
			Controller:     effectiveController(g, attacker),
			PermanentID:    attacker.ObjectID,
			AttackTarget:   declaration.Target,
		})
	}
	return true
}

func canPayAttackTax(g *game.Game, playerID game.PlayerID, declarations []game.AttackDeclaration) bool {
	cost := attackTaxCost(g, declarations)
	_, ok := buildPaymentPlan(g, playerID, cost, 0, attackingPermanentExclusions(declarations))
	return ok
}

func payAttackTax(g *game.Game, playerID game.PlayerID, declarations []game.AttackDeclaration, cost *mana.Cost) bool {
	plan, ok := buildPaymentPlan(g, playerID, cost, 0, attackingPermanentExclusions(declarations))
	if !ok {
		return false
	}
	return applyPaymentPlan(g, playerID, plan)
}

func attackingPermanentExclusions(declarations []game.AttackDeclaration) map[id.ID]bool {
	excluded := make(map[id.ID]bool, len(declarations))
	for _, declaration := range declarations {
		excluded[declaration.Attacker] = true
	}
	return excluded
}

func attackTaxCost(g *game.Game, declarations []game.AttackDeclaration) *mana.Cost {
	total := 0
	for _, declaration := range declarations {
		for _, tax := range g.AttackTaxes {
			if tax.DefendingPlayer == declaration.Target.Player && tax.Amount > 0 {
				total += tax.Amount
			}
		}
	}
	if total <= 0 {
		return nil
	}
	cost := mana.Cost{mana.GenericMana(total)}
	return &cost
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
	if target.Player == attackerController || !isPlayerAlive(g, target.Player) {
		return false
	}
	if target.IsPlayerAttack() {
		return true
	}
	permanent := attackTargetPermanent(g, target)
	if permanent == nil || effectiveController(g, permanent) != target.Player {
		return false
	}
	if target.PlaneswalkerID != 0 {
		return target.BattleID == 0 && permanentHasType(g, permanent, game.TypePlaneswalker)
	}
	if target.BattleID != 0 {
		return permanentHasType(g, permanent, game.TypeBattle)
	}
	return false
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
	if !isLegalAttackTarget(g, playerID, target) {
		return false
	}
	if !isGoaded(attacker) || !hasNonGoadingOpponent(g, playerID, attacker) {
		return true
	}
	return !wasGoadedBy(attacker, target.Player)
}

func legalAttackTargets(g *game.Game, attackerController game.PlayerID) []game.AttackTarget {
	var targets []game.AttackTarget
	for _, opponent := range aliveOpponents(g, attackerController) {
		targets = append(targets, game.AttackTarget{Player: opponent})
	}
	for _, permanent := range g.Battlefield {
		permanentController := effectiveController(g, permanent)
		if permanent == nil || permanent.PhasedOut || permanentController == attackerController || !isPlayerAlive(g, permanentController) {
			continue
		}
		switch {
		case permanentHasType(g, permanent, game.TypePlaneswalker):
			targets = append(targets, game.AttackTarget{Player: permanentController, PlaneswalkerID: permanent.ObjectID})
		case permanentHasType(g, permanent, game.TypeBattle):
			targets = append(targets, game.AttackTarget{Player: permanentController, BattleID: permanent.ObjectID})
		}
	}
	return targets
}

func attackTargetPermanent(g *game.Game, target game.AttackTarget) *game.Permanent {
	switch {
	case target.PlaneswalkerID != 0:
		return permanentByObjectID(g, target.PlaneswalkerID)
	case target.BattleID != 0:
		return permanentByObjectID(g, target.BattleID)
	default:
		return nil
	}
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

func lethalDamageNeeded(g *game.Game, permanent *game.Permanent) (int, bool) {
	toughness, ok := effectiveToughness(g, permanent)
	if !ok || toughness <= 0 {
		return 0, ok
	}
	return toughness, true
}
