package rules

import (
	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

func (e *Engine) applyActivateAbility(g *game.Game, playerID game.PlayerID, activate action.ActivateAbilityAction) bool {
	return e.applyActivateAbilityWithChoices(g, playerID, activate, [game.NumPlayers]PlayerAgent{}, nil)
}

// applyActivateAbilityWithChoices activates an ability (CR 602.2). CR 602.2a:
// the ability is put on top of the stack as an object that isn't a card.
// CR 602.2b: the rest of activation follows the casting steps in CR 601.2b–i —
// choosing targets and modes, then determining and paying the total cost. The
// special activation forms (cycling, ninjutsu, and hand/graveyard abilities) are
// dispatched first; a mana ability is handled inline because it doesn't go on
// the stack and resolves immediately (CR 605.3b). A library top changed during
// activation (or a mana ability's resolution) stays hidden until it finishes
// (CR 401.5).
func (e *Engine) applyActivateAbilityWithChoices(g *game.Game, playerID game.PlayerID, activate action.ActivateAbilityAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	return e.attemptActivateAbilityWithChoices(g, playerID, activate, agents, log) == actionApplied
}

// attemptActivateAbilityWithChoices activates an ability as
// applyActivateAbilityWithChoices does. Once its total cost is locked, when
// that cost includes a mana payment, the player may activate mana abilities
// (CR 602.2b, 601.2g); an activation whose cost then goes unpaid is reversed
// with them (CR 733.1).
func (e *Engine) attemptActivateAbilityWithChoices(g *game.Game, playerID game.PlayerID, activate action.ActivateAbilityAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) actionOutcome {
	defer g.OpenLibraryTopWindow()()
	if outcome := e.applyCyclingAbilityWithChoices(g, playerID, activate, agents, log); outcome != actionFailed {
		return outcome
	}
	if outcome := e.applyNinjutsuAbilityWithChoices(g, playerID, activate, agents, log); outcome != actionFailed {
		return outcome
	}
	if outcome := e.applyHandAbilityWithChoices(g, playerID, activate, agents, log); outcome != actionFailed {
		return outcome
	}
	if e.applyHandManaAbilityWithChoices(g, playerID, activate, agents, log) {
		return actionApplied
	}
	if outcome := e.applyGraveyardAbilityWithChoices(g, playerID, activate, agents, log); outcome != actionFailed {
		return outcome
	}
	return e.applyPermanentAbilityWithChoices(g, playerID, activate, agents, log)
}

// permanentActivation is the cost of activating the activated ability body of
// permanent as activate describes, when it's an equip-like or other activated
// ability the player may begin activating (see activationFeasible), and the
// zero activationCost otherwise.
func (e *Engine) permanentActivation(g *game.Game, playerID game.PlayerID, permanent *game.Permanent, body *game.ActivatedAbility, activate action.ActivateAbilityAction) activationCost {
	if equip := equipActivation(g, playerID, permanent, body, activate.AbilityIndex, activate.Targets, activate.XValue, activate.ChosenModes); e.activationFeasible(g, playerID, equip) {
		return equip
	}
	if general := generalActivation(g, playerID, permanent, body, activate.AbilityIndex, activate.Targets, activate.XValue, activate.ChosenModes); e.activationFeasible(g, playerID, general) {
		return general
	}
	return activationCost{}
}

// applyPermanentAbilityWithChoices activates an ability of a permanent: a mana
// ability, a loyalty ability or another activated ability.
func (e *Engine) applyPermanentAbilityWithChoices(g *game.Game, playerID game.PlayerID, activate action.ActivateAbilityAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) actionOutcome {
	permanent, body, ok := activatedAbilitySource(g, playerID, activate.SourceID, activate.AbilityIndex)
	if !ok {
		return actionFailed
	}

	if manaBody, ok := body.(*game.ManaAbility); ok && canActivateManaAbility(g, playerID, permanent, manaBody, activate.AbilityIndex) {
		if len(activate.Targets) != 0 || len(activate.TargetCounts) != 0 || activate.XValue != 0 || len(activate.ChosenModes) != 0 {
			return actionFailed
		}
		prefs := e.paymentPreferencesForPermanentCost(g, playerID, permanent, manaCostPtr(manaBody.ManaCost), abilityAdditionalCosts(manaBody.AdditionalCosts), 0, agents, log)
		manaSource := captureManaProducedSource(g, permanent)
		eventsBeforePayment := len(g.Events)
		costPaid, ok := paymentOrch.payAbilityCosts(g, payment.AbilityRequest{
			PlayerID:        playerID,
			Source:          permanent,
			ManaCost:        manaBody.ManaCost,
			AdditionalCosts: abilityAdditionalCosts(manaBody.AdditionalCosts),
			XValue:          0,
			Prefs:           prefs,
			ForMana:         true,
		})
		if !ok {
			return actionFailed
		}
		obj := &game.StackObject{
			ID:               g.IDGen.Next(),
			Kind:             game.StackActivatedAbility,
			SourceID:         permanent.ObjectID,
			Face:             permanent.Face,
			SourceCardID:     permanent.CardInstanceID,
			SourceTokenDef:   permanent.TokenDef,
			AbilityIndex:     activate.AbilityIndex,
			Controller:       playerID,
			PaidCostSubjects: costPaid.subjects,
		}
		if len(manaBody.Content.Modes) > 0 {
			seedEntryChoices(obj, permanent)
			before := manaPoolColorSnapshot(g, playerID)
			e.resolveAbilityContentWithChoices(g, obj, manaBody.Content, agents, log)
			produced := producedManaColorsSince(g, playerID, before)
			recordTappedForManaProduced(g, permanent.ObjectID, produced)
			emitManaProducedEvent(g, manaSource, playerID, produced,
				producedManaAmountSince(g, playerID, before),
				manaAbilityTappedSourceSince(g, permanent.ObjectID, eventsBeforePayment))
		}
		emitAbilityActivatedEvent(g, obj, permanent.ObjectID, true)
		recordActivatedAbilityUse(g, permanent.ObjectID, activate.AbilityIndex, manaBody.Timing)
		return actionApplied
	}

	card, ok := permanentCardDef(g, permanent)
	if !ok {
		return actionFailed
	}
	activatedBody, activatedOK := body.(*game.ActivatedAbility)
	loyaltyBody, loyaltyOK := body.(*game.LoyaltyAbility)
	if !activatedOK && !loyaltyOK {
		return actionFailed
	}
	if activatedOK && e.permanentActivation(g, playerID, permanent, activatedBody, activate).payable == nil {
		return actionFailed
	}
	if loyaltyOK && (len(activate.ChosenModes) != 0 || !canActivateLoyaltyAbility(g, playerID, permanent, loyaltyBody, activate.AbilityIndex, activate.Targets, activate.XValue)) {
		return actionFailed
	}
	completedTargets, ok := e.completeAbilityAnnouncementTargetsWithModes(g, playerID, card, permanent.ObjectID, body, activate.ChosenModes, activate.Targets, agents, log)
	if !ok {
		return actionFailed
	}
	activate.Targets = completedTargets
	targetCounts, ok := bodyTargetCountsWithModesAndRecorded(g, playerID, card, permanent.ObjectID, body, activate.ChosenModes, activate.TargetCounts, activate.Targets, activate.XValue)
	if !ok {
		return actionFailed
	}
	// The total cost is locked once the targets are final (CR 601.2f).
	var activation activationCost
	if activatedOK {
		activation = e.permanentActivation(g, playerID, permanent, activatedBody, activate)
		if activation.payable == nil {
			return actionFailed
		}
	}
	if loyaltyOK && (len(activate.ChosenModes) != 0 || !canActivateLoyaltyAbility(g, playerID, permanent, loyaltyBody, activate.AbilityIndex, activate.Targets, activate.XValue)) {
		return actionFailed
	}
	sourceCardID := permanent.CardInstanceID
	sourceTokenDef := permanent.TokenDef
	resolutionUse := permanentActivatedResolutionUse(g, permanent, body, activate.AbilityIndex)
	obj := &game.StackObject{
		ID:                     g.IDGen.Next(),
		Kind:                   game.StackActivatedAbility,
		SourceID:               permanent.ObjectID,
		Face:                   permanent.Face,
		SourceCardID:           sourceCardID,
		SourceTokenDef:         sourceTokenDef,
		AbilityIndex:           activate.AbilityIndex,
		Controller:             playerID,
		Targets:                append([]game.Target(nil), activate.Targets...),
		TargetCounts:           targetCounts,
		ChosenModes:            append([]int(nil), activate.ChosenModes...),
		XValue:                 activate.XValue,
		ActivatedResolutionUse: resolutionUse,
	}
	if activatedOK {
		obj.InlineActivated = activatedBody
	}
	if loyaltyOK {
		obj.InlineLoyalty = loyaltyBody
	}
	targetEvents := prepareTargetEvents(g, obj)
	manaCost := opt.V[cost.Mana]{}
	paymentManaCost := opt.V[cost.Mana]{}
	var costModifiers []game.CostModifier
	var additionalCosts []cost.Additional
	var alternativeCosts []cost.Alternative
	timing := game.NoTimingRestriction
	if activatedOK {
		sourceCard, _ := g.GetCardInstance(permanent.CardInstanceID)
		manaCost = effectiveActivatedAbilityCost(g, playerID, sourceCard, activatedBody)
		paymentManaCost = activatedBody.ManaCost
		costModifiers = activatedAbilityCostModifiers(g, playerID, sourceCard, activatedBody)
		additionalCosts = abilityAdditionalCosts(activatedBody.AdditionalCosts)
		alternativeCosts = append([]cost.Alternative(nil), activatedBody.AlternativeCosts...)
		timing = activatedBody.Timing
	}
	undoMana := e.activationPaymentManaWindow(g, playerID, activation, agents, log)
	if undoMana != nil && !activation.payableNow(g) {
		return reverseActivation(undoMana)
	}
	var tapExclusions []id.ID
	if hasTapCostOf(additionalCosts) {
		tapExclusions = append(tapExclusions, permanent.ObjectID)
	}
	prefs := e.paymentPreferencesForPermanentCost(g, playerID, permanent, manaCostPtr(manaCost), additionalCosts, activate.XValue, agents, log, tapExclusions...)
	costPaid, ok := paymentOrch.payAbilityCosts(g, payment.AbilityRequest{
		PlayerID:         playerID,
		Source:           permanent,
		ManaCost:         paymentManaCost,
		CostModifiers:    costModifiers,
		AdditionalCosts:  additionalCosts,
		AlternativeCosts: alternativeCosts,
		XValue:           activate.XValue,
		Prefs:            prefs,
	})
	if !ok {
		if undoMana != nil {
			return reverseActivation(undoMana)
		}
		return actionFailed
	}
	if loyaltyOK {
		applyLoyaltyCost(permanent, loyaltyBody.LoyaltyCost)
	}
	obj.SacrificedAsCostIDs = costPaid.sacrificedIDs
	obj.PaidCostSubjects = costPaid.subjects
	obj.TappedAsCostIDs = costPaid.tappedIDs
	obj.ExiledAsCostIDs = costPaid.exiledIDs
	pushAbilityToStackWithTargets(g, obj, targetEvents)
	emitAbilityActivatedEvent(g, obj, permanent.ObjectID, false)
	recordActivatedAbilityUse(g, permanent.ObjectID, activate.AbilityIndex, timing)
	if loyaltyOK {
		recordActivatedAbilityUse(g, permanent.ObjectID, -1, game.OncePerTurn)
	}
	return actionApplied
}

func (e *Engine) applyHandAbilityWithChoices(g *game.Game, playerID game.PlayerID, activate action.ActivateAbilityAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) actionOutcome {
	card, ability, ok := handActivatedAbilitySource(g, playerID, activate.SourceID, activate.AbilityIndex)
	if !ok || !e.activationFeasible(g, playerID, handActivation(g, playerID, card.ID, &ability, activate.AbilityIndex, activate.Targets, activate.XValue, activate.ChosenModes)) {
		return actionFailed
	}
	stackBody := activationStackBody(card, activate.AbilityIndex, &ability)
	sourceZoneVersion := card.ZoneVersion
	def := cardFaceOrDefault(card, game.FaceFront)
	resolutionUse := cardActivatedResolutionUse(card, def, activate.AbilityIndex)
	completedTargets, ok := e.completeAbilityAnnouncementTargetsWithModes(g, playerID, def, 0, &ability, activate.ChosenModes, activate.Targets, agents, log)
	if !ok {
		return actionFailed
	}
	// The total cost is locked once the targets are final (CR 601.2f).
	activation := handActivation(g, playerID, card.ID, &ability, activate.AbilityIndex, completedTargets, activate.XValue, activate.ChosenModes)
	if !e.activationFeasible(g, playerID, activation) {
		return actionFailed
	}
	targetCounts, ok := bodyTargetCountsWithModesAndRecorded(g, playerID, def, 0, &ability, activate.ChosenModes, activate.TargetCounts, completedTargets, activate.XValue)
	if !ok {
		return actionFailed
	}
	// The self-entry family (Talon Gates of Madara, Urban Retreat) keeps its
	// source card in the hand until resolution, when the ability's own
	// PutOnBattlefield moves it. It may carry arbitrary additional costs (a
	// return-to-hand, a life payment, etc.), so it pays through the general
	// ability-cost path. The discard-self family (Channel, Transmute) instead
	// removes its own card from the hand as its additional cost and pays only
	// mana here.
	selfEntry := bodyPutsSourceCardOntoBattlefield(&ability)
	discardSelf := !selfEntry && len(ability.AdditionalCosts) == 1 && abilityHasDiscardThisCardCost(ability.AdditionalCosts)
	stackID := g.IDGen.Next()
	targetEvents := prepareTargetEvents(g, &game.StackObject{
		ID: stackID, Kind: game.StackActivatedAbility, Controller: playerID,
		SourceID: card.ID, SourceCardID: card.ID, SourceZone: zone.Hand,
		SourceZoneVersion: sourceZoneVersion, AbilityIndex: activate.AbilityIndex,
		Targets: completedTargets, TargetCounts: targetCounts, ChosenModes: activate.ChosenModes,
		XValue: activate.XValue, InlineActivated: stackBody,
	})
	undoMana := e.activationPaymentManaWindow(g, playerID, activation, agents, log)
	if undoMana != nil && !activation.payableNow(g) {
		return reverseActivation(undoMana)
	}
	unpaid := func() actionOutcome {
		if undoMana != nil {
			return reverseActivation(undoMana)
		}
		return actionFailed
	}
	var additionalCostsPaid []string
	var sacrificedAsCostIDs, tappedAsCostIDs, exiledAsCostIDs []id.ID
	var paidCostSubjects []game.PaidCostSubject
	if selfEntry {
		prefs := e.paymentPreferencesForCostFromSource(g, playerID, manaCostPtr(ability.ManaCost), abilityAdditionalCosts(ability.AdditionalCosts), activate.XValue, card.ID, zone.Hand, agents, log)
		costPaid, ok := paymentOrch.payAbilityCosts(g, payment.AbilityRequest{
			PlayerID:         playerID,
			SourceCardID:     card.ID,
			SourceZone:       zone.Hand,
			ManaCost:         ability.ManaCost,
			AdditionalCosts:  abilityAdditionalCosts(ability.AdditionalCosts),
			AlternativeCosts: append([]cost.Alternative(nil), ability.AlternativeCosts...),
			XValue:           activate.XValue,
			Prefs:            prefs,
		})
		if !ok {
			return unpaid()
		}
		sacrificedAsCostIDs = costPaid.sacrificedIDs
		paidCostSubjects = costPaid.subjects
		tappedAsCostIDs = costPaid.tappedIDs
		exiledAsCostIDs = costPaid.exiledIDs
	} else {
		manaCost := effectiveHandAbilityCost(g, playerID, card, &ability)
		prefs := e.paymentPreferencesForCost(g, playerID, manaCost, nil, activate.XValue, agents, log)
		if !paymentOrch.payGenericCost(g, payment.GenericRequest{
			PlayerID: playerID,
			Cost:     manaCost,
			XValue:   activate.XValue,
			Prefs:    prefs,
		}) {
			return unpaid()
		}
		if discardSelf {
			for _, additional := range ability.AdditionalCosts {
				if additional.SubjectKey != "" {
					paidCostSubjects = append(paidCostSubjects, captureDiscardCostSubject(g, playerID, card.ID, additional.SubjectKey))
				}
			}
			if !discardCardFromHand(g, playerID, card.ID) {
				panic("hand activation source disappeared after validation")
			}
			additionalCostsPaid = []string{"Discard this card"}
		}
	}
	obj := &game.StackObject{
		ID:                     stackID,
		Kind:                   game.StackActivatedAbility,
		SourceID:               card.ID,
		SourceCardID:           card.ID,
		SourceZone:             zone.Hand,
		SourceZoneVersion:      sourceZoneVersion,
		AbilityIndex:           activate.AbilityIndex,
		Controller:             playerID,
		Targets:                append([]game.Target(nil), completedTargets...),
		TargetCounts:           targetCounts,
		ChosenModes:            append([]int(nil), activate.ChosenModes...),
		XValue:                 activate.XValue,
		AdditionalCostsPaid:    additionalCostsPaid,
		SacrificedAsCostIDs:    sacrificedAsCostIDs,
		PaidCostSubjects:       paidCostSubjects,
		TappedAsCostIDs:        tappedAsCostIDs,
		ExiledAsCostIDs:        exiledAsCostIDs,
		InlineActivated:        stackBody,
		ActivatedResolutionUse: resolutionUse,
	}
	pushAbilityToStackWithTargets(g, obj, targetEvents)
	emitAbilityActivatedEvent(g, obj, 0, false)
	recordActivatedAbilityUse(g, card.ID, activate.AbilityIndex, ability.Timing)
	return actionApplied
}

// applyHandManaAbilityWithChoices activates a mana ability printed on a card in
// the player's hand whose cost is exiling that card from hand (Simian/Elvish
// Spirit Guide). The card is exiled as the activation cost and the add-mana
// content resolves immediately into the controller's mana pool; the ability
// never uses the stack, like any mana ability.
func (e *Engine) applyHandManaAbilityWithChoices(g *game.Game, playerID game.PlayerID, activate action.ActivateAbilityAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	card, body, ok := handManaAbilitySource(g, playerID, activate.SourceID, activate.AbilityIndex)
	if !ok || !canActivateHandManaAbility(g, playerID, card.ID, &body, activate.AbilityIndex) {
		return false
	}
	if len(activate.Targets) != 0 || len(activate.TargetCounts) != 0 || activate.XValue != 0 || len(activate.ChosenModes) != 0 {
		return false
	}
	sourceCardID := card.ID
	if !moveCardBetweenZones(g, playerID, card.ID, zone.Hand, zone.Exile) {
		return false
	}
	obj := &game.StackObject{
		ID:                  g.IDGen.Next(),
		Kind:                game.StackActivatedAbility,
		SourceID:            sourceCardID,
		SourceCardID:        sourceCardID,
		SourceZone:          zone.Exile,
		AbilityIndex:        activate.AbilityIndex,
		Controller:          playerID,
		AdditionalCostsPaid: []string{"Exile this card from your hand"},
	}
	if len(body.Content.Modes) > 0 {
		e.resolveAbilityContentWithChoices(g, obj, body.Content, agents, log)
	}
	emitAbilityActivatedEvent(g, obj, 0, true)
	recordActivatedAbilityUse(g, sourceCardID, activate.AbilityIndex, body.Timing)
	return true
}

func (e *Engine) applyGraveyardAbilityWithChoices(g *game.Game, playerID game.PlayerID, activate action.ActivateAbilityAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) actionOutcome {
	card, ability, ok := graveyardAbilitySource(g, playerID, activate.SourceID, activate.AbilityIndex)
	if !ok || !e.activationFeasible(g, playerID, graveyardActivation(g, playerID, card.ID, &ability, activate.AbilityIndex, activate.Targets, activate.XValue, activate.ChosenModes)) {
		return actionFailed
	}
	stackBody := activationStackBody(card, activate.AbilityIndex, &ability)
	sourceZoneVersion := card.ZoneVersion
	def := cardFaceOrDefault(card, game.FaceFront)
	resolutionUse := cardActivatedResolutionUse(card, def, activate.AbilityIndex)
	completedTargets, ok := e.completeAbilityAnnouncementTargetsWithModes(g, playerID, def, 0, &ability, activate.ChosenModes, activate.Targets, agents, log)
	if !ok {
		return actionFailed
	}
	// The total cost is locked once the targets are final (CR 601.2f).
	activation := graveyardActivation(g, playerID, card.ID, &ability, activate.AbilityIndex, completedTargets, activate.XValue, activate.ChosenModes)
	if !e.activationFeasible(g, playerID, activation) {
		return actionFailed
	}
	targetCounts, ok := bodyTargetCountsWithModesAndRecorded(g, playerID, def, 0, &ability, activate.ChosenModes, activate.TargetCounts, completedTargets, activate.XValue)
	if !ok {
		return actionFailed
	}
	undoMana := e.activationPaymentManaWindow(g, playerID, activation, agents, log)
	if undoMana != nil && !activation.payableNow(g) {
		return reverseActivation(undoMana)
	}
	prefs := e.paymentPreferencesForCostFromSource(g, playerID, manaCostPtr(ability.ManaCost), abilityAdditionalCosts(ability.AdditionalCosts), activate.XValue, card.ID, zone.Graveyard, agents, log)
	stackID := g.IDGen.Next()
	targetEvents := prepareTargetEvents(g, &game.StackObject{
		ID: stackID, Kind: game.StackActivatedAbility, Controller: playerID,
		SourceID: card.ID, SourceCardID: card.ID, SourceZone: zone.Graveyard,
		SourceZoneVersion: sourceZoneVersion, AbilityIndex: activate.AbilityIndex,
		Targets: completedTargets, TargetCounts: targetCounts, ChosenModes: activate.ChosenModes,
		XValue: activate.XValue, InlineActivated: stackBody,
	})
	costPaid, ok := paymentOrch.payAbilityCosts(g, payment.AbilityRequest{
		PlayerID:         playerID,
		SourceCardID:     card.ID,
		SourceZone:       zone.Graveyard,
		ManaCost:         ability.ManaCost,
		AdditionalCosts:  abilityAdditionalCosts(ability.AdditionalCosts),
		AlternativeCosts: append([]cost.Alternative(nil), ability.AlternativeCosts...),
		XValue:           activate.XValue,
		Prefs:            prefs,
	})
	if !ok {
		if undoMana != nil {
			return reverseActivation(undoMana)
		}
		return actionFailed
	}
	obj := &game.StackObject{
		ID:                     stackID,
		Kind:                   game.StackActivatedAbility,
		SourceID:               card.ID,
		SourceCardID:           card.ID,
		SourceZone:             zone.Graveyard,
		SourceZoneVersion:      sourceZoneVersion,
		AbilityIndex:           activate.AbilityIndex,
		Controller:             playerID,
		Targets:                append([]game.Target(nil), completedTargets...),
		TargetCounts:           targetCounts,
		ChosenModes:            append([]int(nil), activate.ChosenModes...),
		XValue:                 activate.XValue,
		SacrificedAsCostIDs:    costPaid.sacrificedIDs,
		PaidCostSubjects:       costPaid.subjects,
		TappedAsCostIDs:        costPaid.tappedIDs,
		ExiledAsCostIDs:        costPaid.exiledIDs,
		InlineActivated:        stackBody,
		ActivatedResolutionUse: resolutionUse,
	}
	pushAbilityToStackWithTargets(g, obj, targetEvents)
	emitAbilityActivatedEvent(g, obj, 0, false)
	recordActivatedAbilityUse(g, card.ID, activate.AbilityIndex, ability.Timing)
	return actionApplied
}

func canActivateLoyaltyAbility(g *game.Game, playerID game.PlayerID, permanent *game.Permanent, body *game.LoyaltyAbility, abilityIndex int, targets []game.Target, xValue int) bool {
	_ = abilityIndex
	if body == nil || !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer || permanent.PhasedOut || effectiveController(g, permanent) != playerID {
		return false
	}
	if xValue != 0 || !bodyFunctionsOnBattlefield(body) || !permanentHasType(g, permanent, types.Planeswalker) {
		return false
	}
	if !isSorcerySpeed(g, playerID) || g.ActivatedAbilitiesThisTurn[game.ActivatedAbilityUse{SourceID: permanent.ObjectID, AbilityIndex: -1}] {
		return false
	}
	if body.LoyaltyCost < 0 && permanent.Counters.Get(counter.Loyalty) < -body.LoyaltyCost {
		return false
	}
	if !activationConditionSatisfied(g, playerID, permanent, body.ActivationCondition) {
		return false
	}
	card, ok := permanentCardDef(g, permanent)
	if !ok || !targetsValidForBodyFromSourceObject(g, playerID, card, permanent.ObjectID, body, targets) {
		return false
	}
	return paymentOrch.buildAbilityCostPlan(g, payment.AbilityRequest{PlayerID: playerID, Source: permanent, XValue: xValue})
}

func applyLoyaltyCost(permanent *game.Permanent, loyalty int) {
	if loyalty >= 0 {
		permanent.Counters.Add(counter.Loyalty, loyalty)
		return
	}
	permanent.Counters.Remove(counter.Loyalty, -loyalty)
}

func (e *Engine) applyCyclingAbility(g *game.Game, playerID game.PlayerID, activate action.ActivateAbilityAction) bool {
	return e.applyCyclingAbilityWithChoices(g, playerID, activate, [game.NumPlayers]PlayerAgent{}, nil) == actionApplied
}

func (e *Engine) applyCyclingAbilityWithChoices(g *game.Game, playerID game.PlayerID, activate action.ActivateAbilityAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) actionOutcome {
	if len(activate.TargetCounts) != 0 || len(activate.ChosenModes) != 0 {
		return actionFailed
	}
	card, ability, ok := cyclingAbilitySource(g, playerID, activate.SourceID, activate.AbilityIndex)
	if !ok {
		return actionFailed
	}
	activation := cyclingActivation(g, playerID, activate.SourceID, &ability, activate.AbilityIndex, activate.Targets, activate.XValue)
	if !e.activationFeasible(g, playerID, activation) {
		return actionFailed
	}
	stackBody := activationStackBody(card, activate.AbilityIndex, &ability)
	effectiveCost := effectiveCyclingCost(g, playerID, card, &ability)
	undoMana := e.activationPaymentManaWindow(g, playerID, activation, agents, log)
	if undoMana != nil && !activation.payableNow(g) {
		return reverseActivation(undoMana)
	}
	prefs := e.paymentPreferencesForCost(g, playerID, effectiveCost, nil, activate.XValue, agents, log)
	if !paymentOrch.payGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: effectiveCost, XValue: activate.XValue, Prefs: prefs}) {
		if undoMana != nil {
			return reverseActivation(undoMana)
		}
		return actionFailed
	}
	if !discardCardFromHand(g, playerID, card.ID) {
		panic("cycling card disappeared from hand after validation")
	}
	emitEvent(g, game.Event{
		Kind:       game.EventCycled,
		SourceID:   card.ID,
		Controller: playerID,
		Player:     playerID,
		CardID:     card.ID,
	})
	obj := &game.StackObject{
		ID:                  g.IDGen.Next(),
		Kind:                game.StackActivatedAbility,
		SourceID:            card.ID,
		SourceCardID:        card.ID,
		AbilityIndex:        activate.AbilityIndex,
		Controller:          playerID,
		Targets:             append([]game.Target(nil), activate.Targets...),
		XValue:              activate.XValue,
		AdditionalCostsPaid: []string{"Discard this card"},
		InlineActivated:     stackBody,
	}
	pushAbilityToStack(g, obj)
	emitAbilityActivatedEvent(g, obj, 0, false)
	return actionApplied
}

