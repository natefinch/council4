package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

const maxStateBasedActionPasses = 1000

// The state-based actions in CR 704.5 / 704.6 that this engine does not yet
// implement, recorded so the conformance gap is explicit:
//   - CR 704.5e: spell/card copies in the wrong zone cease to exist. Copies are
//     instead cleaned up where they are created and resolved (stack.go,
//     storm.go) rather than by a dedicated SBA check.
//   - CR 704.5k: the world rule. The world supertype exists (and a few
//     Commander-legal cards such as Concordant Crossroads have it), but the
//     "shortest time as a world permanent" removal is not modeled.
//   - CR 704.5r: removing counters beyond a "can't have more than N" cap.
//   - CR 704.5t/u/w/x/z: dungeon venture, space sculptor, battle protector, and
//     start-your-engines speed actions.
//   - CR 704.6a/b/e/f: Two-Headed Giant, Archenemy, and Planechase variant SBAs.

func (e *Engine) applyStateBasedActions(g *game.Game) []LossLog {
	losses, _ := e.applyStateBasedActionsWithDeaths(g)
	return losses
}

func (e *Engine) applyStateBasedActionsWithLog(g *game.Game, log *TurnLog) []LossLog {
	losses, _ := e.applyStateBasedActionsWithLogResult(g, log)
	return losses
}

func (e *Engine) applyStateBasedActionsWithLogResult(g *game.Game, log *TurnLog) ([]LossLog, bool) {
	losses, deaths, applied := e.applyStateBasedActionsWithOutcome(g)
	for _, loss := range losses {
		log.addLoss(loss)
	}
	for _, death := range deaths {
		log.addDeath(death)
	}
	return losses, applied
}

// applyStateBasedActionsWithDeaths performs all applicable state-based actions
// and returns the player losses and permanent deaths they caused (CR 704).
//
// CR 704.3: whenever a player would get priority, the game checks the SBA
// conditions and performs every applicable action simultaneously as a single
// event; if any were performed, the check repeats. This loop reproduces that
// "check, perform, repeat until stable" cycle. The pass cap is a safety net
// against a non-converging board (a rules bug), not a rules limit.
//
// CR 704.8: a permanent leaving as part of an SBA derives its last known
// information from the game state before any of that pass's actions were
// performed, which is why each check snapshots a permanent before moving it.
func (e *Engine) applyStateBasedActionsWithDeaths(g *game.Game) ([]LossLog, []PermanentDeathLog) {
	losses, deaths, _ := e.applyStateBasedActionsWithOutcome(g)
	return losses, deaths
}

func (*Engine) applyStateBasedActionsWithOutcome(g *game.Game) ([]LossLog, []PermanentDeathLog, bool) {
	var losses []LossLog
	var deaths []PermanentDeathLog
	applied := false
	for range maxStateBasedActionPasses {
		batchID := newPassBatchID(g)
		commanders := pendingCommanderZoneActions(g)
		durationsChanged := expireConditionalControlDurations(g)
		blessingChanged := checkAscendCityBlessing(g)
		pass := prepareStateBasedActionPass(g, commanders)
		changed, passDeaths := pass.apply(g, batchID)
		losses = append(losses, pass.losses...)
		deaths = append(deaths, passDeaths...)
		if !changed && !durationsChanged && !blessingChanged {
			return losses, deaths, applied
		}
		applied = true
	}
	panic("state-based actions did not converge")
}

// newPassBatchID returns a memoizing accessor for the simultaneous event ID
// shared by every permanent that dies during a single state-based-action pass.
// The ID is only generated on first use so passes that move no permanents do
// not consume an ID, and all deaths in the pass share one batch so that
// "another creature dies" and "one or more creatures die" triggers see the
// simultaneous set.
func newPassBatchID(g *game.Game) func() id.ID {
	var assigned id.ID
	return func() id.ID {
		if assigned == 0 {
			assigned = g.IDGen.Next()
		}
		return assigned
	}
}

// checkStateBasedActions performs the player-loss state-based actions:
//   - CR 704.5a: a player with 0 or less life loses.
//   - CR 704.5b: a player who attempted to draw from an empty library loses
//     (tracked in FailedDraws since the last SBA check).
//   - CR 704.5c: a player with ten or more poison counters loses
//     (HasLethalPoison).
//   - CR 704.6c (Commander): a player dealt 21+ combat damage by one commander
//     loses (HasLethalCommanderDamage).
//
// MarkedToLoseGame covers effects that directly state a player loses the game.
// CR 104.5: a player who loses the game leaves it; eliminatePlayer applies the
// CR 800.4 departure cleanup.
func (*Engine) checkStateBasedActions(g *game.Game) (bool, []LossLog) {
	pass := stateBasedActionPass{losses: playerStateBasedLosses(g)}
	changed, _ := pass.apply(g, newPassBatchID(g))
	return changed, pass.losses
}

