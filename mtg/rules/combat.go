package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
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

	g.Turn.Step = game.StepCombatDamage
	e.resolveCombatDamage(g, log)
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
		blocker, blocked := blockersByAttacker[declaration.Attacker]
		if blocked {
			resolveBlockedCombatDamage(g, attacker, blocker, log)
			continue
		}
		if !declaration.Target.IsPlayerAttack() || !isPlayerAlive(g, declaration.Target.Player) {
			continue
		}
		resolveUnblockedCombatDamage(g, attacker, declaration.Target.Player, log)
	}
}

func resolveUnblockedCombatDamage(g *game.Game, attacker *game.Permanent, defendingPlayer game.PlayerID, log *TurnLog) {
	damage := combatDamageAmount(permanentCardDef(g, attacker))
	if damage <= 0 {
		return
	}
	defender := g.Players[defendingPlayer]
	defender.Life -= damage
	if log != nil {
		log.CombatDamage = append(log.CombatDamage, CombatDamageLog{
			Attacker:        attacker.ObjectID,
			SourceID:        attacker.CardInstanceID,
			Controller:      attacker.Controller,
			DefendingPlayer: defendingPlayer,
			Damage:          damage,
		})
	}
}

func resolveBlockedCombatDamage(g *game.Game, attacker *game.Permanent, blocker *game.Permanent, log *TurnLog) {
	if blocker == nil {
		return
	}
	markCreatureCombatDamage(g, attacker, blocker, combatDamageAmount(permanentCardDef(g, attacker)), log)
	markCreatureCombatDamage(g, blocker, attacker, combatDamageAmount(permanentCardDef(g, blocker)), log)
}

func markCreatureCombatDamage(g *game.Game, source *game.Permanent, damaged *game.Permanent, damage int, log *TurnLog) {
	if source == nil || damaged == nil || damage <= 0 {
		return
	}
	damaged.MarkedDamage += damage
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

func blockersByAttacker(g *game.Game) map[id.ID]*game.Permanent {
	blockers := make(map[id.ID]*game.Permanent)
	if g == nil || g.Combat == nil {
		return blockers
	}
	for _, block := range g.Combat.Blockers {
		blockers[block.Blocking] = permanentByObjectID(g, block.Blocker)
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
		for _, blocker := range blockers {
			actions = append(actions, action.DeclareBlockers([]game.BlockDeclaration{
				{
					Blocker:  blocker.ObjectID,
					Blocking: attacker.Attacker,
				},
			}))
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
	alreadyBlocked := make(map[id.ID]bool)
	for _, block := range g.Combat.Blockers {
		alreadyBlocked[block.Blocking] = true
	}

	seenBlockers := make(map[id.ID]bool)
	seenAttackers := make(map[id.ID]bool)
	for _, block := range declare.Blockers {
		if seenBlockers[block.Blocker] || seenAttackers[block.Blocking] {
			return false
		}
		seenBlockers[block.Blocker] = true
		seenAttackers[block.Blocking] = true
		if eligibleByID[block.Blocker] == nil {
			return false
		}
		if !attackersByID[block.Blocking] || alreadyBlocked[block.Blocking] {
			return false
		}
	}

	g.Combat.Blockers = append(g.Combat.Blockers, declare.Blockers...)
	if len(declare.Blockers) > 0 && g.Combat.BlockerOrder == nil {
		g.Combat.BlockerOrder = make(map[id.ID][]id.ID)
	}
	for _, block := range declare.Blockers {
		g.Combat.BlockerOrder[block.Blocking] = []id.ID{block.Blocker}
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
	if card == nil || !card.HasType(game.TypeCreature) || card.HasKeyword(game.Defender) {
		return false
	}
	return !permanent.SummoningSick || card.HasKeyword(game.Haste)
}

func legalDeclareAttackersActions(g *game.Game, playerID game.PlayerID) []action.Action {
	if !canDeclareAttackers(g, playerID) {
		return nil
	}

	attackers := eligibleAttackers(g, playerID)
	opponents := aliveOpponents(g, playerID)
	actions := make([]action.Action, 0, len(opponents)+1)
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
			actions = append(actions, action.DeclareAttackers(declarations))
		}
	}
	actions = append(actions, action.DeclareAttackers(nil))
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

	g.Combat.Attackers = append([]game.AttackDeclaration(nil), declare.Attackers...)
	for _, declaration := range declare.Attackers {
		attacker := eligibleByID[declaration.Attacker]
		card := permanentCardDef(g, attacker)
		if card == nil || !card.HasKeyword(game.Vigilance) {
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

func combatDamageAmount(card *game.CardDef) int {
	if card == nil || card.Power == nil || card.Power.IsStar || card.Power.Value <= 0 {
		return 0
	}
	return card.Power.Value
}

func creatureToughness(card *game.CardDef) (int, bool) {
	if card == nil || card.Toughness == nil || card.Toughness.IsStar {
		return 0, false
	}
	return card.Toughness.Value, true
}
