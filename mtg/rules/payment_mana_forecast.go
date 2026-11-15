package rules

import (
	"maps"
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules/payment"
)

var _ payment.ManaState = (*rulesPaymentState)(nil)

func (s *rulesPaymentState) ProjectResources(p game.ResourceProjection) payment.State {
	return &rulesPaymentState{g: s.g.ResourceView(p), readOnly: true, engine: s.engine}
}

func (s *rulesPaymentState) viewPermanent(p *game.Permanent) *game.Permanent {
	if p == nil || !s.readOnly {
		return p
	}
	projected, _ := permanentByObjectID(s.g, p.ObjectID)
	return projected
}

func (s *rulesPaymentState) assertWritable() {
	if s.readOnly {
		panic("payment resource view is read only")
	}
}

func (s *rulesPaymentState) ResourceReadOnly() bool { return s.readOnly }

func (s *rulesPaymentState) ManaSourceSnapshot(object id.ID) game.ObjectSnapshot {
	permanent, _ := permanentByObjectID(s.g, object)
	return snapshotPermanent(s.g, permanent, zone.Battlefield)
}

func (s *rulesPaymentState) LifePaymentLoss(player game.PlayerID, amount int) int {
	return prepareLifeLoss(s.g, player, amount)
}

func (s *rulesPaymentState) RidersAfterInput(playerID game.PlayerID, sourceID id.ID, spent map[mana.Unit]int) []game.ManaRiderInstance {
	player, _ := playerByID(s.g, playerID)
	source, _ := permanentByObjectID(s.g, sourceID)
	resource := game.Player{ManaRiders: slices.Clone(player.ManaRiders)}
	processManaSpendRiders(&resource, player.ManaPool.Units(), spent, func(rider game.ManaRiderInstance) bool {
		return abilityActivationSatisfiesManaSpendRider(s.g, source, rider)
	}, func(game.ManaRiderInstance) {})
	return resource.ManaRiders
}

func (s *rulesPaymentState) ConsumePaymentRiders(player game.PlayerID, source *game.Permanent, before, spent map[mana.Unit]int) {
	s.assertWritable()
	consumeManaSpendRidersForPayment(s.g, player, source, before, spent)
}

func (s *rulesPaymentState) ManaAbilityLabel(source id.ID, index int) string {
	player := game.Player1
	if permanent, ok := permanentByObjectID(s.g, source); ok {
		player = effectiveController(s.g, permanent)
	} else if card, ok := s.g.GetCardInstance(source); ok {
		player = card.Owner
	}
	options := paymentManaAbilityOptions(s.g, player, []action.Action{action.ActivateAbility(source, index, nil, 0)})
	return options[1].Label
}

func (s *rulesPaymentState) UntapResources(source id.ID) (game.PermanentResources, bool) {
	permanent, _ := permanentByObjectID(s.g, source)
	resource := *permanent
	resource.Counters = permanent.Counters.Clone()
	changed := changePermanentTapped(&resource, false)
	return game.PermanentResources{Tapped: resource.Tapped, Counters: resource.Counters.All()}, changed
}

