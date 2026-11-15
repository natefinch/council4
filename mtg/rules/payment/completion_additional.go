package payment

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

// ComponentChoice fixes all or a prefix of one additional cost's selections.
// Objects are permanent object IDs or card instance IDs, according to the cost.
// Counters identify individual counters; repeated entries remove several.
// Finished distinguishes a completed variable-size selection from a prefix.
type ComponentChoice struct {
	Objects  []id.ID
	Counters []CounterRemovalChoice
	Finished bool
}

func cloneComponent(c ComponentChoice) ComponentChoice {
	c.Objects = slices.Clone(c.Objects)
	c.Counters = slices.Clone(c.Counters)
	return c
}

type additionalWitness struct {
	plan       additionalCostPlan
	components map[int]ComponentChoice
	groups     map[uint8]int
}

type componentReservations struct {
	objects  map[id.ID]bool
	counters map[CounterRemovalChoice]int
}

func buildAdditionalCostPlanForCosts(s State, playerID game.PlayerID, costs []cost.Additional, xValue int, prefs *Preferences, source *game.Permanent, sourceCardID id.ID, sourceZone zone.Type, castingCardID id.ID, tapReservations ...*game.Permanent) (additionalCostPlan, bool) {
	witness, ok := completeAdditional(s, playerID, costs, xValue, prefs, source, sourceCardID, sourceZone, castingCardID, nil, nil, tapReservations, func(additionalWitness) bool { return true })
	return witness.plan, ok
}

// completeAdditional searches selections, not games or mana activation orders.
// A branch is accepted only when all components AND the caller's remaining
// payment are payable. Prefix validation prunes exhausted resources early.
func completeAdditional(s State, playerID game.PlayerID, costs []cost.Additional, xValue int, legacy *Preferences, source *game.Permanent, sourceCardID id.ID, sourceZone zone.Type, castingCardID id.ID, fixed map[int]ComponentChoice, groups map[uint8]int, tapReservations []*game.Permanent, accept func(additionalWitness) bool) (additionalWitness, bool) {
	var result additionalWitness
	components := make(map[int]ComponentChoice)
	selectedGroups := make(map[uint8]int)
	var visit func(int, []cost.Additional, *Preferences) bool
	visit = func(index int, concrete []cost.Additional, prefs *Preferences) bool {
		prefix, ok := buildSelectedAdditionalCostPlan(s, playerID, concrete, xValue, clonePreferences(prefs), source, sourceCardID, sourceZone, castingCardID, tapReservations...)
		if !ok || !remainingFixedResourcesPayable(s, playerID, costs[index:], xValue, source, sourceCardID, sourceZone, castingCardID, prefix, hasTapCostOf(concrete), tapReservations) {
			return false
		}
		if index == len(costs) {
			plan := prefix
			if !additionalLibraryReservationsValid(s, plan) || !additionalResourcesDisjoint(plan) {
				return false
			}
			witness := additionalWitness{plan: plan, components: cloneComponents(components), groups: cloneGroups(selectedGroups)}
			if !accept(witness) {
				return false
			}
			result = witness
			return true
		}
		additional := costs[index]
		group := additional.ChoiceGroup
		if group != 0 {
			if _, chosen := selectedGroups[group]; chosen {
				return visit(index+1, concrete, prefs)
			}
			for member, candidate := range costs {
				if candidate.ChoiceGroup != group {
					continue
				}
				if fixedMember, set := groups[group]; set && fixedMember != member {
					continue
				}
				candidate.ChoiceGroup = 0
				selectedGroups[group] = member
				selection := legacyComponentChoice(s, playerID, legacy, prefs, candidate, xValue, source, fixed[member])
				reserved := componentReservations{objects: componentReservedObjects(prefix, candidate, source, hasTapCostOf(concrete), tapReservations)}
				if candidate.Kind == cost.AdditionalRemoveCounter || candidate.Kind == cost.AdditionalRemoveCounterAmong {
					reserved.counters = counterCompletionReservations(s, playerID, concrete, costs[index+1:], xValue, prefs, source, sourceCardID, sourceZone, castingCardID)
				}
				if selectComponent(s, playerID, candidate, xValue, source, sourceCardID, castingCardID, selection, func(choice ComponentChoice) bool {
					components[member] = cloneComponent(choice)
					nextPrefs := appendComponentPreferences(prefs, candidate, choice)
					return visit(index+1, append(slices.Clone(concrete), candidate), nextPrefs)
				}, reserved) {
					return true
				}
				delete(components, member)
				delete(selectedGroups, group)
			}
			return false
		}
		selection := legacyComponentChoice(s, playerID, legacy, prefs, additional, xValue, source, fixed[index])
		reserved := componentReservations{objects: componentReservedObjects(prefix, additional, source, hasTapCostOf(concrete), tapReservations)}
		if additional.Kind == cost.AdditionalRemoveCounter || additional.Kind == cost.AdditionalRemoveCounterAmong {
			reserved.counters = counterCompletionReservations(s, playerID, concrete, costs[index+1:], xValue, prefs, source, sourceCardID, sourceZone, castingCardID)
		}
		return selectComponent(s, playerID, additional, xValue, source, sourceCardID, castingCardID, selection, func(choice ComponentChoice) bool {
			components[index] = cloneComponent(choice)
			next := append(slices.Clone(concrete), additional)
			nextPrefs := appendComponentPreferences(prefs, additional, choice)
			// Legacy preferences remain hints unless StrictReplay was requested.
			// The new Partial API always fixes choices exactly.
			if !legacyComponentMatches(legacy, prefs, additional, choice) {
				return false
			}
			if _, ok := buildSelectedAdditionalCostPlan(s, playerID, next, xValue, clonePreferences(nextPrefs), source, sourceCardID, sourceZone, castingCardID, tapReservations...); !ok {
				return false
			}
			return visit(index+1, next, nextPrefs)
		}, reserved)
	}
	prefs := &Preferences{StrictReplay: true}
	if legacy != nil {
		prefs.AlternativeIndex = legacy.AlternativeIndex
		prefs.PhyrexianLifeChoices = slices.Clone(legacy.PhyrexianLifeChoices)
		prefs.PhyrexianManaOnly = slices.Clone(legacy.PhyrexianManaOnly)
	}
	ok := visit(0, nil, prefs)
	if !ok && legacy != nil && !legacy.StrictReplay {
		// Preserve the compatibility API's documented stale-hint policy. New
		// callers use Partial, whose fixed choices never take this fallback.
		return completeAdditional(s, playerID, costs, xValue, nil, source, sourceCardID, sourceZone, castingCardID, fixed, groups, tapReservations, accept)
	}
	return result, ok
}

