package payment

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

type additionalCostPlan struct {
	player     game.PlayerID
	paid       []string
	sacrifices []*game.Permanent
	discards   []id.ID
	exiles     []cardZoneSelection
	lifePaid   int
}

type cardZoneSelection struct {
	cardID id.ID
	zone   game.ZoneType
}

func buildAdditionalCostPlanForCosts(s State, playerID game.PlayerID, costs []game.AdditionalCost, prefs *Preferences, source *game.Permanent, sourceCardID id.ID, sourceZone game.ZoneType) (additionalCostPlan, bool) {
	plan := additionalCostPlan{player: playerID}
	for _, cost := range costs {
		amount := AdditionalCostAmount(cost)
		switch cost.Kind {
		case game.AdditionalCostUnknown:
			if cost.Text == "" {
				continue
			}
			return plan, false
		case game.AdditionalCostTap:
			continue
		case game.AdditionalCostSacrifice:
			chosen := preferredSacrificePermanents(s, playerID, cost, amount, plan.sacrifices, prefs)
			if len(chosen) != amount {
				return plan, false
			}
			plan.sacrifices = append(plan.sacrifices, chosen...)
			plan.paid = append(plan.paid, AdditionalCostText(cost))
		case game.AdditionalCostSacrificeSource:
			if amount != 1 || source == nil || s.EffectiveController(source) != playerID || !additionalCostMatchesPermanent(s, source, cost) {
				return plan, false
			}
			plan.sacrifices = append(plan.sacrifices, source)
			plan.paid = append(plan.paid, AdditionalCostText(cost))
		case game.AdditionalCostDiscard:
			chosen := preferredDiscardCards(s, playerID, cost, amount, plan.discards, prefs)
			if len(chosen) != amount {
				return plan, false
			}
			plan.discards = append(plan.discards, chosen...)
			plan.paid = append(plan.paid, AdditionalCostText(cost))
		case game.AdditionalCostPayLife:
			player, ok := s.Player(playerID)
			if !ok || player.Life < amount {
				return plan, false
			}
			plan.lifePaid += amount
			plan.paid = append(plan.paid, AdditionalCostText(cost))
		case game.AdditionalCostExile:
			chosen := preferredExileCards(s, playerID, cost, amount, plan.exiles, prefs)
			if len(chosen) != amount {
				return plan, false
			}
			plan.exiles = append(plan.exiles, chosen...)
			plan.paid = append(plan.paid, AdditionalCostText(cost))
		case game.AdditionalCostExileSource:
			if amount != 1 || sourceCardID == 0 || sourceZone == game.ZoneNone || !zoneContainsCard(s, playerID, sourceZone, sourceCardID) {
				return plan, false
			}
			card, ok := s.CardInstance(sourceCardID)
			if !ok || !additionalCostMatchesCard(s.CardFace(card, game.FaceFront), cost) {
				return plan, false
			}
			plan.exiles = append(plan.exiles, cardZoneSelection{cardID: sourceCardID, zone: sourceZone})
			plan.paid = append(plan.paid, AdditionalCostText(cost))
		default:
			return plan, false
		}
	}

	return plan, true
}

func chooseExileCards(s State, playerID game.PlayerID, cost game.AdditionalCost, amount int, alreadyChosen []cardZoneSelection) []cardZoneSelection {
	player, ok := s.Player(playerID)
	if !ok {
		return nil
	}
	zone := cost.Zone
	if zone == game.ZoneNone {
		zone = game.ZoneGraveyard
	}
	chosenIDs := make(map[id.ID]bool)
	for _, chosen := range alreadyChosen {
		chosenIDs[chosen.cardID] = true
	}
	var chosen []cardZoneSelection
	for _, cardID := range cardIDsInZone(player, zone) {
		if chosenIDs[cardID] {
			continue
		}
		card, ok := s.CardInstance(cardID)
		if !ok || !additionalCostMatchesCard(s.CardFace(card, game.FaceFront), cost) {
			continue
		}
		chosen = append(chosen, cardZoneSelection{cardID: cardID, zone: zone})
		if len(chosen) == amount {
			return chosen
		}
	}
	return chosen
}