func (s *rulesPaymentState) OrdinaryManaOnly(player game.PlayerID) bool {
	if len(s.g.ContinuousEffects) != 0 || len(s.g.DelayedTriggers) != 0 {
		return false
	}
	for _, permanent := range s.g.Battlefield {
		if !activeBattlefieldPermanent(permanent) {
			continue
		}
		for _, ability := range permanentEffectiveAbilitiesView(s.g, permanent) {
			body, ok := ability.(*game.ManaAbility)
			if !ok {
				if other, ok := ability.(*game.TriggeredAbility); ok {
					for _, mode := range other.Content.Modes {
						if !mode.Program.IsZero() {
							return false
						}
						for _, instruction := range mode.Sequence {
							if instruction.Primitive != nil && instruction.Primitive.Kind() == game.PrimitiveAddMana {
								return false
							}
						}
					}
				} else if other, ok := ability.(*game.StaticAbility); ok {
					if len(other.RuleEffects) != 0 || len(other.ContinuousEffects) != 0 {
						return false
					}
				}
				continue
			}
			schema, class := payment.ClassifyManaAbility(body)
			if class != payment.PlannableManaAbility || !schema.Automatic() ||
				permanent.Counters.Has(counter.Stun) && slices.ContainsFunc(body.AdditionalCosts, func(c cost.Additional) bool { return c.Kind == cost.AdditionalUntap }) ||
				abilityActivationProhibited(s.g, player, permanent, true) {
				return false
			}
		}
	}
	payer, ok := playerByID(s.g, player)
	if !ok {
		return false
	}
	for _, cardID := range payer.Hand.All() {
		card, ok := s.g.GetCardInstance(cardID)
		if ok && len(cardFaceOrDefault(card, game.FaceFront).ManaAbilities) != 0 {
			return false
		}
	}
	return true
}

func (s *rulesPaymentState) ManaActivationCandidates(playerID game.PlayerID) []payment.ManaActivationCandidate {
	var result []payment.ManaActivationCandidate
	add := func(source id.ID, cardID id.ID, sourceZone zone.Type, permanent *game.Permanent, index int, body *game.ManaAbility) {
		schema, class := payment.ClassifyManaAbility(body)
		if class != payment.PlannableManaAbility || !activatedAbilityTimingAllows(s.g, playerID, body.Timing) ||
			s.g.ActivatedAbilitiesThisTurn[game.ActivatedAbilityUse{SourceID: source, AbilityIndex: index}] ||
			!activationConditionSatisfied(s.g, playerID, permanent, body.ActivationCondition) {
			return
		}
		if permanent != nil {
			if !bodyFunctionsOnBattlefield(body) || abilityActivationProhibited(s.g, playerID, permanent, true) ||
				!manaBodyEntryChoicesAvailable(permanent, body) {
				return
			}
		} else if game.BodyFunctionZone(body) != zone.Hand || !abilityHasExileThisCardFromHandCost(body.AdditionalCosts) {
			return
		}
		request := payment.ForAbility(payment.AbilityRequest{PlayerID: playerID, Source: permanent, SourceCardID: cardID, SourceZone: sourceZone, ManaCost: body.ManaCost, AdditionalCosts: body.AdditionalCosts, ForMana: true})
		choices := []map[game.ChoiceKey]game.ResolutionChoiceResult{nil}
		obj := &game.StackObject{SourceID: source, SourceCardID: cardID, Controller: playerID}
		if permanent != nil {
			seedEntryChoices(obj, permanent)
		}
		for _, instruction := range body.Content.Modes[0].Sequence {
			choose, ok := instruction.Primitive.(game.Choose)
			if !ok {
				continue
			}
			// The canonical choice evaluator supplies dynamic public color sets.
			if choose.Choice.AtRandom {
				return
			}
			var next []map[game.ChoiceKey]game.ResolutionChoiceResult
			for _, selected := range choices {
				for key, value := range selected {
					rememberResolutionChoice(obj, string(key), value)
				}
				_, values := resolutionChoiceRequest(s.g, obj, &choose.Choice)
				if len(values) > 6 {
					return
				}
				indices := make([]int, 0, len(values))
				for index := range values {
					indices = append(indices, index)
				}
				slices.Sort(indices)
				for _, index := range indices {
					copied := maps.Clone(selected)
					if copied == nil {
						copied = make(map[game.ChoiceKey]game.ResolutionChoiceResult)
					}
					copied[choose.PublishChoice] = values[index]
					next = append(next, copied)
				}
			}
			choices = next
		}
		reusable := !abilityHasOncePerTurnRestriction(body.Timing)
		consumesSource := false
		for _, additional := range body.AdditionalCosts {
			if additional.Kind == cost.AdditionalTap || additional.Kind == cost.AdditionalUntap ||
				additional.Kind == cost.AdditionalSacrificeSource || additional.Kind == cost.AdditionalExileSource {
				reusable = false
			}
			if additional.Kind == cost.AdditionalSacrificeSource || additional.Kind == cost.AdditionalExileSource {
				consumesSource = true
			}
		}
		for _, choice := range choices {
			result = append(result, payment.ManaActivationCandidate{Source: source, CardID: cardID, AbilityIndex: index, SourceZone: sourceZone, Request: request, Body: body, Choices: choice, Automatic: schema.Automatic(), Reusable: reusable, Sacrifice: consumesSource})
		}
	}
	for _, permanent := range s.g.Battlefield {
		if !activeBattlefieldPermanent(permanent) || effectiveController(s.g, permanent) != playerID {
			continue
		}
		for index, ability := range permanentEffectiveAbilitiesView(s.g, permanent) {
			if body, ok := ability.(*game.ManaAbility); ok {
				add(permanent.ObjectID, permanent.CardInstanceID, zone.Battlefield, permanent, index, body)
			}
		}
	}
	player, ok := playerByID(s.g, playerID)
	if !ok {
		return result
	}
	for _, cardID := range player.Hand.All() {
		card, ok := s.g.GetCardInstance(cardID)
		if !ok {
			continue
		}
		def := cardFaceOrDefault(card, game.FaceFront)
		for index := range def.ManaAbilities {
			add(cardID, cardID, zone.Hand, nil, def.ManaAbilityIndex(index), &def.ManaAbilities[index])
		}
	}
	return result
}

