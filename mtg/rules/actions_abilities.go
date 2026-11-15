package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules/payment"
)

func (e *Engine) applyActivateAbility(g *game.Game, player game.PlayerID, activate action.ActivateAbilityAction) bool {
	return e.applyActivateAbilityWithChoices(g, player, activate, [game.NumPlayers]PlayerAgent{}, nil)
}

func (e *Engine) applyActivateAbilityWithChoices(g *game.Game, player game.PlayerID, activate action.ActivateAbilityAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	return e.attemptActivateAbilityWithChoices(g, player, activate, agents, log) == actionApplied
}

func (e *Engine) attemptActivateAbilityWithChoices(g *game.Game, player game.PlayerID, activate action.ActivateAbilityAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) actionOutcome {
	descriptor, ok := e.describeActivation(g, player, activate)
	if !ok {
		return actionFailed
	}
	return e.costedAction(g, player, descriptor, agents, log)
}

// describeActivation supplies identity and rules-specific data; the procedure
// owns all announcement, payment and commit ordering (CR 602.2a–b).
func (e *Engine) describeActivation(g *game.Game, player game.PlayerID, activate action.ActivateAbilityAction) (costedAction, bool) {
	var c activationCost
	var source *game.Permanent
	var card *game.CardInstance
	var def *game.CardDef
	var body game.Ability
	var activated *game.ActivatedAbility
	var loyalty *game.LoyaltyAbility
	var manaBody *game.ManaAbility
	var sourceZone zone.Type
	var cycling, ninjutsu bool

	if permanent, ability, ok := activatedAbilitySource(g, player, activate.SourceID, activate.AbilityIndex); ok {
		source, body = permanent, ability
		def, _ = permanentCardDef(g, permanent)
		card, _ = g.GetCardInstance(permanent.CardInstanceID)
		sourceZone = zone.Battlefield
		if value, ok := ability.(*game.ManaAbility); ok {
			if !manaAbilityAnnouncementLegal(g, player, permanent, value, activate.AbilityIndex) {
				return costedAction{}, false
			}
			manaBody = value
			c = abilityRequestCost(payment.AbilityRequest{
				PlayerID: player, Source: permanent, ManaCost: value.ManaCost,
				AdditionalCosts: abilityAdditionalCosts(value.AdditionalCosts), ForMana: true,
			}, permanentSource(permanent.ObjectID))
		} else if value, ok := ability.(*game.ActivatedAbility); ok {
			activated = value
			c = equipActivation(g, player, permanent, value, activate.AbilityIndex, activate.Targets, activate.XValue, activate.ChosenModes)
			if c.present == nil {
				c = generalActivation(g, player, permanent, value, activate.AbilityIndex, activate.Targets, activate.XValue, activate.ChosenModes)
			}
		} else if value, ok := ability.(*game.LoyaltyAbility); ok {
			if len(activate.ChosenModes) != 0 || !canActivateLoyaltyAbility(g, player, permanent, value, activate.AbilityIndex, activate.Targets, activate.XValue) {
				return costedAction{}, false
			}
			loyalty = value
			var costs []cost.Additional
			if value.LoyaltyCost < 0 {
				costs = []cost.Additional{{Kind: cost.AdditionalRemoveCounter, CounterKind: counter.Loyalty, Amount: -value.LoyaltyCost}}
			}
			c = abilityRequestCost(payment.AbilityRequest{PlayerID: player, Source: permanent, AdditionalCosts: costs}, permanentSource(permanent.ObjectID))
		}
	} else if value, ability, ok := handManaAbilitySource(g, player, activate.SourceID, activate.AbilityIndex); ok {
		if !canActivateHandManaAbility(g, player, value.ID, &ability, activate.AbilityIndex) {
			return costedAction{}, false
		}
		card, manaBody, body = value, &ability, &ability
		def = cardFaceOrDefault(card, game.FaceFront)
		sourceZone = zone.Hand
		c = abilityRequestCost(payment.AbilityRequest{
			PlayerID: player, SourceCardID: card.ID, SourceZone: zone.Hand, AdditionalCosts: ability.AdditionalCosts, ForMana: true,
		}, cardSource(handCardPresent(player, card.ID)))
	} else if value, ability, ok := handActivatedAbilitySource(g, player, activate.SourceID, activate.AbilityIndex); ok {
		card, activated, body = value, &ability, &ability
		def = cardFaceOrDefault(card, game.FaceFront)
		sourceZone = zone.Hand
		cycling, ninjutsu = game.BodyHasKeyword(body, game.Cycling), game.BodyHasKeyword(body, game.Ninjutsu)
		if cycling {
			c = cyclingActivation(g, player, card.ID, activated, activate.AbilityIndex, activate.Targets, activate.XValue)
		} else if ninjutsu {
			c = ninjutsuActivation(g, player, card.ID, activated, activate.AbilityIndex, activate.Targets, activate.XValue)
		} else {
			c = handActivation(g, player, card.ID, activated, activate.AbilityIndex, activate.Targets, activate.XValue, activate.ChosenModes)
		}
	} else if value, ability, ok := graveyardAbilitySource(g, player, activate.SourceID, activate.AbilityIndex); ok {
		card, activated, body = value, &ability, &ability
		def = cardFaceOrDefault(card, game.FaceFront)
		sourceZone = zone.Graveyard
		c = graveyardActivation(g, player, card.ID, activated, activate.AbilityIndex, activate.Targets, activate.XValue, activate.ChosenModes)
	}
	if body == nil || c.present == nil {
		return costedAction{}, false
	}
	if (manaBody != nil || cycling || ninjutsu) &&
		(len(activate.TargetCounts) != 0 || len(activate.ChosenModes) != 0 || len(activate.Targets) != 0 || activate.XValue != 0) {
		return costedAction{}, false
	}
	obj := &game.StackObject{
		Kind: game.StackActivatedAbility, Controller: player, SourceID: activate.SourceID,
		SourceZone: sourceZone, AbilityIndex: activate.AbilityIndex,
		Targets: append([]game.Target(nil), activate.Targets...), ChosenModes: append([]int(nil), activate.ChosenModes...),
		XValue: activate.XValue, InlineLoyalty: loyalty,
	}
	sourceObject := id.ID(0)
	if source != nil {
		obj.SourceZone = zone.None
		sourceObject = source.ObjectID
		obj.Face, obj.SourceCardID, obj.SourceTokenDef = source.Face, source.CardInstanceID, source.TokenDef
		obj.InlineActivated = activated
		obj.ActivatedResolutionUse = permanentActivatedResolutionUse(g, source, body, activate.AbilityIndex)
		seedEntryChoices(obj, source)
	} else {
		obj.SourceCardID, obj.SourceZoneVersion = card.ID, card.ZoneVersion
		if activated != nil {
			obj.InlineActivated = activationStackBody(card, activate.AbilityIndex, activated)
			obj.ActivatedResolutionUse = cardActivatedResolutionUse(card, def, activate.AbilityIndex)
		}
	}
	var manaSource manaProducedSource
	if source != nil {
		manaSource = captureManaProducedSource(g, source)
	}
	var ninjutsuDefender game.AttackTarget
	descriptor := costedAction{
		cost: c, object: obj, usesStack: manaBody == nil,
	}
	if sourceZone == zone.Hand {
		descriptor.disclose = func(g *game.Game) {
			emitEvent(g, game.Event{Kind: game.EventCardRevealed, Announcement: true, Player: player, Controller: player, CardID: card.ID, FromZone: zone.Hand})
		}
	}
	descriptor.targets = func(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
		targets, ok := e.completeAbilityAnnouncementTargetsWithModes(g, player, def, sourceObject, body, activate.ChosenModes, activate.Targets, agents, log)
		if !ok {
			return false
		}
		counts, ok := bodyTargetCountsWithModesAndRecorded(g, player, def, sourceObject, body, activate.ChosenModes, activate.TargetCounts, targets, activate.XValue)
		if !ok {
			return false
		}
		obj.Targets, obj.TargetCounts = targets, counts
		emitPreparedTargetEvents(g, obj, prepareTargetEvents(g, obj))
		return true
	}
	descriptor.commit = func(g *game.Game, paid payment.ApplyResult, agents [game.NumPlayers]PlayerAgent, log *TurnLog, paymentStart int) {
		obj.SacrificedAsCostIDs = paid.Ability.SacrificedIDs
		obj.PaidCostSubjects = paid.Ability.PaidCostSubjects
		obj.TappedAsCostIDs = paid.Ability.TappedIDs
		obj.ExiledAsCostIDs = paid.Ability.ExiledIDs
		obj.AdditionalCostsPaid = paid.Spell.AdditionalCostsPaid
		if manaBody != nil {
			e.resolveManaAbility(g, obj, source, manaBody, manaSource, paymentStart, nil, agents, log)
			return
		}
		if loyalty != nil {
			if loyalty.LoyaltyCost > 0 {
				current, _ := permanentByObjectID(g, sourceObject)
				applyLoyaltyCost(current, loyalty.LoyaltyCost)
			}
			recordActivatedAbilityUse(g, sourceObject, -1, game.OncePerTurn)
		}
		if ninjutsu {
			obj.Ninjutsu = true
			if len(paid.Spell.ReturnedAttackerIDs) != 1 {
				panic("completed ninjutsu payment has no returned attacker")
			}
			obj.NinjutsuAttackTarget = ninjutsuDefender
		}
		if cycling {
			emitEvent(g, game.Event{Kind: game.EventCycled, SourceID: card.ID, Controller: player, Player: player, CardID: card.ID})
		}
		emitAbilityActivatedEvent(g, obj, sourceObject, false)
		if activated != nil {
			recordActivatedAbilityUse(g, activate.SourceID, activate.AbilityIndex, activated.Timing)
		}
	}
	// The attack target is locked while the returning creature is still in
	// combat; the payment operation removes it before the commit hook runs.
	if ninjutsu {
		descriptor.beforePayment = func(g *game.Game, plan payment.Plan) {
			for _, component := range plan.Components() {
				for _, object := range component.Objects {
					if target, ok := attackTargetForAttacker(g, object); ok {
						ninjutsuDefender = target
					}
				}
			}
		}
	}
	return descriptor, true
}