// checkPermanentStateBasedActions destroys or moves the battlefield permanents
// that meet a state-based-action condition (see permanentDeathReason for the
// per-rule mapping). CR 704.3: every permanent that dies this pass shares one
// simultaneous event ID so "another creature dies" / "one or more creatures
// die" triggers see the whole set at once. CR 704.8: each permanent is
// snapshotted before any move so its last known information reflects the board
// before this pass's actions.
func (*Engine) checkPermanentStateBasedActions(g *game.Game, batchID func() id.ID) (bool, []PermanentDeathLog) {
	pass := stateBasedActionPass{departures: permanentStateBasedDepartures(g)}
	return pass.apply(g, batchID)
}

// checkAttachmentStateBasedActions handles the attachment state-based actions.
// CR 704.5m: an Aura attached to an illegal object/player, or attached to
// nothing, is put into its owner's graveyard. CR 704.5n: an Equipment or
// Fortification attached to an illegal permanent or to a player becomes
// unattached but stays on the battlefield. CR 704.5p: any other non-Aura,
// non-Equipment, non-Fortification permanent attached to something likewise
// becomes unattached and remains. detachPermanent covers the 704.5n/704.5p
// cases; illegal Auras are collected and moved to the graveyard.
func checkAttachmentStateBasedActions(g *game.Game, batchID func() id.ID) (bool, []PermanentDeathLog) {
	departures, changes := attachmentStateBasedActions(g)
	pass := stateBasedActionPass{departures: departures, attachments: changes}
	return pass.apply(g, batchID)
}

type legendaryKey struct {
	controller game.PlayerID
	name       string
}

// checkLegendaryRuleStateBasedActions enforces the legend rule. CR 704.5j: if a
// player controls two or more legendary permanents with the same name, they
// choose one to keep and the rest are put into their owners' graveyards. The
// keeper is chosen through the existing choice protocol, with the oldest
// timestamp as the deterministic fallback.
func checkLegendaryRuleStateBasedActions(g *game.Game, batchID func() id.ID) (bool, []PermanentDeathLog) {
	pass := stateBasedActionPass{departures: stateBasedDeparturesForIDs(g,
		legendaryRuleStateBasedActionCandidates(g), PermanentDeathReasonLegendaryRule)}
	return pass.apply(g, batchID)
}

func legendaryRuleStateBasedActionCandidates(g *game.Game) []id.ID {
	groups := legendaryRuleStateBasedActionGroups(g)
	var pending []id.ID
	for _, player := range triggerAPNAPPlayers(g) {
		pending = append(pending, chooseLegendaryRuleStateBasedActions(g, groups, player)...)
	}
	return pending
}

type legendaryRuleGroup struct {
	controller game.PlayerID
	members    []*game.Permanent
	keeper     int
}

func legendaryRuleStateBasedActionGroups(g *game.Game) []legendaryRuleGroup {
	g.BeginStaticSourceFrame()
	defer g.EndStaticSourceFrame()
	exempt := make(map[game.PlayerID]bool, game.NumPlayers)
	for player := range game.PlayerID(game.NumPlayers) {
		exempt[player] = playerRuleEffectActive(g, player, game.RuleEffectLegendRuleDoesNotApply)
	}
	indices := make(map[legendaryKey]int)
	var groups []legendaryRuleGroup
	for _, permanent := range g.Battlefield {
		if !activeBattlefieldPermanent(permanent) {
			continue
		}
		key, ok := permanentLegendaryKey(g, permanent)
		if !ok || exempt[key.controller] {
			continue
		}
		index, exists := indices[key]
		if !exists {
			index = len(groups)
			indices[key] = index
			groups = append(groups, legendaryRuleGroup{controller: key.controller})
		}
		group := &groups[index]
		if len(group.members) > 0 && permanentOlderThan(permanent, group.members[group.keeper]) {
			group.keeper = len(group.members)
		}
		group.members = append(group.members, permanent)
	}
	return groups
}