func deterministicCounterPlacement(g *game.Game, controller game.PlayerID, permanent *game.Permanent, kind counter.Kind, amount int) (int, bool) {
	event := game.Event{Kind: game.EventCountersAdded, Controller: controller, Player: permanent.Owner, CardID: permanent.CardInstanceID, PermanentID: permanent.ObjectID, CounterKind: kind, Amount: amount}
	applied := make(map[id.ID]bool)
	for {
		event.Amount = amount
		matches := matchingCounterPlacementReplacementEffects(g, event, permanent, applied)
		if len(matches) == 0 {
			return amount, true
		}
		if len(matches) != 1 {
			return 0, false
		}
		replacement := matches[0]
		applied[replacement.ID] = true
		amount = counterReplacementAmount(replacement, amount)
	}
}

func counterReplacementAmount(replacement game.ReplacementEffect, amount int) int {
	if replacement.CounterMultiplier > 1 {
		amount *= replacement.CounterMultiplier
	}
	return amount + replacement.CounterAddend
}

func (s *rulesPaymentState) CounterPlacement(controller game.PlayerID, object id.ID, kind counter.Kind, amount int) (int, bool) {
	permanent, ok := permanentByObjectID(s.g, object)
	if !ok {
		return 0, false
	}
	return deterministicCounterPlacement(s.g, controller, permanent, kind, amount)
}

