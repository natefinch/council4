package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

type additionalCostPlan struct {
	player     game.PlayerID
	paid       []string
	sacrifices []*game.Permanent
	discards   []id.ID
	lifePaid   int
}

func buildAdditionalCostPlanForCosts(g *game.Game, playerID game.PlayerID, costs []game.AdditionalCost, prefs *paymentPreferences) (additionalCostPlan, bool) {
	plan := additionalCostPlan{player: playerID}
	for _, cost := range costs {
		amount := additionalCostAmount(cost)
		switch cost.Kind {
		case game.AdditionalCostUnknown:
			if cost.Text == "" {
				continue
			}
			return plan, false
		case game.AdditionalCostTap:
			continue
		case game.AdditionalCostSacrifice:
			chosen := preferredSacrificePermanents(g, playerID, cost, amount, plan.sacrifices, prefs)
			if len(chosen) != amount {
				return plan, false
			}
			plan.sacrifices = append(plan.sacrifices, chosen...)
			plan.paid = append(plan.paid, additionalCostText(cost))
		case game.AdditionalCostDiscard:
			chosen := preferredDiscardCards(g, playerID, cost, amount, plan.discards, prefs)
			if len(chosen) != amount {
				return plan, false
			}
			plan.discards = append(plan.discards, chosen...)
			plan.paid = append(plan.paid, additionalCostText(cost))
		case game.AdditionalCostPayLife:
			player, ok := playerByID(g, playerID)
			if !ok || player.Life < amount {
				return plan, false
			}
			plan.lifePaid += amount
			plan.paid = append(plan.paid, additionalCostText(cost))
		default:
			return plan, false
		}
	}
	return plan, true
}

func chooseSacrificePermanent(g *game.Game, playerID game.PlayerID, matches func(*game.CardDef) bool) (*game.Permanent, bool) {
	if matches == nil {
		return nil, false
	}
	for _, permanent := range g.Battlefield {
		if effectiveController(g, permanent) != playerID {
			continue
		}
		def, ok := permanentCardDef(g, permanent)
		if ok && matches(def) {
			return permanent, true
		}
	}
	return nil, false
}

func chooseSacrificePermanents(g *game.Game, playerID game.PlayerID, cost game.AdditionalCost, amount int, alreadyChosen []*game.Permanent) []*game.Permanent {
	chosenIDs := make(map[id.ID]bool)
	for _, permanent := range alreadyChosen {
		chosenIDs[permanent.ObjectID] = true
	}
	var chosen []*game.Permanent
	for _, permanent := range g.Battlefield {
		if effectiveController(g, permanent) != playerID || chosenIDs[permanent.ObjectID] {
			continue
		}
		if additionalCostMatchesPermanent(g, permanent, cost) {
			chosen = append(chosen, permanent)
			if len(chosen) == amount {
				return chosen
			}
		}
	}
	return chosen
}

func preferredSacrificePermanents(g *game.Game, playerID game.PlayerID, cost game.AdditionalCost, amount int, alreadyChosen []*game.Permanent, prefs *paymentPreferences) []*game.Permanent {
	if prefs == nil || len(prefs.sacrificeChoices) == 0 {
		return chooseSacrificePermanents(g, playerID, cost, amount, alreadyChosen)
	}
	chosenIDs := make(map[id.ID]bool)
	for _, permanent := range alreadyChosen {
		chosenIDs[permanent.ObjectID] = true
	}
	var chosen []*game.Permanent
	var consumed int
	for _, permanentID := range prefs.sacrificeChoices {
		permanent, ok := permanentByObjectID(g, permanentID)
		if !ok || effectiveController(g, permanent) != playerID || chosenIDs[permanentID] || !additionalCostMatchesPermanent(g, permanent, cost) {
			return nil
		}
		chosen = append(chosen, permanent)
		chosenIDs[permanentID] = true
		consumed++
		if len(chosen) == amount {
			prefs.sacrificeChoices = prefs.sacrificeChoices[consumed:]
			return chosen
		}
	}
	return nil
}