// Reject fixed-count contradictions with shared matching before enumerating any
// subset. Tap and consuming costs use separate resource domains because paying
// a tap does not consume the object. Unresolved alternatives are overapproximated
// here and checked exactly when their branches are selected.
func remainingFixedResourcesPayable(s State, playerID game.PlayerID, costs []cost.Additional, x int, source *game.Permanent, sourceCardID id.ID, sourceZone zone.Type, castingCardID id.ID, prefix additionalCostPlan, sourceTap bool, tapReservations []*game.Permanent) bool {
	var consumed, tapped, cards []resourceCostSlot
	tapReserved := reservedPermanentIDs(append(slices.Clone(prefix.permanentsToTap), tapReservations...))
	if source != nil && sourceTap {
		tapReserved[source.ObjectID] = true
	}
	life, energy, mill := prefix.lifePaid, prefix.energyPaid, prefix.millAmount
	libraryExiles := 0
	for _, exile := range prefix.exiles {
		if exile.zone == zone.Library {
			libraryExiles++
		}
	}
	specificCounters := make(map[CounterRemovalChoice]int)
	sourceCounterCount := 0
	for _, removal := range prefix.counterRemovals {
		if !removal.all {
			specificCounters[CounterRemovalChoice{Permanent: removal.source.ObjectID, Kind: removal.kind}] += removal.amount
			if removal.source == source {
				sourceCounterCount += removal.amount
			}
		}
	}
	for _, additional := range costs {
		if additional.ChoiceGroup != 0 {
			continue
		}
		amount := additionalCostAmountIn(s, playerID, additional, x)
		if amount < 0 {
			return false
		}
		if amount == 0 {
			continue
		}
		switch additional.Kind {
		case cost.AdditionalTap:
			if source != nil {
				tapReserved[source.ObjectID] = true
			}
		case cost.AdditionalSacrifice, cost.AdditionalReturnToHand, cost.AdditionalReturnUnblockedAttacker:
			candidates := CandidatePermanentsForCost(s, playerID, additional, returnCostSource(s, additional, source, sourceCardID))
			consumed = append(consumed, resourceCostSlot{candidates: permanentIDs(candidates), amount: amount})
		case cost.AdditionalSacrificeSource:
			if source == nil {
				return false
			}
			consumed = append(consumed, resourceCostSlot{candidates: []id.ID{source.ObjectID}, amount: amount})
		case cost.AdditionalTapPermanents:
			if additional.TotalPowerAtLeast == 0 {
				candidates := CandidatePermanentsForCost(s, playerID, additional, source)
				tapped = append(tapped, resourceCostSlot{candidates: permanentIDs(candidates), amount: amount})
			}
		case cost.AdditionalDiscard, cost.AdditionalExile:
			if additional.TotalManaValueAtLeast > 0 {
				continue
			}
			candidates := CandidateCardsForCost(s, playerID, additional, sourceCardID, castingCardID)
			if additional.Random {
				player, ok := s.Player(playerID)
				if !ok {
					return false
				}
				candidates = slices.DeleteFunc(slices.Clone(player.Hand.All()), func(card id.ID) bool { return card == castingCardID })
			}
			cards = append(cards, resourceCostSlot{candidates: candidates, amount: amount})
			if additional.Kind == cost.AdditionalExile && cardCostZone(additional) == zone.Library {
				libraryExiles += amount
			}
		case cost.AdditionalExileSource:
			if sourceZone == zone.Battlefield {
				if source == nil {
					return false
				}
				consumed = append(consumed, resourceCostSlot{candidates: []id.ID{source.ObjectID}, amount: amount})
			} else {
				if sourceCardID == 0 {
					return false
				}
				cards = append(cards, resourceCostSlot{candidates: []id.ID{sourceCardID}, amount: amount})
				if sourceZone == zone.Library {
					libraryExiles += amount
				}
			}
		case cost.AdditionalPayLife:
			life += amount
		case cost.AdditionalEnergy:
			energy += amount
		case cost.AdditionalMill:
			mill += amount
		case cost.AdditionalRemoveCounter:
			if source == nil {
				return false
			}
			if !additional.RemoveAll {
				sourceCounterCount += amount
				if !additional.AnyCounterKind {
					specificCounters[CounterRemovalChoice{Permanent: source.ObjectID, Kind: additional.CounterKind}] += amount
				}
			}
		}
	}
	player, ok := s.Player(playerID)
	if !ok || player.Life < life || life > 0 && !s.CanPayLife(playerID) || player.EnergyCounters < energy || player.Library.Size() < mill+libraryExiles {
		return false
	}
	for choice, amount := range specificCounters {
		permanent, ok := s.PermanentByObjectID(choice.Permanent)
		if !ok || permanent.Counters.Get(choice.Kind) < amount {
			return false
		}
	}
	if source != nil {
		available := 0
		for _, count := range source.Counters.All() {
			available += count
		}
		if sourceCounterCount > available {
			return false
		}
	}
	if prefix.randomDiscardAmount > 0 {
		candidates := slices.DeleteFunc(slices.Clone(player.Hand.All()), func(card id.ID) bool { return card == castingCardID })
		cards = append(cards, resourceCostSlot{candidates: candidates, amount: prefix.randomDiscardAmount})
	}
	reservedCards := reservedFromSelections(plannedEvidenceCards(prefix))
	for _, card := range prefix.discards {
		reservedCards[card] = true
	}
	return resourceSlotsAssignable(consumed, reservedPermanentIDs(plannedBattlefieldCosts(prefix))) &&
		resourceSlotsAssignable(tapped, tapReserved) && resourceSlotsAssignable(cards, reservedCards)
}

