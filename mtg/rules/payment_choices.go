package rules

import (
	"fmt"
	"maps"
	"slices"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/rules/payment"
)

func (e *Engine) paymentPreferencesForCost(g *game.Game, playerID game.PlayerID, manaCost *cost.Mana, additionalCosts []cost.Additional, xValue int, agents [game.NumPlayers]PlayerAgent, log *TurnLog, tapExclusions ...id.ID) *payment.Preferences {
	return e.paymentPreferencesForCostFromSource(g, playerID, manaCost, additionalCosts, xValue, 0, zone.None, agents, log, tapExclusions...)
}

func (e *Engine) paymentPreferencesForCostFromSource(g *game.Game, playerID game.PlayerID, manaCost *cost.Mana, additionalCosts []cost.Additional, xValue int, sourceCardID id.ID, sourceZone zone.Type, agents [game.NumPlayers]PlayerAgent, log *TurnLog, tapExclusions ...id.ID) *payment.Preferences {
	prefs := &payment.Preferences{}
	prefs.PhyrexianLifeChoices = e.phyrexianPaymentChoices(g, playerID, manaCost, agents, log)
	reservedGraveyardCards := map[id.ID]bool{}
	for i, additionalCost := range additionalCosts {
		amount := payment.AdditionalCostAmountFor(additionalCost, xValue)
		switch additionalCost.Kind {
		case cost.AdditionalSacrifice:
			prefs.SacrificeChoices = append(prefs.SacrificeChoices, e.additionalCostPermanentChoices(g, playerID, additionalCost, amount, agents, log)...)
		case cost.AdditionalTapPermanents:
			prefs.TapChoices = append(prefs.TapChoices, e.additionalCostPermanentChoices(g, playerID, additionalCost, amount, agents, log, tapExclusions...)...)
		case cost.AdditionalReturnToHand:
			prefs.ReturnChoices = append(prefs.ReturnChoices, e.additionalCostPermanentChoices(g, playerID, additionalCost, amount, agents, log)...)
		case cost.AdditionalDiscard:
			prefs.DiscardChoices = append(prefs.DiscardChoices, e.additionalCostCardChoices(g, playerID, additionalCost, amount, nil, 0, 0, zone.None, agents, log)...)
		case cost.AdditionalExile:
			choices := e.additionalCostCardChoices(g, playerID, additionalCost, amount, additionalCosts[i+1:], xValue, sourceCardID, sourceZone, agents, log, reservedCardIDs(reservedGraveyardCards)...)
			prefs.ExileChoices = append(prefs.ExileChoices, choices...)
			reserveIfGraveyard(additionalCost, choices, reservedGraveyardCards)
		case cost.AdditionalReveal:
			prefs.RevealChoices = append(prefs.RevealChoices, e.additionalCostCardChoices(g, playerID, additionalCost, amount, nil, 0, 0, zone.None, agents, log)...)
		case cost.AdditionalExileSource:
			if sourceCardID != 0 && sourceZone == zone.Graveyard {
				reservedGraveyardCards[sourceCardID] = true
			}
		case cost.AdditionalCollectEvidence:
			choices := e.collectEvidenceChoices(g, playerID, additionalCost, amount, additionalCosts[i+1:], xValue, sourceCardID, sourceZone, agents, log, reservedCardIDs(reservedGraveyardCards)...)
			prefs.EvidenceChoices = append(prefs.EvidenceChoices, choices...)
			reserveIfGraveyard(additionalCost, choices, reservedGraveyardCards)
		default:
		}
	}
	return prefs
}

func (e *Engine) paymentPreferencesForSpell(g *game.Game, playerID game.PlayerID, card *game.CardDef, xValue int, agents [game.NumPlayers]PlayerAgent, log *TurnLog) *payment.Preferences {
	return e.paymentPreferencesForSpellFromZone(g, playerID, 0, zone.Hand, card, xValue, agents, log)
}

