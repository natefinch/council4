package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
)

const maxStateBasedActionPasses = 1000

func (e *Engine) applyStateBasedActions(g *game.Game) []LossLog {
	losses, _ := e.applyStateBasedActionsWithDeaths(g)
	return losses
}

func (e *Engine) applyStateBasedActionsWithLog(g *game.Game, log *TurnLog) []LossLog {
	losses, deaths := e.applyStateBasedActionsWithDeaths(g)
	if log != nil {
		log.Losses = append(log.Losses, losses...)
		log.Deaths = append(log.Deaths, deaths...)
	}
	return losses
}

func (e *Engine) applyStateBasedActionsWithDeaths(g *game.Game) ([]LossLog, []PermanentDeathLog) {
	var losses []LossLog
	var deaths []PermanentDeathLog
	for i := 0; i < maxStateBasedActionPasses; i++ {
		changed, passLosses := e.checkStateBasedActions(g)
		permanentsChanged, passDeaths := e.checkPermanentStateBasedActions(g)
		attachmentsChanged, attachmentDeaths := checkAttachmentStateBasedActions(g)
		legendaryChanged, legendaryDeaths := checkLegendaryRuleStateBasedActions(g)
		countersChanged := checkCounterStateBasedActions(g)
		tokensChanged := removeTokensFromNonBattlefieldZones(g)
		losses = append(losses, passLosses...)
		deaths = append(deaths, passDeaths...)
		deaths = append(deaths, attachmentDeaths...)
		deaths = append(deaths, legendaryDeaths...)
		if !changed && !permanentsChanged && !attachmentsChanged && !legendaryChanged && !countersChanged && !tokensChanged {
			return losses, deaths
		}
	}
	panic("state-based actions did not converge")
}

func (e *Engine) checkStateBasedActions(g *game.Game) (bool, []LossLog) {
	if g == nil {
		return false, nil
	}

	changed := false
	var losses []LossLog
	for _, player := range g.Players {
		if player == nil {
			continue
		}
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

func (e *Engine) checkPermanentStateBasedActions(g *game.Game) (bool, []PermanentDeathLog) {
	if g == nil {
		return false, nil
	}

	type pendingDeath struct {
		objectID id.ID
		reason   PermanentDeathReason
	}
	var pending []pendingDeath
	for _, permanent := range g.Battlefield {
		reason, ok := permanentDeathReason(g, permanent)
		if ok {
			pending = append(pending, pendingDeath{
				objectID: permanent.ObjectID,
				reason:   reason,
			})
		}
	}
	if len(pending) == 0 {
		return false, nil
	}

	var deaths []PermanentDeathLog
	for _, death := range pending {
		var permanent *game.Permanent
		if permanentDeathBypassesDestroy(death.reason) {
			permanent = permanentByObjectID(g, death.objectID)
			if permanent == nil || !movePermanentToZone(g, permanent, game.ZoneGraveyard) {
				continue
			}
		} else {
			var ok bool
			permanent, ok = destroyPermanent(g, death.objectID)
			if !ok {
				continue
			}
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

func checkAttachmentStateBasedActions(g *game.Game) (bool, []PermanentDeathLog) {
	if g == nil {
		return false, nil
	}
	var illegalAuras []id.ID
	changed := false
	for _, permanent := range g.Battlefield {
		if permanent == nil {
			continue
		}
		if permanent.AttachedTo == nil {
			if isAuraPermanent(g, permanent) {
				illegalAuras = append(illegalAuras, permanent.ObjectID)
			}
			continue
		}
		target := permanentByObjectID(g, *permanent.AttachedTo)
		if canAttachPermanent(g, permanent, target) {
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
	for _, auraID := range illegalAuras {
		aura := permanentByObjectID(g, auraID)
		if aura == nil || !movePermanentToZone(g, aura, game.ZoneGraveyard) {
			continue
		}
		changed = true
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

func checkLegendaryRuleStateBasedActions(g *game.Game) (bool, []PermanentDeathLog) {
	if g == nil {
		return false, nil
	}
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
	for _, objectID := range pending {
		permanent := permanentByObjectID(g, objectID)
		if permanent == nil || !movePermanentToZone(g, permanent, game.ZoneGraveyard) {
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
	card := permanentCardDef(g, permanent)
	if card == nil || !card.IsLegendary() || card.Name == "" {
		return legendaryKey{}, false
	}
	return legendaryKey{
		controller: permanent.Controller,
		name:       card.Name,
	}, true
}

func permanentOlderThan(left, right *game.Permanent) bool {
	if left.Timestamp != right.Timestamp {
		return left.Timestamp < right.Timestamp
	}
	return left.ObjectID < right.ObjectID
}

func checkCounterStateBasedActions(g *game.Game) bool {
	if g == nil {
		return false
	}
	changed := false
	for _, permanent := range g.Battlefield {
		if permanent == nil {
			continue
		}
		if permanent.Counters.CancelOpposites() > 0 {
			changed = true
		}
	}
	return changed
}

func removeTokensFromNonBattlefieldZones(g *game.Game) bool {
	if g == nil {
		return false
	}
	changed := false
	for _, player := range g.Players {
		if player == nil {
			continue
		}
		for _, zone := range []*game.Zone{&player.Library, &player.Hand, &player.Graveyard, &player.Exile, &player.CommandZone} {
			for _, cardID := range zone.All() {
				if g.CardInstances[cardID] != nil {
					continue
				}
				if zone.Remove(cardID) {
					changed = true
				}
			}
		}
	}
	return changed
}

func permanentTokenName(permanent *game.Permanent) string {
	if permanent == nil || !permanent.Token || permanent.TokenDef == nil {
		return ""
	}
	return permanent.TokenDef.Name
}

func permanentDeathReason(g *game.Game, permanent *game.Permanent) (PermanentDeathReason, bool) {
	card := permanentCardDef(g, permanent)
	if card == nil {
		return "", false
	}
	if card.HasType(game.TypePlaneswalker) && permanent.Counters.Get(counter.Loyalty) <= 0 {
		return PermanentDeathReasonZeroLoyalty, true
	}
	if card.HasType(game.TypeBattle) && permanent.Counters.Get(counter.Defense) <= 0 {
		return PermanentDeathReasonZeroDefense, true
	}
	if !card.HasType(game.TypeCreature) {
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
	case PermanentDeathReasonZeroToughness, PermanentDeathReasonZeroLoyalty, PermanentDeathReasonZeroDefense, PermanentDeathReasonIllegalAura:
		return true
	default:
		return false
	}
}

func (e *Engine) eliminatePlayer(g *game.Game, playerID game.PlayerID) bool {
	if g == nil || playerID < 0 || int(playerID) >= len(g.Players) {
		return false
	}

	player := g.Players[playerID]
	if player == nil {
		return false
	}

	if player.Eliminated && g.TurnOrder.IsEliminated(playerID) {
		return false
	}

	player.Eliminated = true
	g.TurnOrder.Eliminate(playerID)
	return true
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
