package payment

import (
	"slices"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
)

// chooseEvidenceCards greedily selects graveyard cards (highest mana value
// first) whose total mana value reaches the evidence threshold and that leave
// later graveyard costs payable. It enumerates and backtracks through the shared
// card-zone engine so the planner and the choice layer agree on the eligible set
// and the deterministic default selection.
func chooseEvidenceCards(s State, playerID game.PlayerID, threshold int, alreadyChosen []cardZoneSelection, remainingCosts []cost.Additional, xValue int, sourceCardID id.ID, sourceZone zone.Type) []cardZoneSelection {
	reserved := reservedFromSelections(alreadyChosen)
	candidates := candidateCardsForObjectCost(s, playerID, cardCostChoice{domain: zone.Graveyard}, reserved)
	selected := chooseThresholdCardSet(s, candidates, threshold, func(chosen []id.ID) bool {
		reserved := appendCardZoneSelections(alreadyChosen, cardZoneSelectionsFor(chosen, zone.Graveyard)...)
		return remainingGraveyardCostsPayable(s, playerID, remainingCosts, xValue, reserved, sourceCardID, sourceZone)
	})
	return cardZoneSelectionsFor(selected, zone.Graveyard)
}

func preferredEvidenceCards(s State, playerID game.PlayerID, threshold int, alreadyChosen []cardZoneSelection, remainingCosts []cost.Additional, xValue int, sourceCardID id.ID, sourceZone zone.Type, prefs *Preferences) []cardZoneSelection {
	if prefs == nil || len(prefs.EvidenceChoices) == 0 {
		return chooseEvidenceCards(s, playerID, threshold, alreadyChosen, remainingCosts, xValue, sourceCardID, sourceZone)
	}
	chosenIDs := make(map[id.ID]bool, len(alreadyChosen))
	for _, chosen := range alreadyChosen {
		chosenIDs[chosen.cardID] = true
	}
	var chosen []cardZoneSelection
	total := 0
	var consumed int
	for _, cardID := range prefs.EvidenceChoices {
		manaValue, ok := evidenceCardManaValue(s, cardID)
		if !ok || !zoneContainsCard(s, playerID, zone.Graveyard, cardID) || chosenIDs[cardID] {
			return nil
		}
		chosen = append(chosen, cardZoneSelection{cardID: cardID, zone: zone.Graveyard})
		chosenIDs[cardID] = true
		total += manaValue
		consumed++
		if total >= threshold {
			prefs.EvidenceChoices = prefs.EvidenceChoices[consumed:]
			return chosen
		}
	}
	return nil
}

// chooseThresholdExileCards selects a set of cards from the cost's source zone
// matching the cost filter whose mana values total at least
// additional.TotalManaValueAtLeast, backing "exile any number of historic cards
// from your graveyard with total mana value N or greater" (The Capitoline
// Triad). It generalizes chooseEvidenceCards to an arbitrary card filter and is
// payable only when a satisfying set also leaves later graveyard costs payable.
func chooseThresholdExileCards(s State, playerID game.PlayerID, additional cost.Additional, alreadyChosen []cardZoneSelection, remainingCosts []cost.Additional, xValue int, sourceCardID id.ID, sourceZone zone.Type) []cardZoneSelection {
	choice, ok := cardCostChoiceForCost(additional, sourceCardID)
	if !ok {
		return nil
	}
	reserved := reservedFromSelections(alreadyChosen)
	candidates := candidateCardsForObjectCost(s, playerID, choice, reserved)
	selected := chooseThresholdCardSet(s, candidates, additional.TotalManaValueAtLeast, func(chosen []id.ID) bool {
		reserved := appendCardZoneSelections(alreadyChosen, cardZoneSelectionsFor(chosen, choice.domain)...)
		return remainingGraveyardCostsPayable(s, playerID, remainingCosts, xValue, reserved, sourceCardID, sourceZone)
	})
	return cardZoneSelectionsFor(selected, choice.domain)
}

