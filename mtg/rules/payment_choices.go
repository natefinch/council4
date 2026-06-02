package rules

import (
	"fmt"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/rules/payment"
)

func (e *Engine) paymentPreferencesForCost(g *game.Game, playerID game.PlayerID, manaCost *cost.Mana, additionalCosts []game.AdditionalCost, agents [game.NumPlayers]PlayerAgent, log *TurnLog) *payment.Preferences {
	prefs := &payment.Preferences{}
	prefs.PhyrexianLifeChoices = e.phyrexianPaymentChoices(g, playerID, manaCost, agents, log)
	for _, additionalCost := range additionalCosts {
		amount := payment.AdditionalCostAmount(additionalCost)
		switch additionalCost.Kind {
		case game.AdditionalCostSacrifice:
			prefs.SacrificeChoices = append(prefs.SacrificeChoices, e.additionalCostPermanentChoices(g, playerID, additionalCost, amount, agents, log)...)
		case game.AdditionalCostDiscard:
			prefs.DiscardChoices = append(prefs.DiscardChoices, e.additionalCostCardChoices(g, playerID, additionalCost, amount, agents, log)...)
		default:
		}
	}
	return prefs
}

func (e *Engine) paymentPreferencesForSpell(g *game.Game, playerID game.PlayerID, card *game.CardDef, xValue int, agents [game.NumPlayers]PlayerAgent, log *TurnLog) *payment.Preferences {
	return e.paymentPreferencesForSpellFromZone(g, playerID, 0, game.ZoneHand, card, xValue, agents, log)
}

func (e *Engine) paymentPreferencesForSpellFromZone(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone game.ZoneType, card *game.CardDef, xValue int, agents [game.NumPlayers]PlayerAgent, log *TurnLog) *payment.Preferences {
	option := e.chooseSpellCostOptionFromZone(g, playerID, cardID, sourceZone, card, xValue, agents, log)
	prefs := e.paymentPreferencesForCost(g, playerID, option.ManaCost, option.AdditionalCosts, agents, log)
	prefs.AlternativeIndex = option.Index
	return prefs
}

func (e *Engine) chooseSpellCostOptionFromZone(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone game.ZoneType, card *game.CardDef, xValue int, agents [game.NumPlayers]PlayerAgent, log *TurnLog) payment.SpellOptionSummary {
	options := paymentOrch.planner(g).PayableSpellOptions(payment.SpellRequest{PlayerID: playerID, CardID: cardID, SourceZone: sourceZone, Card: card, XValue: xValue})
	if len(options) == 0 {
		return payment.SpellOptionSummary{}
	}
	if len(options) == 1 {
		return options[0]
	}
	choiceOptions := make([]game.ChoiceOption, 0, len(options))
	for _, option := range options {
		choiceOptions = append(choiceOptions, game.ChoiceOption{Index: option.Index, Label: option.Label})
	}
	request := game.ChoiceRequest{
		Kind:             game.ChoicePayment,
		Player:           playerID,
		Prompt:           "Choose spell cost",
		Options:          choiceOptions,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{options[0].Index},
	}
	selected := e.chooseChoice(g, agents, request, log)
	if len(selected) == 1 {
		for _, option := range options {
			if option.Index == selected[0] {
				return option
			}
		}
	}
	return options[0]
}

func (e *Engine) phyrexianPaymentChoices(g *game.Game, playerID game.PlayerID, manaCost *cost.Mana, agents [game.NumPlayers]PlayerAgent, log *TurnLog) []bool {
	if manaCost == nil {
		return nil
	}
	var choices []bool
	availableLife := 0
	if player, ok := playerByID(g, playerID); ok {
		availableLife = player.Life
	}
	for _, symbol := range *manaCost {
		if symbol.Kind != cost.PhyrexianSymbol {
			continue
		}
		options := []game.ChoiceOption{{Index: 0, Label: fmt.Sprintf("Pay %s mana", symbol.Color)}}
		if availableLife >= 2 {
			options = append(options, game.ChoiceOption{Index: 1, Label: "Pay 2 life"})
		}
		request := game.ChoiceRequest{
			Kind:             game.ChoicePayment,
			Player:           playerID,
			Prompt:           fmt.Sprintf("Pay %s", symbol),
			Options:          options,
			MinChoices:       1,
			MaxChoices:       1,
			DefaultSelection: []int{0},
		}
		selected := e.chooseChoice(g, agents, request, log)
		payLife := len(selected) == 1 && selected[0] == 1
		if payLife {
			availableLife -= 2
		}
		choices = append(choices, payLife)
	}
	return choices
}

