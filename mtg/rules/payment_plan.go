package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
)

type paymentPlan struct {
	poolSpend      map[mana.Unit]int
	manaTaps       []manaTap
	convokeTaps    []*game.Permanent
	delveExiles    []id.ID
	lifePayment    int
	symbolPayments []game.SymbolPayment
}

type spellCostPlan struct {
	mana       paymentPlan
	additional additionalCostPlan
	option     spellCostOption
}

type abilityCostPlan struct {
	mana       paymentPlan
	additional additionalCostPlan
	tapSource  bool
}

type manaTap struct {
	permanent *game.Permanent
	color     mana.Color
	amount    int
	snow      bool
}

type manaSource struct {
	permanent *game.Permanent
	color     mana.Color
	amount    int
	snow      bool
}

func canPayCost(g *game.Game, playerID game.PlayerID, cost *mana.Cost) bool {
	return canPayCostWithX(g, playerID, cost, 0)
}

func canPayCostWithX(g *game.Game, playerID game.PlayerID, cost *mana.Cost, xValue int) bool {
	_, ok := buildPaymentPlan(g, playerID, cost, xValue, nil)
	return ok
}

func payCost(g *game.Game, playerID game.PlayerID, cost *mana.Cost) bool {
	return payCostWithX(g, playerID, cost, 0)
}

func payCostWithX(g *game.Game, playerID game.PlayerID, cost *mana.Cost, xValue int) bool {
	plan, ok := buildPaymentPlan(g, playerID, cost, xValue, nil)
	if !ok {
		return false
	}
	return applyPaymentPlan(g, playerID, plan)
}

func canPaySpellCosts(g *game.Game, req spellPaymentRequest) bool {
	for _, option := range spellCostOptionsForZoneAndKicker(req.card, req.sourceZone, req.kickerPaid) {
		if _, ok := buildSpellCostPlanForOption(g, req.playerID, req.cardID, req.sourceZone, option, req.xValue, nil); ok {
			return true
		}
	}
	return false
}

func paySpellCosts(g *game.Game, req spellPaymentRequest) ([]string, bool) {
	plan, ok := buildSpellCostPlan(g, req)
	if !ok {
		return nil, false
	}
	player, ok := playerForCostPayment(g, req.playerID)
	if !ok || !additionalCostPlanStillValid(g, player, plan.additional) || !paymentPlanStillValid(g, player, plan.mana) {
		return nil, false
	}
	if !applyPaymentPlan(g, req.playerID, plan.mana) {
		return nil, false
	}
	if !applyAdditionalCostPlan(g, plan.additional) {
		panic("spell cost plan became invalid while paying additional costs")
	}
	return plan.additional.paid, true
}

func buildSpellCostPlan(g *game.Game, req spellPaymentRequest) (spellCostPlan, bool) {
	options := spellCostOptionsForZoneAndKicker(req.card, req.sourceZone, req.kickerPaid)
	if len(options) == 0 {
		return spellCostPlan{}, false
	}
	if req.prefs != nil {
		for _, option := range options {
			if option.index == req.prefs.alternativeIndex {
				return buildSpellCostPlanForOption(g, req.playerID, req.cardID, req.sourceZone, option, req.xValue, req.prefs)
			}
		}
		return spellCostPlan{}, false
	}
	for _, option := range options {
		if plan, ok := buildSpellCostPlanForOption(g, req.playerID, req.cardID, req.sourceZone, option, req.xValue, nil); ok {
			return plan, true
		}
	}
	return spellCostPlan{}, false
}

func buildAbilityCostPlan(g *game.Game, req abilityPaymentRequest) (abilityCostPlan, bool) {
	plan := abilityCostPlan{}
	if req.source == nil || req.ability == nil {
		return plan, false
	}
	if req.xValue != 0 && !costHasVariableMana(manaCostPtr(req.ability.ManaCost)) {
		return plan, false
	}
	tapSource := hasTapCost(req.ability)
	if tapSource && !canTapPermanentForAbility(g, req.source) {
		return plan, false
	}
	additional, ok := buildAdditionalCostPlanForCosts(g, req.playerID, abilityAdditionalCosts(req.ability), req.prefs)
	if !ok {
		return plan, false
	}
	excluded := make(map[id.ID]bool)
	if tapSource {
		excluded[req.source.ObjectID] = true
	}
	for _, sacrifice := range additional.sacrifices {
		excluded[sacrifice.ObjectID] = true
	}
	manaPlan, ok := buildPaymentPlanWithPreferences(g, req.playerID, manaCostPtr(req.ability.ManaCost), req.xValue, excluded, req.prefs)
	if !ok {
		return plan, false
	}
	plan.mana = manaPlan
	plan.additional = additional
	plan.tapSource = tapSource
	return plan, true
}