func additionalResourcesDisjoint(plan additionalCostPlan) bool {
	permanents := make(map[id.ID]bool)
	for _, permanent := range plannedBattlefieldCosts(plan) {
		if permanents[permanent.ObjectID] {
			return false
		}
		permanents[permanent.ObjectID] = true
	}
	cards := make(map[id.ID]bool)
	for _, card := range plan.discards {
		if cards[card] {
			return false
		}
		cards[card] = true
	}
	for _, card := range plannedEvidenceCards(plan) {
		if cards[card.cardID] {
			return false
		}
		cards[card.cardID] = true
	}
	return true
}

func legacyComponentChoice(s State, playerID game.PlayerID, legacy, before *Preferences, additional cost.Additional, x int, source *game.Permanent, fixed ComponentChoice) ComponentChoice {
	if legacy == nil || fixed.Finished || len(fixed.Objects) > 0 || len(fixed.Counters) > 0 {
		return fixed
	}
	amount := additionalCostAmountIn(s, playerID, additional, x)
	var objects []id.ID
	var offset int
	switch additional.Kind {
	case cost.AdditionalSacrifice:
		objects, offset = legacy.SacrificeChoices, len(before.SacrificeChoices)
	case cost.AdditionalTapPermanents:
		objects, offset = legacy.TapChoices, len(before.TapChoices)
	case cost.AdditionalReturnToHand, cost.AdditionalReturnUnblockedAttacker:
		objects, offset = legacy.ReturnChoices, len(before.ReturnChoices)
	case cost.AdditionalDiscard:
		objects, offset = legacy.DiscardChoices, len(before.DiscardChoices)
	case cost.AdditionalExile:
		objects, offset = legacy.ExileChoices, len(before.ExileChoices)
	case cost.AdditionalReveal:
		objects, offset = legacy.RevealChoices, len(before.RevealChoices)
	case cost.AdditionalCollectEvidence:
		objects, offset = legacy.EvidenceChoices, len(before.EvidenceChoices)
	case cost.AdditionalRemoveCounterAmong:
		offset = len(before.RemoveCounterChoices)
		if offset < len(legacy.RemoveCounterChoices) {
			fixed.Counters = slices.Clone(legacy.RemoveCounterChoices[offset:min(offset+amount, len(legacy.RemoveCounterChoices))])
		}
		return fixed
	case cost.AdditionalRemoveCounter:
		if !additional.AnyCounterKind {
			return fixed
		}
		offset = len(before.SourceCounterKindChoices)
		if additional.AnyCounterKind && source != nil && offset < len(legacy.SourceCounterKindChoices) {
			for _, kind := range legacy.SourceCounterKindChoices[offset:min(offset+amount, len(legacy.SourceCounterKindChoices))] {
				fixed.Counters = append(fixed.Counters, CounterRemovalChoice{Permanent: source.ObjectID, Kind: kind})
			}
		}
		return fixed
	}
	if offset >= len(objects) || amount <= 0 {
		return fixed
	}
	limit := min(offset+amount, len(objects))
	threshold := additional.TotalManaValueAtLeast
	if additional.Kind == cost.AdditionalCollectEvidence {
		threshold = amount
	}
	if additional.TotalPowerAtLeast > 0 || threshold > 0 {
		limit = offset
		total := 0
		for limit < len(objects) {
			object := objects[limit]
			limit++
			if additional.TotalPowerAtLeast > 0 {
				permanent, ok := s.PermanentByObjectID(object)
				if !ok {
					break
				}
				total += permanentPowerContribution(s, permanent, additional.PowerContribution)
				if total >= additional.TotalPowerAtLeast {
					break
				}
			} else {
				value, _ := evidenceCardManaValue(s, object)
				total += value
				if total >= threshold {
					break
				}
			}
		}
	}
	fixed.Objects = slices.Clone(objects[offset:limit])
	return fixed
}