func chooseLegendaryRuleStateBasedActions(g *game.Game, groups []legendaryRuleGroup, player game.PlayerID) []id.ID {
	var pending []id.ID
	for _, group := range groups {
		if group.controller != player || len(group.members) < 2 {
			continue
		}
		options := make([]game.ChoiceOption, len(group.members))
		for index, permanent := range group.members {
			options[index] = game.ChoiceOption{
				Index: index, Label: permanentChoiceLabel(g, permanent),
				Card: permanentChoiceInfo(g, permanent), Targets: []game.Target{game.PermanentTarget(permanent.ObjectID)},
			}
		}
		selected := group.keeper
		if ctx, exists := replacementChoiceContextFor(g); exists {
			selected = ctx.engine.chooseChoice(g, ctx.agents, game.ChoiceRequest{
				Kind: game.ChoiceResolution, Player: player, Prompt: "Choose one legendary permanent to keep.",
				Options: options, MinChoices: 1, MaxChoices: 1, DefaultSelection: []int{group.keeper},
			}, ctx.log)[0]
		}
		for index, permanent := range group.members {
			if index != selected {
				pending = append(pending, permanent.ObjectID)
			}
		}
	}
	return pending
}

func permanentLegendaryKey(g *game.Game, permanent *game.Permanent) (legendaryKey, bool) {
	name := permanentEffectiveName(g, permanent)
	if !permanentHasSupertype(g, permanent, types.Legendary) || name == "" {
		return legendaryKey{}, false
	}
	return legendaryKey{
		controller: effectiveController(g, permanent),
		name:       name,
	}, true
}

func permanentOlderThan(left, right *game.Permanent) bool {
	leftTimestamp := left.Timestamp()
	rightTimestamp := right.Timestamp()
	if leftTimestamp != rightTimestamp {
		return leftTimestamp < rightTimestamp
	}
	return left.ObjectID < right.ObjectID
}

type roleRuleKey struct {
	controller game.PlayerID
	attachedTo id.ID
}

// checkRoleRuleStateBasedActions enforces CR 704.5y. For each player and each
// permanent, only that player's Role Aura with the newest timestamp remains
// attached; Roles controlled by different players are independent and coexist.
func checkRoleRuleStateBasedActions(g *game.Game, batchID func() id.ID) (bool, []PermanentDeathLog) {
	pass := stateBasedActionPass{departures: stateBasedDeparturesForIDs(g,
		roleRuleStateBasedActionCandidates(g), PermanentDeathReasonRoleRule)}
	return pass.apply(g, batchID)
}

func roleRuleStateBasedActionCandidates(g *game.Game) []id.ID {
	g.BeginStaticSourceFrame()
	defer g.EndStaticSourceFrame()

	newest := make(map[roleRuleKey]*game.Permanent)
	counts := make(map[roleRuleKey]int)
	for _, permanent := range g.Battlefield {
		key, ok := permanentRoleRuleKey(g, permanent)
		if !ok {
			continue
		}
		counts[key]++
		if current := newest[key]; current == nil || permanentOlderThan(current, permanent) {
			newest[key] = permanent
		}
	}
	var pending []id.ID
	for _, permanent := range g.Battlefield {
		key, ok := permanentRoleRuleKey(g, permanent)
		if !ok || counts[key] <= 1 || newest[key] == permanent {
			continue
		}
		pending = append(pending, permanent.ObjectID)
	}
	return pending
}

func permanentRoleRuleKey(g *game.Game, permanent *game.Permanent) (roleRuleKey, bool) {
	if !activeBattlefieldPermanent(permanent) ||
		!permanent.AttachedTo.Exists ||
		!permanentHasSubtype(g, permanent, types.Aura) ||
		!permanentHasSubtype(g, permanent, types.Role) {
		return roleRuleKey{}, false
	}
	return roleRuleKey{
		controller: effectiveController(g, permanent),
		attachedTo: permanent.AttachedTo.Val,
	}, true
}

// checkCounterStateBasedActions removes paired +1/+1 and -1/-1 counters.
// CR 704.5q: if a permanent has both a +1/+1 and a -1/-1 counter, N of each are
// removed, where N is the smaller of the two counts (CancelOpposites).
func checkCounterStateBasedActions(g *game.Game) bool {
	pass := stateBasedActionPass{counters: counterStateBasedActions(g)}
	changed, _ := pass.apply(g, newPassBatchID(g))
	return changed
}

