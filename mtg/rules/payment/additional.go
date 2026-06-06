package payment

import (
	"slices"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
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
	zone   zone.Type
}

func buildAdditionalCostPlanForCosts(s State, playerID game.PlayerID, costs []cost.Additional, prefs *Preferences, source *game.Permanent, sourceCardID id.ID, sourceZone zone.Type) (additionalCostPlan, bool) {
	plan := additionalCostPlan{player: playerID}
	for _, additional := range costs {
		amount := AdditionalCostAmount(additional)
		switch additional.Kind {
		case cost.AdditionalUnknown:
			if additional.Text == "" {
				continue
			}
			return plan, false
		case cost.AdditionalTap:
			continue
		case cost.AdditionalSacrifice:
			chosen := preferredSacrificePermanents(s, playerID, additional, amount, plan.sacrifices, prefs)
			if len(chosen) != amount {
				return plan, false
			}
			plan.sacrifices = append(plan.sacrifices, chosen...)
			plan.paid = append(plan.paid, AdditionalCostText(additional))
		case cost.AdditionalSacrificeSource:
			if amount != 1 || source == nil || s.EffectiveController(source) != playerID || !additionalCostMatchesPermanent(s, source, additional) {
				return plan, false
			}
			plan.sacrifices = append(plan.sacrifices, source)
			plan.paid = append(plan.paid, AdditionalCostText(additional))
		case cost.AdditionalDiscard:
			chosen := preferredDiscardCards(s, playerID, additional, amount, plan.discards, prefs)
			if len(chosen) != amount {
				return plan, false
			}
			plan.discards = append(plan.discards, chosen...)
			plan.paid = append(plan.paid, AdditionalCostText(additional))
		case cost.AdditionalPayLife:
			player, ok := s.Player(playerID)
			if !ok || player.Life < amount {
				return plan, false
			}
			plan.lifePaid += amount
			plan.paid = append(plan.paid, AdditionalCostText(additional))
		case cost.AdditionalExile:
			chosen := preferredExileCards(s, playerID, additional, amount, plan.exiles, prefs)
			if len(chosen) != amount {
				return plan, false
			}
			plan.exiles = append(plan.exiles, chosen...)
			plan.paid = append(plan.paid, AdditionalCostText(additional))
		case cost.AdditionalExileSource:
			if amount != 1 || sourceCardID == 0 || sourceZone == zone.None || !zoneContainsCard(s, playerID, sourceZone, sourceCardID) {
				return plan, false
			}
			card, ok := s.CardInstance(sourceCardID)
			if !ok || !additionalCostMatchesCard(s.CardFace(card, game.FaceFront), additional) {
				return plan, false
			}
			plan.exiles = append(plan.exiles, cardZoneSelection{cardID: sourceCardID, zone: sourceZone})
			plan.paid = append(plan.paid, AdditionalCostText(additional))
		default:
			return plan, false
		}
	}

	return plan, true
}

func chooseExileCards(s State, playerID game.PlayerID, additional cost.Additional, amount int, alreadyChosen []cardZoneSelection) []cardZoneSelection {
	player, ok := s.Player(playerID)
	if !ok {
		return nil
	}
	sourceZone := additionalCostSourceZone(additional.Source)
	chosenIDs := make(map[id.ID]bool)
	for _, chosen := range alreadyChosen {
		chosenIDs[chosen.cardID] = true
	}
	var chosen []cardZoneSelection
	for _, cardID := range cardIDsInZone(player, sourceZone) {
		if chosenIDs[cardID] {
			continue
		}
		card, ok := s.CardInstance(cardID)
		if !ok || !additionalCostMatchesCard(s.CardFace(card, game.FaceFront), additional) {
			continue
		}
		chosen = append(chosen, cardZoneSelection{cardID: cardID, zone: sourceZone})
		if len(chosen) == amount {
			return chosen
		}
	}
	return chosen
}

func preferredExileCards(s State, playerID game.PlayerID, additional cost.Additional, amount int, alreadyChosen []cardZoneSelection, prefs *Preferences) []cardZoneSelection {
	if prefs == nil || len(prefs.ExileChoices) == 0 {
		return chooseExileCards(s, playerID, additional, amount, alreadyChosen)
	}
	sourceZone := additionalCostSourceZone(additional.Source)
	chosenIDs := make(map[id.ID]bool)
	for _, chosen := range alreadyChosen {
		chosenIDs[chosen.cardID] = true
	}
	var chosen []cardZoneSelection
	var consumed int
	for _, cardID := range prefs.ExileChoices {
		card, ok := s.CardInstance(cardID)
		if !ok || !zoneContainsCard(s, playerID, sourceZone, cardID) || chosenIDs[cardID] || !additionalCostMatchesCard(s.CardFace(card, game.FaceFront), additional) {
			return nil
		}

		chosen = append(chosen, cardZoneSelection{cardID: cardID, zone: sourceZone})
		chosenIDs[cardID] = true
		consumed++
		if len(chosen) == amount {
			prefs.ExileChoices = prefs.ExileChoices[consumed:]
			return chosen
		}
	}
	return nil
}