func (e *Engine) applyNinjutsuAbilityWithChoices(g *game.Game, playerID game.PlayerID, activate action.ActivateAbilityAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) actionOutcome {
	if len(activate.TargetCounts) != 0 || len(activate.ChosenModes) != 0 {
		return actionFailed
	}
	card, ability, ok := handActivatedAbilitySource(g, playerID, activate.SourceID, activate.AbilityIndex)
	if !ok {
		return actionFailed
	}
	activation := ninjutsuActivation(g, playerID, activate.SourceID, &ability, activate.AbilityIndex, activate.Targets, activate.XValue)
	if !e.activationFeasible(g, playerID, activation) {
		return actionFailed
	}
	attacker := chooseNinjutsuAttacker(e, g, playerID, unblockedAttackers(g, playerID), agents, log)
	if attacker == nil {
		return actionFailed
	}
	attackTarget, ok := attackTargetForAttacker(g, attacker.ObjectID)
	if !ok || attackerWasBlocked(g, attacker.ObjectID) {
		return actionFailed
	}
	undoMana := e.activationPaymentManaWindow(g, playerID, activation, agents, log)
	if undoMana != nil {
		// A mana ability activated while paying may have removed the attacker
		// chosen to return or the card itself.
		if _, attacking := attackTargetForAttacker(g, attacker.ObjectID); !attacking || attackerWasBlocked(g, attacker.ObjectID) || !activation.payableNow(g) {
			return reverseActivation(undoMana)
		}
	}
	prefs := e.paymentPreferencesForCost(g, playerID, manaCostPtr(ability.ManaCost), nil, 0, agents, log)
	if !paymentOrch.payGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: manaCostPtr(ability.ManaCost), Prefs: prefs}) {
		if undoMana != nil {
			return reverseActivation(undoMana)
		}
		return actionFailed
	}
	removePermanentFromCombat(g, attacker.ObjectID)
	if !movePermanentToZone(g, attacker, zone.Hand) {
		panic("Ninjutsu attacker disappeared after validation")
	}
	obj := &game.StackObject{
		ID:                   g.IDGen.Next(),
		Kind:                 game.StackActivatedAbility,
		SourceID:             card.ID,
		SourceCardID:         card.ID,
		SourceZone:           zone.Hand,
		SourceZoneVersion:    card.ZoneVersion,
		AbilityIndex:         activate.AbilityIndex,
		Controller:           playerID,
		Ninjutsu:             true,
		NinjutsuAttackTarget: attackTarget,
		AdditionalCostsPaid:  []string{"Return an unblocked attacker you control to its owner's hand"},
	}
	pushAbilityToStack(g, obj)
	emitAbilityActivatedEvent(g, obj, 0, false)
	return actionApplied
}