func additionalLibraryReservationsValid(s State, plan additionalCostPlan) bool {
	player, ok := s.Player(plan.player)
	if !ok {
		return false
	}
	available := player.Library.Size()
	for _, exile := range plan.exiles {
		if exile.zone == zone.Library {
			available--
		}
	}
	return available >= plan.millAmount
}

func appendComponentPreferences(prefs *Preferences, additional cost.Additional, choice ComponentChoice) *Preferences {
	next := clonePreferences(prefs)
	switch additional.Kind {
	case cost.AdditionalSacrifice:
		next.SacrificeChoices = append(next.SacrificeChoices, choice.Objects...)
	case cost.AdditionalTapPermanents:
		next.TapChoices = append(next.TapChoices, choice.Objects...)
		if additional.TotalPowerAtLeast > 0 {
			next.tapComponentCounts = append(next.tapComponentCounts, len(choice.Objects))
		}
	case cost.AdditionalReturnToHand, cost.AdditionalReturnUnblockedAttacker:
		next.ReturnChoices = append(next.ReturnChoices, choice.Objects...)
	case cost.AdditionalDiscard:
		next.DiscardChoices = append(next.DiscardChoices, choice.Objects...)
	case cost.AdditionalExile:
		next.ExileChoices = append(next.ExileChoices, choice.Objects...)
		if additional.TotalManaValueAtLeast > 0 {
			next.exileComponentCounts = append(next.exileComponentCounts, len(choice.Objects))
		}
	case cost.AdditionalReveal:
		next.RevealChoices = append(next.RevealChoices, choice.Objects...)
	case cost.AdditionalCollectEvidence:
		next.EvidenceChoices = append(next.EvidenceChoices, choice.Objects...)
		next.evidenceComponentCounts = append(next.evidenceComponentCounts, len(choice.Objects))
	case cost.AdditionalRemoveCounterAmong:
		next.RemoveCounterChoices = append(next.RemoveCounterChoices, choice.Counters...)
	case cost.AdditionalRemoveCounter:
		if additional.AnyCounterKind {
			for _, removal := range choice.Counters {
				next.SourceCounterKindChoices = append(next.SourceCounterKindChoices, removal.Kind)
			}
		}
	}
	return next
}