func canActivateLoyaltyAbility(g *game.Game, playerID game.PlayerID, permanent *game.Permanent, body *game.LoyaltyAbility, abilityIndex int, targets []game.Target, xValue int) bool {
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
	return ok && targetsValidForBodyFromSourceObject(g, playerID, card, permanent.ObjectID, body, targets)
}

func applyLoyaltyCost(permanent *game.Permanent, loyalty int) {
	if loyalty >= 0 {
		permanent.Counters.Add(counter.Loyalty, loyalty)
	} else {
		permanent.Counters.Remove(counter.Loyalty, -loyalty)
	}
}

func (e *Engine) applyCyclingAbility(g *game.Game, player game.PlayerID, activate action.ActivateAbilityAction) bool {
	card, body, ok := cyclingAbilitySource(g, player, activate.SourceID, activate.AbilityIndex)
	return ok && game.BodyHasKeyword(&body, game.Cycling) && card != nil && e.applyActivateAbility(g, player, activate)
}

func chooseNinjutsuAttacker(e *Engine, g *game.Game, playerID game.PlayerID, attackers []*game.Permanent, agents [game.NumPlayers]PlayerAgent, log *TurnLog) *game.Permanent {
	if len(attackers) == 0 {
		return nil
	}
	if len(attackers) == 1 {
		return attackers[0]
	}
	options := make([]game.ChoiceOption, len(attackers))
	for i, attacker := range attackers {
		options[i] = game.ChoiceOption{Index: i, Label: permanentChoiceLabel(g, attacker), Card: permanentChoiceInfo(g, attacker), Targets: []game.Target{game.PermanentTarget(attacker.ObjectID)}}
	}
	selected := e.chooseChoice(g, agents, game.ChoiceRequest{
		Kind: game.ChoicePayment, Player: playerID, Prompt: "Choose an unblocked attacker to return",
		Options: options, MinChoices: 1, MaxChoices: 1, DefaultSelection: []int{0},
	}, log)
	return attackers[selected[0]]
}