func (s *rulesPaymentState) ForecastMana(playerID game.PlayerID, candidate payment.ManaActivationCandidate) ([]payment.ManaProduction, game.ResourceProjection, bool) {
	obj := &game.StackObject{Kind: game.StackActivatedAbility, SourceID: candidate.Source, Controller: playerID, AbilityIndex: candidate.AbilityIndex}
	if card, ok := s.g.GetCardInstance(candidate.Source); ok && candidate.SourceZone == zone.Hand {
		obj.SourceCardID = card.ID
	} else if permanent, ok := permanentByObjectID(s.g, candidate.Source); ok {
		obj.SourceCardID, obj.SourceTokenDef, obj.Face = permanent.CardInstanceID, permanent.TokenDef, permanent.Face
		seedEntryChoices(obj, permanent)
	} else {
		// The source may have been sacrificed in the cost projection.
		obj.SourceCardID = candidateCardID(candidate)
		if snapshot, ok := lastKnownObject(s.g, candidate.Source); ok {
			obj.SourceTokenDef, obj.Face = snapshot.TokenDef, snapshot.Face
		}
	}
	for key, choice := range candidate.Choices {
		rememberResolutionChoice(obj, string(key), choice)
	}
	production, effects, ok := forecastManaContent(s.g, obj, candidate.Body.Content, playerID)
	if !ok {
		return nil, game.ResourceProjection{}, false
	}
	colors, total := productionColors(production)
	eventResources := effects
	eventResources.Players = maps.Clone(effects.Players)
	payer, _ := playerByID(s.g, playerID)
	resources, projected := eventResources.Players[playerID]
	if !projected {
		resources = game.PlayerResources{Life: payer.Life, Energy: payer.EnergyCounters, Mana: payer.ManaPool.Units(), Riders: slices.Clone(payer.ManaRiders)}
	}
	resources.Mana = maps.Clone(resources.Mana)
	for _, produced := range production {
		resources.Mana[produced.Unit] += produced.Amount
	}
	eventResources.Players[playerID] = resources
	eventView := s.g.ResourceView(eventResources)
	var events []game.Event
	tapped := slices.ContainsFunc(candidate.Body.AdditionalCosts, func(additional cost.Additional) bool { return additional.Kind == cost.AdditionalTap })
	for index := len(s.g.Events) - 1; tapped && index >= 0; index-- {
		event := s.g.Events[index]
		if event.Kind == game.EventPermanentTapped && event.PermanentID == candidate.Source && event.TappedForMana {
			event.ProducedManaColors = colors
			events = append(events, event)
			break
		}
	}
	objectID := candidate.Source
	isLand := false
	tokenName := ""
	if candidate.SourceZone == zone.Hand {
		objectID = 0
	} else if permanent, ok := permanentByObjectID(s.g, candidate.Source); ok {
		isLand = permanentHasType(s.g, permanent, types.Land)
		tokenName = permanentTokenName(permanent)
	} else if snapshot, ok := lastKnownObject(s.g, candidate.Source); ok {
		isLand = slices.Contains(snapshot.Types, types.Land)
		tokenName = snapshot.TokenName
	}
	source := manaProducedSource{sourceID: obj.SourceCardID, objectID: objectID, controller: playerID, isLand: isLand, tokenName: tokenName, tokenDef: obj.SourceTokenDef}
	event := manaProducedEvent(source, playerID, colors, total, tapped)
	events = append(events, prepareEvent(eventView, event))
	events = append(events, prepareEvent(eventView, abilityActivatedEvent(obj, objectID, true)))
	additional, spentDelayed, ok := forecastTriggeredMana(eventView, events, playerID)
	if !ok {
		return nil, game.ResourceProjection{}, false
	}
	production = append(production, additional...)
	effects.SpentDelayed = spentDelayed
	for _, event := range events {
		if event.Kind == game.EventPermanentTapped {
			continue
		}
		event.ManaTriggersResolved = true
		effects.Events = append(effects.Events, event)
	}
	return production, effects, true
}

func candidateCardID(candidate payment.ManaActivationCandidate) id.ID {
	// Definitions are immutable; the request owns the cost-source identity.
	return candidate.CardID
}

