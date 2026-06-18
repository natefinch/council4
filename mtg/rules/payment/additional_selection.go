package payment

import (
	"slices"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
)

func chooseEvidenceCards(s State, playerID game.PlayerID, threshold int, alreadyChosen []cardZoneSelection, remainingCosts []cost.Additional, xValue int, sourceCardID id.ID, sourceZone zone.Type) []cardZoneSelection {
	player, ok := s.Player(playerID)
	if !ok {
		return nil
	}
	chosenIDs := make(map[id.ID]bool, len(alreadyChosen))
	for _, chosen := range alreadyChosen {
		chosenIDs[chosen.cardID] = true
	}
	type evidenceCandidate struct {
		cardID    id.ID
		manaValue int
	}
	var candidates []evidenceCandidate
	for _, cardID := range player.Graveyard.All() {
		if chosenIDs[cardID] {
			continue
		}
		manaValue, ok := evidenceCardManaValue(s, cardID)
		if !ok || manaValue <= 0 {
			continue
		}
		candidates = append(candidates, evidenceCandidate{cardID: cardID, manaValue: manaValue})
	}
	slices.SortStableFunc(candidates, func(a, b evidenceCandidate) int {
		switch {
		case a.manaValue > b.manaValue:
			return -1
		case a.manaValue < b.manaValue:
			return 1
		default:
			return 0
		}
	})
	var search func(start int, total int, chosen []cardZoneSelection) []cardZoneSelection
	search = func(start int, total int, chosen []cardZoneSelection) []cardZoneSelection {
		if total >= threshold {
			reserved := appendCardZoneSelections(alreadyChosen, chosen...)
			if remainingGraveyardCostsPayable(s, playerID, remainingCosts, xValue, reserved, sourceCardID, sourceZone) {
				return append([]cardZoneSelection(nil), chosen...)
			}
			return nil
		}
		for i := start; i < len(candidates); i++ {
			candidate := candidates[i]
			next := slices.Clone(chosen)
			next = append(next, cardZoneSelection{cardID: candidate.cardID, zone: zone.Graveyard})
			if selected := search(i+1, total+candidate.manaValue, next); len(selected) > 0 {
				return selected
			}
		}
		return nil
	}
	return search(0, 0, nil)
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

func chooseExileCards(s State, playerID game.PlayerID, additional cost.Additional, amount int, alreadyChosen []cardZoneSelection, remainingCosts []cost.Additional, xValue int, sourceCardID id.ID, sourceZone zone.Type) []cardZoneSelection {
	player, ok := s.Player(playerID)
	if !ok {
		return nil
	}
	selectionZone := additionalCostSourceZone(additional.Source)
	chosenIDs := make(map[id.ID]bool)
	for _, chosen := range alreadyChosen {
		chosenIDs[chosen.cardID] = true
	}
	var candidates []cardZoneSelection
	for _, cardID := range cardIDsInZone(player, selectionZone) {
		if chosenIDs[cardID] {
			continue
		}
		card, ok := s.CardInstance(cardID)
		if !ok || !additionalCostMatchesCard(s.CardFace(card, game.FaceFront), additional) {
			continue
		}
		candidates = append(candidates, cardZoneSelection{cardID: cardID, zone: selectionZone})
	}
	return chooseFixedCardZoneSelection(candidates, amount, func(chosen []cardZoneSelection) bool {
		reserved := appendCardZoneSelections(alreadyChosen, chosen...)
		return remainingGraveyardCostsPayable(s, playerID, remainingCosts, xValue, reserved, sourceCardID, sourceZone)
	})
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
		if !ok || !zoneContainsCard(s, playerID, selectionZone, cardID) || chosenIDs[cardID] || !additionalCostMatchesCard(s.CardFace(card, game.FaceFront), additional) {
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

func chooseFixedCardZoneSelection(candidates []cardZoneSelection, amount int, allowsRemaining func([]cardZoneSelection) bool) []cardZoneSelection {
	if amount == 0 {
		if allowsRemaining(nil) {
			return nil
		}
		return nil
	}
	var search func(start int, chosen []cardZoneSelection) []cardZoneSelection
	search = func(start int, chosen []cardZoneSelection) []cardZoneSelection {
		if len(chosen) == amount {
			if allowsRemaining(chosen) {
				return append([]cardZoneSelection(nil), chosen...)
			}
			return nil
		}
		remainingNeeded := amount - len(chosen)
		for i := start; i <= len(candidates)-remainingNeeded; i++ {
			next := slices.Clone(chosen)
			next = append(next, candidates[i])
			if selected := search(i+1, next); len(selected) == amount {
				return selected
			}
		}
		return nil
	}
	return search(0, nil)
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
			if !ok || !zoneContainsCard(s, playerID, sourceZone, sourceCardID) || !additionalCostMatchesCard(s.CardFace(card, game.FaceFront), additional) {
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

func preferredRevealCards(s State, playerID game.PlayerID, additional cost.Additional, amount int, alreadyChosen []cardZoneSelection, prefs *Preferences) []cardZoneSelection {
	sourceZone := additional.Source
	if sourceZone == zone.None {
		sourceZone = zone.Hand
	}
	if prefs == nil || len(prefs.RevealChoices) == 0 {
		additional.Source = sourceZone
		return chooseExileCards(s, playerID, additional, amount, alreadyChosen, nil, 0, 0, zone.None)
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
			!additionalCostMatchesCard(s.CardFace(card, game.FaceFront), additional) {
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
	chosenIDs := make(map[id.ID]bool)
	for _, permanent := range alreadyChosen {
		chosenIDs[permanent.ObjectID] = true
	}
	var chosen []*game.Permanent
	for _, permanent := range s.Battlefield() {
		if s.EffectiveController(permanent) != playerID || chosenIDs[permanent.ObjectID] {
			continue
		}
		if additional.ExcludeSource && source != nil && permanent.ObjectID == source.ObjectID {
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

func preferredSacrificePermanents(s State, playerID game.PlayerID, additional cost.Additional, amount int, alreadyChosen []*game.Permanent, prefs *Preferences, source *game.Permanent) []*game.Permanent {
	if prefs == nil || len(prefs.SacrificeChoices) == 0 {
		return chooseSacrificePermanents(s, playerID, additional, amount, alreadyChosen, source)
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
		if additional.ExcludeSource && source != nil && permanentID == source.ObjectID {
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
	chosenIDs := make(map[id.ID]bool)
	for _, permanent := range alreadyChosen {
		chosenIDs[permanent.ObjectID] = true
	}
	var chosen []*game.Permanent
	for _, permanent := range s.Battlefield() {
		if permanent.Tapped || s.EffectiveController(permanent) != playerID || chosenIDs[permanent.ObjectID] {
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

func preferredTapPermanents(s State, playerID game.PlayerID, additional cost.Additional, amount int, alreadyChosen []*game.Permanent, prefs *Preferences) []*game.Permanent {
	if prefs == nil || len(prefs.TapChoices) == 0 {
		return chooseTapPermanents(s, playerID, additional, amount, alreadyChosen)
	}
	chosenIDs := make(map[id.ID]bool)
	for _, permanent := range alreadyChosen {
		chosenIDs[permanent.ObjectID] = true
	}
	var chosen []*game.Permanent
	var consumed int
	for _, permanentID := range prefs.TapChoices {
		permanent, ok := s.PermanentByObjectID(permanentID)
		if !ok ||
			permanent.Tapped ||
			s.EffectiveController(permanent) != playerID ||
			chosenIDs[permanentID] ||
			!additionalCostMatchesPermanent(s, permanent, additional) {
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

func chooseReturnPermanents(s State, playerID game.PlayerID, additional cost.Additional, amount int, alreadyChosen []*game.Permanent) []*game.Permanent {
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

func preferredReturnPermanents(s State, playerID game.PlayerID, additional cost.Additional, amount int, alreadyChosen []*game.Permanent, prefs *Preferences) []*game.Permanent {
	if prefs == nil || len(prefs.ReturnChoices) == 0 {
		return chooseReturnPermanents(s, playerID, additional, amount, alreadyChosen)
	}
	chosenIDs := make(map[id.ID]bool)
	for _, permanent := range alreadyChosen {
		chosenIDs[permanent.ObjectID] = true
	}
	var chosen []*game.Permanent
	var consumed int
	for _, permanentID := range prefs.ReturnChoices {
		permanent, ok := s.PermanentByObjectID(permanentID)
		if !ok ||
			s.EffectiveController(permanent) != playerID ||
			chosenIDs[permanentID] ||
			!additionalCostMatchesPermanent(s, permanent, additional) {
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
