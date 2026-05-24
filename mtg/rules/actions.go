package rules

import (
	"strings"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
)

const maxLegalXValue = 20

func (e *Engine) legalActions(g *game.Game, playerID game.PlayerID) []action.Action {
	if !canAct(g, playerID) {
		return []action.Action{action.Pass()}
	}
	if splitSecondOnStack(g) {
		actions := e.legalManaAbilityActions(g, playerID)
		actions = append(actions, action.Pass())
		return actions
	}

	actions := e.legalLandActions(g, playerID)
	actions = append(actions, e.legalCastActions(g, playerID)...)
	actions = append(actions, e.legalCommanderCastActions(g, playerID)...)
	actions = append(actions, e.legalActivateAbilityActions(g, playerID)...)
	actions = append(actions, e.legalCyclingActions(g, playerID)...)
	actions = append(actions, action.Pass())
	return actions
}

func normalizedCastSourceZone(cast action.CastSpellAction) game.ZoneType {
	if cast.SourceZone == game.ZoneNone {
		return game.ZoneHand
	}
	return cast.SourceZone
}

func splitSecondOnStack(g *game.Game) bool {
	if g == nil {
		return false
	}
	for _, obj := range g.Stack.Objects() {
		if obj == nil || obj.Kind != game.StackSpell {
			continue
		}
		card := g.GetCardInstance(obj.SourceID)
		if card != nil && card.Def != nil && card.Def.HasKeyword(game.SplitSecond) {
			return true
		}
	}
	return false
}

func (e *Engine) legalLandActions(g *game.Game, playerID game.PlayerID) []action.Action {
	if !canPlayAnyLand(g, playerID) {
		return nil
	}

	player := g.Players[playerID]
	var actions []action.Action
	for _, cardID := range player.Hand.All() {
		if _, ok := landCardInstance(g, player, cardID); ok {
			actions = append(actions, action.PlayLand(cardID))
		}
	}
	return actions
}

func (e *Engine) legalCastActions(g *game.Game, playerID game.PlayerID) []action.Action {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return nil
	}

	player := g.Players[playerID]
	var actions []action.Action
	for _, cardID := range player.Hand.All() {
		card := g.GetCardInstance(cardID)
		if card == nil || card.Def == nil {
			continue
		}
		for _, xValue := range legalXValuesForCost(g, playerID, card.Def.ManaCost) {
			for _, modes := range modeChoicesForSpell(card.Def) {
				for _, targets := range targetChoicesForSpell(g, playerID, card.Def, modes) {
					if e.canCastSpellFromZoneWithKicker(g, playerID, cardID, game.ZoneHand, targets, xValue, modes, false) {
						actions = append(actions, action.CastSpell(cardID, append([]game.Target(nil), targets...), xValue, append([]int(nil), modes...)))
					}
					if spellHasKicker(card.Def) && e.canCastSpellFromZoneWithKicker(g, playerID, cardID, game.ZoneHand, targets, xValue, modes, true) {
						actions = append(actions, action.CastKickedSpell(cardID, append([]game.Target(nil), targets...), xValue, append([]int(nil), modes...)))
					}
				}
			}
		}
	}
	return actions
}

func (e *Engine) legalCommanderCastActions(g *game.Game, playerID game.PlayerID) []action.Action {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return nil
	}
	player := playerByID(g, playerID)
	if player == nil || player.CommanderInstanceID == 0 || !player.CommandZone.Contains(player.CommanderInstanceID) {
		return nil
	}
	card := g.GetCardInstance(player.CommanderInstanceID)
	if card == nil || card.Def == nil {
		return nil
	}
	var actions []action.Action
	for _, xValue := range legalXValuesForCost(g, playerID, card.Def.ManaCost) {
		for _, modes := range modeChoicesForSpell(card.Def) {
			for _, targets := range targetChoicesForSpell(g, playerID, card.Def, modes) {
				if e.canCastSpellFromZoneWithKicker(g, playerID, card.ID, game.ZoneCommand, targets, xValue, modes, false) {
					actions = append(actions, action.CastCommanderSpell(card.ID, append([]game.Target(nil), targets...), xValue, append([]int(nil), modes...)))
				}
				if spellHasKicker(card.Def) && e.canCastSpellFromZoneWithKicker(g, playerID, card.ID, game.ZoneCommand, targets, xValue, modes, true) {
					act := action.CastCommanderSpell(card.ID, append([]game.Target(nil), targets...), xValue, append([]int(nil), modes...))
					act.CastSpell.KickerPaid = true
					actions = append(actions, act)
				}
			}
		}
	}
	return actions
}

