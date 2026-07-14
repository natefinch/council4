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
// the stack and resolves immediately (CR 605.3b).
func (e *Engine) applyActivateAbilityWithChoices(g *game.Game, playerID game.PlayerID, activate action.ActivateAbilityAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	if e.applyCyclingAbilityWithChoices(g, playerID, activate, agents, log) {
		return true
	}
	if e.applyNinjutsuAbilityWithChoices(g, playerID, activate, agents, log) {
		return true
	}
	if e.applyHandAbilityWithChoices(g, playerID, activate, agents, log) {
		return true
	}
	if e.applyHandManaAbilityWithChoices(g, playerID, activate, agents, log) {
		return true
	}
	if e.applyGraveyardAbilityWithChoices(g, playerID, activate, agents, log) {
		return true
	}
	permanent, body, ok := activatedAbilitySource(g, playerID, activate.SourceID, activate.AbilityIndex)
	if !ok {
		return false
	}

	if manaBody, ok := body.(*game.ManaAbility); ok && canActivateManaAbility(g, playerID, permanent, manaBody, activate.AbilityIndex) {
		if len(activate.Targets) != 0 || len(activate.TargetCounts) != 0 || activate.XValue != 0 || len(activate.ChosenModes) != 0 {
			return false
		}
		prefs := e.paymentPreferencesForCost(g, playerID, manaCostPtr(manaBody.ManaCost), abilityAdditionalCosts(manaBody.AdditionalCosts), 0, agents, log)
		manaSource := captureManaProducedSource(g, permanent)
		eventsBeforePayment := len(g.Events)
		if _, ok := paymentOrch.payAbilityCosts(g, payment.AbilityRequest{
			PlayerID:        playerID,
			Source:          permanent,
			ManaCost:        manaBody.ManaCost,
			AdditionalCosts: abilityAdditionalCosts(manaBody.AdditionalCosts),
			XValue:          0,
			Prefs:           prefs,
			ForMana:         true,
		}); !ok {
			return false
		}
		obj := &game.StackObject{
			ID:             g.IDGen.Next(),
			Kind:           game.StackActivatedAbility,
			SourceID:       permanent.ObjectID,
			Face:           permanent.Face,
			SourceCardID:   permanent.CardInstanceID,
			SourceTokenDef: permanent.TokenDef,
			AbilityIndex:   activate.AbilityIndex,
			Controller:     playerID,
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
		return true
	}

	card, ok := permanentCardDef(g, permanent)
	if !ok {
		return false
	}
	activatedBody, activatedOK := body.(*game.ActivatedAbility)
	loyaltyBody, loyaltyOK := body.(*game.LoyaltyAbility)
	if !activatedOK && !loyaltyOK {
		return false
	}
	if activatedOK &&
		!canActivateEquipAbilityWithModes(g, playerID, permanent, activatedBody, activate.AbilityIndex, activate.Targets, activate.XValue, activate.ChosenModes) &&
		!canActivateGeneralAbilityWithModes(g, playerID, permanent, activatedBody, activate.AbilityIndex, activate.Targets, activate.XValue, activate.ChosenModes) &&
		!loyaltyOK {
		return false
	}
	if loyaltyOK && (len(activate.ChosenModes) != 0 || !canActivateLoyaltyAbility(g, playerID, permanent, loyaltyBody, activate.AbilityIndex, activate.Targets, activate.XValue)) {
		return false
	}
	completedTargets, ok := e.completeAbilityAnnouncementTargetsWithModes(g, playerID, card, permanent.ObjectID, body, activate.ChosenModes, activate.Targets, agents, log)
	if !ok {
		return false
	}
	activate.Targets = completedTargets
	targetCounts, ok := bodyTargetCountsWithModesAndRecorded(g, playerID, card, permanent.ObjectID, body, activate.ChosenModes, activate.TargetCounts, activate.Targets)
	if !ok {
		return false
	}
	if activatedOK &&
		!canActivateEquipAbilityWithModes(g, playerID, permanent, activatedBody, activate.AbilityIndex, activate.Targets, activate.XValue, activate.ChosenModes) &&
		!canActivateGeneralAbilityWithModes(g, playerID, permanent, activatedBody, activate.AbilityIndex, activate.Targets, activate.XValue, activate.ChosenModes) &&
		!loyaltyOK {
		return false
	}
	if loyaltyOK && (len(activate.ChosenModes) != 0 || !canActivateLoyaltyAbility(g, playerID, permanent, loyaltyBody, activate.AbilityIndex, activate.Targets, activate.XValue)) {
		return false
	}
	sourceCardID := permanent.CardInstanceID
	sourceTokenDef := permanent.TokenDef
	manaCost := opt.V[cost.Mana]{}
	var additionalCosts []cost.Additional
	var alternativeCosts []cost.Alternative
	timing := game.NoTimingRestriction
	if activatedOK {
		sourceCard, _ := g.GetCardInstance(permanent.CardInstanceID)
		manaCost = effectiveActivatedAbilityCost(g, playerID, sourceCard, activatedBody)
		additionalCosts = abilityAdditionalCosts(activatedBody.AdditionalCosts)
		alternativeCosts = append([]cost.Alternative(nil), activatedBody.AlternativeCosts...)
		timing = activatedBody.Timing
	}
	var tapExclusions []id.ID
	if hasTapCostOf(additionalCosts) {
		tapExclusions = append(tapExclusions, permanent.ObjectID)
	}
	prefs := e.paymentPreferencesForCost(g, playerID, manaCostPtr(manaCost), additionalCosts, activate.XValue, agents, log, tapExclusions...)
	costPaid, ok := paymentOrch.payAbilityCosts(g, payment.AbilityRequest{
		PlayerID:         playerID,
		Source:           permanent,
		ManaCost:         manaCost,
		AdditionalCosts:  additionalCosts,
		AlternativeCosts: alternativeCosts,
		XValue:           activate.XValue,
		Prefs:            prefs,
	})
	if !ok {
		return false
	}
	if loyaltyOK {
		applyLoyaltyCost(permanent, loyaltyBody.LoyaltyCost)
	}
	obj := &game.StackObject{
		ID:                  g.IDGen.Next(),
		Kind:                game.StackActivatedAbility,
		SourceID:            permanent.ObjectID,
		Face:                permanent.Face,
		SourceCardID:        sourceCardID,
		SourceTokenDef:      sourceTokenDef,
		AbilityIndex:        activate.AbilityIndex,
		Controller:          playerID,
		Targets:             append([]game.Target(nil), activate.Targets...),
		TargetCounts:        targetCounts,
		ChosenModes:         append([]int(nil), activate.ChosenModes...),
		XValue:              activate.XValue,
		SacrificedAsCostIDs: costPaid.sacrificedIDs,
		ExiledAsCostIDs:     costPaid.exiledIDs,
	}
	if activatedOK {
		obj.InlineActivated = activatedBody
	}
	if loyaltyOK {
		obj.InlineLoyalty = loyaltyBody
	}
	pushAbilityToStack(g, obj)
	emitAbilityActivatedEvent(g, obj, permanent.ObjectID, false)
	recordActivatedAbilityUse(g, permanent.ObjectID, activate.AbilityIndex, timing)
	if loyaltyOK {
		recordActivatedAbilityUse(g, permanent.ObjectID, -1, game.OncePerTurn)
	}
	return true
}

func (e *Engine) applyHandAbilityWithChoices(g *game.Game, playerID game.PlayerID, activate action.ActivateAbilityAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	card, ability, ok := handActivatedAbilitySource(g, playerID, activate.SourceID, activate.AbilityIndex)
	if !ok || !canActivateHandAbilityWithModes(g, playerID, card.ID, &ability, activate.AbilityIndex, activate.Targets, activate.XValue, activate.ChosenModes) {
		return false
	}
	sourceZoneVersion := card.ZoneVersion
	def := cardFaceOrDefault(card, game.FaceFront)
	completedTargets, ok := e.completeAbilityAnnouncementTargetsWithModes(g, playerID, def, 0, &ability, activate.ChosenModes, activate.Targets, agents, log)
	if !ok || !canActivateHandAbilityWithModes(g, playerID, card.ID, &ability, activate.AbilityIndex, completedTargets, activate.XValue, activate.ChosenModes) {
		return false
	}
	targetCounts, ok := bodyTargetCountsWithModesAndRecorded(g, playerID, def, 0, &ability, activate.ChosenModes, activate.TargetCounts, completedTargets)
	if !ok {
		return false
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
	var additionalCostsPaid []string
	var sacrificedAsCostIDs, exiledAsCostIDs []id.ID
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
			return false
		}
		sacrificedAsCostIDs = costPaid.sacrificedIDs
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
			return false
		}
		if discardSelf {
			if !discardCardFromHand(g, playerID, card.ID) {
				panic("hand activation source disappeared after validation")
			}
			additionalCostsPaid = []string{"Discard this card"}
		}
	}
	obj := &game.StackObject{
		ID:                  g.IDGen.Next(),
		Kind:                game.StackActivatedAbility,
		SourceID:            card.ID,
		SourceCardID:        card.ID,
		SourceZone:          zone.Hand,
		SourceZoneVersion:   sourceZoneVersion,
		AbilityIndex:        activate.AbilityIndex,
		Controller:          playerID,
		Targets:             append([]game.Target(nil), completedTargets...),
		TargetCounts:        targetCounts,
		ChosenModes:         append([]int(nil), activate.ChosenModes...),
		XValue:              activate.XValue,
		AdditionalCostsPaid: additionalCostsPaid,
		SacrificedAsCostIDs: sacrificedAsCostIDs,
		ExiledAsCostIDs:     exiledAsCostIDs,
		InlineActivated:     &ability,
	}
	pushAbilityToStack(g, obj)
	emitAbilityActivatedEvent(g, obj, 0, false)
	recordActivatedAbilityUse(g, card.ID, activate.AbilityIndex, ability.Timing)
	return true
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

