package payment

import (
	"slices"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
)

func additionalCostPlanStillValid(s State, player *game.Player, plan additionalCostPlan) bool {
	if plan.untapSource != nil {
		current, ok := s.PermanentByObjectID(plan.untapSource.ObjectID)
		if !ok ||
			current != plan.untapSource ||
			s.EffectiveController(current) != player.ID ||
			!canUntapForAbility(s, current) {
			return false
		}
	}
	plannedCounters := make(map[*game.Permanent]map[counter.Kind]int)
	for _, removal := range plan.counterRemovals {
		current, ok := s.PermanentByObjectID(removal.source.ObjectID)
		if !ok || current != removal.source || s.EffectiveController(current) != player.ID {
			return false
		}
		if plannedCounters[current] == nil {
			plannedCounters[current] = make(map[counter.Kind]int)
		}
		plannedCounters[current][removal.kind] += removal.amount
		if current.Counters.Get(removal.kind) < plannedCounters[current][removal.kind] {
			return false
		}
	}
	for _, sacrifice := range plan.sacrifices {
		permanent, ok := s.PermanentByObjectID(sacrifice.ObjectID)
		if !ok || s.EffectiveController(permanent) != player.ID || permanent != sacrifice {
			return false
		}
	}
	for _, permanentToTap := range plan.permanentsToTap {
		permanent, ok := s.PermanentByObjectID(permanentToTap.ObjectID)
		if !ok ||
			s.EffectiveController(permanent) != player.ID ||
			permanent != permanentToTap ||
			permanent.Tapped {
			return false
		}
	}
	for _, returned := range plan.returnsToHand {
		permanent, ok := s.PermanentByObjectID(returned.permanent.ObjectID)
		if !ok ||
			s.EffectiveController(permanent) != player.ID ||
			permanent != returned.permanent ||
			!additionalCostMatchesPermanent(s, permanent, returned.additional) {
			return false
		}
	}
	for _, permanent := range plan.exilePermanents {
		current, ok := s.PermanentByObjectID(permanent.ObjectID)
		if !ok || s.EffectiveController(current) != player.ID || current != permanent {
			return false
		}
	}
	if plan.exertSource != nil {
		current, ok := s.PermanentByObjectID(plan.exertSource.ObjectID)
		if !ok ||
			current != plan.exertSource ||
			s.EffectiveController(current) != player.ID {
			return false
		}
	}
	for _, placement := range plan.counterAdds {
		current, ok := s.PermanentByObjectID(placement.source.ObjectID)
		if !ok || current != placement.source || s.EffectiveController(current) != player.ID {
			return false
		}
	}
	for _, cardID := range plan.discards {
		if !player.Hand.Contains(cardID) {
			return false
		}
	}
	for _, exile := range plan.exiles {
		if !zoneContainsCard(s, player.ID, exile.zone, exile.cardID) {
			return false
		}
	}
	if hasDuplicateCardZoneSelections(plannedEvidenceCards(plan)) {
		return false
	}
	for _, reveal := range plan.reveals {
		if !zoneContainsCard(s, player.ID, reveal.zone, reveal.cardID) {
			return false
		}
	}
	for _, evidence := range plan.evidence {
		if !evidenceCardsMeetThreshold(s, player.ID, evidence.cards, evidence.threshold) {
			return false
		}
	}
	if plan.lifePaid > 0 && (player.Life < plan.lifePaid || !s.CanPayLife(player.ID)) {
		return false
	}
	if plan.energyPaid > 0 && player.EnergyCounters < plan.energyPaid {
		return false
	}
	return true
}

func applyAdditionalCostPlan(s State, plan additionalCostPlan) bool {
	if plan.lifePaid > 0 {
		player, ok := s.Player(plan.player)
		if !ok || player.Life < plan.lifePaid || !s.CanPayLife(plan.player) {
			return false
		}
	}
	if plan.untapSource != nil {
		s.SetTapped(plan.untapSource, false)
	}
	for _, removal := range plan.counterRemovals {
		if !s.RemoveCounters(removal.source, removal.kind, removal.amount) {
			return false
		}
	}
	if plan.exertSource != nil && !s.ExertPermanent(plan.exertSource) {
		return false
	}
	for _, placement := range plan.counterAdds {
		if !s.AddCounters(plan.player, placement.source, placement.kind, placement.amount) {
			return false
		}
	}
	if plan.millAmount > 0 {
		s.MillCards(plan.player, plan.millAmount)
	}
	for _, sacrifice := range plan.sacrifices {
		if !s.SacrificePermanent(sacrifice) {
			return false
		}
	}
	for _, permanentToTap := range plan.permanentsToTap {
		s.SetTapped(permanentToTap, true)
	}
	for _, returned := range plan.returnsToHand {
		if !s.MovePermanentToZone(returned.permanent, zone.Hand) {
			return false
		}
	}
	for _, permanent := range plan.exilePermanents {
		if !s.MovePermanentToZone(permanent, zone.Exile) {
			return false
		}
	}
	for _, cardID := range plan.discards {
		if !s.DiscardFromHand(plan.player, cardID) {
			return false
		}
	}
	for _, exile := range plan.exiles {
		if !s.MoveCard(plan.player, exile.cardID, exile.zone, zone.Exile) {
			return false
		}
	}
	for _, reveal := range plan.reveals {
		s.EmitCardReveal(plan.player, plan.sourceCardID, reveal.cardID, reveal.zone)
	}
	for _, evidence := range plan.evidence {
		for _, card := range evidence.cards {
			if !s.MoveCard(plan.player, card.cardID, zone.Graveyard, zone.Exile) {
				return false
			}
		}
	}
	if plan.lifePaid > 0 {
		s.LoseLife(plan.player, plan.lifePaid)
	}
	if plan.energyPaid > 0 {
		player, ok := s.Player(plan.player)
		if !ok || player.EnergyCounters < plan.energyPaid {
			return false
		}
		if !s.SetPlayerEnergyCounters(plan.player, player.EnergyCounters-plan.energyPaid) {
			return false
		}
	}
	return true
}

func zoneContainsCard(s State, playerID game.PlayerID, zoneType zone.Type, cardID id.ID) bool {
	player, ok := s.Player(playerID)
	if !ok {
		return false
	}
	return slices.Contains(cardIDsInZone(player, zoneType), cardID)
}

func cardIDsInZone(player *game.Player, zoneType zone.Type) []id.ID {
	switch zoneType {
	case zone.Library:
		return player.Library.All()
	case zone.Hand:
		return player.Hand.All()
	case zone.Graveyard:
		return player.Graveyard.All()
	case zone.Exile:
		return player.Exile.All()
	case zone.Command:
		return player.CommandZone.All()
	default:
		return nil
	}
}