func (e *Engine) applyAction(g *game.Game, playerID game.PlayerID, act action.Action) bool {
	return e.applyActionWithChoices(g, playerID, act, [game.NumPlayers]PlayerAgent{}, nil)
}

func (e *Engine) applyActionWithChoices(g *game.Game, playerID game.PlayerID, act action.Action, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	switch act.Kind {
	case action.ActionPass:
		return true
	case action.ActionPlayLand:
		return e.applyPlayLand(g, playerID, act.PlayLand.CardID)
	case action.ActionCastSpell:
		return e.applyCastSpellWithChoices(g, playerID, act.CastSpell, agents, log)
	case action.ActionActivateAbility:
		return e.applyActivateAbilityWithChoices(g, playerID, act.ActivateAbility, agents, log)
	case action.ActionDeclareAttackers:
		return e.applyDeclareAttackers(g, playerID, act.DeclareAttackers)
	case action.ActionDeclareBlockers:
		return e.applyDeclareBlockers(g, playerID, act.DeclareBlockers)
	default:
		return false
	}
}

func (e *Engine) legalActivateAbilityActions(g *game.Game, playerID game.PlayerID) []action.Action {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return nil
	}

	var actions []action.Action
	for _, permanent := range g.Battlefield {
		if permanent == nil || permanent.PhasedOut || effectiveController(g, permanent) != playerID {
			continue
		}
		card := permanentCardDef(g, permanent)
		if card == nil {
			continue
		}
		for i := range card.Abilities {
			ability := &card.Abilities[i]
			if canActivateManaAbility(g, playerID, permanent, ability, i) {
				actions = append(actions, action.ActivateAbility(permanent.ObjectID, i, nil, 0))
				continue
			}
			for _, xValue := range legalXValuesForCost(g, playerID, ability.ManaCost) {
				for _, targets := range targetChoicesForAbilityFromSourceObject(g, playerID, card, permanent.ObjectID, ability) {
					if canActivateEquipAbility(g, playerID, permanent, ability, i, targets, xValue) ||
						canActivateLoyaltyAbility(g, playerID, permanent, ability, i, targets, xValue) ||
						canActivateGeneralAbility(g, playerID, permanent, ability, i, targets, xValue) {
						actions = append(actions, action.ActivateAbility(permanent.ObjectID, i, append([]game.Target(nil), targets...), xValue))
					}
				}
			}
		}
	}
	return actions
}

func (e *Engine) legalManaAbilityActions(g *game.Game, playerID game.PlayerID) []action.Action {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return nil
	}
	var actions []action.Action
	for _, permanent := range g.Battlefield {
		if permanent == nil || permanent.PhasedOut || effectiveController(g, permanent) != playerID {
			continue
		}
		card := permanentCardDef(g, permanent)
		if card == nil {
			continue
		}
		for i := range card.Abilities {
			if canActivateManaAbility(g, playerID, permanent, &card.Abilities[i], i) {
				actions = append(actions, action.ActivateAbility(permanent.ObjectID, i, nil, 0))
			}
		}
	}
	return actions
}

