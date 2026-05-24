package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
)

const maxLegalXValue = 20

func (e *Engine) legalActions(g *game.Game, playerID game.PlayerID) []action.Action {
	if !canAct(g, playerID) {
		return []action.Action{action.Pass()}
	}

	actions := e.legalLandActions(g, playerID)
	actions = append(actions, e.legalCastActions(g, playerID)...)
	actions = append(actions, e.legalActivateAbilityActions(g, playerID)...)
	actions = append(actions, action.Pass())
	return actions
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
					if e.canCastSpell(g, playerID, cardID, targets, xValue, modes) {
						actions = append(actions, action.CastSpell(cardID, append([]game.Target(nil), targets...), xValue, append([]int(nil), modes...)))
					}
				}
			}
		}
	}
	return actions
}

func (e *Engine) applyAction(g *game.Game, playerID game.PlayerID, act action.Action) bool {
	switch act.Kind {
	case action.ActionPass:
		return true
	case action.ActionPlayLand:
		return e.applyPlayLand(g, playerID, act.PlayLand.CardID)
	case action.ActionCastSpell:
		return e.applyCastSpell(g, playerID, act.CastSpell)
	case action.ActionActivateAbility:
		return e.applyActivateAbility(g, playerID, act.ActivateAbility)
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
		if permanent == nil || permanent.Controller != playerID {
			continue
		}
		card := permanentCardDef(g, permanent)
		if card == nil {
			continue
		}
		for i := range card.Abilities {
			if canActivateManaAbility(g, playerID, permanent, &card.Abilities[i]) {
				actions = append(actions, action.ActivateAbility(permanent.ObjectID, i, nil, 0))
				continue
			}
			for _, targets := range targetChoicesForAbility(g, playerID, &card.Abilities[i]) {
				if canActivateEquipAbility(g, playerID, permanent, &card.Abilities[i], targets) {
					actions = append(actions, action.ActivateAbility(permanent.ObjectID, i, append([]game.Target(nil), targets...), 0))
				}
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

	createCardPermanent(g, card, playerID)
	g.Turn.LandsPlayedThisTurn++
	return true
}

func (e *Engine) applyCastSpell(g *game.Game, playerID game.PlayerID, cast action.CastSpellAction) bool {
	if !e.canCastSpell(g, playerID, cast.CardID, cast.Targets, cast.XValue, cast.ChosenModes) {
		return false
	}

	player := g.Players[playerID]
	card := g.GetCardInstance(cast.CardID)
	additionalCostsPaid, ok := paySpellCosts(g, playerID, card.Def, cast.XValue)
	if !ok {
		return false
	}
	if !player.Hand.Remove(cast.CardID) {
		panic("cast spell disappeared from hand after validation")
	}
	g.Stack.Push(&game.StackObject{
		ID:                  g.IDGen.Next(),
		Kind:                game.StackSpell,
		SourceID:            cast.CardID,
		Controller:          playerID,
		Targets:             append([]game.Target(nil), cast.Targets...),
		ChosenModes:         append([]int(nil), cast.ChosenModes...),
		XValue:              cast.XValue,
		AdditionalCostsPaid: additionalCostsPaid,
	})
	return true
}

func (e *Engine) applyActivateAbility(g *game.Game, playerID game.PlayerID, activate action.ActivateAbilityAction) bool {
	permanent, ability, ok := activatedAbilitySource(g, playerID, activate.SourceID, activate.AbilityIndex)
	if !ok || activate.XValue != 0 {
		return false
	}
	if canActivateManaAbility(g, playerID, permanent, ability) {
		if len(activate.Targets) != 0 {
			return false
		}
		if !payCost(g, playerID, ability.ManaCost) {
			return false
		}
		if hasTapCost(ability) {
			if permanent.Tapped {
				return false
			}
			permanent.Tapped = true
		}
		obj := &game.StackObject{
			ID:           g.IDGen.Next(),
			Kind:         game.StackActivatedAbility,
			SourceID:     permanent.ObjectID,
			AbilityIndex: activate.AbilityIndex,
			Controller:   playerID,
		}
		for _, effect := range ability.Effects {
			e.resolveEffect(g, obj, effect, nil)
		}
		return true
	}
	if !canActivateEquipAbility(g, playerID, permanent, ability, activate.Targets) {
		return false
	}
	if !payCost(g, playerID, ability.ManaCost) {
		return false
	}
	g.Stack.Push(&game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackActivatedAbility,
		SourceID:     permanent.ObjectID,
		AbilityIndex: activate.AbilityIndex,
		Controller:   playerID,
		Targets:      append([]game.Target(nil), activate.Targets...),
	})
	return true
}

func (e *Engine) canCastSpell(g *game.Game, playerID game.PlayerID, cardID id.ID, targets []game.Target, xValue int, chosenModes []int) bool {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return false
	}
	if xValue < 0 {
		return false
	}
	player := g.Players[playerID]
	card := g.GetCardInstance(cardID)
	if card == nil || card.Def == nil || !player.Hand.Contains(cardID) {
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
	return canPaySpellCosts(g, playerID, card.Def, xValue)
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
	if permanent == nil || permanent.Controller != playerID {
		return nil, nil, false
	}
	card := permanentCardDef(g, permanent)
	if card == nil || abilityIndex >= len(card.Abilities) {
		return nil, nil, false
	}
	return permanent, &card.Abilities[abilityIndex], true
}

func canActivateEquipAbility(g *game.Game, playerID game.PlayerID, permanent *game.Permanent, ability *game.AbilityDef, targets []game.Target) bool {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer || permanent == nil || permanent.Controller != playerID || ability == nil {
		return false
	}
	if ability.Kind != game.ActivatedAbility || ability.IsManaAbility || !isEquipmentPermanent(g, permanent) {
		return false
	}
	if !abilityHasKeyword(ability, game.Equip) && ability.Timing != game.SorceryOnly {
		return false
	}
	if !isSorcerySpeed(g, playerID) || ability.AdditionalCost != "" {
		return false
	}
	if !targetsValidForAbility(g, playerID, ability, targets) {
		return false
	}
	if len(targets) != 1 || targets[0].Kind != game.TargetPermanent {
		return false
	}
	target := permanentByObjectID(g, targets[0].PermanentID)
	if target == nil || target.Controller != playerID || !canAttachPermanent(g, permanent, target) {
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

func canActivateManaAbility(g *game.Game, playerID game.PlayerID, permanent *game.Permanent, ability *game.AbilityDef) bool {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer || permanent == nil || permanent.Controller != playerID || ability == nil {
		return false
	}
	if ability.Kind != game.ActivatedAbility || !ability.IsManaAbility || ability.IsLoyaltyAbility {
		return false
	}
	if len(ability.Targets) != 0 || !manaAbilityHasAddManaEffect(ability) {
		return false
	}
	if ability.Timing != game.NoTimingRestriction {
		return false
	}
	if hasTapCost(ability) {
		if permanent.Tapped {
			return false
		}
		card := permanentCardDef(g, permanent)
		if card != nil && card.HasType(game.TypeCreature) && permanent.SummoningSick {
			return false
		}
	} else if ability.AdditionalCost != "" {
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
	switch ability.AdditionalCost {
	case "Tap", "{T}", "T":
		return true
	default:
		return false
	}
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