// removeTokensFromNonBattlefieldZones makes tokens that have left the
// battlefield cease to exist. CR 704.5d: a token in a zone other than the
// battlefield ceases to exist. A token that has moved out of play leaves a
// dangling card ID in its new zone with no backing CardInstance, so any such ID
// in a non-battlefield zone is removed here.
func removeTokensFromNonBattlefieldZones(g *game.Game) bool {
	pass := stateBasedActionPass{tokens: tokenStateBasedActions(g)}
	changed, _ := pass.apply(g, newPassBatchID(g))
	return changed
}

func permanentTokenName(permanent *game.Permanent) string {
	if !permanent.Token || permanent.TokenDef == nil {
		return ""
	}
	return permanent.TokenDef.Name
}

// permanentDeathReason reports whether a permanent meets a state-based-action
// condition that removes it, and which one. The conditions are independent —
// each applies to a distinct permanent type (Saga, planeswalker, battle,
// creature) — so the check order below does not affect the outcome; it does not
// match the numeric CR 704.5 ordering:
//   - CR 704.5s: a Saga at or past its final chapter number, not awaiting a
//     chapter ability still on the stack, is sacrificed.
//   - CR 704.5i: a planeswalker with 0 loyalty is put into its graveyard.
//   - CR 704.5v: a battle with 0 defense, not the source of an ability still on
//     the stack, is put into its graveyard.
//   - CR 704.5f: a creature with toughness 0 or less is put into its graveyard.
//     This is checked before indestructible and regeneration because neither
//     can replace it.
//   - CR 704.5h: a creature dealt damage by a deathtouch source is destroyed.
//   - CR 704.5g: a creature with lethal marked damage (>= its toughness) is
//     destroyed.
//
// Indestructible exempts a creature from the destroy-based rules (704.5g/h) but
// not from the toughness-0 rule (704.5f), matching the ordering above.
func permanentDeathReason(g *game.Game, permanent *game.Permanent) (PermanentDeathReason, bool) {
	if permanentHasSubtype(g, permanent, types.Saga) {
		final := finalSagaChapter(g, permanent)
		if final > 0 &&
			permanent.Counters.Get(counter.Lore) >= final &&
			!sagaAwaitingChapterAbility(g, permanent, final) &&
			!permanentCantBeSacrificed(g, permanent) {
			return PermanentDeathReasonSagaComplete, true
		}
	}
	if permanentHasType(g, permanent, types.Planeswalker) && permanent.Counters.Get(counter.Loyalty) <= 0 {
		return PermanentDeathReasonZeroLoyalty, true
	}
	if permanentHasType(g, permanent, types.Battle) &&
		permanent.Counters.Get(counter.Defense) <= 0 &&
		!battleAwaitingAbility(g, permanent) {
		return PermanentDeathReasonZeroDefense, true
	}
	if !permanentHasType(g, permanent, types.Creature) {
		return "", false
	}
	toughness, ok := effectiveToughness(g, permanent)
	if !ok {
		return "", false
	}

	if toughness <= 0 {
		return PermanentDeathReasonZeroToughness, true
	}
	if hasKeyword(g, permanent, game.Indestructible) {
		return "", false
	}
	if permanent.MarkedDeathtouchDamage {
		return PermanentDeathReasonLethalDamage, true
	}
	lethal, ok := lethalDamageNeeded(g, permanent)
	if ok && permanent.MarkedDamage >= lethal {
		return PermanentDeathReasonLethalDamage, true
	}
	return "", false
}

// battleAwaitingAbility reports whether a battle is the source of a triggered
// ability currently on the stack. CR 704.5v keeps a battle with 0 defense from
// being put into its graveyard while it "is the source of an ability that has
// triggered but not yet left the stack" — for example a "when this battle's
// defense becomes 0" trigger that has not finished resolving. Abilities that
// have triggered but have not yet been put on the stack are not covered here.
func battleAwaitingAbility(g *game.Game, permanent *game.Permanent) bool {
	for _, object := range g.Stack.Objects() {
		if object.Kind == game.StackTriggeredAbility && object.SourceID == permanent.ObjectID {
			return true
		}
	}
	return false
}

// permanent directly into the graveyard rather than destroying it. The
// CR 704.5 actions that move a permanent without destroying it can't be
// replaced by regeneration or a destruction-replacement (toughness 0 per
// CR 704.5f, 0 loyalty per 704.5i, 0 defense per 704.5v, illegal Aura per
// 704.5m, completed Saga per 704.5s); lethal and deathtouch damage (704.5g/h)
// destroy and therefore route through the regeneration-aware destroy path.
func permanentDeathBypassesDestroy(reason PermanentDeathReason) bool {
	switch reason {
	case PermanentDeathReasonZeroToughness, PermanentDeathReasonZeroLoyalty, PermanentDeathReasonZeroDefense, PermanentDeathReasonIllegalAura, PermanentDeathReasonSagaComplete:
		return true
	default:
		return false
	}
}