func (e *Engine) legalCyclingActions(g *game.Game, playerID game.PlayerID) []action.Action {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return nil
	}
	player := playerByID(g, playerID)
	if player == nil {
		return nil
	}
	var actions []action.Action
	for _, cardID := range player.Hand.All() {
		card := g.GetCardInstance(cardID)
		if card == nil || card.Def == nil {
			continue
		}
		for i := range card.Def.Abilities {
			ability := &card.Def.Abilities[i]
			if canActivateCyclingAbility(g, playerID, cardID, ability, i, nil, 0) {
				actions = append(actions, action.ActivateAbility(cardID, i, nil, 0))
			}
		}
	}
	return actions
}

func (e *Engine) applyPlayLand(g *game.Game, playerID game.PlayerID, cardID id.ID) bool {
	if !canPlayAnyLand(g, playerID) {
		return false
	}

	player := g.Players[playerID]
	card, ok := landCardInstance(g, player, cardID)
	if !ok {
		return false
	}
	if !player.Hand.Remove(cardID) {
		return false
	}

	createCardPermanent(g, card, playerID, game.ZoneHand)
	g.Turn.LandsPlayedThisTurn++
	return true
}

func (e *Engine) applyCastSpell(g *game.Game, playerID game.PlayerID, cast action.CastSpellAction) bool {
	return e.applyCastSpellWithChoices(g, playerID, cast, [game.NumPlayers]PlayerAgent{}, nil)
}

func (e *Engine) applyCastSpellWithChoices(g *game.Game, playerID game.PlayerID, cast action.CastSpellAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	sourceZone := normalizedCastSourceZone(cast)
	if !e.canCastSpellFromZoneWithKicker(g, playerID, cast.CardID, sourceZone, cast.Targets, cast.XValue, cast.ChosenModes, cast.KickerPaid) {
		return false
	}

	player := g.Players[playerID]
	card := g.GetCardInstance(cast.CardID)
	prefs := e.paymentPreferencesForSpellFromZone(g, playerID, card.ID, sourceZone, card.Def, cast.XValue, agents, log)
	additionalCostsPaid, ok := paySpellCostsWithKickerFromZoneAndPreferences(g, playerID, card.ID, sourceZone, card.Def, cast.XValue, cast.KickerPaid, prefs)
	if !ok {
		return false
	}
	if !removeCastSourceCard(player, cast.CardID, sourceZone) {
		panic("cast spell disappeared from source zone after validation")
	}
	if sourceZone == game.ZoneCommand && player.CommanderInstanceID == cast.CardID {
		player.CommanderCastCount++
	}
	obj := &game.StackObject{
		ID:                  g.IDGen.Next(),
		Kind:                game.StackSpell,
		SourceID:            cast.CardID,
		Controller:          playerID,
		Targets:             append([]game.Target(nil), cast.Targets...),
		ChosenModes:         append([]int(nil), cast.ChosenModes...),
		XValue:              cast.XValue,
		KickerPaid:          cast.KickerPaid,
		AdditionalCostsPaid: additionalCostsPaid,
	}
	g.Stack.Push(obj)
	event := game.GameEvent{
		SourceID:      cast.CardID,
		StackObjectID: obj.ID,
		Controller:    playerID,
		CardID:        cast.CardID,
		FromZone:      sourceZone,
		ToZone:        game.ZoneStack,
	}
	emitZoneChangeEvent(g, event)
	event.Kind = game.EventSpellCast
	emitEvent(g, event)
	return true
}

func (e *Engine) applyActivateAbility(g *game.Game, playerID game.PlayerID, activate action.ActivateAbilityAction) bool {
	return e.applyActivateAbilityWithChoices(g, playerID, activate, [game.NumPlayers]PlayerAgent{}, nil)
}