func unblockedAttackers(g *game.Game, playerID game.PlayerID) []*game.Permanent {
	if g.Combat == nil {
		return nil
	}
	var attackers []*game.Permanent
	for _, attack := range g.Combat.Attackers {
		if !attackIsUnblocked(g, attack) {
			continue
		}
		permanent, ok := permanentByObjectID(g, attack.Attacker)
		if ok && effectiveController(g, permanent) == playerID {
			attackers = append(attackers, permanent)
		}
	}
	return attackers
}

func attackIsUnblocked(g *game.Game, attack game.AttackDeclaration) bool {
	if g.Combat == nil || g.Turn.Phase != game.PhaseCombat || g.Turn.Step < game.StepDeclareBlockers || g.Turn.Step > game.StepEndOfCombat {
		return false
	}
	if g.Turn.Step == game.StepDeclareBlockers && !g.Combat.BlockDeclarationsComplete[attack.Target.Player] {
		return false
	}
	return !attackerWasBlocked(g, attack.Attacker)
}

func attackTargetForAttacker(g *game.Game, attackerID id.ID) (game.AttackTarget, bool) {
	if g.Combat != nil {
		for _, attack := range g.Combat.Attackers {
			if attack.Attacker == attackerID {
				return attack.Target, true
			}
		}
	}
	return game.AttackTarget{}, false
}

func ninjutsuActivation(g *game.Game, playerID game.PlayerID, cardID id.ID, body *game.ActivatedAbility, abilityIndex int, targets []game.Target, xValue int) activationCost {
	if body == nil || !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer || xValue != 0 || abilityIndex < 0 ||
		!game.BodyHasKeyword(body, game.Ninjutsu) || game.BodyFunctionZone(body) != zone.Hand || body.Timing != game.DuringCombat ||
		len(targets) != 0 || len(game.BodyTargets(body)) != 0 || !abilityHasReturnUnblockedAttackerCost(body.AdditionalCosts) ||
		len(unblockedAttackers(g, playerID)) == 0 {
		return activationCost{}
	}
	card, gotAbility, ok := handActivatedAbilitySource(g, playerID, cardID, abilityIndex)
	if !ok || !game.BodyHasKeyword(&gotAbility, game.Ninjutsu) {
		return activationCost{}
	}
	return abilityRequestCost(payment.AbilityRequest{
		PlayerID: playerID, SourceCardID: cardID, SourceZone: zone.Hand, ManaCost: body.ManaCost,
		CostModifiers:   activatedAbilityCostModifiers(g, playerID, card, body),
		AdditionalCosts: abilityAdditionalCosts(body.AdditionalCosts),
	}, cardSource(cardSourcePresent(handActivatedAbilitySource, playerID, cardID, abilityIndex)))
}