func (*Engine) eliminatePlayer(g *game.Game, playerID game.PlayerID) bool {
	return eliminatePlayer(g, playerID)
}

func eliminatePlayer(g *game.Game, playerID game.PlayerID) bool {
	if !markPlayerEliminated(g, playerID) {
		return false
	}
	passInitiativeOnElimination(g, playerID)
	cleanupEliminatedPlayer(g, playerID)
	return true
}

func markPlayerEliminated(g *game.Game, playerID game.PlayerID) bool {
	if playerID < 0 || int(playerID) >= len(g.Players) {
		return false
	}

	player := g.Players[playerID]

	if player.Eliminated && g.TurnOrder.IsEliminated(playerID) {
		return false
	}

	player.Eliminated = true
	g.TurnOrder.Eliminate(playerID)
	return true
}

// cleanupEliminatedPlayer applies the consequences of a player leaving the game.
// CR 800.4a: all objects owned by the departing player leave the game, effects
// granting them control of objects end, their stack objects not represented by
// cards cease to exist, and any objects they still control are exiled.
func cleanupEliminatedPlayer(g *game.Game, playerID game.PlayerID) {
	before := make(map[id.ID]game.PlayerID, len(g.Battlefield))
	for _, permanent := range g.Battlefield {
		before[permanent.ObjectID] = effectiveController(g, permanent)
	}
	leaveDepartedOwnedObjects(g)
	for _, player := range g.Players {
		if player.Eliminated {
			g.EndPlayerControls(player.ID)
			g.ControlGrantsEnded[player.ID] = true
		}
	}
	kept := g.ContinuousEffects[:0]
	for _, effect := range g.ContinuousEffects {
		if effect.Layer == game.LayerControl && effect.NewController.Exists && departedPlayer(g, effect.NewController.Val) {
			continue
		}
		kept = append(kept, effect)
	}
	g.ContinuousEffects = kept
	for _, player := range g.Players {
		if player.Eliminated {
			cleanupDepartedStackObjects(g, player.ID)
		}
	}
	settleDepartedControl(g)
	for _, permanent := range g.Battlefield {
		if before[permanent.ObjectID] != effectiveController(g, permanent) {
			permanent.SummoningSick = true
			removePermanentFromCombat(g, permanent.ObjectID)
		}
	}
	if g.Turn.PriorityPlayer == playerID && len(g.AlivePlayers()) > 0 {
		g.Turn.PriorityPlayer = g.TurnOrder.NextActivePlayer(playerID)
	}
	if g.Combat == nil {
		return
	}
	var removeFromCombat []id.ID
	for _, attack := range g.Combat.Attackers {
		attacker, ok := permanentByObjectID(g, attack.Attacker)
		if (!attack.Target.NoTarget && attack.Target.Player == playerID) ||
			(ok && effectiveController(g, attacker) == playerID) {
			removeFromCombat = append(removeFromCombat, attack.Attacker)
		}
	}
	for _, block := range g.Combat.Blockers {
		blocker, ok := permanentByObjectID(g, block.Blocker)
		if ok && effectiveController(g, blocker) == playerID {
			removeFromCombat = append(removeFromCombat, block.Blocker)
		}
	}
	for blockerID := range g.Combat.BlockersOfRemovedAttackers {
		blocker, ok := permanentByObjectID(g, blockerID)
		if ok && effectiveController(g, blocker) == playerID {
			removeFromCombat = append(removeFromCombat, blockerID)
		}
	}
	for _, objectID := range removeFromCombat {
		removePermanentFromCombat(g, objectID)
	}
}

func lossReason(g *game.Game, player *game.Player) LossReason {
	if g.MarkedToLoseGame[player.ID] {
		return LossReasonGameLossEffect
	}
	if g.FailedDraws[player.ID] {
		return LossReasonEmptyLibraryDraw
	}
	if player.Life <= 0 {
		return LossReasonZeroLife
	}
	if player.HasLethalPoison() {
		return LossReasonPoisonCounters
	}
	if player.HasLethalCommanderDamage() {
		return LossReasonCommanderDamage
	}
	return LossReasonStateBasedEliminate
}