func (e *Engine) applyActivateAbilityWithChoices(g *game.Game, playerID game.PlayerID, activate action.ActivateAbilityAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	if e.applyCyclingAbilityWithChoices(g, playerID, activate, agents, log) {
		return true
	}
	permanent, ability, ok := activatedAbilitySource(g, playerID, activate.SourceID, activate.AbilityIndex)
	if !ok {
		return false
	}
	if canActivateManaAbility(g, playerID, permanent, ability, activate.AbilityIndex) {
		if len(activate.Targets) != 0 || activate.XValue != 0 {
			return false
		}
		prefs := e.paymentPreferencesForCost(g, playerID, ability.ManaCost, abilityAdditionalCosts(ability), agents, log)
		if _, ok := payAbilityCostsWithPreferences(g, playerID, permanent, ability, 0, prefs); !ok {
			return false
		}
		obj := &game.StackObject{
			ID:             g.IDGen.Next(),
			Kind:           game.StackActivatedAbility,
			SourceID:       permanent.ObjectID,
			SourceCardID:   permanent.CardInstanceID,
			SourceTokenDef: permanent.TokenDef,
			AbilityIndex:   activate.AbilityIndex,
			Controller:     playerID,
		}
		for _, effect := range ability.Effects {
			e.resolveEffect(g, obj, effect, nil)
		}
		recordActivatedAbilityUse(g, permanent.ObjectID, activate.AbilityIndex, ability)
		return true
	}

	if !canActivateEquipAbility(g, playerID, permanent, ability, activate.AbilityIndex, activate.Targets, activate.XValue) &&
		!canActivateLoyaltyAbility(g, playerID, permanent, ability, activate.AbilityIndex, activate.Targets, activate.XValue) &&
		!canActivateGeneralAbility(g, playerID, permanent, ability, activate.AbilityIndex, activate.Targets, activate.XValue) {
		return false
	}
	sourceCardID := permanent.CardInstanceID
	sourceTokenDef := permanent.TokenDef
	prefs := e.paymentPreferencesForCost(g, playerID, ability.ManaCost, abilityAdditionalCosts(ability), agents, log)
	if _, ok := payAbilityCostsWithPreferences(g, playerID, permanent, ability, activate.XValue, prefs); !ok {
		return false
	}
	if ability.IsLoyaltyAbility {
		applyLoyaltyCost(permanent, ability.LoyaltyCost)
	}
	g.Stack.Push(&game.StackObject{
		ID:             g.IDGen.Next(),
		Kind:           game.StackActivatedAbility,
		SourceID:       permanent.ObjectID,
		SourceCardID:   sourceCardID,
		SourceTokenDef: sourceTokenDef,
		AbilityIndex:   activate.AbilityIndex,
		Controller:     playerID,
		Targets:        append([]game.Target(nil), activate.Targets...),
		XValue:         activate.XValue,
	})
	recordActivatedAbilityUse(g, permanent.ObjectID, activate.AbilityIndex, ability)
	if ability.IsLoyaltyAbility {
		recordActivatedAbilityUse(g, permanent.ObjectID, -1, &game.AbilityDef{Timing: game.OncePerTurn})
	}
	return true
}

func canActivateLoyaltyAbility(g *game.Game, playerID game.PlayerID, permanent *game.Permanent, ability *game.AbilityDef, abilityIndex int, targets []game.Target, xValue int) bool {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer || permanent == nil || permanent.PhasedOut || effectiveController(g, permanent) != playerID || ability == nil {
		return false
	}
	if xValue != 0 || ability.Kind != game.ActivatedAbility || !ability.IsLoyaltyAbility || !permanentHasType(g, permanent, game.TypePlaneswalker) {
		return false
	}
	if !isSorcerySpeed(g, playerID) || activatedAbilityUsedThisTurn(g, permanent.ObjectID, abilityIndex, ability) || g.ActivatedAbilitiesThisTurn[game.ActivatedAbilityUse{SourceID: permanent.ObjectID, AbilityIndex: -1}] {
		return false
	}
	if ability.LoyaltyCost < 0 && permanent.Counters.Get(counter.Loyalty) < -ability.LoyaltyCost {
		return false
	}
	if !targetsValidForAbilityFromSourceObject(g, playerID, permanentCardDef(g, permanent), permanent.ObjectID, ability, targets) {
		return false
	}
	_, ok := buildAbilityCostPlan(g, playerID, permanent, ability, xValue)
	return ok
}