func preferredThresholdExileCards(s State, playerID game.PlayerID, additional cost.Additional, alreadyChosen []cardZoneSelection, remainingCosts []cost.Additional, xValue int, sourceCardID id.ID, sourceZone zone.Type, prefs *Preferences) []cardZoneSelection {
	if prefs == nil || len(prefs.ExileChoices) == 0 {
		return chooseThresholdExileCards(s, playerID, additional, alreadyChosen, remainingCosts, xValue, sourceCardID, sourceZone)
	}
	selectionZone := additionalCostSourceZone(additional.Source)
	chosenIDs := make(map[id.ID]bool, len(alreadyChosen))
	for _, chosen := range alreadyChosen {
		chosenIDs[chosen.cardID] = true
	}
	var chosen []cardZoneSelection
	total := 0
	var consumed int
	for _, cardID := range prefs.ExileChoices {
		card, ok := s.CardInstance(cardID)
		if !ok || !zoneContainsCard(s, playerID, selectionZone, cardID) || chosenIDs[cardID] ||
			!additionalCostMatchesCard(s, s.CardFace(card, game.FaceFront), additional) ||
			(additional.ExcludeSource && cardID == sourceCardID) {
			return nil
		}
		manaValue, ok := evidenceCardManaValue(s, cardID)
		if !ok {
			return nil
		}
		chosen = append(chosen, cardZoneSelection{cardID: cardID, zone: selectionZone})
		chosenIDs[cardID] = true
		total += manaValue
		consumed++
		if total >= additional.TotalManaValueAtLeast {
			prefs.ExileChoices = prefs.ExileChoices[consumed:]
			return chosen
		}
	}
	return nil
}

func chooseExileCards(s State, playerID game.PlayerID, additional cost.Additional, amount int, alreadyChosen []cardZoneSelection, remainingCosts []cost.Additional, xValue int, sourceCardID id.ID, sourceZone zone.Type) []cardZoneSelection {
	choice, ok := cardCostChoiceForCost(additional, sourceCardID)
	if !ok {
		return nil
	}
	reserved := reservedFromSelections(alreadyChosen)
	candidates := candidateCardsForObjectCost(s, playerID, choice, reserved)
	selected := chooseFixedCardSet(candidates, amount, func(chosen []id.ID) bool {
		reserved := appendCardZoneSelections(alreadyChosen, cardZoneSelectionsFor(chosen, choice.domain)...)
		return remainingGraveyardCostsPayable(s, playerID, remainingCosts, xValue, reserved, sourceCardID, sourceZone)
	})
	return cardZoneSelectionsFor(selected, choice.domain)
}

func preferredExileCards(s State, playerID game.PlayerID, additional cost.Additional, amount int, alreadyChosen []cardZoneSelection, remainingCosts []cost.Additional, xValue int, sourceCardID id.ID, sourceZone zone.Type, prefs *Preferences) []cardZoneSelection {
	if prefs == nil || len(prefs.ExileChoices) == 0 {
		return chooseExileCards(s, playerID, additional, amount, alreadyChosen, remainingCosts, xValue, sourceCardID, sourceZone)
	}
	selectionZone := additionalCostSourceZone(additional.Source)
	chosenIDs := make(map[id.ID]bool)
	for _, chosen := range alreadyChosen {
		chosenIDs[chosen.cardID] = true
	}
	var chosen []cardZoneSelection
	var consumed int
	for _, cardID := range prefs.ExileChoices {
		card, ok := s.CardInstance(cardID)
		if !ok || !zoneContainsCard(s, playerID, selectionZone, cardID) || chosenIDs[cardID] || !additionalCostMatchesCard(s, s.CardFace(card, game.FaceFront), additional) || (additional.ExcludeSource && cardID == sourceCardID) {
			return nil
		}
		chosen = append(chosen, cardZoneSelection{cardID: cardID, zone: selectionZone})
		chosenIDs[cardID] = true
		consumed++
		if len(chosen) == amount {
			prefs.ExileChoices = prefs.ExileChoices[consumed:]
			return chosen
		}
	}
	return nil
}