func chooseDiscardCards(g *game.Game, playerID game.PlayerID, cost game.AdditionalCost, amount int, alreadyChosen []id.ID) []id.ID {
	player, ok := playerByID(g, playerID)
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
		card, ok := g.GetCardInstance(cardID)
		if !ok || !additionalCostMatchesCard(cardFaceOrDefault(card, game.FaceFront), cost) {
			continue
		}
		chosen = append(chosen, cardID)
		if len(chosen) == amount {
			return chosen
		}
	}
	return chosen
}

func preferredDiscardCards(g *game.Game, playerID game.PlayerID, cost game.AdditionalCost, amount int, alreadyChosen []id.ID, prefs *paymentPreferences) []id.ID {
	if prefs == nil || len(prefs.discardChoices) == 0 {
		return chooseDiscardCards(g, playerID, cost, amount, alreadyChosen)
	}
	player, ok := playerByID(g, playerID)
	if !ok {
		return nil
	}
	chosenIDs := make(map[id.ID]bool)
	for _, cardID := range alreadyChosen {
		chosenIDs[cardID] = true
	}
	var chosen []id.ID
	var consumed int
	for _, cardID := range prefs.discardChoices {
		card, ok := g.GetCardInstance(cardID)
		if !ok || !player.Hand.Contains(cardID) || chosenIDs[cardID] || !additionalCostMatchesCard(cardFaceOrDefault(card, game.FaceFront), cost) {
			return nil
		}
		chosen = append(chosen, cardID)
		chosenIDs[cardID] = true
		consumed++
		if len(chosen) == amount {
			prefs.discardChoices = prefs.discardChoices[consumed:]
			return chosen
		}
	}
	return nil
}

func additionalCostMatchesPermanent(g *game.Game, permanent *game.Permanent, cost game.AdditionalCost) bool {
	if cost.MatchPermanentType && !permanentHasType(g, permanent, cost.PermanentType) {
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

func additionalCostCardMatcher(cost game.AdditionalCost) func(*game.CardDef) bool {
	return func(card *game.CardDef) bool {
		return additionalCostMatchesCard(card, game.AdditionalCost{
			MatchCardType: cost.MatchPermanentType,
			CardType:      cost.PermanentType,
		})
	}
}

func additionalCostAmount(cost game.AdditionalCost) int {
	if cost.Amount > 0 {
		return cost.Amount
	}
	return 1
}

func additionalCostText(cost game.AdditionalCost) string {
	if cost.Text != "" {
		return cost.Text
	}
	switch cost.Kind {
	case game.AdditionalCostSacrifice:
		return "Sacrifice a permanent"
	case game.AdditionalCostDiscard:
		return "Discard a card"
	case game.AdditionalCostPayLife:
		return "Pay life"
	case game.AdditionalCostExile:
		return "Exile a card"
	case game.AdditionalCostReveal:
		return "Reveal a card"
	case game.AdditionalCostTap:
		return "{T}"
	default:
		return "Additional cost"
	}
}

func additionalCostPlanStillValid(g *game.Game, player *game.Player, plan additionalCostPlan) bool {
	for _, sacrifice := range plan.sacrifices {
		permanent, ok := permanentByObjectID(g, sacrifice.ObjectID)
		if !ok || effectiveController(g, permanent) != player.ID || permanent != sacrifice {
			return false
		}
	}
	for _, cardID := range plan.discards {
		if !player.Hand.Contains(cardID) {
			return false
		}
	}
	if plan.lifePaid > 0 && player.Life < plan.lifePaid {
		return false
	}
	return true
}

func applyAdditionalCostPlan(g *game.Game, plan additionalCostPlan) bool {
	for _, sacrifice := range plan.sacrifices {
		if !movePermanentToZone(g, sacrifice, game.ZoneGraveyard) {
			return false
		}
	}
	for _, cardID := range plan.discards {
		card, ok := g.GetCardInstance(cardID)
		if !ok || !discardCardFromHand(g, card.Owner, cardID) {
			return false
		}
	}
	if plan.lifePaid > 0 {
		player, ok := playerByID(g, plan.player)
		if !ok || player.Life < plan.lifePaid {
			return false
		}
		loseLife(g, plan.player, plan.lifePaid)
	}
	return true
}