func applyLoyaltyCost(permanent *game.Permanent, cost int) {
	if cost >= 0 {
		permanent.Counters.Add(counter.Loyalty, cost)
		return
	}
	permanent.Counters.Remove(counter.Loyalty, -cost)
}

func (e *Engine) applyCyclingAbility(g *game.Game, playerID game.PlayerID, activate action.ActivateAbilityAction) bool {
	return e.applyCyclingAbilityWithChoices(g, playerID, activate, [game.NumPlayers]PlayerAgent{}, nil)
}

func (e *Engine) applyCyclingAbilityWithChoices(g *game.Game, playerID game.PlayerID, activate action.ActivateAbilityAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	card, ability, ok := cyclingAbilitySource(g, playerID, activate.SourceID, activate.AbilityIndex)
	if !ok {
		return false
	}
	if !canActivateCyclingAbility(g, playerID, activate.SourceID, ability, activate.AbilityIndex, activate.Targets, activate.XValue) {
		return false
	}
	prefs := e.paymentPreferencesForCost(g, playerID, ability.ManaCost, nil, agents, log)
	plan, ok := buildPaymentPlanWithPreferences(g, playerID, ability.ManaCost, activate.XValue, nil, prefs)
	if !ok || !applyPaymentPlan(g, playerID, plan) {
		return false
	}
	if !discardCardFromHand(g, playerID, card.ID) {
		panic("cycling card disappeared from hand after validation")
	}
	g.Stack.Push(&game.StackObject{
		ID:                  g.IDGen.Next(),
		Kind:                game.StackActivatedAbility,
		SourceID:            card.ID,
		SourceCardID:        card.ID,
		AbilityIndex:        activate.AbilityIndex,
		Controller:          playerID,
		Targets:             append([]game.Target(nil), activate.Targets...),
		XValue:              activate.XValue,
		AdditionalCostsPaid: []string{"Discard this card"},
	})
	return true
}

func (e *Engine) canCastSpell(g *game.Game, playerID game.PlayerID, cardID id.ID, targets []game.Target, xValue int, chosenModes []int) bool {
	return e.canCastSpellWithKicker(g, playerID, cardID, targets, xValue, chosenModes, false)
}

func (e *Engine) canCastSpellWithKicker(g *game.Game, playerID game.PlayerID, cardID id.ID, targets []game.Target, xValue int, chosenModes []int, kickerPaid bool) bool {
	return e.canCastSpellFromZoneWithKicker(g, playerID, cardID, game.ZoneHand, targets, xValue, chosenModes, kickerPaid)
}

func (e *Engine) canCastSpellFromZoneWithKicker(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone game.ZoneType, targets []game.Target, xValue int, chosenModes []int, kickerPaid bool) bool {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return false
	}
	if xValue < 0 {
		return false
	}
	player := g.Players[playerID]
	card := g.GetCardInstance(cardID)
	if card == nil || card.Def == nil || !castSourceContains(player, cardID, sourceZone) {
		return false
	}
	if sourceZone == game.ZoneCommand && player.CommanderInstanceID != cardID {
		return false
	}
	if xValue != 0 && !costHasVariableMana(card.Def.ManaCost) {
		return false
	}
	if !modesValidForSpell(card.Def, chosenModes) || !isSupportedSpell(card.Def) || !targetsValidForSpell(g, playerID, card.Def, chosenModes, targets) {
		return false
	}
	if !canCastAtCurrentTiming(g, playerID, card.Def) {
		return false
	}
	if kickerPaid && !spellHasKicker(card.Def) {
		return false
	}
	if !canPaySpellCostsWithKickerFromZone(g, playerID, card.ID, sourceZone, card.Def, xValue, kickerPaid) {
		return false
	}
	return true
}

func castSourceContains(player *game.Player, cardID id.ID, sourceZone game.ZoneType) bool {
	if player == nil {
		return false
	}
	switch sourceZone {
	case game.ZoneHand:
		return player.Hand.Contains(cardID)
	case game.ZoneCommand:
		return player.CommandZone.Contains(cardID)
	default:
		return false
	}
}