func remainingGraveyardCostsPayable(s State, playerID game.PlayerID, remainingCosts []cost.Additional, xValue int, alreadyChosen []cardZoneSelection, sourceCardID id.ID, sourceZone zone.Type) bool {
	for i, additional := range remainingCosts {
		amount := AdditionalCostAmountFor(additional, xValue)
		if amount < 0 {
			return false
		}
		switch additional.Kind {
		case cost.AdditionalCollectEvidence:
			if amount <= 0 {
				return false
			}
			return len(chooseEvidenceCards(s, playerID, amount, alreadyChosen, remainingCosts[i+1:], xValue, sourceCardID, sourceZone)) > 0
		case cost.AdditionalExile:
			if additional.TotalManaValueAtLeast > 0 {
				if additionalCostSourceZone(additional.Source) != zone.Graveyard {
					continue
				}
				return len(chooseThresholdExileCards(s, playerID, additional, alreadyChosen, remainingCosts[i+1:], xValue, sourceCardID, sourceZone)) > 0
			}
			if amount == 0 || additionalCostSourceZone(additional.Source) != zone.Graveyard {
				continue
			}
			return len(chooseExileCards(s, playerID, additional, amount, alreadyChosen, remainingCosts[i+1:], xValue, sourceCardID, sourceZone)) == amount
		case cost.AdditionalExileSource:
			if amount != 1 || sourceCardID == 0 || sourceZone != zone.Graveyard {
				continue
			}
			for _, chosen := range alreadyChosen {
				if chosen.cardID == sourceCardID {
					return false
				}
			}
			card, ok := s.CardInstance(sourceCardID)
			if !ok || !zoneContainsCard(s, playerID, sourceZone, sourceCardID) || !additionalCostMatchesCard(s, s.CardFace(card, game.FaceFront), additional) {
				return false
			}
			return remainingGraveyardCostsPayable(
				s,
				playerID,
				remainingCosts[i+1:],
				xValue,
				appendCardZoneSelections(alreadyChosen, cardZoneSelection{cardID: sourceCardID, zone: sourceZone}),
				sourceCardID,
				sourceZone,
			)
		default:
			// Only the selection-bearing additional cost kinds above carry card
			// selection; other kinds need no handling here.
		}
	}
	return true
}

func appendCardZoneSelections(base []cardZoneSelection, added ...cardZoneSelection) []cardZoneSelection {
	result := make([]cardZoneSelection, 0, len(base)+len(added))
	result = append(result, base...)
	result = append(result, added...)
	return result
}

func cardZoneSelectionsInclude(selections []cardZoneSelection, cardID id.ID) bool {
	return slices.ContainsFunc(selections, func(selection cardZoneSelection) bool {
		return selection.cardID == cardID
	})
}

func hasDuplicateCardZoneSelections(selections []cardZoneSelection) bool {
	seen := make(map[id.ID]bool, len(selections))
	for _, selection := range selections {
		if seen[selection.cardID] {
			return true
		}
		seen[selection.cardID] = true
	}
	return false
}

// chooseRevealCards deterministically selects amount cards to reveal for a
// reveal cost, drawing from the cost's source zone (defaulting to the hand) and
// honoring the cost filter. It backs both the no-preference path and the
// invalid-preference fallback so a stale reveal preference degrades to a legal
// reveal rather than rejecting the payment.
func chooseRevealCards(s State, playerID game.PlayerID, additional cost.Additional, amount int, alreadyChosen []cardZoneSelection) []cardZoneSelection {
	sourceZone := additional.Source
	if sourceZone == zone.None {
		sourceZone = zone.Hand
	}
	additional.Source = sourceZone
	return chooseExileCards(s, playerID, additional, amount, alreadyChosen, nil, 0, 0, zone.None)
}