func chooseNinjutsuAttacker(e *Engine, g *game.Game, playerID game.PlayerID, attackers []*game.Permanent, agents [game.NumPlayers]PlayerAgent, log *TurnLog) *game.Permanent {
	if len(attackers) == 0 {
		return nil
	}
	if len(attackers) == 1 {
		return attackers[0]
	}
	options := make([]game.ChoiceOption, 0, len(attackers))
	for i, attacker := range attackers {
		options = append(options, game.ChoiceOption{
			Index: i, Label: permanentChoiceLabel(g, attacker),
			Card: permanentChoiceInfo(g, attacker), Targets: []game.Target{game.PermanentTarget(attacker.ObjectID)},
		})
	}
	selected := e.chooseChoice(g, agents, game.ChoiceRequest{
		Kind:             game.ChoicePayment,
		Player:           playerID,
		Prompt:           "Choose an unblocked attacker to return",
		Options:          options,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{0},
	}, log)
	if len(selected) != 1 || selected[0] < 0 || selected[0] >= len(attackers) {
		return nil
	}
	return attackers[selected[0]]
}

func unblockedAttackers(g *game.Game, playerID game.PlayerID) []*game.Permanent {
	if g.Combat == nil ||
		g.Turn.Phase != game.PhaseCombat ||
		g.Turn.Step < game.StepDeclareBlockers ||
		g.Turn.Step > game.StepEndOfCombat {
		return nil
	}
	var attackers []*game.Permanent
	for _, attack := range g.Combat.Attackers {
		permanent, ok := permanentByObjectID(g, attack.Attacker)
		if !ok || effectiveController(g, permanent) != playerID || attackerWasBlocked(g, attack.Attacker) {
			continue
		}
		attackers = append(attackers, permanent)
	}
	return attackers
}