func (e *Engine) additionalCostPermanentChoices(g *game.Game, playerID game.PlayerID, addCost game.AdditionalCost, amount int, agents [game.NumPlayers]PlayerAgent, log *TurnLog) []id.ID {
	candidates := candidateSacrificePermanents(g, playerID, addCost)
	if len(candidates) <= amount {
		return paymentPermanentIDs(candidates)
	}
	options := make([]game.ChoiceOption, 0, len(candidates))
	for i, permanent := range candidates {
		options = append(options, game.ChoiceOption{Index: i, Label: permanentChoiceLabel(g, permanent)})
	}
	request := game.ChoiceRequest{
		Kind:             game.ChoicePayment,
		Player:           playerID,
		Prompt:           payment.AdditionalCostText(addCost),
		Options:          options,
		MinChoices:       amount,
		MaxChoices:       amount,
		DefaultSelection: firstChoiceIndices(amount),
	}
	selected := e.chooseChoice(g, agents, request, log)
	return selectedPaymentPermanentIDs(candidates, selected)
}

func (e *Engine) additionalCostCardChoices(g *game.Game, playerID game.PlayerID, addCost game.AdditionalCost, amount int, agents [game.NumPlayers]PlayerAgent, log *TurnLog) []id.ID {
	candidates := candidateDiscardCards(g, playerID, addCost)
	if len(candidates) <= amount {
		return candidates
	}
	options := make([]game.ChoiceOption, 0, len(candidates))
	for i, cardID := range candidates {
		options = append(options, game.ChoiceOption{Index: i, Label: cardChoiceLabel(g, cardID)})
	}
	request := game.ChoiceRequest{
		Kind:             game.ChoicePayment,
		Player:           playerID,
		Prompt:           payment.AdditionalCostText(addCost),
		Options:          options,
		MinChoices:       amount,
		MaxChoices:       amount,
		DefaultSelection: firstChoiceIndices(amount),
	}
	selected := e.chooseChoice(g, agents, request, log)
	return selectedCardIDs(candidates, selected)
}

func firstChoiceIndices(amount int) []int {
	selected := make([]int, 0, amount)
	for i := range amount {
		selected = append(selected, i)
	}
	return selected
}

func candidateSacrificePermanents(g *game.Game, playerID game.PlayerID, addCost game.AdditionalCost) []*game.Permanent {
	var candidates []*game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Controller == playerID && localAdditionalCostMatchesPermanent(g, permanent, addCost) {
			candidates = append(candidates, permanent)
		}
	}
	return candidates
}

func localAdditionalCostMatchesPermanent(g *game.Game, permanent *game.Permanent, addCost game.AdditionalCost) bool {
	if addCost.MatchPermanentType && !permanentHasType(g, permanent, addCost.PermanentType) {
		return false
	}
	return true
}

func candidateDiscardCards(g *game.Game, playerID game.PlayerID, addCost game.AdditionalCost) []id.ID {
	player, ok := playerByID(g, playerID)
	if !ok {
		return nil
	}
	var candidates []id.ID
	for _, cardID := range player.Hand.All() {
		card, ok := g.GetCardInstance(cardID)
		if ok && localAdditionalCostMatchesCard(cardFaceOrDefault(card, game.FaceFront), addCost) {
			candidates = append(candidates, cardID)
		}
	}
	return candidates
}

func localAdditionalCostMatchesCard(face *game.CardDef, addCost game.AdditionalCost) bool {
	if face == nil {
		return false
	}
	if addCost.MatchCardType && !face.HasType(addCost.CardType) {
		return false
	}
	return true
}

func paymentPermanentIDs(permanents []*game.Permanent) []id.ID {
	ids := make([]id.ID, 0, len(permanents))
	for _, permanent := range permanents {
		ids = append(ids, permanent.ObjectID)
	}
	return ids
}

func selectedPaymentPermanentIDs(candidates []*game.Permanent, selected []int) []id.ID {
	var ids []id.ID
	for _, index := range selected {
		if index >= 0 && index < len(candidates) {
			ids = append(ids, candidates[index].ObjectID)
		}
	}
	return ids
}

func selectedCardIDs(candidates []id.ID, selected []int) []id.ID {
	var ids []id.ID
	for _, index := range selected {
		if index >= 0 && index < len(candidates) {
			ids = append(ids, candidates[index])
		}
	}
	return ids
}

func permanentChoiceLabel(g *game.Game, permanent *game.Permanent) string {
	card, ok := permanentCardDef(g, permanent)
	if !ok {
		return fmt.Sprintf("Permanent %d", permanent.ObjectID)
	}
	return card.Name
}

func cardChoiceLabel(g *game.Game, cardID id.ID) string {
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return fmt.Sprintf("Card %d", cardID)
	}
	return cardFaceOrDefault(card, game.FaceFront).Name
}