func (e *Engine) paymentPreferencesForSpellFromZone(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, card *game.CardDef, xValue int, agents [game.NumPlayers]PlayerAgent, log *TurnLog) *payment.Preferences {
	option := e.chooseSpellCostOptionFromZone(g, playerID, cardID, sourceZone, card, xValue, agents, log)
	prefs := e.paymentPreferencesForCostFromSource(g, playerID, option.ManaCost, option.AdditionalCosts, xValue, cardID, sourceZone, agents, log)
	prefs.AlternativeIndex = option.Index
	return prefs
}

func reservedCardIDs(reserved map[id.ID]bool) []id.ID {
	ids := make([]id.ID, 0, len(reserved))
	for cardID := range reserved {
		ids = append(ids, cardID)
	}
	return ids
}

func reserveIfGraveyard(additionalCost cost.Additional, choices []id.ID, reserved map[id.ID]bool) {
	if additionalCostSourceZone(additionalCost) != zone.Graveyard {
		return
	}
	for _, cardID := range choices {
		reserved[cardID] = true
	}
}

func additionalCostSourceZone(additionalCost cost.Additional) zone.Type {
	if additionalCost.Source != zone.None {
		return additionalCost.Source
	}
	if additionalCost.Kind == cost.AdditionalExile || additionalCost.Kind == cost.AdditionalCollectEvidence {
		return zone.Graveyard
	}
	return zone.Hand
}