func legacyComponentMatches(legacy, before *Preferences, additional cost.Additional, choice ComponentChoice) bool {
	if legacy == nil {
		return true
	}
	var wanted []id.ID
	var offset int
	switch additional.Kind {
	case cost.AdditionalSacrifice:
		wanted, offset = legacy.SacrificeChoices, len(before.SacrificeChoices)
	case cost.AdditionalTapPermanents:
		wanted, offset = legacy.TapChoices, len(before.TapChoices)
	case cost.AdditionalReturnToHand, cost.AdditionalReturnUnblockedAttacker:
		wanted, offset = legacy.ReturnChoices, len(before.ReturnChoices)
	case cost.AdditionalDiscard:
		wanted, offset = legacy.DiscardChoices, len(before.DiscardChoices)
	case cost.AdditionalExile:
		wanted, offset = legacy.ExileChoices, len(before.ExileChoices)
	case cost.AdditionalReveal:
		wanted, offset = legacy.RevealChoices, len(before.RevealChoices)
	case cost.AdditionalCollectEvidence:
		wanted, offset = legacy.EvidenceChoices, len(before.EvidenceChoices)
	case cost.AdditionalRemoveCounterAmong:
		offset = len(before.RemoveCounterChoices)
		if offset < len(legacy.RemoveCounterChoices) {
			return slices.Equal(choice.Counters, legacy.RemoveCounterChoices[offset:min(offset+len(choice.Counters), len(legacy.RemoveCounterChoices))])
		}
	case cost.AdditionalRemoveCounter:
		offset = len(before.SourceCounterKindChoices)
		if !additional.AnyCounterKind {
			return true
		}
		for i, removal := range choice.Counters {
			if offset+i < len(legacy.SourceCounterKindChoices) && removal.Kind != legacy.SourceCounterKindChoices[offset+i] {
				return false
			}
		}
	}
	if offset >= len(wanted) {
		return true
	}
	return slices.Equal(choice.Objects, wanted[offset:min(offset+len(choice.Objects), len(wanted))])
}

