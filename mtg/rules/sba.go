package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
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
//   - CR 704.5y: keeping only the most recent of multiple same-controller Roles.
//   - CR 704.6a/b/e/f: Two-Headed Giant, Archenemy, and Planechase variant SBAs.

func (e *Engine) applyStateBasedActions(g *game.Game) []LossLog {
	losses, _ := e.applyStateBasedActionsWithDeaths(g)
	return losses
}

func (e *Engine) applyStateBasedActionsWithLog(g *game.Game, log *TurnLog) []LossLog {
	losses, deaths := e.applyStateBasedActionsWithDeaths(g)
	for _, loss := range losses {
		log.addLoss(loss)
	}
	for _, death := range deaths {
		log.addDeath(death)
	}
	return losses
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
	var losses []LossLog
	var deaths []PermanentDeathLog
	for range maxStateBasedActionPasses {
		batchID := newPassBatchID(g)
		durationsChanged := expireConditionalControlDurations(g)
		changed, passLosses := e.checkStateBasedActions(g)
		permanentsChanged, passDeaths := e.checkPermanentStateBasedActions(g, batchID)
		attachmentsChanged, attachmentDeaths := checkAttachmentStateBasedActions(g, batchID)
		legendaryChanged, legendaryDeaths := checkLegendaryRuleStateBasedActions(g, batchID)
		countersChanged := checkCounterStateBasedActions(g)
		tokensChanged := removeTokensFromNonBattlefieldZones(g)
		blessingChanged := checkAscendCityBlessing(g)
		losses = append(losses, passLosses...)
		deaths = append(deaths, passDeaths...)
		deaths = append(deaths, attachmentDeaths...)
		deaths = append(deaths, legendaryDeaths...)
		if !changed && !permanentsChanged && !attachmentsChanged && !legendaryChanged && !countersChanged && !tokensChanged && !durationsChanged && !blessingChanged {
			return losses, deaths
		}
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
func (e *Engine) checkStateBasedActions(g *game.Game) (bool, []LossLog) {
	changed := false
	var losses []LossLog
	for _, player := range g.Players {
		if player.Eliminated {
			delete(g.FailedDraws, player.ID)
			delete(g.MarkedToLoseGame, player.ID)
			continue
		}
		if player.Life <= 0 ||
			player.HasLethalPoison() ||
			player.HasLethalCommanderDamage() ||
			g.FailedDraws[player.ID] ||
			g.MarkedToLoseGame[player.ID] {
			reason := lossReason(g, player)
			if e.eliminatePlayer(g, player.ID) {
				changed = true
				losses = append(losses, LossLog{
					Player: player.ID,
					Reason: reason,
				})
			}
			delete(g.FailedDraws, player.ID)
			delete(g.MarkedToLoseGame, player.ID)
		}
	}
	return changed, losses
}

// checkPermanentStateBasedActions destroys or moves the battlefield permanents
// that meet a state-based-action condition (see permanentDeathReason for the
// per-rule mapping). CR 704.3: every permanent that dies this pass shares one
// simultaneous event ID so "another creature dies" / "one or more creatures
// die" triggers see the whole set at once. CR 704.8: each permanent is
// snapshotted before any move so its last known information reflects the board
// before this pass's actions.
func (*Engine) checkPermanentStateBasedActions(g *game.Game, batchID func() id.ID) (bool, []PermanentDeathLog) {
	type pendingDeath struct {
		objectID id.ID
		reason   PermanentDeathReason
		snapshot game.ObjectSnapshot
	}
	var pending []pendingDeath
	// The death-reason scan is a pure read over every permanent; frame it so the
	// static-ability source set is built once, and close the frame before the
	// mutation phase below so the cache never spans a state change.
	func() {
		g.BeginStaticSourceFrame()
		defer g.EndStaticSourceFrame()
		for _, permanent := range g.Battlefield {
			if !activeBattlefieldPermanent(permanent) {
				continue
			}
			reason, ok := permanentDeathReason(g, permanent)
			if ok {
				pending = append(pending, pendingDeath{
					objectID: permanent.ObjectID,
					reason:   reason,
					snapshot: snapshotPermanent(g, permanent, zone.Battlefield),
				})
			}
		}
	}()
	if len(pending) == 0 {
		return false, nil
	}

	simultaneousID := batchID()
	var deaths []PermanentDeathLog
	for _, death := range pending {
		var permanent *game.Permanent
		if permanentDeathBypassesDestroy(death.reason) {
			var ok bool
			permanent, ok = permanentByObjectID(g, death.objectID)
			replacedToCommand := ok && commanderReplacementDestination(g, permanent.CardInstanceID, zone.Graveyard) == zone.Command
			if !ok || !movePermanentToZoneInBatch(g, permanent, zone.Graveyard, simultaneousID) {
				continue
			}
			rememberLastKnown(g, &death.snapshot)
			if replacedToCommand {
				continue
			}
		} else {
			var ok bool
			permanent, ok = destroyPermanentInBatch(g, death.objectID, simultaneousID, false)
			if !ok {
				if _, remains := permanentByObjectID(g, death.objectID); !remains {
					rememberLastKnown(g, &death.snapshot)
				}
				continue
			}
			rememberLastKnown(g, &death.snapshot)
		}
		deaths = append(deaths, PermanentDeathLog{
			Permanent:  permanent.ObjectID,
			SourceID:   permanent.CardInstanceID,
			TokenName:  permanentTokenName(permanent),
			Owner:      permanent.Owner,
			Controller: permanent.Controller,
			Reason:     death.reason,
		})
	}
	return len(deaths) > 0, deaths
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
	var illegalAuras []id.ID
	changed := false
	// Most permanents are not attachments, but identifying them still requires
	// effective characteristics. Reuse those values across the battlefield scan,
	// closing and reopening the frame around the uncommon detach mutation so the
	// cache never outlives the state it describes.
	func() {
		g.BeginStaticSourceFrame()
		defer g.EndStaticSourceFrame()
		for _, permanent := range g.Battlefield {
			if permanent.PhasedOut {
				continue
			}
			if permanent.AttachedToPlayer.Exists {
				// CR 704.5m: an Aura attached to a player stays on the
				// battlefield only while its Enchant restriction still allows
				// that player (the player is in the game and, for "Enchant
				// opponent", still an opponent of the Aura's controller). A
				// player attachment never applies to Equipment or a bestowed
				// creature, so an illegal one is always a 704.5m illegal Aura
				// bound for its owner's graveyard.
				if auraCanAttachToPlayer(g, permanent, permanent.AttachedToPlayer.Val) {
					continue
				}
				illegalAuras = append(illegalAuras, permanent.ObjectID)
				continue
			}
			if !permanent.AttachedTo.Exists {
				if permanent.Bestowed {
					// CR 702.103f: a bestowed Aura that is unattached stops being
					// a bestowed Aura and stays on the battlefield as a creature,
					// so it is exempt from the 704.5m illegal-Aura graveyard rule.
					g.EndStaticSourceFrame()
					permanent.Bestowed = false
					changed = true
					g.BeginStaticSourceFrame()
					continue
				}
				if isAuraPermanent(g, permanent) {
					illegalAuras = append(illegalAuras, permanent.ObjectID)
				}
				continue
			}
			target, ok := permanentByObjectID(g, permanent.AttachedTo.Val)
			if ok && canAttachPermanent(g, permanent, target) {
				continue
			}
			if permanent.Bestowed {
				// CR 702.103f: a bestowed Aura attached to an illegal object
				// becomes unattached, stops being a bestowed Aura, and stays on
				// the battlefield as a creature rather than being put into the
				// graveyard as an illegal Aura.
				g.EndStaticSourceFrame()
				permanent.Bestowed = false
				detachPermanent(g, permanent)
				changed = true
				g.BeginStaticSourceFrame()
				continue
			}
			if isAuraPermanent(g, permanent) {
				illegalAuras = append(illegalAuras, permanent.ObjectID)
				continue
			}
			g.EndStaticSourceFrame()
			detachPermanent(g, permanent)
			changed = true
			g.BeginStaticSourceFrame()
		}
	}()
	if len(illegalAuras) == 0 {
		return changed, nil
	}
	var deaths []PermanentDeathLog
	simultaneousID := batchID()
	for _, auraID := range illegalAuras {
		aura, ok := permanentByObjectID(g, auraID)
		replacedToCommand := ok && commanderReplacementDestination(g, aura.CardInstanceID, zone.Graveyard) == zone.Command
		if !ok || !movePermanentToZoneInBatch(g, aura, zone.Graveyard, simultaneousID) {
			continue
		}
		changed = true
		if replacedToCommand {
			continue
		}
		deaths = append(deaths, PermanentDeathLog{
			Permanent:  aura.ObjectID,
			SourceID:   aura.CardInstanceID,
			TokenName:  permanentTokenName(aura),
			Owner:      aura.Owner,
			Controller: aura.Controller,
			Reason:     PermanentDeathReasonIllegalAura,
		})
	}
	return changed, deaths
}

type legendaryKey struct {
	controller game.PlayerID
	name       string
}

// checkLegendaryRuleStateBasedActions enforces the legend rule. CR 704.5j: if a
// player controls two or more legendary permanents with the same name, they
// choose one to keep and the rest are put into their owners' graveyards. The
// keeper here is chosen deterministically (oldest timestamp) rather than by a
// player choice, which is a simplification of the "that player chooses" wording
// but yields a legal resulting board.
func checkLegendaryRuleStateBasedActions(g *game.Game, batchID func() id.ID) (bool, []PermanentDeathLog) {
	pending := legendaryRuleStateBasedActionCandidates(g)
	if len(pending) == 0 {
		return false, nil
	}

	var deaths []PermanentDeathLog
	simultaneousID := batchID()
	for _, objectID := range pending {
		permanent, ok := permanentByObjectID(g, objectID)
		replacedToCommand := ok && commanderReplacementDestination(g, permanent.CardInstanceID, zone.Graveyard) == zone.Command
		if !ok || !movePermanentToZoneInBatch(g, permanent, zone.Graveyard, simultaneousID) {
			continue
		}
		if replacedToCommand {
			continue
		}
		deaths = append(deaths, PermanentDeathLog{
			Permanent:  permanent.ObjectID,
			SourceID:   permanent.CardInstanceID,
			TokenName:  permanentTokenName(permanent),
			Owner:      permanent.Owner,
			Controller: permanent.Controller,
			Reason:     PermanentDeathReasonLegendaryRule,
		})
	}
	return len(deaths) > 0, deaths
}

func legendaryRuleStateBasedActionCandidates(g *game.Game) []id.ID {
	// Both passes are pure reads over the same battlefield. One frame lets every
	// effective-name, supertype, and controller query share its derived values;
	// the caller closes this frame before moving any duplicate permanents.
	g.BeginStaticSourceFrame()
	defer g.EndStaticSourceFrame()

	keepers := make(map[legendaryKey]*game.Permanent)
	counts := make(map[legendaryKey]int)
	for _, permanent := range g.Battlefield {
		if !activeBattlefieldPermanent(permanent) {
			continue
		}
		key, ok := permanentLegendaryKey(g, permanent)
		if !ok {
			continue
		}
		counts[key]++
		if current := keepers[key]; current == nil || permanentOlderThan(permanent, current) {
			keepers[key] = permanent
		}
	}

	var pending []id.ID
	for _, permanent := range g.Battlefield {
		if !activeBattlefieldPermanent(permanent) {
			continue
		}
		key, ok := permanentLegendaryKey(g, permanent)
		if !ok || counts[key] <= 1 || keepers[key] == permanent {
			continue
		}
		pending = append(pending, permanent.ObjectID)
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

// checkCounterStateBasedActions removes paired +1/+1 and -1/-1 counters.
// CR 704.5q: if a permanent has both a +1/+1 and a -1/-1 counter, N of each are
// removed, where N is the smaller of the two counts (CancelOpposites).
func checkCounterStateBasedActions(g *game.Game) bool {
	changed := false
	for _, permanent := range g.Battlefield {
		if !activeBattlefieldPermanent(permanent) {
			continue
		}
		if permanent.Counters.CancelOpposites() > 0 {
			changed = true
		}
	}
	return changed
}

// removeTokensFromNonBattlefieldZones makes tokens that have left the
// battlefield cease to exist. CR 704.5d: a token in a zone other than the
// battlefield ceases to exist. A token that has moved out of play leaves a
// dangling card ID in its new zone with no backing CardInstance, so any such ID
// in a non-battlefield zone is removed here.
func removeTokensFromNonBattlefieldZones(g *game.Game) bool {
	changed := false
	for _, player := range g.Players {
		for _, cards := range []*zone.Zone{&player.Library, &player.Hand, &player.Graveyard, &player.Exile, &player.CommandZone} {
			for _, cardID := range cards.All() {
				if g.CardInstances[cardID] != nil {
					continue
				}
				if cards.Remove(cardID) {
					changed = true
				}
			}
		}
	}
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
			!sagaAwaitingChapterAbility(g, permanent, final) {
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
	if playerID < 0 || int(playerID) >= len(g.Players) {
		return false
	}

	player := g.Players[playerID]

	if player.Eliminated && g.TurnOrder.IsEliminated(playerID) {
		return false
	}

	player.Eliminated = true
	g.TurnOrder.Eliminate(playerID)
	passInitiativeOnElimination(g, playerID)
	cleanupEliminatedPlayer(g, playerID)
	return true
}

// cleanupEliminatedPlayer applies the consequences of a player leaving the game.
// CR 800.4a: all objects owned by the departing player leave the game, effects
// granting them control of objects end, their stack objects not represented by
// cards cease to exist, and any objects they still control are exiled. Here
// their stack objects are removed, their permanents exiled, and control of
// permanents they controlled but did not own reverts to the owner. Permanents
// that leave the battlefield this way are removed from combat (CR 506.4).
func cleanupEliminatedPlayer(g *game.Game, playerID game.PlayerID) {
	g.Stack.RemoveControlledBy(playerID)
	cleanupEliminatedPlayerPermanents(g, playerID)
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
	for _, objectID := range removeFromCombat {
		removePermanentFromCombat(g, objectID)
	}
}

func cleanupEliminatedPlayerPermanents(g *game.Game, playerID game.PlayerID) {
	for {
		var owned *game.Permanent
		for _, permanent := range g.Battlefield {
			if permanent.Owner == playerID {
				owned = permanent
				break
			}
			if effectiveController(g, permanent) == playerID {
				permanent.Controller = permanent.Owner
			}
		}
		if owned == nil {
			return
		}
		movePermanentToZone(g, owned, zone.Exile)
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