func (e *Engine) chooseSpellCostOptionFromZone(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, card *game.CardDef, xValue int, agents [game.NumPlayers]PlayerAgent, log *TurnLog) payment.SpellOptionSummary {
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

func (e *Engine) additionalCostPermanentChoices(g *game.Game, playerID game.PlayerID, addCost cost.Additional, amount int, agents [game.NumPlayers]PlayerAgent, log *TurnLog, excludedTapIDs ...id.ID) []id.ID {
	candidates := candidateSacrificePermanents(g, playerID, addCost, excludedTapIDs)
	if len(candidates) <= amount {
		return paymentPermanentIDs(candidates)
	}
	options := make([]game.ChoiceOption, 0, len(candidates))
	for i, permanent := range candidates {
		options = append(options, game.ChoiceOption{Index: i, Label: permanentChoiceLabel(g, permanent), Card: permanentChoiceInfo(g, permanent)})
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

func (e *Engine) additionalCostCardChoices(g *game.Game, playerID game.PlayerID, addCost cost.Additional, amount int, remainingCosts []cost.Additional, xValue int, sourceCardID id.ID, sourceZone zone.Type, agents [game.NumPlayers]PlayerAgent, log *TurnLog, excludedCardIDs ...id.ID) []id.ID {
	if amount == 0 {
		return nil
	}
	candidates := candidateAdditionalCostCards(g, playerID, addCost, excludedCardIDs...)
	if len(candidates) <= amount {
		return candidates
	}
	defaultSelection := firstChoiceIndices(amount)
	if additionalCostSourceZone(addCost) == zone.Graveyard {
		defaultSelection = exileDefaultSelection(g, playerID, candidates, addCost, amount, remainingCosts, xValue, sourceCardID, sourceZone, excludedCardIDs...)
	}
	if len(defaultSelection) == 0 {
		return nil
	}
	options := make([]game.ChoiceOption, 0, len(candidates))
	for i, cardID := range candidates {
		options = append(options, game.ChoiceOption{Index: i, Label: cardChoiceLabel(g, cardID), Card: cardChoiceInfo(g, cardID)})
	}
	request := game.ChoiceRequest{
		Kind:             game.ChoicePayment,
		Player:           playerID,
		Prompt:           payment.AdditionalCostText(addCost),
		Options:          options,
		MinChoices:       amount,
		MaxChoices:       amount,
		DefaultSelection: defaultSelection,
	}
	selected := e.chooseChoice(g, agents, request, log)
	return selectedCardIDs(candidates, selected)
}

func (e *Engine) collectEvidenceChoices(g *game.Game, playerID game.PlayerID, addCost cost.Additional, threshold int, remainingCosts []cost.Additional, xValue int, sourceCardID id.ID, sourceZone zone.Type, agents [game.NumPlayers]PlayerAgent, log *TurnLog, excludedCardIDs ...id.ID) []id.ID {
	if threshold <= 0 {
		return nil
	}
	candidates := candidateAdditionalCostCards(g, playerID, addCost, excludedCardIDs...)
	defaultSelection := evidenceDefaultSelection(g, playerID, candidates, threshold, remainingCosts, xValue, sourceCardID, sourceZone, excludedCardIDs...)
	if len(defaultSelection) == 0 {
		return nil
	}
	options := make([]game.ChoiceOption, 0, len(candidates))
	for i, cardID := range candidates {
		options = append(options, game.ChoiceOption{Index: i, Label: evidenceChoiceLabel(g, cardID)})
	}
	request := game.ChoiceRequest{
		Kind:             game.ChoicePayment,
		Player:           playerID,
		Prompt:           payment.AdditionalCostText(addCost),
		Options:          options,
		MinChoices:       1,
		MaxChoices:       len(candidates),
		DefaultSelection: defaultSelection,
	}
	selected := e.chooseChoice(g, agents, request, log)
	return selectedCardIDs(candidates, selected)
}

func evidenceDefaultSelection(g *game.Game, playerID game.PlayerID, candidates []id.ID, threshold int, remainingCosts []cost.Additional, xValue int, sourceCardID id.ID, sourceZone zone.Type, excludedCardIDs ...id.ID) []int {
	reserved := cardIDSet(excludedCardIDs...)
	return evidenceDefaultSelectionWithReserved(g, playerID, candidates, threshold, remainingCosts, xValue, sourceCardID, sourceZone, reserved)
}

func evidenceDefaultSelectionWithReserved(g *game.Game, playerID game.PlayerID, candidates []id.ID, threshold int, remainingCosts []cost.Additional, xValue int, sourceCardID id.ID, sourceZone zone.Type, reserved map[id.ID]bool) []int {
	type evidenceOption struct {
		index     int
		manaValue int
	}
	var options []evidenceOption
	for i, cardID := range candidates {
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			continue
		}
		face := cardFaceOrDefault(card, game.FaceFront)
		if !evidenceFaceHasSupportedManaValue(face) {
			continue
		}
		manaValue := face.ManaValue()
		if manaValue <= 0 {
			continue
		}
		options = append(options, evidenceOption{index: i, manaValue: manaValue})
	}
	slices.SortStableFunc(options, func(a, b evidenceOption) int {
		switch {
		case a.manaValue > b.manaValue:
			return -1
		case a.manaValue < b.manaValue:
			return 1
		default:
			return 0
		}
	})
	var search func(start int, total int, selected []int) []int
	search = func(start int, total int, selected []int) []int {
		if total >= threshold {
			nextReserved := reserveSelectedCardIDs(candidates, selected, reserved)
			if remainingGraveyardPreferenceCostsPayable(g, playerID, remainingCosts, xValue, sourceCardID, sourceZone, nextReserved) {
				return append([]int(nil), selected...)
			}
			return nil
		}
		for i := start; i < len(options); i++ {
			option := options[i]
			next := slices.Clone(selected)
			next = append(next, option.index)
			if selection := search(i+1, total+option.manaValue, next); len(selection) > 0 {
				return selection
			}
		}
		return nil
	}
	return search(0, 0, nil)
}

func exileDefaultSelection(g *game.Game, playerID game.PlayerID, candidates []id.ID, addCost cost.Additional, amount int, remainingCosts []cost.Additional, xValue int, sourceCardID id.ID, sourceZone zone.Type, excludedCardIDs ...id.ID) []int {
	reserved := cardIDSet(excludedCardIDs...)
	return exileDefaultSelectionWithReserved(g, playerID, candidates, addCost, amount, remainingCosts, xValue, sourceCardID, sourceZone, reserved)
}

func exileDefaultSelectionWithReserved(g *game.Game, playerID game.PlayerID, candidates []id.ID, addCost cost.Additional, amount int, remainingCosts []cost.Additional, xValue int, sourceCardID id.ID, sourceZone zone.Type, reserved map[id.ID]bool) []int {
	return chooseFixedChoiceIndices(len(candidates), amount, func(selected []int) bool {
		nextReserved := reserveSelectedCardIDs(candidates, selected, reserved)
		return remainingGraveyardPreferenceCostsPayable(g, playerID, remainingCosts, xValue, sourceCardID, sourceZone, nextReserved)
	})
}

func chooseFixedChoiceIndices(candidateCount, amount int, allowsRemaining func([]int) bool) []int {
	var search func(start int, selected []int) []int
	search = func(start int, selected []int) []int {
		if len(selected) == amount {
			if allowsRemaining(selected) {
				return append([]int(nil), selected...)
			}
			return nil
		}
		remainingNeeded := amount - len(selected)
		for i := start; i <= candidateCount-remainingNeeded; i++ {
			next := slices.Clone(selected)
			next = append(next, i)
			if selection := search(i+1, next); len(selection) == amount {
				return selection
			}
		}
		return nil
	}
	return search(0, nil)
}

func remainingGraveyardPreferenceCostsPayable(g *game.Game, playerID game.PlayerID, remainingCosts []cost.Additional, xValue int, sourceCardID id.ID, sourceZone zone.Type, reserved map[id.ID]bool) bool {
	for i, additional := range remainingCosts {
		amount := payment.AdditionalCostAmountFor(additional, xValue)
		if amount < 0 {
			return false
		}
		switch additional.Kind {
		case cost.AdditionalCollectEvidence:
			if amount <= 0 {
				return false
			}
			candidates := candidateAdditionalCostCards(g, playerID, additional, reservedCardIDs(reserved)...)
			return len(evidenceDefaultSelectionWithReserved(g, playerID, candidates, amount, remainingCosts[i+1:], xValue, sourceCardID, sourceZone, reserved)) > 0
		case cost.AdditionalExile:
			if amount == 0 || additionalCostSourceZone(additional) != zone.Graveyard {
				continue
			}
			candidates := candidateAdditionalCostCards(g, playerID, additional, reservedCardIDs(reserved)...)
			return len(exileDefaultSelectionWithReserved(g, playerID, candidates, additional, amount, remainingCosts[i+1:], xValue, sourceCardID, sourceZone, reserved)) == amount
		case cost.AdditionalExileSource:
			if amount != 1 || sourceCardID == 0 || sourceZone != zone.Graveyard {
				continue
			}
			if reserved[sourceCardID] {
				return false
			}
			card, ok := g.GetCardInstance(sourceCardID)
			if !ok || !g.Players[playerID].Graveyard.Contains(sourceCardID) || !localAdditionalCostMatchesCard(cardFaceOrDefault(card, game.FaceFront), additional) {
				return false
			}
			nextReserved := cardIDSet(reservedCardIDs(reserved)...)
			nextReserved[sourceCardID] = true
			return remainingGraveyardPreferenceCostsPayable(g, playerID, remainingCosts[i+1:], xValue, sourceCardID, sourceZone, nextReserved)
		}
	}
	return true
}

func cardIDSet(cardIDs ...id.ID) map[id.ID]bool {
	result := make(map[id.ID]bool, len(cardIDs))
	for _, cardID := range cardIDs {
		result[cardID] = true
	}
	return result
}

func reserveSelectedCardIDs(candidates []id.ID, selected []int, reserved map[id.ID]bool) map[id.ID]bool {
	next := make(map[id.ID]bool, len(reserved)+len(selected))
	maps.Copy(next, reserved)
	for _, selection := range selected {
		if selection >= 0 && selection < len(candidates) {
			next[candidates[selection]] = true
		}
	}
	return next
}

func evidenceFaceHasSupportedManaValue(face *game.CardDef) bool {
	if face == nil {
		return false
	}
	if !face.ManaCost.Exists {
		return true
	}
	for _, symbol := range face.ManaCost.Val {
		if symbol.Kind == cost.VariableSymbol {
			return false
		}
	}
	return true
}

func firstChoiceIndices(amount int) []int {
	selected := make([]int, 0, amount)
	for i := range amount {
		selected = append(selected, i)
	}
	return selected
}

func candidateSacrificePermanents(g *game.Game, playerID game.PlayerID, addCost cost.Additional, excludedTapIDs []id.ID) []*game.Permanent {
	excluded := map[id.ID]bool{}
	if addCost.Kind == cost.AdditionalTapPermanents {
		for _, permanentID := range excludedTapIDs {
			excluded[permanentID] = true
		}
	}
	var candidates []*game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Controller != playerID || !localAdditionalCostMatchesPermanent(g, permanent, addCost) {
			continue
		}
		if excluded[permanent.ObjectID] {
			continue
		}
		if addCost.Kind == cost.AdditionalTapPermanents && permanent.Tapped {
			continue
		}
		if addCost.RequireTapped && !permanent.Tapped {
			continue
		}
		candidates = append(candidates, permanent)
	}
	return candidates
}

func localAdditionalCostMatchesPermanent(g *game.Game, permanent *game.Permanent, addCost cost.Additional) bool {
	if addCost.MatchPermanentType && !permanentHasType(g, permanent, addCost.PermanentType) {
		return false
	}
	if addCost.RequireSupertype != "" && !permanentHasSupertype(g, permanent, addCost.RequireSupertype) {
		return false
	}
	if addCost.SubtypesAny != (cost.SubtypeSet{}) {
		for _, subtype := range addCost.SubtypesAny {
			if subtype != "" && permanentHasSubtype(g, permanent, subtype) {
				return true
			}
		}
		return false
	}
	return true
}

func candidateAdditionalCostCards(g *game.Game, playerID game.PlayerID, addCost cost.Additional, excludedCardIDs ...id.ID) []id.ID {
	player, ok := playerByID(g, playerID)
	if !ok {
		return nil
	}
	source := additionalCostSourceZone(addCost)
	excluded := make(map[id.ID]bool, len(excludedCardIDs))
	for _, cardID := range excludedCardIDs {
		excluded[cardID] = true
	}
	var cardIDs []id.ID
	switch source {
	case zone.Hand:
		cardIDs = player.Hand.All()
	case zone.Graveyard:
		cardIDs = player.Graveyard.All()
	case zone.Exile:
		cardIDs = player.Exile.All()
	case zone.Command:
		cardIDs = player.CommandZone.All()
	default:
		return nil
	}
	var candidates []id.ID
	for _, cardID := range cardIDs {
		if excluded[cardID] {
			continue
		}
		card, ok := g.GetCardInstance(cardID)
		if ok && localAdditionalCostMatchesCard(cardFaceOrDefault(card, game.FaceFront), addCost) {
			candidates = append(candidates, cardID)
		}
	}
	return candidates
}

func localAdditionalCostMatchesCard(face *game.CardDef, addCost cost.Additional) bool {
	if face == nil {
		return false
	}
	if addCost.MatchCardType && !face.HasType(addCost.CardType) {
		return false
	}
	if addCost.MatchCardColor && !slices.Contains(face.Colors, addCost.CardColor) {
		return false
	}
	if addCost.SubtypesAny != (cost.SubtypeSet{}) {
		for _, subtype := range addCost.SubtypesAny {
			if subtype != "" && face.HasSubtype(subtype) {
				return true
			}
		}
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

func evidenceChoiceLabel(g *game.Game, cardID id.ID) string {
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return fmt.Sprintf("Card %d", cardID)
	}
	face := cardFaceOrDefault(card, game.FaceFront)
	return fmt.Sprintf("%s (mana value %d)", face.Name, face.ManaValue())
}