func additionalCostSourceZone(source zone.Type) zone.Type {
	if source == zone.None {
		return zone.Graveyard
	}
	return source
}

func chooseSacrificePermanents(s State, playerID game.PlayerID, additional cost.Additional, amount int, alreadyChosen []*game.Permanent) []*game.Permanent {
	chosenIDs := make(map[id.ID]bool)
	for _, permanent := range alreadyChosen {
		chosenIDs[permanent.ObjectID] = true
	}
	var chosen []*game.Permanent
	for _, permanent := range s.Battlefield() {
		if s.EffectiveController(permanent) != playerID || chosenIDs[permanent.ObjectID] {
			continue
		}
		if additionalCostMatchesPermanent(s, permanent, additional) {
			chosen = append(chosen, permanent)
			if len(chosen) == amount {
				return chosen
			}
		}
	}
	return chosen
}

func preferredSacrificePermanents(s State, playerID game.PlayerID, additional cost.Additional, amount int, alreadyChosen []*game.Permanent, prefs *Preferences) []*game.Permanent {
	if prefs == nil || len(prefs.SacrificeChoices) == 0 {
		return chooseSacrificePermanents(s, playerID, additional, amount, alreadyChosen)
	}
	chosenIDs := make(map[id.ID]bool)
	for _, permanent := range alreadyChosen {
		chosenIDs[permanent.ObjectID] = true
	}
	var chosen []*game.Permanent
	var consumed int
	for _, permanentID := range prefs.SacrificeChoices {
		permanent, ok := s.PermanentByObjectID(permanentID)
		if !ok || s.EffectiveController(permanent) != playerID || chosenIDs[permanentID] || !additionalCostMatchesPermanent(s, permanent, additional) {
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

func chooseDiscardCards(s State, playerID game.PlayerID, additional cost.Additional, amount int, alreadyChosen []id.ID) []id.ID {
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
		if !ok || !additionalCostMatchesCard(s.CardFace(card, game.FaceFront), additional) {
			continue
		}
		chosen = append(chosen, cardID)
		if len(chosen) == amount {
			return chosen
		}
	}
	return chosen
}

func preferredDiscardCards(s State, playerID game.PlayerID, additional cost.Additional, amount int, alreadyChosen []id.ID, prefs *Preferences) []id.ID {
	if prefs == nil || len(prefs.DiscardChoices) == 0 {
		return chooseDiscardCards(s, playerID, additional, amount, alreadyChosen)
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
		if !ok || !player.Hand.Contains(cardID) || chosenIDs[cardID] || !additionalCostMatchesCard(s.CardFace(card, game.FaceFront), additional) {
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

func additionalCostMatchesPermanent(s State, permanent *game.Permanent, additional cost.Additional) bool {
	if additional.MatchPermanentType && !s.PermanentHasType(permanent, additional.PermanentType) {
		return false
	}
	return true
}

func additionalCostMatchesCard(card *game.CardDef, additional cost.Additional) bool {
	if card == nil {
		return false
	}
	if additional.MatchCardType && !card.HasType(additional.CardType) {
		return false
	}
	return true
}

// AdditionalCostAmount returns the effective amount for an additional cost.
func AdditionalCostAmount(additional cost.Additional) int {
	if additional.Amount > 0 {
		return additional.Amount
	}
	return 1
}

// AdditionalCostText returns display text for an additional cost.
func AdditionalCostText(additional cost.Additional) string {
	if additional.Text != "" {
		return additional.Text
	}
	switch additional.Kind {
	case cost.AdditionalSacrifice:
		return "Sacrifice a permanent"
	case cost.AdditionalSacrificeSource:
		return "Sacrifice this permanent"
	case cost.AdditionalDiscard:
		return "Discard a card"
	case cost.AdditionalPayLife:
		return "Pay life"
	case cost.AdditionalExile:
		return "Exile a card"
	case cost.AdditionalExileSource:
		return "Exile this card"
	case cost.AdditionalReveal:
		return "Reveal a card"
	case cost.AdditionalTap:
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
		if !s.MovePermanentToZone(sacrifice, zone.Graveyard) {
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
	if plan.lifePaid > 0 {
		player, ok := s.Player(plan.player)
		if !ok || player.Life < plan.lifePaid {
			return false
		}
		s.LoseLife(plan.player, plan.lifePaid)
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