func preferredRevealCards(s State, playerID game.PlayerID, additional cost.Additional, amount int, alreadyChosen []cardZoneSelection, prefs *Preferences) []cardZoneSelection {
	sourceZone := additional.Source
	if sourceZone == zone.None {
		sourceZone = zone.Hand
	}
	if prefs == nil || len(prefs.RevealChoices) == 0 {
		return chooseRevealCards(s, playerID, additional, amount, alreadyChosen)
	}
	chosenIDs := make(map[id.ID]bool, len(alreadyChosen))
	for _, chosen := range alreadyChosen {
		chosenIDs[chosen.cardID] = true
	}
	var chosen []cardZoneSelection
	var consumed int
	for _, cardID := range prefs.RevealChoices {
		card, ok := s.CardInstance(cardID)
		if !ok ||
			!zoneContainsCard(s, playerID, sourceZone, cardID) ||
			chosenIDs[cardID] ||
			!additionalCostMatchesCard(s, s.CardFace(card, game.FaceFront), additional) {
			return nil
		}
		chosen = append(chosen, cardZoneSelection{cardID: cardID, zone: sourceZone})
		chosenIDs[cardID] = true
		consumed++
		if len(chosen) == amount {
			prefs.RevealChoices = prefs.RevealChoices[consumed:]
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

func chooseSacrificePermanents(s State, playerID game.PlayerID, additional cost.Additional, amount int, alreadyChosen []*game.Permanent, source *game.Permanent) []*game.Permanent {
	choice, ok := objectCostChoiceForCost(additional)
	if !ok {
		return nil
	}
	candidates := candidatePermanentsForObjectCost(s, playerID, choice, source, reservedPermanentIDs(alreadyChosen))
	return truncatePermanents(candidates, amount)
}

func preferredSacrificePermanents(s State, playerID game.PlayerID, additional cost.Additional, amount int, alreadyChosen []*game.Permanent, prefs *Preferences, source *game.Permanent) []*game.Permanent {
	if prefs == nil || len(prefs.SacrificeChoices) == 0 {
		return chooseSacrificePermanents(s, playerID, additional, amount, alreadyChosen, source)
	}
	choice, ok := objectCostChoiceForCost(additional)
	if !ok {
		return nil
	}
	chosenIDs := reservedPermanentIDs(alreadyChosen)
	var chosen []*game.Permanent
	var consumed int
	for _, permanentID := range prefs.SacrificeChoices {
		permanent, ok := s.PermanentByObjectID(permanentID)
		if !ok || chosenIDs[permanentID] || !permanentSatisfiesObjectCost(s, playerID, permanent, choice, source) {
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

func chooseTapPermanents(s State, playerID game.PlayerID, additional cost.Additional, amount int, alreadyChosen []*game.Permanent) []*game.Permanent {
	choice, ok := objectCostChoiceForCost(additional)
	if !ok {
		return nil
	}
	candidates := candidatePermanentsForObjectCost(s, playerID, choice, nil, reservedPermanentIDs(alreadyChosen))
	return truncatePermanents(candidates, amount)
}

func preferredTapPermanents(s State, playerID game.PlayerID, additional cost.Additional, amount int, alreadyChosen []*game.Permanent, prefs *Preferences) []*game.Permanent {
	if prefs == nil || len(prefs.TapChoices) == 0 {
		return chooseTapPermanents(s, playerID, additional, amount, alreadyChosen)
	}
	choice, ok := objectCostChoiceForCost(additional)
	if !ok {
		return nil
	}
	chosenIDs := reservedPermanentIDs(alreadyChosen)
	var chosen []*game.Permanent
	var consumed int
	for _, permanentID := range prefs.TapChoices {
		permanent, ok := s.PermanentByObjectID(permanentID)
		if !ok || chosenIDs[permanentID] || !permanentSatisfiesObjectCost(s, playerID, permanent, choice, nil) {
			return nil
		}
		chosen = append(chosen, permanent)
		chosenIDs[permanentID] = true
		consumed++
		if len(chosen) == amount {
			prefs.TapChoices = prefs.TapChoices[consumed:]
			return chosen
		}
	}
	return nil
}

// chooseTapPermanentsTotalPower selects untapped matching permanents the player
// controls whose total power reaches additional.TotalPowerAtLeast, honoring
// ExcludeSource. It prefers tapping the fewest creatures by taking the highest
// power first, and returns nil when the threshold cannot be reached.
func chooseTapPermanentsTotalPower(s State, playerID game.PlayerID, additional cost.Additional, alreadyChosen []*game.Permanent, source *game.Permanent) []*game.Permanent {
	choice, ok := objectCostChoiceForCost(additional)
	if !ok {
		return nil
	}
	candidates := candidatePermanentsForObjectCost(s, playerID, choice, source, reservedPermanentIDs(alreadyChosen))
	slices.SortStableFunc(candidates, func(a, b *game.Permanent) int {
		return permanentPowerContribution(s, b, additional.PowerContribution) -
			permanentPowerContribution(s, a, additional.PowerContribution)
	})
	var chosen []*game.Permanent
	total := 0
	for _, permanent := range candidates {
		if total >= additional.TotalPowerAtLeast {
			break
		}
		chosen = append(chosen, permanent)
		total += permanentPowerContribution(s, permanent, additional.PowerContribution)
	}
	if total < additional.TotalPowerAtLeast {
		return nil
	}
	return chosen
}

// preferredTapPermanentsTotalPower honors an explicit tap selection from prefs
// for a total-power tap cost, falling back to chooseTapPermanentsTotalPower.
func preferredTapPermanentsTotalPower(s State, playerID game.PlayerID, additional cost.Additional, alreadyChosen []*game.Permanent, source *game.Permanent, prefs *Preferences) []*game.Permanent {
	if prefs == nil || len(prefs.TapChoices) == 0 {
		return chooseTapPermanentsTotalPower(s, playerID, additional, alreadyChosen, source)
	}
	choice, ok := objectCostChoiceForCost(additional)
	if !ok {
		return nil
	}
	chosenIDs := reservedPermanentIDs(alreadyChosen)
	var chosen []*game.Permanent
	total := 0
	var consumed int
	for _, permanentID := range prefs.TapChoices {
		permanent, ok := s.PermanentByObjectID(permanentID)
		if !ok || chosenIDs[permanentID] || !permanentSatisfiesObjectCost(s, playerID, permanent, choice, source) {
			return nil
		}
		chosen = append(chosen, permanent)
		chosenIDs[permanentID] = true
		total += permanentPowerContribution(s, permanent, additional.PowerContribution)
		consumed++
		if total >= additional.TotalPowerAtLeast {
			prefs.TapChoices = prefs.TapChoices[consumed:]
			return chosen
		}
	}
	return nil
}

func permanentPowerContribution(s State, permanent *game.Permanent, kind cost.PowerContributionKind) int {
	power := s.PermanentPower(permanent)
	if kind != cost.PowerContributionCrew {
		return power
	}
	for _, ability := range s.PermanentEffectiveAbilities(permanent) {
		if static, ok := ability.(*game.StaticAbility); ok {
			power += static.CrewPowerBonus
		}
	}
	return power
}

// costSourcePermanentByCardID resolves an object cost's ExcludeSource source
// permanent for a generic payment that carries only the source card ID.
func costSourcePermanentByCardID(s State, sourceCardID id.ID) *game.Permanent {
	if sourceCardID == 0 {
		return nil
	}
	for _, permanent := range s.Battlefield() {
		if permanent.CardInstanceID == sourceCardID {
			return permanent
		}
	}
	return nil
}

func chooseReturnPermanents(s State, playerID game.PlayerID, additional cost.Additional, amount int, alreadyChosen []*game.Permanent, source *game.Permanent) []*game.Permanent {
	choice, ok := objectCostChoiceForCost(additional)
	if !ok {
		return nil
	}
	candidates := candidatePermanentsForObjectCost(s, playerID, choice, source, reservedPermanentIDs(alreadyChosen))
	return truncatePermanents(candidates, amount)
}

func preferredReturnPermanents(s State, playerID game.PlayerID, additional cost.Additional, amount int, alreadyChosen []*game.Permanent, prefs *Preferences, source *game.Permanent) []*game.Permanent {
	if prefs == nil || len(prefs.ReturnChoices) == 0 {
		return chooseReturnPermanents(s, playerID, additional, amount, alreadyChosen, source)
	}
	choice, ok := objectCostChoiceForCost(additional)
	if !ok {
		return nil
	}
	chosenIDs := reservedPermanentIDs(alreadyChosen)
	var chosen []*game.Permanent
	var consumed int
	for _, permanentID := range prefs.ReturnChoices {
		permanent, ok := s.PermanentByObjectID(permanentID)
		if !ok || chosenIDs[permanentID] || !permanentSatisfiesObjectCost(s, playerID, permanent, choice, source) {
			return nil
		}
		chosen = append(chosen, permanent)
		chosenIDs[permanentID] = true
		consumed++
		if len(chosen) == amount {
			prefs.ReturnChoices = prefs.ReturnChoices[consumed:]
			return chosen
		}
	}
	return nil
}

func chooseDiscardCards(s State, playerID game.PlayerID, additional cost.Additional, amount int, alreadyChosen []id.ID) []id.ID {
	choice, ok := cardCostChoiceForCost(additional, 0)
	if !ok {
		return nil
	}
	candidates := candidateCardsForObjectCost(s, playerID, choice, reservedCardIDSet(alreadyChosen))
	return truncateCardIDs(candidates, amount)
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
		if !ok || !player.Hand.Contains(cardID) || chosenIDs[cardID] || !additionalCostMatchesCard(s, s.CardFace(card, game.FaceFront), additional) {
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

// planRemoveCounterAmong plans the counter removals for an
// AdditionalRemoveCounterAmong cost, spreading amount counters of the cost's
// kind across permanents the player controls that match the cost constraint.
// When the preferences carry an explicit selection it is honored; if that
// selection is stale or illegal it falls back to the greedy choice (unless
// strict replay is demanded), matching the engine's uniform invalid-preference
// policy. With no preference the removals are chosen greedily. It returns false
// when the player cannot supply amount matching counters.
func planRemoveCounterAmong(s State, playerID game.PlayerID, additional cost.Additional, amount int, alreadyPlanned []counterRemoval, prefs *Preferences) ([]counterRemoval, bool) {
	reserved := plannedCounterRemovalsBySourceKind(alreadyPlanned)
	hadPreference := prefs != nil && len(prefs.RemoveCounterChoices) > 0
	if hadPreference {
		if removals, ok := preferredRemoveCounterAmong(s, playerID, additional, amount, reserved, prefs); ok {
			return removals, true
		}
		if !preferenceFallbackAllowed(prefs, hadPreference) {
			return nil, false
		}
	}
	return greedyRemoveCounterAmong(s, playerID, additional, amount, reserved)
}

// counterSourceKind keys a planned counter removal by both its source permanent
// and counter kind, so a generic any-kind among-removal can reserve counters of
// several kinds on the same permanent without conflating their counts.
type counterSourceKind struct {
	source *game.Permanent
	kind   counter.Kind
}

// plannedCounterRemovalsBySourceKind totals, per (source, kind), the counters
// already reserved by earlier planned removals so the same counters are not
// spent twice.
func plannedCounterRemovalsBySourceKind(planned []counterRemoval) map[counterSourceKind]int {
	reserved := make(map[counterSourceKind]int)
	for _, removal := range planned {
		reserved[counterSourceKind{source: removal.source, kind: removal.kind}] += removal.amount
	}
	return reserved
}

// presentCounterKinds returns every counter kind present on a permanent in a
// stable order, sorted by kind value so selection is deterministic.
func presentCounterKinds(permanent *game.Permanent) []counter.Kind {
	present := permanent.Counters.All()
	kinds := make([]counter.Kind, 0, len(present))
	for kind := range present {
		kinds = append(kinds, kind)
	}
	slices.Sort(kinds)
	return kinds
}

// amongCounterKinds returns the counter kinds an among-removal cost may take
// from a permanent, in a stable order. A kind-specific cost yields the single
// named kind; the generic any-kind cost yields every kind present on the
// permanent, sorted by kind value so selection is deterministic.
func amongCounterKinds(permanent *game.Permanent, additional cost.Additional) []counter.Kind {
	if !additional.AnyCounterKind {
		return []counter.Kind{additional.CounterKind}
	}
	return presentCounterKinds(permanent)
}

// planRemoveCounterFromSource plans the removal of amount counters from a single
// source permanent for a generic any-kind "remove N counters from this
// permanent" cost. It takes counters across the kinds present in stable order,
// honoring counters already reserved by earlier planned removals, and returns
// false when the source cannot supply amount counters.
func planRemoveCounterFromSource(source *game.Permanent, amount int, alreadyPlanned []counterRemoval) ([]counterRemoval, bool) {
	if amount <= 0 {
		return nil, false
	}
	reserved := plannedCounterRemovalsBySourceKind(alreadyPlanned)
	var removals []counterRemoval
	remaining := amount
	for _, kind := range presentCounterKinds(source) {
		if remaining == 0 {
			break
		}
		available := source.Counters.Get(kind) - reserved[counterSourceKind{source: source, kind: kind}]
		if available <= 0 {
			continue
		}
		take := min(available, remaining)
		removals = append(removals, counterRemoval{source: source, kind: kind, amount: take})
		remaining -= take
	}
	if remaining > 0 {
		return nil, false
	}
	return removals, true
}

// RemovableAmongCounterCount reports how many counters a permanent can supply
// toward an among-removal cost: the named kind's count for a kind-specific cost,
// or the total of all counters for the generic any-kind cost.
func RemovableAmongCounterCount(permanent *game.Permanent, additional cost.Additional) int {
	if !additional.AnyCounterKind {
		return permanent.Counters.Get(additional.CounterKind)
	}
	total := 0
	for _, count := range permanent.Counters.All() {
		total += count
	}
	return total
}

// greedyRemoveCounterAmong removes counters from matching controlled permanents
// in battlefield order, taking as many as available from each until amount is
// reached.
func greedyRemoveCounterAmong(s State, playerID game.PlayerID, additional cost.Additional, amount int, reserved map[counterSourceKind]int) ([]counterRemoval, bool) {
	choice, ok := objectCostChoiceForCost(additional)
	if !ok {
		return nil, false
	}
	var removals []counterRemoval
	remaining := amount
	for _, permanent := range candidatePermanentsForObjectCost(s, playerID, choice, nil, nil) {
		if remaining == 0 {
			break
		}
		for _, kind := range amongCounterKinds(permanent, additional) {
			if remaining == 0 {
				break
			}
			available := permanent.Counters.Get(kind) - reserved[counterSourceKind{source: permanent, kind: kind}]
			if available <= 0 {
				continue
			}
			take := min(available, remaining)
			removals = append(removals, counterRemoval{source: permanent, kind: kind, amount: take})
			remaining -= take
		}
	}
	if remaining > 0 {
		return nil, false
	}
	return removals, true
}

// preferredRemoveCounterAmong honors an explicit per-counter selection from
// prefs, consuming amount entries from RemoveCounterChoices. Each entry names a
// permanent to lose one counter; repeated entries remove several counters from
// the same permanent. For an any-kind cost the kind removed from each chosen
// permanent is the first still-available kind in stable order. It fails closed
// when an entry is invalid or insufficient entries are supplied.
func preferredRemoveCounterAmong(s State, playerID game.PlayerID, additional cost.Additional, amount int, reserved map[counterSourceKind]int, prefs *Preferences) ([]counterRemoval, bool) {
	choice, ok := objectCostChoiceForCost(additional)
	if !ok {
		return nil, false
	}
	used := make(map[counterSourceKind]int)
	var order []counterSourceKind
	consumed := 0
	for _, permanentID := range prefs.RemoveCounterChoices {
		if consumed == amount {
			break
		}
		permanent, ok := s.PermanentByObjectID(permanentID)
		if !ok || !permanentSatisfiesObjectCost(s, playerID, permanent, choice, nil) {
			return nil, false
		}
		kind, ok := chooseAmongCounterKind(permanent, additional, reserved, used)
		if !ok {
			return nil, false
		}
		key := counterSourceKind{source: permanent, kind: kind}
		if used[key] == 0 {
			order = append(order, key)
		}
		used[key]++
		consumed++
	}
	if consumed != amount {
		return nil, false
	}
	prefs.RemoveCounterChoices = prefs.RemoveCounterChoices[consumed:]
	removals := make([]counterRemoval, 0, len(order))
	for _, key := range order {
		removals = append(removals, counterRemoval{source: key.source, kind: key.kind, amount: used[key]})
	}
	return removals, true
}

// chooseAmongCounterKind picks the kind to remove from a permanent for an
// among-removal cost, returning the first kind in stable order that still has an
// unreserved, unused counter. It fails closed when none remains.
func chooseAmongCounterKind(permanent *game.Permanent, additional cost.Additional, reserved, used map[counterSourceKind]int) (counter.Kind, bool) {
	for _, kind := range amongCounterKinds(permanent, additional) {
		key := counterSourceKind{source: permanent, kind: kind}
		if permanent.Counters.Get(kind) > reserved[key]+used[key] {
			return kind, true
		}
	}
	return 0, false
}