func preferredExileCards(s State, playerID game.PlayerID, cost game.AdditionalCost, amount int, alreadyChosen []cardZoneSelection, prefs *Preferences) []cardZoneSelection {
	if prefs == nil || len(prefs.ExileChoices) == 0 {
		return chooseExileCards(s, playerID, cost, amount, alreadyChosen)
	}
	zone := cost.Zone
	if zone == game.ZoneNone {
		zone = game.ZoneGraveyard
	}
	chosenIDs := make(map[id.ID]bool)
	for _, chosen := range alreadyChosen {
		chosenIDs[chosen.cardID] = true
	}
	var chosen []cardZoneSelection
	var consumed int
	for _, cardID := range prefs.ExileChoices {
		card, ok := s.CardInstance(cardID)
		if !ok || !zoneContainsCard(s, playerID, zone, cardID) || chosenIDs[cardID] || !additionalCostMatchesCard(s.CardFace(card, game.FaceFront), cost) {
			return nil
		}
		chosen = append(chosen, cardZoneSelection{cardID: cardID, zone: zone})
		chosenIDs[cardID] = true
		consumed++
		if len(chosen) == amount {
			prefs.ExileChoices = prefs.ExileChoices[consumed:]
			return chosen
		}
	}
	return nil
}

func chooseSacrificePermanents(s State, playerID game.PlayerID, cost game.AdditionalCost, amount int, alreadyChosen []*game.Permanent) []*game.Permanent {
	chosenIDs := make(map[id.ID]bool)
	for _, permanent := range alreadyChosen {
		chosenIDs[permanent.ObjectID] = true
	}
	var chosen []*game.Permanent
	for _, permanent := range s.Battlefield() {
		if s.EffectiveController(permanent) != playerID || chosenIDs[permanent.ObjectID] {
			continue
		}
		if additionalCostMatchesPermanent(s, permanent, cost) {
			chosen = append(chosen, permanent)
			if len(chosen) == amount {
				return chosen
			}
		}
	}
	return chosen
}

func preferredSacrificePermanents(s State, playerID game.PlayerID, cost game.AdditionalCost, amount int, alreadyChosen []*game.Permanent, prefs *Preferences) []*game.Permanent {
	if prefs == nil || len(prefs.SacrificeChoices) == 0 {
		return chooseSacrificePermanents(s, playerID, cost, amount, alreadyChosen)
	}
	chosenIDs := make(map[id.ID]bool)
	for _, permanent := range alreadyChosen {
		chosenIDs[permanent.ObjectID] = true
	}
	var chosen []*game.Permanent
	var consumed int
	for _, permanentID := range prefs.SacrificeChoices {
		permanent, ok := s.PermanentByObjectID(permanentID)
		if !ok || s.EffectiveController(permanent) != playerID || chosenIDs[permanentID] || !additionalCostMatchesPermanent(s, permanent, cost) {
			return nil
		}
		chosen = append(chosen, permanent)
		chosenIDs[permanentID] = true
		consumed++
		if len(chosen) == amount {
			prefs.SacrificeChoices = prefs.SacrificeChoices[consumed:]
			return chosen
		}
	}
	return nil
}

func chooseDiscardCards(s State, playerID game.PlayerID, cost game.AdditionalCost, amount int, alreadyChosen []id.ID) []id.ID {
	player, ok := s.Player(playerID)
	if !ok {
		return nil
	}
	chosenIDs := make(map[id.ID]bool)
	for _, cardID := range alreadyChosen {
		chosenIDs[cardID] = true
	}
	var chosen []id.ID
	for _, cardID := range player.Hand.All() {
		if chosenIDs[cardID] {
			continue
		}
		card, ok := s.CardInstance(cardID)
		if !ok || !additionalCostMatchesCard(s.CardFace(card, game.FaceFront), cost) {
			continue
		}
		chosen = append(chosen, cardID)
		if len(chosen) == amount {
			return chosen
		}
	}
	return chosen
}

