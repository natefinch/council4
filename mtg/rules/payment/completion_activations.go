package payment

import (
	"fmt"
	"maps"
	"reflect"
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

func cloneProjection(p game.ResourceProjection) game.ResourceProjection {
	p.Permanents = maps.Clone(p.Permanents)
	for object, resources := range p.Permanents {
		resources.Counters = maps.Clone(resources.Counters)
		p.Permanents[object] = resources
	}
	p.Players = maps.Clone(p.Players)
	for player, resources := range p.Players {
		resources.Mana = maps.Clone(resources.Mana)
		resources.Riders = slices.Clone(resources.Riders)
		p.Players[player] = resources
	}
	p.Used = maps.Clone(p.Used)
	p.UnavailableCards = maps.Clone(p.UnavailableCards)
	p.Events = slices.Clone(p.Events)
	for i := range p.Events {
		p.Events[i] = p.Events[i].Copy()
	}
	p.LastKnown = maps.Clone(p.LastKnown)
	p.SpentDelayed = maps.Clone(p.SpentDelayed)
	return p
}

func mergeProjection(base, delta game.ResourceProjection) game.ResourceProjection {
	p := cloneProjection(base)
	delta = cloneProjection(delta)
	if p.Permanents == nil {
		p.Permanents = make(map[id.ID]game.PermanentResources)
	}
	if p.Players == nil {
		p.Players = make(map[game.PlayerID]game.PlayerResources)
	}
	if p.Used == nil {
		p.Used = make(map[game.ActivatedAbilityUse]bool)
	}
	if p.UnavailableCards == nil {
		p.UnavailableCards = make(map[id.ID]bool)
	}
	if p.LastKnown == nil {
		p.LastKnown = make(map[id.ID]game.ObjectSnapshot)
	}
	if p.SpentDelayed == nil {
		p.SpentDelayed = make(map[id.ID]bool)
	}
	maps.Copy(p.Permanents, delta.Permanents)
	maps.Copy(p.Players, delta.Players)
	maps.Copy(p.Used, delta.Used)
	maps.Copy(p.UnavailableCards, delta.UnavailableCards)
	maps.Copy(p.LastKnown, delta.LastKnown)
	maps.Copy(p.SpentDelayed, delta.SpentDelayed)
	p.Events = append(p.Events, delta.Events...)
	return p
}

func permanentResources(p *game.Permanent) game.PermanentResources {
	counters := make(map[counter.Kind]int)
	maps.Copy(counters, p.Counters.All())
	return game.PermanentResources{Tapped: p.Tapped, Counters: counters}
}

func projectActivationCosts(s ManaState, playerID game.PlayerID, witness additionalWitness, tap bool, source id.ID, pool map[mana.Unit]int, inputLife int) (game.ResourceProjection, bool) {
	player, _ := s.Player(playerID)
	p := mergeProjection(game.ResourceProjection{}, game.ResourceProjection{})
	p.Players[playerID] = game.PlayerResources{
		Life: player.Life - s.LifePaymentLoss(playerID, inputLife), Energy: player.EnergyCounters - witness.plan.energyPaid,
		Mana: clonePoolSpend(pool), Riders: slices.Clone(player.ManaRiders),
	}
	edit := func(object id.ID, f func(*game.PermanentResources)) {
		resources, ok := p.Permanents[object]
		if !ok {
			permanent, _ := s.PermanentByObjectID(object)
			resources = permanentResources(permanent)
		}
		f(&resources)
		p.Permanents[object] = resources
	}
	if tap {
		edit(source, func(r *game.PermanentResources) { r.Tapped = true })
		p.Events = append(p.Events, s.ProjectResources(p).(ManaState).PrepareManaEvent(game.Event{Kind: game.EventPermanentTapped, PermanentID: source, Player: playerID, TappedForMana: true}))
	}
	for _, removal := range witness.plan.counterRemovals {
		edit(removal.source.ObjectID, func(r *game.PermanentResources) { r.Counters[removal.kind] -= removal.amount })
	}
	if witness.plan.untapSource != nil {
		object := witness.plan.untapSource.ObjectID
		resources, changed := s.ProjectResources(p).(ManaState).UntapResources(object)
		p.Permanents[object] = resources
		if changed {
			p.Events = append(p.Events, game.Event{Kind: game.EventPermanentUntapped, PermanentID: object, Player: playerID})
		}
	}
	for _, addition := range witness.plan.counterAdds {
		amount, ok := s.CounterPlacement(playerID, addition.source.ObjectID, addition.kind, addition.amount)
		if !ok {
			return game.ResourceProjection{}, false
		}
		edit(addition.source.ObjectID, func(r *game.PermanentResources) { r.Counters[addition.kind] += amount })
	}
	for _, permanent := range append(slices.Clone(witness.plan.sacrifices), witness.plan.exilePermanents...) {
		p.LastKnown[permanent.ObjectID] = s.ProjectResources(p).(ManaState).ManaSourceSnapshot(permanent.ObjectID)
		edit(permanent.ObjectID, func(r *game.PermanentResources) { r.Removed = true })
	}
	if lost := s.ProjectResources(p).(ManaState).LifePaymentLoss(playerID, witness.plan.lifePaid); lost > 0 {
		resources := p.Players[playerID]
		resources.Life -= lost
		p.Players[playerID] = resources
		p.Events = append(p.Events, s.ProjectResources(p).(ManaState).PrepareManaEvent(game.Event{Kind: game.EventLifeLost, Player: playerID, Amount: lost}))
	}
	for _, exile := range witness.plan.exiles {
		p.UnavailableCards[exile.cardID] = true
	}
	return p, true
}

func jointPaymentReady(s State, playerID game.PlayerID, plan paymentPlan, additional additionalCostPlan) bool {
	if len(plan.activations) == 0 {
		player, ok := s.Player(playerID)
		return ok && paymentApplicationReady(s, player, plan, additional)
	}
	state := s.(ManaState).ProjectResources(plan.projection)
	player, ok := state.Player(playerID)
	additional = translateAdditionalPlan(state, additional)
	return ok && player.Life >= plan.lifePayment+additional.lifePaid &&
		additionalCostPlanStillValid(state, player, additional)
}

func translateAdditionalPlan(s State, p additionalCostPlan) additionalCostPlan {
	resolve := func(permanent *game.Permanent) *game.Permanent {
		if permanent == nil {
			return nil
		}
		current, ok := s.PermanentByObjectID(permanent.ObjectID)
		if !ok {
			return permanent
		}
		return current
	}
	translate := func(permanents []*game.Permanent) []*game.Permanent {
		copy := slices.Clone(permanents)
		for i := range copy {
			copy[i] = resolve(copy[i])
		}
		return copy
	}
	p.sacrifices = translate(p.sacrifices)
	p.permanentsToTap = translate(p.permanentsToTap)
	p.exilePermanents = translate(p.exilePermanents)
	p.untapSource, p.exertSource = resolve(p.untapSource), resolve(p.exertSource)
	p.returnsToHand = slices.Clone(p.returnsToHand)
	for i := range p.returnsToHand {
		p.returnsToHand[i].permanent = resolve(p.returnsToHand[i].permanent)
	}
	p.counterRemovals = slices.Clone(p.counterRemovals)
	for i := range p.counterRemovals {
		p.counterRemovals[i].source = resolve(p.counterRemovals[i].source)
	}
	p.counterAdds = slices.Clone(p.counterAdds)
	for i := range p.counterAdds {
		p.counterAdds[i].source = resolve(p.counterAdds[i].source)
	}
	return p
}
func manaOccurrenceCounts(activations []ManaActivation) map[ManaAbilityKey]int {
	counts := make(map[ManaAbilityKey]int)
	for _, activation := range activations {
		counts[ManaAbilityKey{Source: activation.Source, AbilityIndex: activation.AbilityIndex}]++
	}
	return counts
}

func manaActivationOccurrence(activations []ManaActivation, key ManaActivationKey) (ManaActivation, bool) {
	occurrence := 0
	for _, activation := range activations {
		if activation.Source != key.Source || activation.AbilityIndex != key.AbilityIndex {
			continue
		}
		if occurrence == key.Occurrence {
			return activation, true
		}
		occurrence++
	}
	return ManaActivation{}, false
}

func remainingManaOccurrences[V any](fixed map[ManaActivationKey]V, counts map[ManaAbilityKey]int) map[ManaActivationKey]V {
	if fixed == nil {
		return nil
	}
	remaining := make(map[ManaActivationKey]V)
	for key, value := range fixed {
		if key.Occurrence < counts[key.ManaAbilityKey] {
			continue
		}
		key.Occurrence -= counts[key.ManaAbilityKey]
		remaining[key] = value
	}
	return remaining
}

func bindManaCostOccurrence(partial Partial, plan Plan) (Partial, bool) {
	for index, fixed := range partial.Components {
		component, exists := plan.components[index]
		if !exists || len(fixed.Objects) > len(component.Objects) || len(fixed.Counters) > len(component.Counters) ||
			!slices.Equal(fixed.Objects, component.Objects[:len(fixed.Objects)]) ||
			!slices.Equal(fixed.Counters, component.Counters[:len(fixed.Counters)]) ||
			fixed.Finished && (len(fixed.Objects) != len(component.Objects) || len(fixed.Counters) != len(component.Counters)) {
			return partial, false
		}
	}
	for group, selected := range partial.Groups {
		if actual, exists := plan.groups[group]; !exists || actual != selected {
			return partial, false
		}
	}
	allocation := plan.Mana()
	if partial.Allocation != nil && !reflect.DeepEqual(*partial.Allocation, allocation) {
		return partial, false
	}
	partial.Components, partial.Groups = plan.Components(), cloneGroups(plan.groups)
	partial.Allocation = &allocation
	return partial, true
}

func completeManaActivations(s ManaState, playerID game.PlayerID, manaCost *cost.Mana, x int, restrictions manaSourceRestrictions, ctx spendContext, prefs *Preferences, lifeReserved int, spell *spellCostOption, cardID id.ID, sourceZone zone.Type, partial Partial) (paymentPlan, bool) {
	if view, ok := s.(interface{ ManaAbilityUnreachable(ManaAbilityKey) bool }); ok {
		for key, use := range partial.ManaAbilities {
			if use && view.ManaAbilityUnreachable(key) {
				return paymentPlan{}, false
			}
		}
	}
	schemaAllocation := partial.Allocation != nil && slices.ContainsFunc(partial.Allocation.Activations, func(a ManaActivation) bool { return a.Effects.Players != nil })
	if restrictions.resourceCompletion == nil && len(partial.ManaAbilities) == 0 && len(partial.ManaOccurrences) == 0 && !schemaAllocation && s.OrdinaryManaOnly(playerID) {
		return completeOrdinaryMana(s, playerID, manaCost, x, restrictions, ctx, prefs, lifeReserved, spell, cardID, sourceZone, partial)
	}
	failed := make(map[string]bool)
	historySensitive := false
	for _, permanent := range s.Battlefield() {
		for _, ability := range s.PermanentEffectiveAbilities(permanent) {
			if body, ok := ability.(*game.ManaAbility); ok {
				historySensitive = historySensitive || body.ActivationCondition.Exists
			} else if body, ok := ability.(*game.TriggeredAbility); ok {
				historySensitive = historySensitive || body.Trigger.InterveningCondition.Exists
			} else if _, ok := ability.(*game.StaticAbility); ok {
				historySensitive = true
			}
		}
	}
	var result paymentPlan
	depth := 0
	var search func(game.ResourceProjection, []ManaActivation) bool
	search = func(projection game.ResourceProjection, activations []ManaActivation) bool {
		current := s
		if len(activations) != 0 {
			current = s.ProjectResources(projection).(ManaState)
		}
		player, ok := current.Player(playerID)
		if !ok {
			return false
		}
		counts := manaOccurrenceCounts(activations)
		key := fmt.Sprint(depth-len(activations), player.Life, player.EnergyCounters, player.ManaPool.Units(), player.ManaRiders, projection.Permanents, projection.Used, projection.UnavailableCards, projection.SpentDelayed)
		if historySensitive {
			events := make(map[string]int)
			for _, event := range projection.Events {
				events[fmt.Sprint(event)]++
			}
			key += fmt.Sprint(events)
		}
		for choice, use := range partial.ManaAbilities {
			if use {
				key += fmt.Sprint(choice, slices.ContainsFunc(activations, func(a ManaActivation) bool { return a.Source == choice.Source && a.AbilityIndex == choice.AbilityIndex }))
			}
		}
		seenOccurrences := make(map[ManaActivationKey]int)
		for occurrence := range partial.ManaPayments {
			seenOccurrences[occurrence] = min(counts[occurrence.ManaAbilityKey], occurrence.Occurrence+1)
		}
		for occurrence := range partial.ManaChoices {
			seenOccurrences[occurrence] = min(counts[occurrence.ManaAbilityKey], occurrence.Occurrence+1)
		}
		for occurrence := range partial.ManaOccurrences {
			seenOccurrences[occurrence] = min(counts[occurrence.ManaAbilityKey], occurrence.Occurrence+1)
		}
		key += fmt.Sprint(seenOccurrences)
		if failed[key] {
			return false
		}
		failed[key] = true
		if restrictions.resourceCompletion != nil {
			if plan, ok := restrictions.resourceCompletion(current, activations, projection); ok {
				result = plan
				return true
			}
			if len(activations) == depth {
				return false
			}
		}
		// Only pool and cost mechanics are considered here. All mana abilities
		// (including basic-land taps) go through the same canonical forecast.
		exclude := cloneManaSourceRestrictions(restrictions)
		fixed := partial
		fixed.joint = false
		fixed.poolOnly = true
		fixed.Allocation = nil
		fixed.ManaAbilities = nil
		fixed.ManaOccurrences = nil
		fixed.ManaPayments = nil
		fixed.ManaChoices = nil
		var plan paymentPlan
		payable := false
		if restrictions.resourceCompletion == nil {
			plan, payable = completeMana(current, playerID, manaCost, x, exclude, ctx, prefs, lifeReserved, spell, cardID, sourceZone, fixed)
		}
		if payable {
			plan.activations = slices.Clone(activations)
			plan.projection = projection
			if partial.acceptsMana(plan) && (restrictions.resourceValid == nil || restrictions.resourceValid(current)) && jointPaymentReady(s, playerID, plan, reservedAdditional(restrictions, playerID)) {
				result = plan
				return true
			}
		}
		if len(activations) == depth {
			return false
		}
		candidates := current.ManaActivationCandidates(playerID)
		slices.SortStableFunc(candidates, func(a, b ManaActivationCandidate) int {
			if a.Automatic != b.Automatic {
				if a.Automatic {
					return -1
				}
				return 1
			}
			return 0
		})
		for _, candidate := range candidates {
			abilityKey := ManaAbilityKey{Source: candidate.Source, AbilityIndex: candidate.AbilityIndex}
			activationKey := ManaActivationKey{ManaAbilityKey: abilityKey, Occurrence: counts[abilityKey]}
			if use, fixed := partial.ManaOccurrences[activationKey]; fixed && !use {
				continue
			}
			if use, fixed := partial.ManaAbilities[abilityKey]; fixed && !use ||
				restrictions.excluded[candidate.Source] ||
				restrictions.nonSacrificingOnly[candidate.Source] && candidate.Sacrifice ||
				projection.UnavailableCards[candidate.Source] {
				continue
			}
			if choices, fixed := partial.ManaChoices[activationKey]; fixed && !maps.Equal(choices, candidate.Choices) {
				continue
			}
			if partial.Allocation != nil {
				index := len(activations)
				if index >= len(partial.Allocation.Activations) {
					continue
				}
				expected := partial.Allocation.Activations[index]
				if expected.Source != candidate.Source || expected.AbilityIndex != candidate.AbilityIndex ||
					!reflectChoices(expected.Choices, candidate.Choices) {
					continue
				}
			}
			costState := current
			ability := candidate.Request.ability
			if ability.SourceCardID != 0 && restrictions.reserved != nil && slices.Contains(restrictions.reserved.discards, ability.SourceCardID) {
				continue
			}
			var source *game.Permanent
			if candidate.SourceZone == zone.Battlefield {
				source, _ = current.PermanentByObjectID(candidate.Source)
			}
			tap := hasTapCostOf(ability.AdditionalCosts)
			if tap && (source == nil || !canTapForAbility(current, source)) {
				continue
			}
			costExclude := manaSourceRestrictions{excluded: make(map[id.ID]bool)}
			costPartial := partial.ManaPayments[activationKey]
			if partial.Allocation != nil {
				expected := partial.Allocation.Activations[len(activations)]
				if expected.CostPlan != nil {
					var valid bool
					costPartial, valid = bindManaCostOccurrence(costPartial, *expected.CostPlan)
					if !valid {
						continue
					}
				}
			}
			costPartial.poolOnly = true
			input, ok := completeMana(current, playerID, manaCostPtr(ability.ManaCost), 0, costExclude, spendContext{abilitySource: source}, nil, lifeReserved, nil, 0, zone.None, costPartial)
			if !ok {
				continue
			}
			_, completed := completeAdditional(current, playerID, ability.AdditionalCosts, 0, nil, source, ability.SourceCardID, candidate.SourceZone, 0, costPartial.Components, costPartial.Groups, nil, func(w additionalWitness) bool {
				for _, permanent := range append(slices.Clone(w.plan.sacrifices), w.plan.exilePermanents...) {
					if restrictions.nonSacrificingOnly[permanent.ObjectID] || restrictions.excluded[permanent.ObjectID] {
						return false
					}
				}
				pool := player.ManaPool.Units()
				for unit, amount := range input.poolSpend {
					pool[unit] -= amount
				}
				delta, good := projectActivationCosts(current, playerID, w, tap, candidate.Source, pool, input.lifePayment)
				if !good {
					return false
				}
				resources := delta.Players[playerID]
				resources.Riders = current.RidersAfterInput(playerID, candidate.Source, input.poolSpend)
				delta.Players[playerID] = resources
				if resources.Life < lifeReserved {
					return false
				}
				costState = current.ProjectResources(delta).(ManaState)
				production, effects, ok := costState.ForecastMana(playerID, candidate)
				if !ok {
					return false
				}
				next := mergeProjection(projection, delta)
				next = mergeProjection(next, effects)
				resources = next.Players[playerID]
				for _, produced := range production {
					resources.Mana[produced.Unit] += produced.Amount
					if produced.Rider != nil {
						for range produced.Amount {
							resources.Riders = append(resources.Riders, *produced.Rider)
						}
					}
				}
				next.Players[playerID] = resources
				if !candidate.Reusable {
					next.Used[game.ActivatedAbilityUse{SourceID: candidate.Source, AbilityIndex: candidate.AbilityIndex}] = true
				}
				costPlan := Plan{request: candidate.Request, option: 0, components: w.components, groups: w.groups, allocation: freezeMana(input), valid: true}
				activation := ManaActivation{Source: candidate.Source, AbilityIndex: candidate.AbilityIndex, SourceZone: candidate.SourceZone, CostPlan: &costPlan, Choices: maps.Clone(candidate.Choices), Production: production, Effects: effects, Before: cloneProjection(projection)}
				if len(production) > 0 && !slices.ContainsFunc(production, func(p ManaProduction) bool { return p.Unit != production[0].Unit }) {
					activation.Color, activation.Amount = production[0].Unit.Color, production[0].Amount
					activation.Snow, activation.FromCreature = production[0].Unit.Snow, production[0].Unit.FromCreature
					for _, extra := range production[1:] {
						activation.Amount += extra.Amount
					}
				}
				activation.Sacrifice = candidate.Sacrifice
				activation.Untap = slices.ContainsFunc(candidate.Body.AdditionalCosts, func(additional cost.Additional) bool { return additional.Kind == cost.AdditionalUntap })
				return search(next, append(slices.Clone(activations), activation))
			})
			if completed {
				return true
			}
		}
		failed[key] = true
		return false
	}
	candidates := s.ManaActivationCandidates(playerID)
	limit := len(candidates)
	required := make(map[ManaAbilityKey]int)
	for key, use := range partial.ManaOccurrences {
		if use {
			required[key.ManaAbilityKey] = max(required[key.ManaAbilityKey], key.Occurrence+1)
		}
	}
	for key := range partial.ManaPayments {
		required[key.ManaAbilityKey] = max(required[key.ManaAbilityKey], key.Occurrence+1)
	}
	for key := range partial.ManaChoices {
		required[key.ManaAbilityKey] = max(required[key.ManaAbilityKey], key.Occurrence+1)
	}
	for _, count := range required {
		limit += count
	}
	player, ok := s.Player(playerID)
	if !ok {
		return paymentPlan{}, false
	}
	for _, candidate := range candidates {
		if !candidate.Reusable {
			continue
		}
		for _, additional := range candidate.Body.AdditionalCosts {
			switch additional.Kind {
			case cost.AdditionalPayLife:
				if additional.Amount > 0 {
					limit += player.Life / additional.Amount
				}
			case cost.AdditionalRemoveCounter:
				if source, ok := s.PermanentByObjectID(candidate.Source); ok {
					for _, count := range source.Counters.All() {
						limit += count
					}
				}
			case cost.AdditionalEnergy:
				limit += player.EnergyCounters
			case cost.AdditionalSacrifice:
				if amount := AdditionalCostAmount(additional); amount > 0 {
					source, _ := s.PermanentByObjectID(candidate.Source)
					limit += len(CandidatePermanentsForCost(s, playerID, additional, source)) / amount
				}
			}
		}
	}
	for depth = 0; depth <= limit; depth++ {
		failed = make(map[string]bool)
		if search(game.ResourceProjection{}, nil) {
			return result, true
		}
	}
	return result, false
}

func completeOrdinaryMana(s ManaState, player game.PlayerID, manaCost *cost.Mana, x int, restrictions manaSourceRestrictions, ctx spendContext, prefs *Preferences, life int, spell *spellCostOption, cardID id.ID, sourceZone zone.Type, partial Partial) (paymentPlan, bool) {
	fixed := partial
	fixed.joint = false
	fixed.Allocation = nil
	plan, ok := completeMana(s, player, manaCost, x, restrictions, ctx, prefs, life, spell, cardID, sourceZone, fixed)
	if !ok {
		return paymentPlan{}, false
	}
	payer, _ := s.Player(player)
	projection := mergeProjection(game.ResourceProjection{}, game.ResourceProjection{})
	projection.Players[player] = game.PlayerResources{Life: payer.Life, Energy: payer.EnergyCounters, Mana: payer.ManaPool.Units(), Riders: slices.Clone(payer.ManaRiders)}
	for _, tap := range plan.manaTaps {
		abilities := s.PermanentEffectiveAbilities(tap.permanent)
		if tap.abilityIndex < 0 || tap.abilityIndex >= len(abilities) {
			return paymentPlan{}, false
		}
		body, ok := abilities[tap.abilityIndex].(*game.ManaAbility)
		if !ok {
			return paymentPlan{}, false
		}
		request := ForAbility(AbilityRequest{PlayerID: player, Source: tap.permanent, SourceCardID: tap.permanent.CardInstanceID, SourceZone: zone.Battlefield, AdditionalCosts: body.AdditionalCosts, ForMana: true})
		witness, ok := completeAdditional(s, player, body.AdditionalCosts, 0, nil, tap.permanent, tap.permanent.CardInstanceID, zone.Battlefield, 0, nil, nil, nil, func(additionalWitness) bool { return true })
		if !ok {
			return paymentPlan{}, false
		}
		input := paymentPlan{poolSpend: make(map[mana.Unit]int)}
		costPlan := Plan{request: request, components: witness.components, groups: witness.groups, allocation: freezeMana(input), valid: true}
		unit := mana.Unit{Color: tap.color, Snow: tap.snow, FromCreature: tap.fromCreature}
		plan.activations = append(plan.activations, ManaActivation{Source: tap.permanent.ObjectID, AbilityIndex: tap.abilityIndex, SourceZone: zone.Battlefield, Color: tap.color, Amount: tap.amount, Snow: tap.snow, FromCreature: tap.fromCreature, Untap: tap.untap, CostPlan: &costPlan, Production: []ManaProduction{{Unit: unit, Amount: tap.amount}}, Before: cloneProjection(projection)})
		resources := permanentResources(tap.permanent)
		resources.Tapped = !tap.untap
		projection.Permanents[tap.permanent.ObjectID] = resources
		pr := projection.Players[player]
		pr.Mana[unit] += tap.amount
		projection.Players[player] = pr
	}
	plan.manaTaps = nil
	plan.projection = projection
	return plan, partial.acceptsMana(plan) && (restrictions.resourceValid == nil || restrictions.resourceValid(s.ProjectResources(projection)))
}
func reservedAdditional(r manaSourceRestrictions, player game.PlayerID) additionalCostPlan {
	if r.reserved != nil {
		return *r.reserved
	}
	return additionalCostPlan{player: player}
}

func reflectChoices(a, b map[game.ChoiceKey]game.ResolutionChoiceResult) bool {
	return maps.Equal(a, b)
}
