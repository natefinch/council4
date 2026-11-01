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
	phyrexianLifeChoices := e.phyrexianPaymentChoices(g, playerID, manaCost, agents, log)
	prefs := e.additionalCostPaymentPreferences(g, playerID, additionalCosts, xValue, sourceCardID, sourceZone, agents, log, tapExclusions...)
	prefs.PhyrexianLifeChoices = phyrexianLifeChoices
	return prefs
}

// additionalCostPaymentPreferences gathers the caster's choices for paying
// additionalCosts, without the Phyrexian mana announcements.
func (e *Engine) additionalCostPaymentPreferences(g *game.Game, playerID game.PlayerID, additionalCosts []cost.Additional, xValue int, sourceCardID id.ID, sourceZone zone.Type, agents [game.NumPlayers]PlayerAgent, log *TurnLog, tapExclusions ...id.ID) *payment.Preferences {
	prefs := &payment.Preferences{}
	reservedGraveyardCards := map[id.ID]bool{}
	for i, additionalCost := range additionalCosts {
		amount := payment.AdditionalCostAmountFor(additionalCost, xValue)
		if additionalCost.AmountDynamic != cost.AdditionalDynamicAmountNone {
			amount = (&rulesPaymentState{g: g}).AdditionalDynamicAmountValue(playerID, additionalCost.AmountDynamic)
		}
		switch additionalCost.Kind {
		case cost.AdditionalSacrifice:
			prefs.SacrificeChoices = append(prefs.SacrificeChoices, e.additionalCostPermanentChoices(g, playerID, additionalCost, amount, agents, log)...)
		case cost.AdditionalTapPermanents:
			if additionalCost.TotalPowerAtLeast > 0 {
				// The payment planner greedily selects which creatures to tap to
				// reach the Saddle power threshold; no explicit preference is
				// gathered here.
				continue
			}
			prefs.TapChoices = append(prefs.TapChoices, e.additionalCostPermanentChoices(g, playerID, additionalCost, amount, agents, log, tapExclusions...)...)
		case cost.AdditionalReturnToHand:
			prefs.ReturnChoices = append(prefs.ReturnChoices, e.additionalCostPermanentChoices(g, playerID, additionalCost, amount, agents, log)...)
		case cost.AdditionalRemoveCounterAmong:
			prefs.RemoveCounterChoices = append(prefs.RemoveCounterChoices, e.additionalCostRemoveCounterAmongChoices(g, playerID, additionalCost, amount, agents, log)...)
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
	return e.paymentPreferencesForSpellFromZone(g, playerID, 0, zone.Hand, game.FaceFront, card, xValue, agents, log)
}

func (e *Engine) paymentPreferencesForSpellFromZone(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex, card *game.CardDef, xValue int, agents [game.NumPlayers]PlayerAgent, log *TurnLog) *payment.Preferences {
	return e.paymentPreferencesForSpellFromZoneWithKicker(g, playerID, cardID, sourceZone, face, card, xValue, false, 0, agents, log)
}

func (e *Engine) paymentPreferencesForSpellFromZoneWithKicker(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex, card *game.CardDef, xValue int, kickerPaid bool, kickerCount int, agents [game.NumPlayers]PlayerAgent, log *TurnLog) *payment.Preferences {
	request := payment.SpellRequest{
		PlayerID:        playerID,
		CardID:          cardID,
		SourceZone:      sourceZone,
		Card:            card,
		XValue:          xValue,
		KickerPaid:      kickerPaid,
		KickerCount:     kickerCount,
		CastPermissions: castPermissionsForZone(g, playerID, cardID, sourceZone, face),
	}
	return e.paymentPreferencesForSpellRequest(g, request, agents, log)
}

func (e *Engine) paymentPreferencesForSpellRequest(g *game.Game, request payment.SpellRequest, agents [game.NumPlayers]PlayerAgent, log *TurnLog) *payment.Preferences {
	return e.paymentPreferencesForSpellOption(g, request, e.chooseSpellCostOptionForRequest(g, request, agents, log), agents, log)
}

// paymentPreferencesForSpellOption returns payment preferences for an already
// chosen spell cost option of request.
func (e *Engine) paymentPreferencesForSpellOption(g *game.Game, request payment.SpellRequest, option payment.SpellOptionSummary, agents [game.NumPlayers]PlayerAgent, log *TurnLog) *payment.Preferences {
	prefs := e.paymentPreferencesForCostFromSource(g, request.PlayerID, option.ManaCost, option.AdditionalCosts, request.XValue, request.CardID, request.SourceZone, agents, log)
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

func (e *Engine) chooseSpellCostOptionFromZone(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex, card *game.CardDef, xValue int, agents [game.NumPlayers]PlayerAgent, log *TurnLog) payment.SpellOptionSummary {
	return e.chooseSpellCostOptionFromZoneWithKicker(g, playerID, cardID, sourceZone, face, card, xValue, false, 0, agents, log)
}

func (e *Engine) chooseSpellCostOptionFromZoneWithKicker(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex, card *game.CardDef, xValue int, kickerPaid bool, kickerCount int, agents [game.NumPlayers]PlayerAgent, log *TurnLog) payment.SpellOptionSummary {
	return e.chooseSpellCostOptionForRequest(g, payment.SpellRequest{
		PlayerID:        playerID,
		CardID:          cardID,
		SourceZone:      sourceZone,
		Card:            card,
		XValue:          xValue,
		KickerPaid:      kickerPaid,
		KickerCount:     kickerCount,
		CastPermissions: castPermissionsForZone(g, playerID, cardID, sourceZone, face),
	}, agents, log)
}

func (e *Engine) chooseSpellCostOptionForRequest(g *game.Game, request payment.SpellRequest, agents [game.NumPlayers]PlayerAgent, log *TurnLog) payment.SpellOptionSummary {
	return e.chooseSpellCostOption(g, request.PlayerID, paymentOrch.planner(g).PayableSpellOptions(request), agents, log)
}

// chooseSpellCostOption lets player choose one of the payable cost options of
// a spell (CR 601.2b).
func (e *Engine) chooseSpellCostOption(g *game.Game, player game.PlayerID, options []payment.SpellOptionSummary, agents [game.NumPlayers]PlayerAgent, log *TurnLog) payment.SpellOptionSummary {
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
	choiceRequest := game.ChoiceRequest{
		Kind:             game.ChoicePayment,
		Player:           player,
		Prompt:           "Choose spell cost",
		Options:          choiceOptions,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{options[0].Index},
	}
	selected := e.chooseChoice(g, agents, choiceRequest, log)
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
	choices, _ := e.phyrexianPaymentChoicesAfter(g, playerID, manaCost, nil, agents, log)
	return choices
}

// announcePhyrexianPayments asks how the caster will pay each Phyrexian symbol
// of a proposed spell cost, in cost order, before its targets are chosen and
// before any mana ability window (CR 601.2b). manaFeasible reports whether the
// proposed cost can still be paid with the given announcements; each symbol's
// choice defaults to mana when announcing it as mana keeps the cost payable,
// and otherwise to 2 life, since an announced mana payment is binding.
// withheld, when non-nil, reports whether a complete set of announcements was
// withheld after an earlier attempt was reversed (see castAttempts); a payment
// is offered only while some announcements including it remain open.
func (e *Engine) announcePhyrexianPayments(g *game.Game, playerID game.PlayerID, manaCost *cost.Mana, manaFeasible, withheld func([]phyrexianAnnouncement) bool, agents [game.NumPlayers]PlayerAgent, log *TurnLog) []phyrexianAnnouncement {
	symbols := phyrexianSymbols(manaCost)
	availableLife := playerLife(g, playerID)
	var announced []phyrexianAnnouncement
	for _, symbol := range symbols {
		withMana := append(slices.Clone(announced), phyrexianAnnouncement{symbol: symbol})
		withLife := append(slices.Clone(announced), phyrexianAnnouncement{symbol: symbol, payLife: true})
		offer := phyrexianOffer{
			mana: announcementsOpen(symbols, availableLife, withMana, withheld),
			life: availableLife >= 2 && announcementsOpen(symbols, availableLife-2, withLife, withheld),
		}
		if !offer.mana && !offer.life {
			offer = phyrexianOffer{mana: true, life: availableLife >= 2}
		}
		preferLife := manaFeasible != nil && !manaFeasible(withMana)
		payLife := e.choosePhyrexianPaymentOf(g, playerID, symbol, offer, preferLife, agents, log)
		if payLife {
			availableLife -= 2
		}
		announced = append(announced, phyrexianAnnouncement{symbol: symbol, payLife: payLife})
	}
	return announced
}

// phyrexianPaymentChoicesAfter returns one payment choice per Phyrexian symbol
// of manaCost's effective cost, in cost order, true paying 2 life, and whether
// each must be paid with mana. announced maps cost positions to announced
// payments (true is 2 life). A symbol announced as mana must be paid with mana
// unless an effect such as K'rrik, Son of Yawgmoth lets the caster pay 2 life
// for that color: then, as for each unannounced symbol, the caster chooses
// while paying (CR 601.2h).
func (e *Engine) phyrexianPaymentChoicesAfter(g *game.Game, playerID game.PlayerID, manaCost *cost.Mana, announced map[int]bool, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (lifeChoices, manaOnly []bool) {
	if manaCost == nil {
		return nil, nil
	}
	state := &rulesPaymentState{g: g}
	effective := payment.EffectiveManaCost(state, playerID, *manaCost)
	availableLife := playerLife(g, playerID)
	for _, payLife := range announced {
		if payLife {
			availableLife -= 2
		}
	}
	for i, symbol := range effective {
		if !isPhyrexianSymbol(symbol) {
			continue
		}
		payLife, ok := announced[i]
		locked := ok && !payLife
		if locked && symbol.Kind == cost.PhyrexianSymbol && state.PayLifeForManaColor(playerID, symbol.Color) {
			ok, locked = false, false
		}
		if !ok {
			payLife = e.choosePhyrexianPayment(g, playerID, symbol, availableLife, false, agents, log)
			if payLife {
				availableLife -= 2
			}
		}
		lifeChoices = append(lifeChoices, payLife)
		manaOnly = append(manaOnly, locked)
	}
	return lifeChoices, manaOnly
}

// choosePhyrexianPayment asks whether to pay a Phyrexian symbol with 2 life
// rather than mana; paying life is offered only with at least 2 life available.
// The choice defaults to life when preferLife and life is offered, and
// otherwise to mana.
func (e *Engine) choosePhyrexianPayment(g *game.Game, playerID game.PlayerID, symbol cost.Symbol, availableLife int, preferLife bool, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	return e.choosePhyrexianPaymentOf(g, playerID, symbol, phyrexianOffer{mana: true, life: availableLife >= 2}, preferLife, agents, log)
}

// phyrexianOffer is which payments a Phyrexian symbol's choice offers.
type phyrexianOffer struct {
	mana, life bool
}

// choosePhyrexianPaymentOf asks whether to pay a Phyrexian symbol with 2 life
// rather than mana, offering the payments in offer, at least one of them. The
// choice defaults to life when it is offered and either preferLife or mana
// isn't offered, and otherwise to mana.
func (e *Engine) choosePhyrexianPaymentOf(g *game.Game, playerID game.PlayerID, symbol cost.Symbol, offer phyrexianOffer, preferLife bool, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	manaLabel := fmt.Sprintf("Pay %s mana", symbol.Color)
	if symbol.Kind == cost.PhyrexianGenericSymbol {
		manaLabel = fmt.Sprintf("Pay {%d}", symbol.Generic)
	}
	var options []game.ChoiceOption
	if offer.mana {
		options = append(options, game.ChoiceOption{Index: 0, Label: manaLabel})
	}
	if offer.life {
		options = append(options, game.ChoiceOption{Index: 1, Label: "Pay 2 life"})
	}
	defaultSelection := 0
	if offer.life && (preferLife || !offer.mana) {
		defaultSelection = 1
	}
	selected := e.chooseChoice(g, agents, game.ChoiceRequest{
		Kind:             game.ChoicePayment,
		Player:           playerID,
		Prompt:           fmt.Sprintf("Pay %s", symbol),
		Options:          options,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{defaultSelection},
	}, log)
	return len(selected) == 1 && selected[0] == 1
}

func playerLife(g *game.Game, playerID game.PlayerID) int {
	if player, ok := playerByID(g, playerID); ok {
		return player.Life
	}
	return 0
}

func isPhyrexianSymbol(symbol cost.Symbol) bool {
	return symbol.Kind == cost.PhyrexianSymbol || symbol.Kind == cost.PhyrexianGenericSymbol
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

// additionalCostRemoveCounterAmongChoices gathers the per-counter selection for
// an AdditionalRemoveCounterAmong cost. Each removable counter on a matching
// controlled permanent becomes one selectable option, so the player may take
// several counters from the same permanent. The returned slice names one
// permanent per counter removed.
func (e *Engine) additionalCostRemoveCounterAmongChoices(g *game.Game, playerID game.PlayerID, addCost cost.Additional, amount int, agents [game.NumPlayers]PlayerAgent, log *TurnLog) []id.ID {
	if amount <= 0 {
		return nil
	}
	candidates := candidateSacrificePermanents(g, playerID, addCost, nil)
	var expanded []id.ID
	options := make([]game.ChoiceOption, 0, len(candidates))
	for _, permanent := range candidates {
		for range payment.RemovableAmongCounterCount(permanent, addCost) {
			options = append(options, game.ChoiceOption{Index: len(expanded), Label: permanentChoiceLabel(g, permanent), Card: permanentChoiceInfo(g, permanent)})
			expanded = append(expanded, permanent.ObjectID)
		}
	}
	if len(expanded) <= amount {
		return expanded
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
	result := make([]id.ID, 0, len(selected))
	for _, index := range selected {
		if index < 0 || index >= len(expanded) {
			continue
		}
		result = append(result, expanded[index])
	}
	return result
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
			if !ok || !g.Players[playerID].Graveyard.Contains(sourceCardID) || !localAdditionalCostMatchesCard(g, card, additional) {
				return false
			}
			nextReserved := cardIDSet(reservedCardIDs(reserved)...)
			nextReserved[sourceCardID] = true
			return remainingGraveyardPreferenceCostsPayable(g, playerID, remainingCosts[i+1:], xValue, sourceCardID, sourceZone, nextReserved)
		default:
			// Only the selection-bearing additional cost kinds above carry card
			// selection; other kinds need no handling here.
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
	var excludedIDs []id.ID
	if addCost.Kind == cost.AdditionalTapPermanents {
		excludedIDs = excludedTapIDs
	}
	candidates := payment.CandidatePermanentsForCost(&rulesPaymentState{g: g}, playerID, addCost, nil, excludedIDs...)
	if addCost.Kind != cost.AdditionalSacrifice {
		return candidates
	}
	kept := candidates[:0]
	for _, permanent := range candidates {
		if permanentCantBeSacrificed(g, permanent) {
			continue
		}
		kept = append(kept, permanent)
	}
	return kept
}

func candidateAdditionalCostCards(g *game.Game, playerID game.PlayerID, addCost cost.Additional, excludedCardIDs ...id.ID) []id.ID {
	// The choice layer never excludes the cost's own source card here (sourceCardID
	// is zero): escape's "another" exclusion is enforced by the planner, and
	// widening it to candidate presentation would change the offered set. Candidate
	// enumeration otherwise shares the planner's reservation-aware engine.
	return payment.CandidateCardsForCost(&rulesPaymentState{g: g}, playerID, addCost, 0, excludedCardIDs...)
}

func localAdditionalCostMatchesCard(g *game.Game, card *game.CardInstance, addCost cost.Additional) bool {
	sel, ok := payment.SelectionForAdditionalCost(addCost)
	if !ok {
		return false
	}
	return cardMatchesCostSelection(g, card, sel)
}

// permanentMatchesCostSelection reports whether a battlefield permanent
// satisfies a cost's converted Selection. It builds the same subjectPermanent
// the effect side uses so additional-cost candidate filtering and the payment
// planner share one matcher.
func permanentMatchesCostSelection(g *game.Game, permanent *game.Permanent, sel game.Selection) bool {
	values := effectivePermanentValues(g, permanent)
	subject := selectionSubject{
		kind:      subjectPermanent,
		g:         g,
		permanent: permanent,
		values:    values,
	}
	if sel.Controller != game.ControllerAny {
		subject.controller = effectiveController(g, permanent)
	}
	return matchSelection(&subject, &sel)
}

// cardMatchesCostSelection reports whether a real card in a non-battlefield zone
// (a card to discard, exile or reveal for a cost, or a card whose ability cost
// is modified) satisfies a cost's converted Selection. It reads the instance
// itself through the subjectCard the effect side uses, so the card's own
// characteristic-defining abilities, its counters and the type-changing
// effects that apply to cards its owner owns outside the battlefield count
// (Arcane Adaptation makes a creature card in hand the chosen type, CR 613.1d).
func cardMatchesCostSelection(g *game.Game, card *game.CardInstance, sel game.Selection) bool {
	if card == nil || card.Def == nil {
		return false
	}
	return matchSelection(&selectionSubject{
		kind: subjectCard,
		g:    g,
		card: card,
	}, &sel)
}

// cardDefMatchesCostSelection reports whether a card face that is not a card
// in a zone (a spell's cast characteristics, castSelectionFace) satisfies a
// cost's converted Selection, reading the face through the same subjectCard
// the effect side uses. A card in a zone uses cardMatchesCostSelection.
func cardDefMatchesCostSelection(g *game.Game, face *game.CardDef, sel game.Selection) bool {
	if face == nil {
		return false
	}
	return matchSelection(&selectionSubject{
		kind: subjectCard,
		g:    g,
		card: &game.CardInstance{Def: face},
	}, &sel)
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

// permanentChoiceLabel labels a permanent in a choice request or logged
// decision by its effective (public) name. A nameless face-down permanent (CR
// 708.2a) is labeled as a face-down permanent with its object ID, so several
// face-down permanents stay distinguishable without revealing the card
// underneath; its controller can match the ID against the private face-down
// card views in its observation (CR 708.5).
func permanentChoiceLabel(g *game.Game, permanent *game.Permanent) string {
	if name := permanentEffectiveName(g, permanent); name != "" {
		return name
	}
	if permanent.FaceDown {
		return fmt.Sprintf("Face-down permanent %d", permanent.ObjectID)
	}
	return fmt.Sprintf("Permanent %d", permanent.ObjectID)
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