func selectComponent(s State, player game.PlayerID, additional cost.Additional, x int, source *game.Permanent, sourceCardID, castingCardID id.ID, fixed ComponentChoice, accept func(ComponentChoice) bool, reservations ...componentReservations) bool {
	amount := additionalCostAmountIn(s, player, additional, x)
	if amount < 0 {
		return false
	}
	var candidates []id.ID
	var weight func(id.ID) int
	threshold := 0
	switch additional.Kind {
	case cost.AdditionalSacrifice, cost.AdditionalTapPermanents, cost.AdditionalReturnToHand, cost.AdditionalReturnUnblockedAttacker:
		for _, permanent := range CandidatePermanentsForCost(s, player, additional, returnCostSource(s, additional, source, sourceCardID)) {
			candidates = append(candidates, permanent.ObjectID)
		}
		if additional.TotalPowerAtLeast > 0 {
			threshold = additional.TotalPowerAtLeast
			weight = func(object id.ID) int {
				permanent, ok := s.PermanentByObjectID(object)
				if !ok {
					return 0
				}
				return permanentPowerContribution(s, permanent, additional.PowerContribution)
			}
		}
	case cost.AdditionalDiscard, cost.AdditionalExile, cost.AdditionalReveal, cost.AdditionalCollectEvidence:
		if additional.Random {
			return len(fixed.Objects) == 0 && len(fixed.Counters) == 0 && accept(ComponentChoice{Finished: true})
		}
		candidates = CandidateCardsForCost(s, player, additional, sourceCardID, castingCardID)
		threshold = additional.TotalManaValueAtLeast
		if additional.Kind == cost.AdditionalCollectEvidence {
			threshold = amount
		}
		if threshold > 0 {
			weight = func(card id.ID) int {
				value, _ := evidenceCardManaValue(s, card)
				return value
			}
		}
	case cost.AdditionalRemoveCounter, cost.AdditionalRemoveCounterAmong:
		var counterReservations []map[CounterRemovalChoice]int
		for _, reserved := range reservations {
			counterReservations = append(counterReservations, reserved.counters)
		}
		return selectCounterComponent(s, player, additional, amount, source, fixed, accept, counterReservations...)
	default:
		return len(fixed.Objects) == 0 && len(fixed.Counters) == 0 && accept(ComponentChoice{Finished: true})
	}
	for _, reserved := range reservations {
		candidates = slices.DeleteFunc(candidates, func(object id.ID) bool { return reserved.objects[object] })
	}
	if len(fixed.Counters) != 0 {
		return false
	}
	if threshold == 0 && len(fixed.Objects) > amount {
		return false
	}
	for i, object := range fixed.Objects {
		if !slices.Contains(candidates, object) || slices.Contains(fixed.Objects[:i], object) {
			return false
		}
	}
	total := 0
	for _, object := range fixed.Objects {
		if weight != nil {
			total += weight(object)
		}
	}
	var visit func(int, []id.ID, int) bool
	visit = func(start int, selected []id.ID, total int) bool {
		enough := len(selected) == amount
		if threshold > 0 {
			enough = total >= threshold
		}
		if enough && accept(ComponentChoice{Objects: slices.Clone(selected), Finished: true}) {
			return true
		}
		if fixed.Finished || threshold == 0 && enough {
			return false
		}
		for i := start; i < len(candidates); i++ {
			object := candidates[i]
			if slices.Contains(selected, object) {
				continue
			}
			nextTotal := total
			if weight != nil {
				nextTotal += weight(object)
			}
			if visit(i+1, append(slices.Clone(selected), object), nextTotal) {
				return true
			}
		}
		return false
	}
	return visit(0, slices.Clone(fixed.Objects), total)
}

func componentReservedObjects(prefix additionalCostPlan, additional cost.Additional, source *game.Permanent, sourceTap bool, tapReservations []*game.Permanent) map[id.ID]bool {
	switch additional.Kind {
	case cost.AdditionalSacrifice, cost.AdditionalReturnToHand, cost.AdditionalReturnUnblockedAttacker:
		return reservedPermanentIDs(plannedBattlefieldCosts(prefix))
	case cost.AdditionalTapPermanents:
		reserved := reservedPermanentIDs(append(slices.Clone(prefix.permanentsToTap), tapReservations...))
		if sourceTap && source != nil {
			reserved[source.ObjectID] = true
		}
		return reserved
	case cost.AdditionalDiscard, cost.AdditionalExile, cost.AdditionalCollectEvidence:
		reserved := reservedFromSelections(plannedEvidenceCards(prefix))
		for _, card := range prefix.discards {
			reserved[card] = true
		}
		return reserved
	default:
		return nil
	}
}