func removeCastSourceCard(player *game.Player, cardID id.ID, sourceZone game.ZoneType) bool {
	switch sourceZone {
	case game.ZoneHand:
		return player.Hand.Remove(cardID)
	case game.ZoneCommand:
		return player.CommandZone.Remove(cardID)
	default:
		return false
	}
}

func canAct(g *game.Game, playerID game.PlayerID) bool {
	return isPlayerAlive(g, playerID)
}

func canPlayAnyLand(g *game.Game, playerID game.PlayerID) bool {
	return canAct(g, playerID) &&
		playerID == g.Turn.ActivePlayer &&
		playerID == g.Turn.PriorityPlayer &&
		isSorcerySpeed(g, playerID) &&
		g.Turn.CanPlayLand()
}

func canCastAtCurrentTiming(g *game.Game, playerID game.PlayerID, card *game.CardDef) bool {
	if card.HasType(game.TypeInstant) || card.HasKeyword(game.Flash) {
		return true
	}
	return isSorcerySpeed(g, playerID)
}

func legalXValuesForCost(g *game.Game, playerID game.PlayerID, cost *mana.Cost) []int {
	if !costHasVariableMana(cost) {
		return []int{0}
	}
	var values []int
	for x := 0; x <= maxLegalXValue; x++ {
		if !canPayCostWithX(g, playerID, cost, x) {
			break
		}
		values = append(values, x)
	}
	return values
}

func costHasVariableMana(cost *mana.Cost) bool {
	if cost == nil {
		return false
	}
	for _, symbol := range *cost {
		if symbol.Kind == mana.VariableSymbol {
			return true
		}
	}
	return false
}

func activatedAbilitySource(g *game.Game, playerID game.PlayerID, sourceID id.ID, abilityIndex int) (*game.Permanent, *game.AbilityDef, bool) {
	if abilityIndex < 0 {
		return nil, nil, false
	}
	permanent := permanentByObjectID(g, sourceID)
	if permanent == nil || permanent.PhasedOut || effectiveController(g, permanent) != playerID {
		return nil, nil, false
	}
	card := permanentCardDef(g, permanent)
	if card == nil || abilityIndex >= len(card.Abilities) {
		return nil, nil, false
	}
	return permanent, &card.Abilities[abilityIndex], true
}

func cyclingAbilitySource(g *game.Game, playerID game.PlayerID, sourceID id.ID, abilityIndex int) (*game.CardInstance, *game.AbilityDef, bool) {
	if abilityIndex < 0 {
		return nil, nil, false
	}
	player := playerByID(g, playerID)
	if player == nil || !player.Hand.Contains(sourceID) {
		return nil, nil, false
	}
	card := g.GetCardInstance(sourceID)
	if card == nil || card.Def == nil || abilityIndex >= len(card.Def.Abilities) {
		return nil, nil, false
	}
	return card, &card.Def.Abilities[abilityIndex], true
}

func canActivateEquipAbility(g *game.Game, playerID game.PlayerID, permanent *game.Permanent, ability *game.AbilityDef, abilityIndex int, targets []game.Target, xValue int) bool {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer || permanent == nil || permanent.PhasedOut || effectiveController(g, permanent) != playerID || ability == nil {
		return false
	}
	if xValue != 0 || ability.Kind != game.ActivatedAbility || ability.IsManaAbility || !isEquipmentPermanent(g, permanent) {
		return false
	}
	if !abilityHasKeyword(ability, game.Equip) && ability.Timing != game.SorceryOnly {
		return false
	}
	if !isSorcerySpeed(g, playerID) || abilityHasNonTapAdditionalCosts(ability) || activatedAbilityUsedThisTurn(g, permanent.ObjectID, abilityIndex, ability) {
		return false
	}
	if !targetsValidForAbilityFromSourceObject(g, playerID, permanentCardDef(g, permanent), permanent.ObjectID, ability, targets) {
		return false
	}
	if len(targets) != 1 || targets[0].Kind != game.TargetPermanent {
		return false
	}
	target := permanentByObjectID(g, targets[0].PermanentID)
	if target == nil || effectiveController(g, target) != playerID || !canAttachPermanent(g, permanent, target) {
		return false
	}
	return canPayCost(g, playerID, ability.ManaCost)
}