func forecastManaContent(g *game.Game, obj *game.StackObject, content game.AbilityContent, payer game.PlayerID) ([]payment.ManaProduction, game.ResourceProjection, bool) {
	if len(content.Modes) != 1 || content.IsModal() || !content.Modes[0].Program.IsZero() {
		return nil, game.ResourceProjection{}, false
	}
	effects := game.ResourceProjection{Players: make(map[game.PlayerID]game.PlayerResources), Permanents: make(map[id.ID]game.PermanentResources)}
	var production []payment.ManaProduction
	for _, instruction := range content.Modes[0].Sequence {
		if instruction.Primitive == nil || !forecastInstructionPlain(instruction) {
			return nil, game.ResourceProjection{}, false
		}
		view := g.ResourceView(effects)
		r := &effectResolver{game: view, obj: obj}
		player, ok := playerByID(view, obj.Controller)
		if !ok {
			return nil, game.ResourceProjection{}, false
		}
		resources := game.PlayerResources{Life: player.Life, Energy: player.EnergyCounters, Mana: player.ManaPool.Units(), Riders: slices.Clone(player.ManaRiders)}
		switch instruction.Primitive.Kind() {
		case game.PrimitiveChoose:
			prim := instruction.Primitive.(game.Choose)
			if _, ok := linkedResolutionChoice(obj, string(prim.PublishChoice)); !ok {
				return nil, game.ResourceProjection{}, false
			}
		case game.PrimitiveAddMana:
			prim := instruction.Primitive.(game.AddMana)
			if len(prim.CombinationColors) != 0 || prim.EachControlledColor != nil || prim.Amount.IsDynamic() {
				return nil, game.ResourceProjection{}, false
			}
			_, recipient, units := evaluateAddMana(r, prim)
			if recipient != payer || len(units) == 0 {
				return nil, game.ResourceProjection{}, false
			}
			for _, unit := range units {
				production = append(production, payment.ManaProduction{Unit: unit.unit, Amount: unit.amount, Rider: unit.rider})
			}
		case game.PrimitiveGainLife:
			prim := instruction.Primitive.(game.GainLife)
			gain := prepareLifeGain(view, obj.Controller, prim.Amount.Value())
			resources.Life += gain
			if gain > 0 {
				effects.Events = append(effects.Events, prepareEvent(view, game.Event{Kind: game.EventLifeGained, Player: obj.Controller, Amount: gain}))
			}
		case game.PrimitiveLoseLife:
			prim := instruction.Primitive.(game.LoseLife)
			loss := prepareLifeLoss(view, obj.Controller, prim.Amount.Value())
			resources.Life -= loss
			if loss > 0 {
				effects.Events = append(effects.Events, prepareEvent(view, game.Event{Kind: game.EventLifeLost, Player: obj.Controller, Amount: loss}))
			}
		case game.PrimitiveDamage:
			prim := instruction.Primitive.(game.Damage)
			if damageSourceHasKeyword(view, obj.SourceCardID, obj.SourceID, game.Infect) ||
				damageSourceHasKeyword(view, obj.SourceCardID, obj.SourceID, game.Lifelink) {
				return nil, game.ResourceProjection{}, false
			}
			event := damageEvent{sourceID: obj.SourceCardID, sourceObjectID: obj.SourceID, controller: obj.Controller, player: obj.Controller, amount: prim.Amount.Value()}
			amount, ok := deterministicDamageAmount(view, event)
			if !ok {
				return nil, game.ResourceProjection{}, false
			}
			if !playerRuleEffectActive(view, obj.Controller, game.RuleEffectDamageDoesntCauseLifeLoss) {
				lost := prepareLifeLoss(view, obj.Controller, amount)
				resources.Life -= lost
				if lost > 0 {
					effects.Events = append(effects.Events, prepareEvent(view, game.Event{Kind: game.EventLifeLost, Player: obj.Controller, Amount: lost}))
				}
			}
			if amount > 0 {
				effects.Events = append(effects.Events, prepareEvent(view, damageReplacementEvent(view, damageEvent{sourceID: obj.SourceCardID, sourceObjectID: obj.SourceID, controller: obj.Controller, player: obj.Controller, amount: amount})))
			}
		case game.PrimitiveAddCounter:
			prim := instruction.Primitive.(game.AddCounter)
			permanent, ok := permanentByObjectID(view, obj.SourceID)
			if !ok {
				return nil, game.ResourceProjection{}, false
			}
			amount, ok := deterministicCounterPlacement(view, obj.Controller, permanent, prim.CounterKind, prim.Amount.Value())
			if !ok {
				return nil, game.ResourceProjection{}, false
			}
			counters := permanent.Counters.All()
			counters[prim.CounterKind] += amount
			effects.Permanents[obj.SourceID] = game.PermanentResources{Tapped: permanent.Tapped, Counters: counters}
			if amount > 0 {
				effects.Events = append(effects.Events, prepareEvent(view, game.Event{Kind: game.EventCountersAdded, Controller: obj.Controller, Player: permanent.Owner, CardID: permanent.CardInstanceID, PermanentID: obj.SourceID, CounterKind: prim.CounterKind, Amount: amount}))
			}
		case game.PrimitiveRemoveCounter:
			prim := instruction.Primitive.(game.RemoveCounter)
			permanent, ok := permanentByObjectID(view, obj.SourceID)
			if !ok {
				return nil, game.ResourceProjection{}, false
			}
			counters := permanent.Counters.All()
			removed := min(counters[prim.CounterKind], prim.Amount.Value())
			counters[prim.CounterKind] -= removed
			effects.Permanents[obj.SourceID] = game.PermanentResources{Tapped: permanent.Tapped, Counters: counters}
			if removed > 0 {
				effects.Events = append(effects.Events, prepareEvent(view, game.Event{Kind: game.EventCountersRemoved, Controller: obj.Controller, Player: permanent.Owner, CardID: permanent.CardInstanceID, PermanentID: obj.SourceID, CounterKind: prim.CounterKind, Amount: removed}))
			}
		default:
			return nil, game.ResourceProjection{}, false
		}
		effects.Players[obj.Controller] = resources
	}
	return production, effects, true
}