func preferredDiscardCards(s State, playerID game.PlayerID, cost game.AdditionalCost, amount int, alreadyChosen []id.ID, prefs *Preferences) []id.ID {
	if prefs == nil || len(prefs.DiscardChoices) == 0 {
		return chooseDiscardCards(s, playerID, cost, amount, alreadyChosen)
	}
	player, ok := s.Player(playerID)
	if !ok {
		return nil
	}
	chosenIDs := make(map[id.ID]bool)
	for _, cardID := range alreadyChosen {
		chosenIDs[cardID] = true
	}
	var chosen []id.ID
	var consumed int
	for _, cardID := range prefs.DiscardChoices {
		card, ok := s.CardInstance(cardID)
		if !ok || !player.Hand.Contains(cardID) || chosenIDs[cardID] || !additionalCostMatchesCard(s.CardFace(card, game.FaceFront), cost) {
			return nil
		}
		chosen = append(chosen, cardID)
		chosenIDs[cardID] = true
		consumed++
		if len(chosen) == amount {
			prefs.DiscardChoices = prefs.DiscardChoices[consumed:]
			return chosen
		}
	}
	return nil
}

func additionalCostMatchesPermanent(s State, permanent *game.Permanent, cost game.AdditionalCost) bool {
	if cost.MatchPermanentType && !s.PermanentHasType(permanent, cost.PermanentType) {
		return false
	}
	return true
}

func additionalCostMatchesCard(card *game.CardDef, cost game.AdditionalCost) bool {
	if card == nil {
		return false
	}
	if cost.MatchCardType && !card.HasType(cost.CardType) {
		return false
	}
	return true
}

// AdditionalCostAmount returns the effective amount for an additional cost.
func AdditionalCostAmount(cost game.AdditionalCost) int {
	if cost.Amount > 0 {
		return cost.Amount
	}
	return 1
}

// AdditionalCostText returns display text for an additional cost.
func AdditionalCostText(cost game.AdditionalCost) string {
	if cost.Text != "" {
		return cost.Text
	}
	switch cost.Kind {
	case game.AdditionalCostSacrifice:
		return "Sacrifice a permanent"
	case game.AdditionalCostSacrificeSource:
		return "Sacrifice this permanent"
	case game.AdditionalCostDiscard:
		return "Discard a card"
	case game.AdditionalCostPayLife:
		return "Pay life"
	case game.AdditionalCostExile:
		return "Exile a card"
	case game.AdditionalCostExileSource:
		return "Exile this card"
	case game.AdditionalCostReveal:
		return "Reveal a card"
	case game.AdditionalCostTap:
		return "{T}"
	default:
		return "Additional cost"
	}
}

func additionalCostPlanStillValid(s State, player *game.Player, plan additionalCostPlan) bool {
	for _, sacrifice := range plan.sacrifices {
		permanent, ok := s.PermanentByObjectID(sacrifice.ObjectID)
		if !ok || s.EffectiveController(permanent) != player.ID || permanent != sacrifice {
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
	if plan.lifePaid > 0 && player.Life < plan.lifePaid {
		return false
	}
	return true
}

func applyAdditionalCostPlan(s State, plan additionalCostPlan) bool {
	for _, sacrifice := range plan.sacrifices {
		if !s.MovePermanentToZone(sacrifice, game.ZoneGraveyard) {
			return false
		}
	}
	for _, cardID := range plan.discards {
		if !s.DiscardFromHand(plan.player, cardID) {
			return false
		}
	}
	for _, exile := range plan.exiles {
		if !s.MoveCard(plan.player, exile.cardID, exile.zone, game.ZoneExile) {
			return false
		}
	}
	if plan.lifePaid > 0 {
		player, ok := s.Player(plan.player)
		if !ok || player.Life < plan.lifePaid {
			return false
		}
		s.LoseLife(plan.player, plan.lifePaid)
	}
	return true
}

func zoneContainsCard(s State, playerID game.PlayerID, zone game.ZoneType, cardID id.ID) bool {
	player, ok := s.Player(playerID)
	if !ok {
		return false
	}
	return slices.Contains(cardIDsInZone(player, zone), cardID)
}

func cardIDsInZone(player *game.Player, zone game.ZoneType) []id.ID {
	switch zone {
	case game.ZoneLibrary:
		return player.Library.All()
	case game.ZoneHand:
		return player.Hand.All()
	case game.ZoneGraveyard:
		return player.Graveyard.All()
	case game.ZoneExile:
		return player.Exile.All()
	case game.ZoneCommand:
		return player.CommandZone.All()
	default:
		return nil
	}
}
