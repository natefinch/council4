package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

const maxStateBasedActionPasses = 1000

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

func (e *Engine) applyStateBasedActionsWithDeaths(g *game.Game) ([]LossLog, []PermanentDeathLog) {
	var losses []LossLog
	var deaths []PermanentDeathLog
	for range maxStateBasedActionPasses {
		batchID := newPassBatchID(g)
		changed, passLosses := e.checkStateBasedActions(g)
		permanentsChanged, passDeaths := e.checkPermanentStateBasedActions(g, batchID)
		attachmentsChanged, attachmentDeaths := checkAttachmentStateBasedActions(g, batchID)
		legendaryChanged, legendaryDeaths := checkLegendaryRuleStateBasedActions(g, batchID)
		countersChanged := checkCounterStateBasedActions(g)
		tokensChanged := removeTokensFromNonBattlefieldZones(g)
		durationsChanged := expireConditionalControlDurations(g)
		losses = append(losses, passLosses...)
		deaths = append(deaths, passDeaths...)
		deaths = append(deaths, attachmentDeaths...)
		deaths = append(deaths, legendaryDeaths...)
		if !changed && !permanentsChanged && !attachmentsChanged && !legendaryChanged && !countersChanged && !tokensChanged && !durationsChanged {
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

func (e *Engine) checkStateBasedActions(g *game.Game) (bool, []LossLog) {
	changed := false
	var losses []LossLog
	for _, player := range g.Players {
		if player.Eliminated {
			delete(g.FailedDraws, player.ID)
			continue
		}
		if player.Life <= 0 ||
			player.HasLethalPoison() ||
			player.HasLethalCommanderDamage() ||
			g.FailedDraws[player.ID] {
			reason := lossReason(g, player)
			if e.eliminatePlayer(g, player.ID) {
				changed = true
				losses = append(losses, LossLog{
					Player: player.ID,
					Reason: reason,
				})
			}
			delete(g.FailedDraws, player.ID)
		}
	}
	return changed, losses
}

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

func checkAttachmentStateBasedActions(g *game.Game, batchID func() id.ID) (bool, []PermanentDeathLog) {
	var illegalAuras []id.ID
	changed := false
	for _, permanent := range g.Battlefield {
		if !permanent.AttachedTo.Exists {
			if isAuraPermanent(g, permanent) {
				illegalAuras = append(illegalAuras, permanent.ObjectID)
			}
			continue
		}
		target, ok := permanentByObjectID(g, permanent.AttachedTo.Val)
		if ok && canAttachPermanent(g, permanent, target) {
			continue
		}
		if isAuraPermanent(g, permanent) {
			illegalAuras = append(illegalAuras, permanent.ObjectID)
			continue
		}
		detachPermanent(g, permanent)
		changed = true
	}
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

func checkLegendaryRuleStateBasedActions(g *game.Game, batchID func() id.ID) (bool, []PermanentDeathLog) {
	keepers := make(map[legendaryKey]*game.Permanent)
	counts := make(map[legendaryKey]int)
	for _, permanent := range g.Battlefield {
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
		key, ok := permanentLegendaryKey(g, permanent)
		if !ok || counts[key] <= 1 || keepers[key] == permanent {
			continue
		}
		pending = append(pending, permanent.ObjectID)
	}
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

func checkCounterStateBasedActions(g *game.Game) bool {
	changed := false
	for _, permanent := range g.Battlefield {
		if permanent.Counters.CancelOpposites() > 0 {
			changed = true
		}
	}
	return changed
}

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
	if permanentHasType(g, permanent, types.Battle) && permanent.Counters.Get(counter.Defense) <= 0 {
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
	cleanupEliminatedPlayer(g, playerID)
	return true
}

func cleanupEliminatedPlayer(g *game.Game, playerID game.PlayerID) {
	g.Stack.RemoveControlledBy(playerID)
	cleanupEliminatedPlayerPermanents(g, playerID)
	if g.Combat == nil {
		return
	}
	var removeFromCombat []id.ID
	for _, attack := range g.Combat.Attackers {
		attacker, ok := permanentByObjectID(g, attack.Attacker)
		if attack.Target.Player == playerID || (ok && effectiveController(g, attacker) == playerID) {
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