func (e *Engine) applyGraveyardAbilityWithChoices(g *game.Game, playerID game.PlayerID, activate action.ActivateAbilityAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	card, ability, ok := graveyardAbilitySource(g, playerID, activate.SourceID, activate.AbilityIndex)
	if !ok || !canActivateGraveyardAbilityWithModes(g, playerID, card.ID, &ability, activate.AbilityIndex, activate.Targets, activate.XValue, activate.ChosenModes) {
		return false
	}
	sourceZoneVersion := card.ZoneVersion
	def := cardFaceOrDefault(card, game.FaceFront)
	completedTargets, ok := e.completeAbilityAnnouncementTargetsWithModes(g, playerID, def, 0, &ability, activate.ChosenModes, activate.Targets, agents, log)
	if !ok || !canActivateGraveyardAbilityWithModes(g, playerID, card.ID, &ability, activate.AbilityIndex, completedTargets, activate.XValue, activate.ChosenModes) {
		return false
	}
	targetCounts, ok := bodyTargetCountsWithModesAndRecorded(g, playerID, def, 0, &ability, activate.ChosenModes, activate.TargetCounts, completedTargets)
	if !ok {
		return false
	}
	prefs := e.paymentPreferencesForCostFromSource(g, playerID, manaCostPtr(ability.ManaCost), abilityAdditionalCosts(ability.AdditionalCosts), activate.XValue, card.ID, zone.Graveyard, agents, log)
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
		return false
	}
	obj := &game.StackObject{
		ID:                  g.IDGen.Next(),
		Kind:                game.StackActivatedAbility,
		SourceID:            card.ID,
		SourceCardID:        card.ID,
		SourceZone:          zone.Graveyard,
		SourceZoneVersion:   sourceZoneVersion,
		AbilityIndex:        activate.AbilityIndex,
		Controller:          playerID,
		Targets:             append([]game.Target(nil), completedTargets...),
		TargetCounts:        targetCounts,
		ChosenModes:         append([]int(nil), activate.ChosenModes...),
		XValue:              activate.XValue,
		SacrificedAsCostIDs: costPaid.sacrificedIDs,
		ExiledAsCostIDs:     costPaid.exiledIDs,
	}
	pushAbilityToStack(g, obj)
	emitAbilityActivatedEvent(g, obj, 0, false)
	recordActivatedAbilityUse(g, card.ID, activate.AbilityIndex, ability.Timing)
	return true
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
	return e.applyCyclingAbilityWithChoices(g, playerID, activate, [game.NumPlayers]PlayerAgent{}, nil)
}