func attackTargetForAttacker(g *game.Game, attackerID id.ID) (game.AttackTarget, bool) {
	if g.Combat == nil {
		return game.AttackTarget{}, false
	}
	for _, attack := range g.Combat.Attackers {
		if attack.Attacker == attackerID {
			return attack.Target, true
		}
	}
	return game.AttackTarget{}, false
}

func canActivateNinjutsuAbility(g *game.Game, playerID game.PlayerID, cardID id.ID, body *game.ActivatedAbility, abilityIndex int, targets []game.Target, xValue int) bool {
	return ninjutsuActivation(g, playerID, cardID, body, abilityIndex, targets, xValue).payableNow(g)
}

// ninjutsuActivation is the cost of activating the ninjutsu ability body of the
// card cardID, or the zero activationCost when that isn't legal. Its other
// component, returning an unblocked attacker, can be paid while there is one.
func ninjutsuActivation(g *game.Game, playerID game.PlayerID, cardID id.ID, body *game.ActivatedAbility, abilityIndex int, targets []game.Target, xValue int) activationCost {
	if body == nil || !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return activationCost{}
	}
	if xValue != 0 ||
		abilityIndex < 0 ||
		!game.BodyHasKeyword(body, game.Ninjutsu) ||
		game.BodyFunctionZone(body) != zone.Hand ||
		body.Timing != game.DuringCombat ||
		len(targets) != 0 ||
		len(game.BodyTargets(body)) != 0 ||
		!abilityHasReturnUnblockedAttackerCost(body.AdditionalCosts) ||
		len(unblockedAttackers(g, playerID)) == 0 {
		return activationCost{}
	}
	_, gotAbility, ok := handActivatedAbilitySource(g, playerID, cardID, abilityIndex)
	if !ok || !game.BodyHasKeyword(&gotAbility, game.Ninjutsu) {
		return activationCost{}
	}
	inHand := cardSourcePresent(handActivatedAbilitySource, playerID, cardID, abilityIndex)
	return genericRequestCost(payment.GenericRequest{PlayerID: playerID, Cost: manaCostPtr(body.ManaCost)}, func(g *game.Game) bool {
		return inHand(g) && len(unblockedAttackers(g, playerID)) != 0
	})
}