func selectCounterComponent(s State, player game.PlayerID, additional cost.Additional, amount int, source *game.Permanent, fixed ComponentChoice, accept func(ComponentChoice) bool, reserved ...map[CounterRemovalChoice]int) bool {
	if additional.RemoveAll {
		if source == nil || additional.AnyCounterKind {
			return false
		}
		amount = source.Counters.Get(additional.CounterKind)
		for _, reservations := range reserved {
			amount -= reservations[CounterRemovalChoice{Permanent: source.ObjectID, Kind: additional.CounterKind}]
		}
		amount = max(amount, 0)
	}
	if len(fixed.Objects) != 0 || len(fixed.Counters) > amount || amount < 0 ||
		amount == 0 && additional.AmountAtLeastOne {
		return false
	}
	permanents := CandidatePermanentsForCost(s, player, additional, source)
	if additional.Kind == cost.AdditionalRemoveCounter {
		if source == nil {
			return false
		}
		permanents = []*game.Permanent{source}
	}
	var candidates []CounterRemovalChoice
	capacity := make(map[CounterRemovalChoice]int)
	for _, permanent := range permanents {
		for _, kind := range RemovableCounterKinds(permanent, additional) {
			choice := CounterRemovalChoice{Permanent: permanent.ObjectID, Kind: kind}
			available := permanent.Counters.Get(kind)
			for _, reservations := range reserved {
				available -= reservations[choice]
			}
			if available > 0 {
				candidates = append(candidates, choice)
				capacity[choice] = available
			}
		}
	}
	used := make(map[CounterRemovalChoice]int)
	for _, choice := range fixed.Counters {
		used[choice]++
		if used[choice] > capacity[choice] {
			return false
		}
	}
	var visit func(int, []CounterRemovalChoice) bool
	visit = func(start int, selected []CounterRemovalChoice) bool {
		if len(selected) == amount {
			return accept(ComponentChoice{Counters: slices.Clone(selected), Finished: true})
		}
		if fixed.Finished {
			return false
		}
		for i := start; i < len(candidates); i++ {
			choice := candidates[i]
			if used[choice] >= capacity[choice] {
				continue
			}
			used[choice]++
			if visit(i, append(slices.Clone(selected), choice)) {
				return true
			}
			used[choice]--
		}
		return false
	}
	return visit(0, slices.Clone(fixed.Counters))
}

// Mandatory later source-kind removals are reserved before an any-kind choice.
// This avoids enumerating combinations that steal a required counter (#3712).
func counterCompletionReservations(s State, player game.PlayerID, before, after []cost.Additional, x int, prefs *Preferences, source *game.Permanent, sourceCardID id.ID, sourceZone zone.Type, castingCardID id.ID) map[CounterRemovalChoice]int {
	reserved := make(map[CounterRemovalChoice]int)
	if prefix, ok := buildSelectedAdditionalCostPlan(s, player, before, x, clonePreferences(prefs), source, sourceCardID, sourceZone, castingCardID); ok {
		for _, removal := range prefix.counterRemovals {
			if !removal.all {
				reserved[CounterRemovalChoice{Permanent: removal.source.ObjectID, Kind: removal.kind}] += removal.amount
			}
		}
	}
	if source != nil {
		for _, additional := range after {
			if additional.ChoiceGroup == 0 && additional.Kind == cost.AdditionalRemoveCounter && !additional.AnyCounterKind && !additional.RemoveAll {
				reserved[CounterRemovalChoice{Permanent: source.ObjectID, Kind: additional.CounterKind}] += additionalCostAmountIn(s, player, additional, x)
			}
		}
	}
	return reserved
}

func cloneComponents(components map[int]ComponentChoice) map[int]ComponentChoice {
	result := make(map[int]ComponentChoice, len(components))
	for index, component := range components {
		result[index] = cloneComponent(component)
	}
	return result
}

func cloneGroups(groups map[uint8]int) map[uint8]int {
	result := make(map[uint8]int, len(groups))
	for group, index := range groups {
		result[group] = index
	}
	return result
}
