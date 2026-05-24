package rules

import (
	"fmt"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
)

func (e *Engine) paymentPreferencesForCost(g *game.Game, playerID game.PlayerID, cost *mana.Cost, additionalCosts []game.AdditionalCost, agents [game.NumPlayers]PlayerAgent, log *TurnLog) *paymentPreferences {
	prefs := &paymentPreferences{}
	prefs.phyrexianLifeChoices = e.phyrexianPaymentChoices(g, playerID, cost, agents, log)
	for _, additionalCost := range additionalCosts {
		amount := additionalCostAmount(additionalCost)
		switch additionalCost.Kind {
		case game.AdditionalCostSacrifice:
			prefs.sacrificeChoices = append(prefs.sacrificeChoices, e.additionalCostPermanentChoices(g, playerID, additionalCost, amount, agents, log)...)
		case game.AdditionalCostDiscard:
			prefs.discardChoices = append(prefs.discardChoices, e.additionalCostCardChoices(g, playerID, additionalCost, amount, agents, log)...)
		}
	}
	return prefs
}

func (e *Engine) paymentPreferencesForSpell(g *game.Game, playerID game.PlayerID, card *game.CardDef, xValue int, agents [game.NumPlayers]PlayerAgent, log *TurnLog) *paymentPreferences {
	return e.paymentPreferencesForSpellFromZone(g, playerID, 0, game.ZoneHand, card, xValue, agents, log)
}

func (e *Engine) paymentPreferencesForSpellFromZone(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone game.ZoneType, card *game.CardDef, xValue int, agents [game.NumPlayers]PlayerAgent, log *TurnLog) *paymentPreferences {
	option := e.chooseSpellCostOptionFromZone(g, playerID, cardID, sourceZone, card, xValue, agents, log)
	prefs := e.paymentPreferencesForCost(g, playerID, option.manaCost, option.additionalCosts, agents, log)
	prefs.alternativeIndex = option.index
	return prefs
}

func (e *Engine) chooseSpellCostOption(g *game.Game, playerID game.PlayerID, card *game.CardDef, xValue int, agents [game.NumPlayers]PlayerAgent, log *TurnLog) spellCostOption {
	return e.chooseSpellCostOptionFromZone(g, playerID, 0, game.ZoneHand, card, xValue, agents, log)
}

func (e *Engine) chooseSpellCostOptionFromZone(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone game.ZoneType, card *game.CardDef, xValue int, agents [game.NumPlayers]PlayerAgent, log *TurnLog) spellCostOption {
	options := payableSpellCostOptionsFromZone(g, playerID, cardID, sourceZone, card, xValue)
	if len(options) == 0 {
		return spellCostOption{}
	}
	if len(options) == 1 {
		return options[0]
	}
	choiceOptions := make([]game.ChoiceOption, 0, len(options))
	for _, option := range options {
		choiceOptions = append(choiceOptions, game.ChoiceOption{Index: option.index, Label: option.label})
	}
	request := game.ChoiceRequest{
		Kind:             game.ChoicePayment,
		Player:           playerID,
		Prompt:           "Choose spell cost",
		Options:          choiceOptions,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{options[0].index},
	}
	selected := e.chooseChoice(g, agents, request, log)
	if len(selected) == 1 {
		for _, option := range options {
			if option.index == selected[0] {
				return option
			}
		}
	}
	return options[0]
}

func payableSpellCostOptions(g *game.Game, playerID game.PlayerID, card *game.CardDef, xValue int) []spellCostOption {
	return payableSpellCostOptionsFromZone(g, playerID, 0, game.ZoneHand, card, xValue)
}

func payableSpellCostOptionsFromZone(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone game.ZoneType, card *game.CardDef, xValue int) []spellCostOption {
	var payable []spellCostOption
	for _, option := range spellCostOptions(card) {
		if _, ok := buildSpellCostPlanForOption(g, playerID, cardID, sourceZone, option, xValue, nil); ok {
			payable = append(payable, option)
		}
	}
	return payable
}

func (e *Engine) phyrexianPaymentChoices(g *game.Game, playerID game.PlayerID, cost *mana.Cost, agents [game.NumPlayers]PlayerAgent, log *TurnLog) []bool {
	if cost == nil {
		return nil
	}
	var choices []bool
	availableLife := 0
	if player := playerByID(g, playerID); player != nil {
		availableLife = player.Life
	}
	for _, symbol := range *cost {
		if symbol.Kind != mana.PhyrexianSymbol {
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

func (e *Engine) additionalCostPermanentChoices(g *game.Game, playerID game.PlayerID, cost game.AdditionalCost, amount int, agents [game.NumPlayers]PlayerAgent, log *TurnLog) []id.ID {
	candidates := candidateSacrificePermanents(g, playerID, cost)
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
		Prompt:           additionalCostText(cost),
		Options:          options,
		MinChoices:       amount,
		MaxChoices:       amount,
		DefaultSelection: firstChoiceIndices(amount),
	}
	selected := e.chooseChoice(g, agents, request, log)
	return selectedPaymentPermanentIDs(candidates, selected)
}

func (e *Engine) additionalCostCardChoices(g *game.Game, playerID game.PlayerID, cost game.AdditionalCost, amount int, agents [game.NumPlayers]PlayerAgent, log *TurnLog) []id.ID {
	candidates := candidateDiscardCards(g, playerID, cost)
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
		Prompt:           additionalCostText(cost),
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
	for i := 0; i < amount; i++ {
		selected = append(selected, i)
	}
	return selected
}

func candidateSacrificePermanents(g *game.Game, playerID game.PlayerID, cost game.AdditionalCost) []*game.Permanent {
	if g == nil {
		return nil
	}
	var candidates []*game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent != nil && permanent.Controller == playerID && additionalCostMatchesPermanent(g, permanent, cost) {
			candidates = append(candidates, permanent)
		}
	}
	return candidates
}

func candidateDiscardCards(g *game.Game, playerID game.PlayerID, cost game.AdditionalCost) []id.ID {
	player := playerByID(g, playerID)
	if player == nil {
		return nil
	}
	var candidates []id.ID
	for _, cardID := range player.Hand.All() {
		card := g.GetCardInstance(cardID)
		if card != nil && additionalCostMatchesCard(card.Def, cost) {
			candidates = append(candidates, cardID)
		}
	}
	return candidates
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
	card := permanentCardDef(g, permanent)
	if card == nil {
		return fmt.Sprintf("Permanent %d", permanent.ObjectID)
	}
	return card.Name
}

func cardChoiceLabel(g *game.Game, cardID id.ID) string {
	card := g.GetCardInstance(cardID)
	if card == nil || card.Def == nil {
		return fmt.Sprintf("Card %d", cardID)
	}
	return card.Def.Name
}
