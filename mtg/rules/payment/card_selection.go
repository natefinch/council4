package payment

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

// cardCostChoice is the shared descriptor for the cards an additional card-zone
// cost (discard/exile/threshold-exile/reveal/collect-evidence) may select.
// Candidate enumeration (agent-facing choice presentation) and payment planning
// both derive their eligible set from one cardCostChoice, so the cards offered to
// the agent are exactly the cards the planner will accept (CR 601.2b/602.2b).
// Only card eligibility is unified here; action execution (discard vs exile vs
// reveal vs collect-evidence) stays distinct because those have different
// authoritative rules behavior. It is the card-zone analogue of objectCostChoice.
type cardCostChoice struct {
	// selection is the canonical characteristic predicate the cost imposes,
	// converted once via SelectionForAdditionalCost (the #1730 matcher).
	selection game.Selection
	// domain is the zone the cost draws cards from (hand or graveyard, or an
	// explicit source zone), resolved once via cardCostZone.
	domain zone.Type
	// excludeSource drops the cost's own source card, backing escape's "exile N
	// other cards from your graveyard".
	excludeSource bool
	// sourceCardID identifies the source card excluded when excludeSource is set.
	sourceCardID id.ID
}

// cardCostZone resolves the zone a card-zone cost draws from: an explicit Source,
// else the graveyard for exile/collect-evidence costs, else the hand. It mirrors
// the rules-layer additionalCostSourceZone so the choice layer and the planner
// enumerate from the same zone.
func cardCostZone(additional cost.Additional) zone.Type {
	if additional.Source != zone.None {
		return additional.Source
	}
	if additional.Kind == cost.AdditionalExile || additional.Kind == cost.AdditionalCollectEvidence {
		return zone.Graveyard
	}
	return zone.Hand
}

// cardCostChoiceForCost converts an additional card-zone cost's constraint into
// the shared descriptor. It returns false when the constraint is not
// representable as a Selection so callers fail closed, mirroring
// SelectionForAdditionalCost.
func cardCostChoiceForCost(additional cost.Additional, sourceCardID id.ID) (cardCostChoice, bool) {
	sel, ok := SelectionForAdditionalCost(additional)
	if !ok {
		return cardCostChoice{}, false
	}
	return cardCostChoice{
		selection:     sel,
		domain:        cardCostZone(additional),
		excludeSource: additional.ExcludeSource,
		sourceCardID:  sourceCardID,
	}, true
}

// cardSatisfiesObjectCost reports whether one card is an eligible object for the
// cost: it is not the excluded source, it has a live instance, and it matches the
// canonical selection. It is the single per-card gate shared by candidate
// enumeration, the deterministic backtrackers, and preference fallback.
func cardSatisfiesObjectCost(s State, cardID id.ID, choice cardCostChoice) bool {
	if choice.excludeSource && cardID == choice.sourceCardID {
		return false
	}
	card, ok := s.CardInstance(cardID)
	if !ok {
		return false
	}
	return s.CardMatchesSelection(s.CardFace(card, game.FaceFront), choice.selection)
}

// candidateCardsForObjectCost enumerates, in zone order, every card eligible for
// the cost that is not already reserved by an earlier selection in the same
// payment. It is the one reservation-aware card-zone enumerator used by both the
// choice layer and the planner.
func candidateCardsForObjectCost(s State, playerID game.PlayerID, choice cardCostChoice, reserved map[id.ID]bool) []id.ID {
	player, ok := s.Player(playerID)
	if !ok {
		return nil
	}
	var candidates []id.ID
	for _, cardID := range cardIDsInZone(player, choice.domain) {
		if reserved[cardID] {
			continue
		}
		if cardSatisfiesObjectCost(s, cardID, choice) {
			candidates = append(candidates, cardID)
		}
	}
	return candidates
}

// reservedCardIDSet collects card IDs already selected by earlier costs into a
// lookup set so the shared enumerator and backtrackers never offer the same card
// twice.
func reservedCardIDSet(cardIDs []id.ID) map[id.ID]bool {
	reserved := make(map[id.ID]bool, len(cardIDs))
	for _, cardID := range cardIDs {
		reserved[cardID] = true
	}
	return reserved
}

// reservedFromSelections is reservedCardIDSet over card-zone selections, used by
// the planner choosers whose already-chosen cards carry their zone.
func reservedFromSelections(selections []cardZoneSelection) map[id.ID]bool {
	reserved := make(map[id.ID]bool, len(selections))
	for _, selection := range selections {
		reserved[selection.cardID] = true
	}
	return reserved
}

// cardZoneSelectionsFor tags a backtracker's chosen card IDs with the cost's zone
// so the planner can act on them (exile/evidence record the source zone).
func cardZoneSelectionsFor(cardIDs []id.ID, z zone.Type) []cardZoneSelection {
	if len(cardIDs) == 0 {
		return nil
	}
	selections := make([]cardZoneSelection, 0, len(cardIDs))
	for _, cardID := range cardIDs {
		selections = append(selections, cardZoneSelection{cardID: cardID, zone: z})
	}
	return selections
}

// cardZoneSelectionCardIDs projects card-zone selections to their card IDs for
// the choice layer, which keys candidate presentation on card IDs.
func cardZoneSelectionCardIDs(selections []cardZoneSelection) []id.ID {
	if len(selections) == 0 {
		return nil
	}
	cardIDs := make([]id.ID, 0, len(selections))
	for _, selection := range selections {
		cardIDs = append(cardIDs, selection.cardID)
	}
	return cardIDs
}