func canActivateGeneralAbility(g *game.Game, playerID game.PlayerID, permanent *game.Permanent, ability *game.AbilityDef, abilityIndex int, targets []game.Target, xValue int) bool {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer || permanent == nil || permanent.PhasedOut || effectiveController(g, permanent) != playerID || ability == nil {
		return false
	}
	if ability.Kind != game.ActivatedAbility || ability.IsManaAbility || ability.IsLoyaltyAbility || abilityHasKeyword(ability, game.Equip) {
		return false
	}
	if !activatedAbilityTimingAllows(g, playerID, ability) || activatedAbilityUsedThisTurn(g, permanent.ObjectID, abilityIndex, ability) {
		return false
	}
	if !targetsValidForAbilityFromSourceObject(g, playerID, permanentCardDef(g, permanent), permanent.ObjectID, ability, targets) {
		return false
	}
	_, ok := buildAbilityCostPlan(g, playerID, permanent, ability, xValue)
	return ok
}

func canActivateCyclingAbility(g *game.Game, playerID game.PlayerID, cardID id.ID, ability *game.AbilityDef, abilityIndex int, targets []game.Target, xValue int) bool {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer || ability == nil {
		return false
	}
	if xValue != 0 || abilityIndex < 0 || ability.Kind != game.ActivatedAbility || ability.IsManaAbility || !abilityHasKeyword(ability, game.Cycling) {
		return false
	}
	if ability.Timing != game.NoTimingRestriction || !abilityHasDiscardThisCardCost(ability) {
		return false
	}
	if len(targets) != 0 || len(ability.Targets) != 0 {
		return false
	}
	card, gotAbility, ok := cyclingAbilitySource(g, playerID, cardID, abilityIndex)
	if !ok || gotAbility != ability || card == nil {
		return false
	}
	return canPayCost(g, playerID, ability.ManaCost)
}

func abilityHasKeyword(ability *game.AbilityDef, keyword game.Keyword) bool {
	if ability == nil {
		return false
	}
	for _, got := range ability.Keywords {
		if got == keyword {
			return true
		}
	}
	return false
}

func canActivateManaAbility(g *game.Game, playerID game.PlayerID, permanent *game.Permanent, ability *game.AbilityDef, abilityIndex int) bool {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer || permanent == nil || permanent.PhasedOut || effectiveController(g, permanent) != playerID || ability == nil {
		return false
	}
	if ability.Kind != game.ActivatedAbility || !ability.IsManaAbility || ability.IsLoyaltyAbility {
		return false
	}
	if len(ability.Targets) != 0 || !manaAbilityHasAddManaEffect(ability) {
		return false
	}
	if ability.Timing != game.NoTimingRestriction || activatedAbilityUsedThisTurn(g, permanent.ObjectID, abilityIndex, ability) {
		return false
	}
	if hasTapCost(ability) {
		if !canTapPermanentForAbility(g, permanent) {
			return false
		}
	} else if abilityHasNonTapAdditionalCosts(ability) {
		return false
	}
	return canPayCost(g, playerID, ability.ManaCost)
}

func manaAbilityHasAddManaEffect(ability *game.AbilityDef) bool {
	if ability == nil || len(ability.Effects) == 0 {
		return false
	}
	for _, effect := range ability.Effects {
		if effect.Type != game.EffectAddMana {
			return false
		}
	}
	return true
}

func hasTapCost(ability *game.AbilityDef) bool {
	if ability == nil {
		return false
	}
	for _, cost := range ability.AdditionalCosts {
		if cost.Kind == game.AdditionalCostTap {
			return true
		}
	}
	return false
}