func payAbilityCosts(g *game.Game, req abilityPaymentRequest) (abilityCostPlan, bool) {
	plan, ok := buildAbilityCostPlan(g, req)
	if !ok {
		return plan, false
	}
	player, ok := playerForCostPayment(g, req.playerID)
	if !ok || !abilityCostPlanStillValid(g, player, req.source, plan) {
		return plan, false
	}
	if !applyPaymentPlan(g, req.playerID, plan.mana) {
		return plan, false
	}
	if plan.tapSource {
		if !tapPermanentForAbility(g, req.source) {
			return plan, false
		}
	}
	if !applyAdditionalCostPlan(g, plan.additional) {
		panic("ability cost plan became invalid while paying additional costs")
	}
	return plan, true
}

func buildSpellCostPlanForOption(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone game.ZoneType, option spellCostOption, xValue int, prefs *paymentPreferences) (spellCostPlan, bool) {
	option = applyCostModifiers(g, costModificationContext{player: playerID, card: option.card, cardID: cardID, sourceZone: sourceZone, option: option})
	plan := spellCostPlan{option: option}
	additional, ok := buildAdditionalCostPlanForCosts(g, playerID, option.additionalCosts, prefs)
	if !ok {
		return plan, false
	}
	excluded := make(map[id.ID]bool)
	for _, sacrifice := range additional.sacrifices {
		excluded[sacrifice.ObjectID] = true
	}
	manaPlan, ok := buildPaymentPlanWithPreferences(g, playerID, option.manaCost, xValue, excluded, prefs)
	if !ok {
		convokeTaps, convokedCost, convokeOK := convokePayment(g, playerID, option.manaCost, xValue, excluded)
		if option.card.HasKeyword(game.Convoke) && convokeOK {
			for _, permanent := range convokeTaps {
				excluded[permanent.ObjectID] = true
			}
			manaPlan, ok = buildPaymentPlanWithPreferences(g, playerID, convokedCost, xValue, excluded, prefs)
			if ok {
				manaPlan.convokeTaps = convokeTaps
			}
		}
		if !ok && option.card.HasKeyword(game.Delve) {
			delveExiles, generic, delveOK := delveCandidates(g, playerID, option.manaCost, xValue, cardID, sourceZone)
			for exiledCount := 1; delveOK && exiledCount <= min(generic, len(delveExiles)); exiledCount++ {
				delvedCost := costWithGenericRequirement(option.manaCost, generic-exiledCount)
				manaPlan, ok = buildPaymentPlanWithPreferences(g, playerID, delvedCost, 0, excluded, prefs)
				if ok {
					manaPlan.delveExiles = append([]id.ID(nil), delveExiles[:exiledCount]...)
					break
				}
			}
		}
		if !ok {
			return plan, false
		}
	}
	plan.additional = additional
	plan.mana = manaPlan
	return plan, true
}

func buildPaymentPlan(g *game.Game, playerID game.PlayerID, cost *mana.Cost, xValue int, exclude map[id.ID]bool) (paymentPlan, bool) {
	return buildPaymentPlanWithPreferences(g, playerID, cost, xValue, exclude, nil)
}

func buildPaymentPlanWithPreferences(g *game.Game, playerID game.PlayerID, cost *mana.Cost, xValue int, exclude map[id.ID]bool, prefs *paymentPreferences) (paymentPlan, bool) {
	plan := paymentPlan{poolSpend: make(map[mana.Unit]int)}
	player, ok := playerForCostPayment(g, playerID)
	if !ok {
		return plan, false
	}
	pool := snapshotPool(player)
	manaSources := availableManaSources(g, playerID, exclude)
	if xValue < 0 {
		return plan, false
	}
	if cost == nil {
		return plan, true
	}

	for _, symbol := range *cost {
		switch symbol.Kind {
		case mana.ColoredSymbol:
			if !payColoredSymbol(&plan, pool, manaSources, symbol, symbol.Color, game.SymbolPaymentMana) {
				return plan, false
			}
		case mana.ColorlessSymbol:
			if !payColoredSymbol(&plan, pool, manaSources, symbol, mana.Colorless, game.SymbolPaymentMana) {
				return plan, false
			}
		}
	}
	for _, symbol := range *cost {
		if symbol.Kind == mana.SnowSymbol {
			if !paySnowSymbol(&plan, pool, manaSources, symbol) {
				return plan, false
			}
		}
	}
	for _, symbol := range *cost {
		switch symbol.Kind {
		case mana.HybridSymbol:
			if !payHybridSymbol(&plan, pool, manaSources, symbol) {
				return plan, false
			}
		case mana.MonoHybridSymbol:
			if !payMonoHybridSymbol(&plan, pool, manaSources, symbol) {
				return plan, false
			}
		case mana.PhyrexianSymbol:
			if !payPhyrexianSymbol(player, &plan, pool, manaSources, symbol, prefs) {
				return plan, false
			}
		}
	}
	for _, symbol := range *cost {
		switch symbol.Kind {
		case mana.GenericSymbol:
			if !payGenericSymbol(&plan, pool, manaSources, symbol, symbol.Generic, game.SymbolPaymentGeneric) {
				return plan, false
			}
		case mana.VariableSymbol:
			if !payGenericSymbol(&plan, pool, manaSources, symbol, xValue, game.SymbolPaymentX) {
				return plan, false
			}
		default:
			if symbol.Kind != mana.ColoredSymbol &&
				symbol.Kind != mana.ColorlessSymbol &&
				symbol.Kind != mana.SnowSymbol &&
				symbol.Kind != mana.HybridSymbol &&
				symbol.Kind != mana.MonoHybridSymbol &&
				symbol.Kind != mana.PhyrexianSymbol {
				return plan, false
			}
		}
	}
	return plan, true
}