// truncateCardIDs returns at most amount card IDs from a zone-order candidate
// list, used by the fixed-count discard cost that takes the first eligible cards.
// A shorter result signals the player cannot supply enough cards, which the
// caller treats as unpayable.
func truncateCardIDs(cardIDs []id.ID, amount int) []id.ID {
	if len(cardIDs) > amount {
		return cardIDs[:amount]
	}
	return cardIDs
}

// chooseFixedCardSet selects exactly amount cards (in candidate order) such that
// allowsRemaining holds for the chosen set, backtracking when an earlier choice
// would leave a later graveyard cost unpayable. It is the card-zone analogue of
// the permanent fixed-count selection and returns nil when no satisfying set of
// the required size exists.
func chooseFixedCardSet(candidates []id.ID, amount int, allowsRemaining func(chosen []id.ID) bool) []id.ID {
	if amount == 0 {
		return nil
	}
	var search func(start int, chosen []id.ID) []id.ID
	search = func(start int, chosen []id.ID) []id.ID {
		if len(chosen) == amount {
			if allowsRemaining(chosen) {
				return slices.Clone(chosen)
			}
			return nil
		}
		remainingNeeded := amount - len(chosen)
		for i := start; i <= len(candidates)-remainingNeeded; i++ {
			next := append(slices.Clone(chosen), candidates[i])
			if selected := search(i+1, next); len(selected) == amount {
				return selected
			}
		}
		return nil
	}
	return search(0, nil)
}

// chooseThresholdCardSet selects cards (highest mana value first) whose total
// mana value reaches threshold, backtracking so a satisfying set also leaves
// later graveyard costs payable. It backs collect-evidence and threshold-exile
// ("exile cards with total mana value N or greater"). Candidates with no usable
// mana value are skipped. It returns nil when no satisfying set exists.
func chooseThresholdCardSet(s State, candidates []id.ID, threshold int, allowsRemaining func(chosen []id.ID) bool) []id.ID {
	type thresholdCandidate struct {
		cardID    id.ID
		manaValue int
	}
	var ranked []thresholdCandidate
	for _, cardID := range candidates {
		manaValue, ok := evidenceCardManaValue(s, cardID)
		if !ok || manaValue <= 0 {
			continue
		}
		ranked = append(ranked, thresholdCandidate{cardID: cardID, manaValue: manaValue})
	}
	slices.SortStableFunc(ranked, func(a, b thresholdCandidate) int {
		switch {
		case a.manaValue > b.manaValue:
			return -1
		case a.manaValue < b.manaValue:
			return 1
		default:
			return 0
		}
	})
	var search func(start int, total int, chosen []id.ID) []id.ID
	search = func(start int, total int, chosen []id.ID) []id.ID {
		if total >= threshold {
			if allowsRemaining(chosen) {
				return slices.Clone(chosen)
			}
			return nil
		}
		for i := start; i < len(ranked); i++ {
			next := append(slices.Clone(chosen), ranked[i].cardID)
			if selected := search(i+1, total+ranked[i].manaValue, next); len(selected) > 0 {
				return selected
			}
		}
		return nil
	}
	return search(0, 0, nil)
}

// CandidateCardsForCost returns the cards the player may choose for an additional
// card-zone cost — the same eligible set the planner enumerates. The choice layer
// calls it through the payment State so candidate presentation and payment
// planning share one selection pipeline. excludedIDs are cards already reserved
// by earlier costs in the same payment; sourceCardID is the cost's source card,
// excluded when the cost is "another" (escape).
func CandidateCardsForCost(s State, playerID game.PlayerID, additional cost.Additional, sourceCardID id.ID, excludedIDs ...id.ID) []id.ID {
	choice, ok := cardCostChoiceForCost(additional, sourceCardID)
	if !ok {
		return nil
	}
	return candidateCardsForObjectCost(s, playerID, choice, reservedCardIDSet(excludedIDs))
}

// ChooseExileCardIDs returns the card IDs the deterministic fixed-count exile
// backtracker would select, so the choice layer can present them as the default
// selection while sharing the planner's reservation-aware backtracking.
// reservedIDs are cards already reserved by earlier costs in the same payment.
func ChooseExileCardIDs(s State, playerID game.PlayerID, additional cost.Additional, amount int, reservedIDs []id.ID, remainingCosts []cost.Additional, xValue int, sourceCardID id.ID, sourceZone zone.Type) []id.ID {
	alreadyChosen := cardZoneSelectionsFor(reservedIDs, cardCostZone(additional))
	return cardZoneSelectionCardIDs(chooseExileCards(s, playerID, additional, amount, alreadyChosen, remainingCosts, xValue, sourceCardID, sourceZone))
}

// ChooseEvidenceCardIDs returns the card IDs the deterministic collect-evidence
// backtracker would select, so the choice layer can present them as the default
// selection while sharing the planner's mana-value-ordered backtracking.
func ChooseEvidenceCardIDs(s State, playerID game.PlayerID, threshold int, reservedIDs []id.ID, remainingCosts []cost.Additional, xValue int, sourceCardID id.ID, sourceZone zone.Type) []id.ID {
	alreadyChosen := cardZoneSelectionsFor(reservedIDs, zone.Graveyard)
	return cardZoneSelectionCardIDs(chooseEvidenceCards(s, playerID, threshold, alreadyChosen, remainingCosts, xValue, sourceCardID, sourceZone))
}
