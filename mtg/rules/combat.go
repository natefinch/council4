package rules

import (
	"fmt"
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

// runCombatPhase drives the combat phase. It delegates to combatEngine so all
// combat orchestration, declaration, and damage logic is concentrated there.
func (e *Engine) runCombatPhase(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	combatEngine{e}.runPhase(g, agents, log)
}

// resolveCombatDamage applies a normal (non-first-strike) combat damage pass.
func (e *Engine) resolveCombatDamage(g *game.Game, log *TurnLog) {
	combatEngine{e}.resolveDamagePass(g, normalCombatDamage, log)
}

func combatActionLog(g *game.Game, playerID game.PlayerID, act action.Action) ActionLog {
	logged := ActionLog{
		Player: playerID,
		Action: act,
	}
	switch act.Kind {
	case action.ActionDeclareAttackers:
		payload, ok := act.DeclareAttackersPayload()
		if !ok {
			return logged
		}
		for _, declaration := range payload.Attackers {
			logged.addPermanentSnapshot(g, declaration.Attacker)
		}
	case action.ActionDeclareBlockers:
		payload, ok := act.DeclareBlockersPayload()
		if !ok {
			return logged
		}
		for _, declaration := range payload.Blockers {
			logged.addPermanentSnapshot(g, declaration.Blocker)
			logged.addPermanentSnapshot(g, declaration.Blocking)
		}
	default:
	}
	return logged
}

func (log *ActionLog) addPermanentSnapshot(g *game.Game, objectID id.ID) {
	permanent, ok := permanentByObjectID(g, objectID)
	if !ok {
		return
	}
	if permanent.Token {
		if log.PermanentTokenNames == nil {
			log.PermanentTokenNames = make(map[id.ID]string)
		}
		log.PermanentTokenNames[objectID] = permanentTokenName(permanent)
		return
	}
	if log.PermanentSources == nil {
		log.PermanentSources = make(map[id.ID]id.ID)
	}
	log.PermanentSources[objectID] = permanent.CardInstanceID
}

type combatDamagePass int

const (
	firstStrikeCombatDamage combatDamagePass = iota
	normalCombatDamage
)

func (e *Engine) resolveCombatDamagePass(g *game.Game, pass combatDamagePass, log *TurnLog) {
	combatEngine{e}.resolveDamagePass(g, pass, log)
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
	blockerDamage := make([]int, len(blockers))
	for i, blocker := range blockers {
		if dealsCombatDamageInPass(g, blocker, pass) {
			blockerDamage[i] = effectivePower(g, blocker)
		}
	}
	if dealsCombatDamageInPass(g, attacker, pass) {
		assignments, tramplingDamage := assignAttackerCombatDamage(g, attacker, blockers)
		for _, assignment := range assignments {
			markCreatureCombatDamage(g, attacker, assignment.permanent, assignment.damage, log)
		}
		markAttackTargetCombatDamage(g, attacker, target, tramplingDamage, log)
	}
	for i, blocker := range blockers {
		if blockerDamage[i] > 0 {
			markCreatureCombatDamage(g, blocker, attacker, blockerDamage[i], log)
		}
	}
}

func markAttackTargetCombatDamage(g *game.Game, source *game.Permanent, target game.AttackTarget, damage int, log *TurnLog) {
	if target.IsPlayerAttack() {
		markPlayerCombatDamage(g, source, target.Player, damage, log)
		return
	}
	permanent, ok := attackTargetPermanent(g, target)
	if !ok || damage <= 0 {
		return
	}
	sourceController := effectiveController(g, source)
	dealt := dealPermanentDamage(g, source.CardInstanceID, source.ObjectID, sourceController, permanent, damage, true)
	applyLifelink(g, source, dealt)
	if dealt <= 0 {
		return
	}
	log.addCreatureDamage(CreatureDamageLog{
		SourcePermanent:   source.ObjectID,
		SourceID:          source.CardInstanceID,
		Controller:        sourceController,
		DamagedPermanent:  permanent.ObjectID,
		DamagedSourceID:   permanent.CardInstanceID,
		DamagedController: effectiveController(g, permanent),
		Damage:            dealt,
	})
}

func markCreatureCombatDamage(g *game.Game, source, damaged *game.Permanent, damage int, log *TurnLog) {
	if damage <= 0 {
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
	log.addCreatureDamage(CreatureDamageLog{
		SourcePermanent:   source.ObjectID,
		SourceID:          source.CardInstanceID,
		Controller:        sourceController,
		DamagedPermanent:  damaged.ObjectID,
		DamagedSourceID:   damaged.CardInstanceID,
		DamagedController: effectiveController(g, damaged),
		Damage:            dealt,
	})
}

func markPlayerCombatDamage(g *game.Game, source *game.Permanent, defendingPlayer game.PlayerID, damage int, log *TurnLog) {
	if damage <= 0 || !isPlayerAlive(g, defendingPlayer) {
		return
	}
	defender := g.Players[defendingPlayer]
	sourceController := effectiveController(g, source)
	dealt := dealPlayerDamage(g, source.CardInstanceID, source.ObjectID, sourceController, defendingPlayer, damage, true)
	for _, commanderID := range sourceCommanderIDs(g, source) {
		defender.CommanderDamage[commanderID] += dealt
	}
	applyToxic(g, source, defendingPlayer, dealt)
	applyLifelink(g, source, dealt)
	if dealt <= 0 {
		return
	}
	log.addCombatDamage(CombatDamageLog{
		Attacker:        source.ObjectID,
		SourceID:        source.CardInstanceID,
		Controller:      sourceController,
		DefendingPlayer: defendingPlayer,
		Damage:          dealt,
	})
	stealMonarchByCombatDamage(g, sourceController, defendingPlayer)
}

func applyToxic(g *game.Game, source *game.Permanent, defendingPlayer game.PlayerID, dealt int) {
	if dealt <= 0 {
		return
	}
	total := 0
	for _, body := range permanentEffectiveAbilities(g, source) {
		amount, ok := game.BodyToxicAmount(body)
		if !ok {
			continue
		}
		total += amount
	}
	if total > 0 {
		addCountersToPlayerControlledBy(g, effectiveController(g, source), g.Players[defendingPlayer], counter.Poison, total)
	}
}

func markPermanentDamage(g *game.Game, placementController game.PlayerID, permanent *game.Permanent, damage int, minusOneCounters bool) {
	if damage <= 0 {
		return
	}
	switch {
	case minusOneCounters && permanentHasType(g, permanent, types.Creature):
		addCountersToPermanentControlledBy(g, placementController, permanent, counter.MinusOneMinusOne, damage)
	case permanentHasType(g, permanent, types.Planeswalker):
		permanent.Counters.Remove(counter.Loyalty, damage)
	case permanentHasType(g, permanent, types.Battle):
		permanent.Counters.Remove(counter.Defense, damage)
	default:
		permanent.MarkedDamage += damage
	}
}

func dealPlayerDamage(g *game.Game, sourceID, sourceObjectID id.ID, controller, playerID game.PlayerID, damage int, combatDamage bool) int {
	if damage <= 0 || !isPlayerAlive(g, playerID) {
		return 0
	}
	event := damageEvent{
		sourceID:       sourceID,
		sourceObjectID: sourceObjectID,
		controller:     controller,
		player:         playerID,
		amount:         damage,
		combatDamage:   combatDamage,
	}
	dealt := applyDamagePrevention(g, event)
	if dealt <= 0 {
		return 0
	}
	event.amount = dealt
	dealt = replacementDamageAmount(g, event)
	if dealt <= 0 {
		return 0
	}
	loseLife(g, playerID, dealt)
	emitEvent(g, game.Event{
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
	if damage <= 0 || !activeBattlefieldPermanent(permanent) {
		return 0
	}
	event := damageEvent{
		sourceID:       sourceID,
		sourceObjectID: sourceObjectID,
		controller:     controller,
		permanent:      permanent,
		amount:         damage,
		combatDamage:   combatDamage,
	}
	dealt := applyDamagePrevention(g, event)
	if dealt <= 0 {
		return 0
	}
	event.amount = dealt
	dealt = replacementDamageAmount(g, event)
	if dealt <= 0 {
		return 0
	}
	markPermanentDamage(g, controller, permanent, dealt, damageSourceUsesMinusOneCounters(g, sourceObjectID))
	emitEvent(g, game.Event{
		Kind:            game.EventDamageDealt,
		SourceID:        sourceID,
		SourceObjectID:  sourceObjectID,
		Controller:      controller,
		Player:          effectiveController(g, permanent),
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

func damageSourceUsesMinusOneCounters(g *game.Game, sourceObjectID id.ID) bool {
	if source, ok := permanentByObjectID(g, sourceObjectID); ok {
		return hasKeyword(g, source, game.Wither) || hasKeyword(g, source, game.Infect)
	}
	snapshot, ok := lastKnownObject(g, sourceObjectID)
	if !ok {
		return false
	}
	for _, keyword := range snapshot.Keywords {
		if keyword == game.Wither || keyword == game.Infect {
			return true
		}
	}
	return false
}

func applyLifelink(g *game.Game, source *game.Permanent, damage int) {
	if damage <= 0 || !hasKeyword(g, source, game.Lifelink) {
		return
	}
	controllerID := effectiveController(g, source)
	if controllerID < 0 || int(controllerID) >= len(g.Players) {
		return
	}
	gainLife(g, controllerID, damage)
}

func sourceCommanderIDs(g *game.Game, source *game.Permanent) []id.ID {
	cardIDs := make([]id.ID, 0, len(source.MergedCards)+1)
	cardIDs = append(cardIDs, source.CardInstanceID)
	for _, component := range source.MergedCards {
		cardIDs = append(cardIDs, component.CardInstanceID)
	}
	var commanders []id.ID
	for _, cardID := range cardIDs {
		if cardID == 0 {
			continue
		}
		if g.CommanderIDs[cardID] {
			commanders = append(commanders, cardID)
			continue
		}
		for _, player := range g.Players {
			if player.CommanderInstanceID == cardID {
				commanders = append(commanders, cardID)
				break
			}
		}
	}
	return commanders
}

func combatHasFirstStrikeDamage(g *game.Game) bool {
	if g.Combat == nil {
		return false
	}
	for _, attack := range g.Combat.Attackers {
		if permanent, ok := permanentByObjectID(g, attack.Attacker); ok && hasFirstOrDoubleStrike(g, permanent) {
			return true
		}
	}
	for _, block := range g.Combat.Blockers {
		if permanent, ok := permanentByObjectID(g, block.Blocker); ok && hasFirstOrDoubleStrike(g, permanent) {
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
	if g.Combat == nil {
		return false
	}
	if g.Combat.BlockedAttackers[attackerID] {
		return true
	}
	for _, block := range g.Combat.Blockers {
		if block.Blocking == attackerID {
			return true
		}
	}
	return false
}

func removePermanentFromCombat(g *game.Game, permanentID id.ID) {
	if g.Combat == nil || permanentID == 0 {
		return
	}
	var attackers []game.AttackDeclaration
	removedAttackers := make(map[id.ID]bool)
	for _, attack := range g.Combat.Attackers {
		if attack.Attacker == permanentID {
			removedAttackers[attack.Attacker] = true
			continue
		}
		if attackTargetsPermanent(attack.Target, permanentID) {
			attack.Target.NoTarget = true
			attack.Target.PlaneswalkerID = 0
			attack.Target.BattleID = 0
		}
		attackers = append(attackers, attack)
	}
	g.Combat.Attackers = attackers

	var blockers []game.BlockDeclaration
	for _, block := range g.Combat.Blockers {
		if block.Blocker == permanentID || removedAttackers[block.Blocking] {
			continue
		}
		blockers = append(blockers, block)
	}
	g.Combat.Blockers = blockers
	for attackerID, order := range g.Combat.BlockerOrder {
		if removedAttackers[attackerID] {
			delete(g.Combat.BlockerOrder, attackerID)
			continue
		}
		g.Combat.BlockerOrder[attackerID] = removePermanentID(order, permanentID)
	}
	for attackerID := range removedAttackers {
		delete(g.Combat.BlockedAttackers, attackerID)
		delete(g.Combat.DamageAssignment, attackerID)
	}
	delete(g.Combat.DamageAssignment, permanentID)
}

func attackTargetsPermanent(target game.AttackTarget, permanentID id.ID) bool {
	return target.PlaneswalkerID == permanentID || target.BattleID == permanentID
}

type creatureDamageAssignment struct {
	permanent *game.Permanent
	damage    int
}

func assignAttackerCombatDamage(g *game.Game, attacker *game.Permanent, blockers []*game.Permanent) (assignments []creatureDamageAssignment, tramplingDamage int) {
	damageRemaining := effectivePower(g, attacker)
	if damageRemaining <= 0 {
		return nil, 0
	}
	if assignments, tramplingDamage, ok := attackerChosenDamageAssignments(g, attacker, blockers, damageRemaining); ok {
		return assignments, tramplingDamage
	}
	assignments = make([]creatureDamageAssignment, 0, len(blockers))
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
	if g.Combat == nil || attacker == nil || len(g.Combat.DamageAssignment) == 0 {
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
	if g.Combat == nil || attacker == nil {
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

func lethalDamageRemainingFromSource(g *game.Game, source, permanent *game.Permanent) int {
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
	if g.Combat == nil {
		return blockers
	}
	blockerByID := make(map[id.ID]*game.Permanent)
	for _, block := range g.Combat.Blockers {
		blocker, ok := permanentByObjectID(g, block.Blocker)
		if !ok {
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

// legalDeclareBlockersActions returns the legal declare-blockers actions for
// playerID. It delegates to combatEngine.legalBlockers.
func legalDeclareBlockersActions(g *game.Game, playerID game.PlayerID) []action.Action {
	return combatEngine{}.legalBlockers(g, playerID)
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
	if effectiveController(g, permanent) != playerID || permanent.Tapped {
		return false
	}
	if permanent.PhasedOut {
		return false
	}
	if ruleEffectProhibitsBlock(g, permanent) {
		return false
	}
	// CR 702.86b: a creature with unleash can't block while it has a +1/+1
	// counter on it.
	if hasKeyword(g, permanent, game.Unleash) && permanent.Counters.Get(counter.PlusOnePlusOne) > 0 {
		return false
	}
	return permanentHasType(g, permanent, types.Creature)
}

func canAttackTarget(g *game.Game, attacker *game.Permanent, target game.AttackTarget) bool {
	return !ruleEffectProhibitsAttack(g, attacker, &target)
}

func canBlockAttacker(g *game.Game, blocker, attacker *game.Permanent) bool {
	if ruleEffectProhibitsBeingBlocked(g, attacker) {
		return false
	}
	if ruleEffectProhibitsBlockingAttacker(g, blocker, attacker) {
		return false
	}
	if ruleEffectRestrictsBlocker(g, attacker, blocker) {
		return false
	}
	if hasKeyword(g, attacker, game.Flying) && !hasKeyword(g, blocker, game.Flying) && !hasKeyword(g, blocker, game.Reach) {
		return false
	}
	// CR 702.31c: a creature with horsemanship can't be blocked except by
	// creatures with horsemanship.
	if hasKeyword(g, attacker, game.Horsemanship) && !hasKeyword(g, blocker, game.Horsemanship) {
		return false
	}
	// CR 702.28c: a creature with shadow can block or be blocked by only
	// creatures with shadow, so shadow and non-shadow creatures can't block
	// each other in either direction.
	if hasKeyword(g, attacker, game.Shadow) != hasKeyword(g, blocker, game.Shadow) {
		return false
	}
	// CR 702.36c: a creature with fear can't be blocked except by artifact
	// creatures and/or black creatures.
	if hasKeyword(g, attacker, game.Fear) &&
		!permanentHasType(g, blocker, types.Artifact) &&
		!slices.Contains(permanentEffectiveColors(g, blocker), color.Black) {
		return false
	}
	// CR 702.72b: a creature with skulk can't be blocked by creatures with
	// greater power than it.
	if hasKeyword(g, attacker, game.Skulk) &&
		effectivePower(g, blocker) > effectivePower(g, attacker) {
		return false
	}
	// CR 702.13b: a creature with intimidate can't be blocked except by artifact
	// creatures and/or creatures that share a color with it.
	if hasKeyword(g, attacker, game.Intimidate) &&
		!permanentHasType(g, blocker, types.Artifact) &&
		!sharesColor(permanentEffectiveColors(g, attacker), permanentEffectiveColors(g, blocker)) {
		return false
	}
	// CR 702.16b: the attacker can't be blocked by a permanent it has protection from.
	if permanentProtectedFromPermanentEffective(g, attacker, blocker) {
		return false
	}
	// CR 702.14c: a creature with landwalk can't be blocked as long as the
	// defending player (the blocker's controller) controls a matching land.
	if attackerLandwalkUnblockableBy(g, attacker, blocker) {
		return false
	}
	return true
}

func attackerRequiresMultipleBlockers(g *game.Game, attacker *game.Permanent) bool {
	return hasKeyword(g, attacker, game.Menace)
}

// sharesColor reports whether the two color sets have at least one color in
// common. Colorless creatures (empty color sets) share no color.
func sharesColor(a, b []color.Color) bool {
	for _, c := range a {
		if slices.Contains(b, c) {
			return true
		}
	}
	return false
}

// applyDeclareBlockers validates and applies the declare-blockers action.
// It delegates to combatEngine.applyBlockers.
func (e *Engine) applyDeclareBlockers(g *game.Game, playerID game.PlayerID, declare action.DeclareBlockersAction) bool {
	return combatEngine{e}.applyBlockers(g, playerID, declare)
}

func canDeclareBlockers(g *game.Game, playerID game.PlayerID) bool {
	return g.Combat != nil &&
		g.Turn.Phase == game.PhaseCombat &&
		g.Turn.Step == game.StepDeclareBlockers &&
		isPlayerAlive(g, playerID)
}

func attacksAgainstPlayer(g *game.Game, playerID game.PlayerID) []game.AttackDeclaration {
	if g.Combat == nil {
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
	if g.Combat == nil {
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
	if effectiveController(g, permanent) != playerID || permanent.Tapped || permanent.PhasedOut {
		return false
	}
	if !permanentHasType(g, permanent, types.Creature) || hasKeyword(g, permanent, game.Defender) {
		return false
	}
	if ruleEffectProhibitsAttack(g, permanent, nil) {
		return false
	}
	return !permanent.SummoningSick || hasKeyword(g, permanent, game.Haste) || suspendHasteApplies(g, permanent)
}

func suspendHasteApplies(g *game.Game, permanent *game.Permanent) bool {
	return permanent.SuspendHasteController.Exists && permanent.SuspendHasteController.Val == effectiveController(g, permanent)
}

// legalDeclareAttackersActions returns the legal declare-attackers actions for
// playerID. It delegates to combatEngine.legalAttackers.
func legalDeclareAttackersActions(g *game.Game, playerID game.PlayerID) []action.Action {
	return combatEngine{}.legalAttackers(g, playerID)
}

// applyDeclareAttackers validates and applies the declare-attackers action.
// It delegates to combatEngine.applyAttackers.
func (e *Engine) applyDeclareAttackers(g *game.Game, playerID game.PlayerID, declare action.DeclareAttackersAction) bool {
	return combatEngine{e}.applyAttackers(g, playerID, declare)
}

func canDeclareAttackers(g *game.Game, playerID game.PlayerID) bool {
	return g.Combat != nil &&
		g.Turn.Phase == game.PhaseCombat &&
		g.Turn.Step == game.StepDeclareAttackers &&
		playerID == g.Turn.ActivePlayer &&
		isPlayerAlive(g, playerID)
}

func isLegalAttackTarget(g *game.Game, attackerController game.PlayerID, target game.AttackTarget) bool {
	if target.NoTarget {
		return false
	}
	if target.Player == attackerController || !isPlayerAlive(g, target.Player) {
		return false
	}
	if target.IsPlayerAttack() {
		return true
	}
	permanent, ok := attackTargetPermanent(g, target)
	if !ok || effectiveController(g, permanent) != target.Player {
		return false
	}
	if target.PlaneswalkerID != 0 {
		return target.BattleID == 0 && permanentHasType(g, permanent, types.Planeswalker)
	}
	if target.BattleID != 0 {
		return permanentHasType(g, permanent, types.Battle)
	}
	return false
}

func (combatEngine) declareAttackersSatisfiesRequirements(g *game.Game, playerID game.PlayerID, declarations []game.AttackDeclaration, eligibleByID map[id.ID]*game.Permanent) bool {
	declared := make(map[id.ID]game.AttackTarget, len(declarations))
	for _, declaration := range declarations {
		declared[declaration.Attacker] = declaration.Target
	}
	for _, attacker := range eligibleByID {
		target, isAttacking := declared[attacker.ObjectID]
		if !attackerMustAttack(g, attacker) {
			continue
		}
		if !isAttacking {
			if requiredAttackerCanAttackWithoutTax(g, playerID, attacker) {
				return false
			}
			continue
		}
		if !goadAllowsAttackTarget(g, playerID, attacker, target) {
			return false
		}
	}
	return true
}

func requiredAttackerCanAttackWithoutTax(g *game.Game, playerID game.PlayerID, attacker *game.Permanent) bool {
	_, ok := preferredRequiredAttackTarget(g, playerID, attacker)
	return ok
}

func preferredRequiredAttackTarget(g *game.Game, playerID game.PlayerID, attacker *game.Permanent) (game.AttackTarget, bool) {
	for _, target := range legalAttackTargets(g, playerID) {
		if !canAttackTarget(g, attacker, target) || !goadAllowsAttackTarget(g, playerID, attacker, target) {
			continue
		}
		declaration := []game.AttackDeclaration{{
			Attacker: attacker.ObjectID,
			Target:   target,
		}}
		if _, taxed := (combatEngine{}).attackTaxCost(g, declaration); !taxed {
			return target, true
		}
	}
	return game.AttackTarget{}, false
}

func preferredRequiredAttackDeclarations(g *game.Game, playerID game.PlayerID, attackers []*game.Permanent) []game.AttackDeclaration {
	declarations := make([]game.AttackDeclaration, 0, len(attackers))
	for _, attacker := range attackers {
		if !attackerMustAttack(g, attacker) {
			continue
		}
		target, ok := preferredRequiredAttackTarget(g, playerID, attacker)
		if !ok {
			continue
		}
		declarations = append(declarations, game.AttackDeclaration{
			Attacker: attacker.ObjectID,
			Target:   target,
		})
	}
	return declarations
}

func attackerMustAttack(g *game.Game, attacker *game.Permanent) bool {
	return isGoaded(attacker) || ruleEffectRequiresAttack(g, attacker)
}

func goadAllowsAttackTarget(g *game.Game, playerID game.PlayerID, attacker *game.Permanent, target game.AttackTarget) bool {
	if !isLegalAttackTarget(g, playerID, target) {
		return false
	}
	if !isGoaded(attacker) || !canAttackNonGoadingOpponentWithoutTax(g, playerID, attacker) {
		return true
	}
	return target.IsPlayerAttack() && !wasGoadedBy(attacker, target.Player)
}

func canAttackNonGoadingOpponentWithoutTax(g *game.Game, playerID game.PlayerID, attacker *game.Permanent) bool {
	for _, opponent := range aliveOpponents(g, playerID) {
		if wasGoadedBy(attacker, opponent) {
			continue
		}
		target := game.AttackTarget{Player: opponent}
		if !canAttackTarget(g, attacker, target) {
			continue
		}
		declaration := []game.AttackDeclaration{{
			Attacker: attacker.ObjectID,
			Target:   target,
		}}
		if _, taxed := (combatEngine{}).attackTaxCost(g, declaration); !taxed {
			return true
		}
	}
	return false
}

func legalAttackTargets(g *game.Game, attackerController game.PlayerID) []game.AttackTarget {
	var targets []game.AttackTarget
	for _, opponent := range aliveOpponents(g, attackerController) {
		targets = append(targets, game.AttackTarget{Player: opponent})
	}
	for _, permanent := range g.Battlefield {
		permanentController := effectiveController(g, permanent)
		if permanent.PhasedOut || permanentController == attackerController || !isPlayerAlive(g, permanentController) {
			continue
		}
		switch {
		case permanentHasType(g, permanent, types.Planeswalker):
			targets = append(targets, game.AttackTarget{Player: permanentController, PlaneswalkerID: permanent.ObjectID})
		case permanentHasType(g, permanent, types.Battle):
			targets = append(targets, game.AttackTarget{Player: permanentController, BattleID: permanent.ObjectID})
		default:
		}
	}
	return targets
}

func attackTargetPermanent(g *game.Game, target game.AttackTarget) (*game.Permanent, bool) {
	var permanent *game.Permanent
	var ok bool
	switch {
	case target.PlaneswalkerID != 0:
		permanent, ok = permanentByObjectID(g, target.PlaneswalkerID)
	case target.BattleID != 0:
		permanent, ok = permanentByObjectID(g, target.BattleID)
	default:
		return nil, false
	}
	if !ok || !activeBattlefieldPermanent(permanent) {
		return nil, false
	}
	return permanent, true
}

func isGoaded(permanent *game.Permanent) bool {
	for _, status := range permanent.Goaded {
		if status.ExpiresFor >= 0 {
			return true
		}
	}
	return false
}

func wasGoadedBy(permanent *game.Permanent, player game.PlayerID) bool {
	_, ok := permanent.Goaded[player]
	return ok
}

func goadPermanent(g *game.Game, permanent *game.Permanent, player game.PlayerID) {
	if permanent.Goaded == nil {
		permanent.Goaded = make(map[game.PlayerID]game.GoadStatus)
	}
	permanent.Goaded[player] = game.GoadStatus{CreatedTurn: g.Turn.TurnNumber, ExpiresFor: player}
}

func expireGoadForActivePlayer(g *game.Game) {
	for _, permanent := range g.Battlefield {
		if len(permanent.Goaded) == 0 {
			continue
		}
		for player, status := range permanent.Goaded {
			if status.ExpiresFor == g.Turn.ActivePlayer && status.CreatedTurn < g.Turn.TurnNumber {
				delete(permanent.Goaded, player)
			}
		}
	}
}

func permanentMapByObjectID(permanents []*game.Permanent) map[id.ID]*game.Permanent {
	byID := make(map[id.ID]*game.Permanent, len(permanents))
	for _, permanent := range permanents {
		byID[permanent.ObjectID] = permanent
	}
	return byID
}

func aliveOpponents(g *game.Game, playerID game.PlayerID) []game.PlayerID {
	var opponents []game.PlayerID
	for opponent := range game.PlayerID(game.NumPlayers) {
		if opponent != playerID && isPlayerAlive(g, opponent) {
			opponents = append(opponents, opponent)
		}
	}
	return opponents
}

func permanentCardDef(g *game.Game, permanent *game.Permanent) (*game.CardDef, bool) {
	return permanentFaceDef(g, permanent)
}

func permanentByObjectID(g *game.Game, objectID id.ID) (*game.Permanent, bool) {
	for _, permanent := range g.Battlefield {
		if permanent.ObjectID == objectID {
			return permanent, true
		}
	}
	return nil, false
}

// activeBattlefieldPermanent is the ordinary-rules view of a stored battlefield
// object. Identity, phasing, and last-known-information paths intentionally use
// permanentByObjectID directly so they can still find phased-out permanents.
func activeBattlefieldPermanent(permanent *game.Permanent) bool {
	return permanent != nil && !permanent.PhasedOut
}

func lethalDamageNeeded(g *game.Game, permanent *game.Permanent) (int, bool) {
	toughness, ok := effectiveToughness(g, permanent)
	if !ok || toughness <= 0 {
		return 0, ok
	}
	return toughness, true
}

// declareCreatedTokensAttacking puts freshly created tokens onto the battlefield
// attacking (CR 508.4), used for "... token that's tapped and attacking." style
// effects. A creature can only be put onto the battlefield attacking during a
// combat in which its controller is the attacking player, so this is a no-op
// outside combat or when the tokens' controller is not the active player; in
// that case the tokens simply remain on the battlefield without attacking. Each
// token's controller chooses which defending player it attacks; the entry never
// emits an attacker-declared event, so "whenever a creature attacks" abilities do
// not trigger for it (CR 508.4 declares it attacking without it being declared an
// attacker).
func declareCreatedTokensAttacking(e *Engine, g *game.Game, controller game.PlayerID, tokens []*game.Permanent, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	if g.Combat == nil || g.Turn.ActivePlayer != controller {
		return
	}
	for _, token := range tokens {
		if token == nil {
			continue
		}
		defender, ok := chooseEntryAttackDefender(e, g, controller, agents, log)
		if !ok {
			return
		}
		g.Combat.Attackers = append(g.Combat.Attackers, game.AttackDeclaration{
			Attacker: token.ObjectID,
			Target:   game.AttackTarget{Player: defender},
		})
	}
}

// chooseEntryAttackDefender selects which player a token entering the
// battlefield attacking is attacking. It prefers the players already being
// attacked this combat and otherwise falls back to the controller's living
// opponents. With no eligible defender it reports false and the token enters
// without attacking; with a single eligible defender that defender is chosen
// without a prompt; otherwise the controller chooses.
func chooseEntryAttackDefender(e *Engine, g *game.Game, controller game.PlayerID, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (game.PlayerID, bool) {
	defenders := defendingPlayersInOrder(g)
	if len(defenders) == 0 {
		defenders = aliveOpponents(g, controller)
	}
	defenders = aliveDefenders(g, defenders)
	switch len(defenders) {
	case 0:
		return 0, false
	case 1:
		return defenders[0], true
	}
	options := make([]game.ChoiceOption, 0, len(defenders))
	for i, defender := range defenders {
		options = append(options, game.ChoiceOption{Index: i, Label: fmt.Sprintf("Player %d", defender+1)})
	}
	selected := e.chooseChoice(g, agents, game.ChoiceRequest{
		Kind:       game.ChoicePlayer,
		Player:     controller,
		Prompt:     "Choose a player for the attacking token to attack",
		Options:    options,
		MinChoices: 1,
		MaxChoices: 1,
	}, log)
	if len(selected) != 1 || selected[0] < 0 || selected[0] >= len(defenders) {
		return 0, false
	}
	return defenders[selected[0]], true
}

// aliveDefenders filters a defender list down to the players still alive.
func aliveDefenders(g *game.Game, defenders []game.PlayerID) []game.PlayerID {
	alive := make([]game.PlayerID, 0, len(defenders))
	for _, defender := range defenders {
		if isPlayerAlive(g, defender) {
			alive = append(alive, defender)
		}
	}
	return alive
}