func (e *Engine) applyCyclingAbilityWithChoices(g *game.Game, playerID game.PlayerID, activate action.ActivateAbilityAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	if len(activate.TargetCounts) != 0 || len(activate.ChosenModes) != 0 {
		return false
	}
	card, ability, ok := cyclingAbilitySource(g, playerID, activate.SourceID, activate.AbilityIndex)
	if !ok {
		return false
	}
	if !canActivateCyclingAbility(g, playerID, activate.SourceID, &ability, activate.AbilityIndex, activate.Targets, activate.XValue) {
		return false
	}
	effectiveCost := effectiveCyclingCost(g, playerID, card, &ability)
	prefs := e.paymentPreferencesForCost(g, playerID, effectiveCost, nil, activate.XValue, agents, log)
	if !paymentOrch.payGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: effectiveCost, XValue: activate.XValue, Prefs: prefs}) {
		return false
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
		InlineActivated:     &ability,
	}
	pushAbilityToStack(g, obj)
	emitAbilityActivatedEvent(g, obj, 0, false)
	return true
}

func (e *Engine) applyNinjutsuAbilityWithChoices(g *game.Game, playerID game.PlayerID, activate action.ActivateAbilityAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	if len(activate.TargetCounts) != 0 || len(activate.ChosenModes) != 0 {
		return false
	}
	card, ability, ok := handActivatedAbilitySource(g, playerID, activate.SourceID, activate.AbilityIndex)
	if !ok || !canActivateNinjutsuAbility(g, playerID, activate.SourceID, &ability, activate.AbilityIndex, activate.Targets, activate.XValue) {
		return false
	}
	attacker := chooseNinjutsuAttacker(e, g, playerID, unblockedAttackers(g, playerID), agents, log)
	if attacker == nil {
		return false
	}
	attackTarget, ok := attackTargetForAttacker(g, attacker.ObjectID)
	if !ok || attackerWasBlocked(g, attacker.ObjectID) {
		return false
	}
	prefs := e.paymentPreferencesForCost(g, playerID, manaCostPtr(ability.ManaCost), nil, 0, agents, log)
	if !paymentOrch.payGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: manaCostPtr(ability.ManaCost), Prefs: prefs}) {
		return false
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
	return true
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
		options = append(options, game.ChoiceOption{Index: i, Label: permanentEffectiveName(g, attacker)})
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
	if body == nil || !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return false
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
		return false
	}
	_, gotAbility, ok := handActivatedAbilitySource(g, playerID, cardID, abilityIndex)
	if !ok || !game.BodyHasKeyword(&gotAbility, game.Ninjutsu) {
		return false
	}
	return paymentOrch.canPayGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: manaCostPtr(body.ManaCost)})
}