func abilityHasNonTapAdditionalCosts(ability *game.AbilityDef) bool {
	for _, cost := range abilityAdditionalCosts(ability) {
		if cost.Kind != game.AdditionalCostTap {
			return true
		}
	}
	return false
}

func abilityHasDiscardThisCardCost(ability *game.AbilityDef) bool {
	costs := abilityAdditionalCosts(ability)
	if len(costs) != 1 {
		return false
	}
	cost := costs[0]
	if cost.Kind != game.AdditionalCostDiscard || additionalCostAmount(cost) != 1 {
		return false
	}
	if cost.Text != "" {
		return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(cost.Text)), ".") == "discard this card"
	}
	return cost.Zone == game.ZoneHand
}

func canTapPermanentForAbility(g *game.Game, permanent *game.Permanent) bool {
	if permanent == nil || permanent.Tapped {
		return false
	}
	return !permanentHasType(g, permanent, game.TypeCreature) || !permanent.SummoningSick
}

func tapPermanentForAbility(g *game.Game, permanent *game.Permanent) bool {
	if !canTapPermanentForAbility(g, permanent) {
		return false
	}
	permanent.Tapped = true
	return true
}

func activatedAbilityTimingAllows(g *game.Game, playerID game.PlayerID, ability *game.AbilityDef) bool {
	if ability == nil {
		return false
	}
	switch ability.Timing {
	case game.NoTimingRestriction, game.OncePerTurn:
		return true
	case game.SorceryOnly, game.SorceryOncePerTurn:
		return isSorcerySpeed(g, playerID)
	case game.DuringCombat:
		return g != nil && g.Turn.Phase == game.PhaseCombat
	case game.DuringUpkeep:
		return g != nil && g.Turn.Phase == game.PhaseBeginning && g.Turn.Step == game.StepUpkeep
	default:
		return false
	}
}

func activatedAbilityUsedThisTurn(g *game.Game, sourceID id.ID, abilityIndex int, ability *game.AbilityDef) bool {
	if g == nil || ability == nil || !abilityHasOncePerTurnRestriction(ability) {
		return false
	}
	return g.ActivatedAbilitiesThisTurn[game.ActivatedAbilityUse{
		SourceID:     sourceID,
		AbilityIndex: abilityIndex,
	}]
}

func recordActivatedAbilityUse(g *game.Game, sourceID id.ID, abilityIndex int, ability *game.AbilityDef) {
	if g == nil || ability == nil || !abilityHasOncePerTurnRestriction(ability) {
		return
	}
	if g.ActivatedAbilitiesThisTurn == nil {
		g.ActivatedAbilitiesThisTurn = make(map[game.ActivatedAbilityUse]bool)
	}
	g.ActivatedAbilitiesThisTurn[game.ActivatedAbilityUse{SourceID: sourceID, AbilityIndex: abilityIndex}] = true
}

func abilityHasOncePerTurnRestriction(ability *game.AbilityDef) bool {
	return ability.Timing == game.OncePerTurn || ability.Timing == game.SorceryOncePerTurn
}

func isSorcerySpeed(g *game.Game, playerID game.PlayerID) bool {
	return playerID == g.Turn.ActivePlayer &&
		g.Turn.IsMainPhase() &&
		g.Turn.Step == game.StepNone &&
		g.Stack.IsEmpty()
}

func landCardInstance(g *game.Game, player *game.Player, cardID id.ID) (*game.CardInstance, bool) {
	if player == nil || !player.Hand.Contains(cardID) {
		return nil, false
	}
	card := g.GetCardInstance(cardID)
	if card == nil || card.Def == nil || !card.Def.HasType(game.TypeLand) {
		return nil, false
	}
	return card, true
}

func entersSummoningSick(card *game.CardDef) bool {
	return card != nil && !card.HasKeyword(game.Haste)
}

func isSupportedSpell(card *game.CardDef) bool {
	return !card.HasType(game.TypeLand) &&
		(card.HasType(game.TypeCreature) ||
			card.HasType(game.TypeInstant) ||
			card.HasType(game.TypeSorcery))
}