func forecastInstructionPlain(i game.Instruction) bool {
	return !i.Condition.Exists && i.ConditionGate == "" && i.PublishCondition == "" && !i.CardCondition.Exists &&
		!i.ResultGate.Exists && !i.Optional && !i.OptionalActor.Exists && i.PublishResult == "" && len(i.TemptingOfferBody) == 0 &&
		!i.OptionalActorGroup.Exists && !i.ForEachPlayerGroup.Exists && i.OptionalDecisionGate == "" && i.PublishOptionalDecision == "" && !i.TemptingOffer
}

func deterministicDamageAmount(g *game.Game, event damageEvent) (int, bool) {
	batch := damageBatch{game: g, usedCounters: make(map[id.ID]bool), members: []pendingDamage{{event: event, applied: make(map[id.ID]bool), shields: make(map[int]bool)}}}
	for {
		modifiers := batch.modifiers()
		if len(modifiers) == 0 {
			return batch.members[0].event.amount, true
		}
		if len(modifiers) != 1 {
			return 0, false
		}
		modifier := modifiers[0]
		switch modifier.kind {
		case damageProtection:
			return 0, true
		case damageReplacement:
			replacement := modifier.replacement
			if replacement.DamagePreventedBecomesPlusOneCounters || replacement.DamagePreventedRemovesPlusOneCounter {
				return 0, false
			}
			batch.members[0].applied[replacement.ID] = true
			batch.members[0].event.amount = damageReplacementAmount(replacement, batch.members[0].event.amount)
		default:
			// Finite shields, redirection and their consequential events are not
			// resource-only shapes. Never execute a damage batch to probe them.
			return 0, false
		}
	}
}

func damageReplacementAmount(replacement game.ReplacementEffect, amount int) int {
	if replacement.DamagePreventAll {
		return 0
	}
	if replacement.DamagePreventAmount > 0 {
		return max(0, amount-replacement.DamagePreventAmount)
	}
	if replacement.DamageMultiplier > 1 {
		amount *= replacement.DamageMultiplier
	}
	return max(0, amount+replacement.DamageAddend)
}

func (s *rulesPaymentState) PrepareManaEvent(event game.Event) game.Event {
	return prepareEvent(s.g, event)
}