func paymentPlanStillValid(g *game.Game, player *game.Player, plan paymentPlan) bool {
	tappedMana := make(map[mana.Unit]int)
	for _, tap := range plan.manaTaps {
		if tap.permanent.Tapped || effectiveController(g, tap.permanent) != player.ID {
			return false
		}
		output, ok := permanentManaOutput(g, tap.permanent)
		if !ok || output.color != tap.color || output.amount != tap.amount || output.snow != tap.snow {
			return false
		}
		tappedMana[mana.Unit{Color: tap.color, Snow: tap.snow}] += tap.amount
	}
	for _, permanent := range plan.convokeTaps {
		if !canConvokeWith(g, player.ID, permanent, nil) {
			return false
		}
	}
	for _, cardID := range plan.delveExiles {
		if !player.Graveyard.Contains(cardID) {
			return false
		}
	}
	for _, color := range paymentColors {
		for _, snow := range []bool{false, true} {
			unit := mana.Unit{Color: color, Snow: snow}
			if player.ManaPool.Units()[unit]+tappedMana[unit] < plan.poolSpend[unit] {
				return false
			}
		}
	}
	if player.Life < plan.lifePayment {
		return false
	}
	return true
}

func abilityCostPlanStillValid(g *game.Game, player *game.Player, source *game.Permanent, plan abilityCostPlan) bool {
	if source == nil {
		return false
	}
	if plan.tapSource && !canTapPermanentForAbility(g, source) {
		return false
	}
	return additionalCostPlanStillValid(g, player, plan.additional) &&
		paymentPlanStillValid(g, player, plan.mana)
}

func clonePaymentPlan(plan paymentPlan) paymentPlan {
	plan.poolSpend = cloneUnitCounts(plan.poolSpend)
	plan.manaTaps = append([]manaTap(nil), plan.manaTaps...)
	plan.symbolPayments = append([]game.SymbolPayment(nil), plan.symbolPayments...)
	return plan
}

func cloneUnitCounts(units map[mana.Unit]int) map[mana.Unit]int {
	clone := make(map[mana.Unit]int, len(units))
	for unit, amount := range units {
		clone[unit] = amount
	}
	return clone
}

func replaceUnitCounts(dst, src map[mana.Unit]int) {
	for unit := range dst {
		delete(dst, unit)
	}
	for unit, amount := range src {
		dst[unit] = amount
	}
}

func cloneManaSources(sources map[mana.Color][]manaSource) map[mana.Color][]manaSource {
	clone := make(map[mana.Color][]manaSource, len(sources))
	for color, colorSources := range sources {
		clone[color] = append([]manaSource(nil), colorSources...)
	}
	return clone
}

func replaceManaSources(dst, src map[mana.Color][]manaSource) {
	for color := range dst {
		delete(dst, color)
	}
	for color, colorSources := range src {
		dst[color] = append([]manaSource(nil), colorSources...)
	}
}

func costRequirements(cost *mana.Cost, xValue int) (map[mana.Color]int, int, bool) {
	colored := make(map[mana.Color]int)
	if xValue < 0 {
		return nil, 0, false
	}
	if cost == nil {
		return colored, 0, true
	}

	generic := 0
	for _, symbol := range *cost {
		switch symbol.Kind {
		case mana.ColoredSymbol:
			colored[symbol.Color]++
		case mana.ColorlessSymbol:
			colored[mana.Colorless]++
		case mana.GenericSymbol:
			generic += symbol.Generic
		case mana.VariableSymbol:
			generic += xValue
		default:
			return nil, 0, false
		}
	}
	return colored, generic, true
}

func snapshotPool(player *game.Player) map[mana.Unit]int {
	return player.ManaPool.Units()
}
