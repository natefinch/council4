package payment

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/opt"
)

// paymentPlan describes the concrete steps needed to pay a mana cost.
type paymentPlan struct {
	poolSpend      map[mana.Unit]int
	manaTaps       []manaTap
	convokeTaps    []*game.Permanent
	delveExiles    []id.ID
	lifePayment    int
	symbolPayments []game.SymbolPayment
}

// spellCostPlan is the full payment plan for a spell, including mana and
// additional costs.
type spellCostPlan struct {
	mana       paymentPlan
	additional additionalCostPlan
	option     spellCostOption
}

// abilityCostPlan is the full payment plan for an activated ability.
type abilityCostPlan struct {
	mana       paymentPlan
	additional additionalCostPlan
	tapSource  bool
}

// manaTap records a planned tap of a mana-producing permanent.
type manaTap struct {
	permanent *game.Permanent
	color     mana.Color
	amount    int
	snow      bool
}

// manaSource is a candidate mana-producing permanent used during plan building.
type manaSource struct {
	permanent *game.Permanent
	color     mana.Color
	amount    int
	snow      bool
}

// paymentColors is the deterministic ordering used when spending mana. Callers
// must consume mana sources through this slice rather than ranging over maps.
var paymentColors = []mana.Color{
	mana.White,
	mana.Blue,
	mana.Black,
	mana.Red,
	mana.Green,
	mana.Colorless,
}

func canPayCostWithX(s State, playerID game.PlayerID, cost *mana.Cost, xValue int) bool {
	_, ok := buildPaymentPlan(s, playerID, cost, xValue, nil)
	return ok
}

func canPaySpellCosts(s State, req SpellRequest) bool {
	for _, option := range spellCostOptionsForZoneAndKicker(req.Card, req.SourceZone, req.KickerPaid) {
		if _, ok := buildSpellCostPlanForOption(s, req.PlayerID, req.CardID, req.SourceZone, option, req.XValue, nil); ok {
			return true
		}
	}
	return false
}

func paySpellCosts(s State, req SpellRequest) ([]string, bool) {
	plan, ok := buildSpellCostPlan(s, req)
	if !ok {
		return nil, false
	}
	player, ok := s.Player(req.PlayerID)
	if !ok || !additionalCostPlanStillValid(s, player, plan.additional) || !paymentPlanStillValid(s, player, plan.mana) {
		return nil, false
	}
	if !applyPaymentPlan(s, req.PlayerID, plan.mana) {
		return nil, false
	}
	if !applyAdditionalCostPlan(s, plan.additional) {
		panic("spell cost plan became invalid while paying additional costs")
	}
	return plan.additional.paid, true
}

func buildSpellCostPlan(s State, req SpellRequest) (spellCostPlan, bool) {
	options := spellCostOptionsForZoneAndKicker(req.Card, req.SourceZone, req.KickerPaid)
	if len(options) == 0 {
		return spellCostPlan{}, false
	}
	if req.Prefs != nil {
		for _, option := range options {
			if option.index == req.Prefs.AlternativeIndex {
				return buildSpellCostPlanForOption(s, req.PlayerID, req.CardID, req.SourceZone, option, req.XValue, req.Prefs)
			}
		}
		return spellCostPlan{}, false
	}
	for _, option := range options {
		if plan, ok := buildSpellCostPlanForOption(s, req.PlayerID, req.CardID, req.SourceZone, option, req.XValue, nil); ok {
			return plan, true
		}
	}
	return spellCostPlan{}, false
}

func buildAbilityCostPlan(s State, req AbilityRequest) (abilityCostPlan, bool) {
	plan := abilityCostPlan{}
	if req.Source == nil || req.Ability == nil {
		return plan, false
	}
	if req.XValue != 0 && !costHasVariableMana(manaCostPtr(req.Ability.ManaCost)) {
		return plan, false
	}
	tapSource := hasTapCost(req.Ability)
	if tapSource && !canTapForAbility(s, req.Source) {
		return plan, false
	}
	additional, ok := buildAdditionalCostPlanForCosts(s, req.PlayerID, abilityAdditionalCosts(req.Ability), req.Prefs, req.Source)
	if !ok {
		return plan, false
	}
	excluded := make(map[id.ID]bool)
	if tapSource {
		excluded[req.Source.ObjectID] = true
	}
	for _, sacrifice := range additional.sacrifices {
		excluded[sacrifice.ObjectID] = true
	}
	manaPlan, ok := buildPaymentPlanWithPreferences(s, req.PlayerID, manaCostPtr(req.Ability.ManaCost), req.XValue, excluded, req.Prefs)
	if !ok {
		return plan, false
	}
	plan.mana = manaPlan
	plan.additional = additional
	plan.tapSource = tapSource
	return plan, true
}

func payAbilityCosts(s State, req AbilityRequest) (abilityCostPlan, bool) {
	plan, ok := buildAbilityCostPlan(s, req)
	if !ok {
		return plan, false
	}
	player, ok := s.Player(req.PlayerID)
	if !ok || !abilityCostPlanStillValid(s, player, req.Source, plan) {
		return plan, false
	}
	if !applyPaymentPlan(s, req.PlayerID, plan.mana) {
		return plan, false
	}
	if plan.tapSource {
		if !tapForAbility(s, req.Source) {
			return plan, false
		}
	}
	if !applyAdditionalCostPlan(s, plan.additional) {
		panic("ability cost plan became invalid while paying additional costs")
	}
	return plan, true
}

func canPayGenericCost(s State, req GenericRequest) bool {
	if len(req.AdditionalCosts) > 0 {
		if _, ok := buildGenericCostPlan(s, req); !ok {
			return false
		}
		return true
	}
	if len(req.Exclude) > 0 {
		_, ok := buildPaymentPlan(s, req.PlayerID, req.Cost, req.XValue, req.Exclude)
		return ok
	}
	return canPayCostWithX(s, req.PlayerID, req.Cost, req.XValue)
}

func payGenericCost(s State, req GenericRequest) bool {
	if len(req.AdditionalCosts) > 0 {
		plan, ok := buildGenericCostPlan(s, req)
		if !ok {
			return false
		}
		player, ok := s.Player(req.PlayerID)
		if !ok || !additionalCostPlanStillValid(s, player, plan.additional) || !paymentPlanStillValid(s, player, plan.mana) {
			return false
		}
		if !applyPaymentPlan(s, req.PlayerID, plan.mana) {
			return false
		}
		if !applyAdditionalCostPlan(s, plan.additional) {
			panic("generic cost plan became invalid while paying additional costs")
		}
		return true
	}
	plan, ok := buildPaymentPlanWithPreferences(s, req.PlayerID, req.Cost, req.XValue, req.Exclude, req.Prefs)
	if !ok {
		return false
	}
	player, ok := s.Player(req.PlayerID)
	if !ok || !paymentPlanStillValid(s, player, plan) {
		return false
	}
	return applyPaymentPlan(s, req.PlayerID, plan)
}

func buildGenericCostPlan(s State, req GenericRequest) (spellCostPlan, bool) {
	plan := spellCostPlan{}
	additional, ok := buildAdditionalCostPlanForCosts(s, req.PlayerID, req.AdditionalCosts, req.Prefs, nil)
	if !ok {
		return plan, false
	}
	excluded := make(map[id.ID]bool)
	for k, v := range req.Exclude {
		excluded[k] = v
	}
	for _, sacrifice := range additional.sacrifices {
		excluded[sacrifice.ObjectID] = true
	}
	manaPlan, ok := buildPaymentPlanWithPreferences(s, req.PlayerID, req.Cost, req.XValue, excluded, req.Prefs)
	if !ok {
		return plan, false
	}
	plan.additional = additional
	plan.mana = manaPlan
	return plan, true
}

func buildSpellCostPlanForOption(s State, playerID game.PlayerID, cardID id.ID, sourceZone game.ZoneType, option spellCostOption, xValue int, prefs *Preferences) (spellCostPlan, bool) {
	option = applyCostModifiers(s, costModificationContext{player: playerID, card: option.card, cardID: cardID, sourceZone: sourceZone, option: option})
	plan := spellCostPlan{option: option}
	additional, ok := buildAdditionalCostPlanForCosts(s, playerID, option.additionalCosts, prefs, nil)
	if !ok {
		return plan, false
	}
	excluded := make(map[id.ID]bool)
	for _, sacrifice := range additional.sacrifices {
		excluded[sacrifice.ObjectID] = true
	}
	manaPlan, ok := buildPaymentPlanWithPreferences(s, playerID, option.manaCost, xValue, excluded, prefs)
	if !ok {
		convokeTaps, convokedCost, convokeOK := convokePayment(s, playerID, option.manaCost, xValue, excluded)
		if option.card.HasKeyword(game.Convoke) && convokeOK {
			for _, permanent := range convokeTaps {
				excluded[permanent.ObjectID] = true
			}
			manaPlan, ok = buildPaymentPlanWithPreferences(s, playerID, convokedCost, xValue, excluded, prefs)
			if ok {
				manaPlan.convokeTaps = convokeTaps
			}
		}
		if !ok && option.card.HasKeyword(game.Delve) {
			delveExiles, generic, delveOK := delveCandidates(s, playerID, option.manaCost, xValue, cardID, sourceZone)
			for exiledCount := 1; delveOK && exiledCount <= min(generic, len(delveExiles)); exiledCount++ {
				delvedCost := costWithGenericRequirement(option.manaCost, generic-exiledCount)
				manaPlan, ok = buildPaymentPlanWithPreferences(s, playerID, delvedCost, 0, excluded, prefs)
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

func buildPaymentPlan(s State, playerID game.PlayerID, cost *mana.Cost, xValue int, exclude map[id.ID]bool) (paymentPlan, bool) {
	return buildPaymentPlanWithPreferences(s, playerID, cost, xValue, exclude, nil)
}

func buildPaymentPlanWithPreferences(s State, playerID game.PlayerID, cost *mana.Cost, xValue int, exclude map[id.ID]bool, prefs *Preferences) (paymentPlan, bool) {
	plan := paymentPlan{poolSpend: make(map[mana.Unit]int)}
	player, ok := s.Player(playerID)
	if !ok {
		return plan, false
	}
	pool := snapshotPool(player)
	manaSources := availableManaSources(s, playerID, exclude)
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

func paymentPlanStillValid(s State, player *game.Player, plan paymentPlan) bool {
	tappedMana := make(map[mana.Unit]int)
	for _, tap := range plan.manaTaps {
		if tap.permanent.Tapped || s.EffectiveController(tap.permanent) != player.ID {
			return false
		}
		color, amount, snow, ok := permanentManaOutput(s, tap.permanent)
		if !ok || color != tap.color || amount != tap.amount || snow != tap.snow {
			return false
		}
		tappedMana[mana.Unit{Color: tap.color, Snow: tap.snow}] += tap.amount
	}
	for _, permanent := range plan.convokeTaps {
		if !canConvokeWith(s, player.ID, permanent, nil) {
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

func abilityCostPlanStillValid(s State, player *game.Player, source *game.Permanent, plan abilityCostPlan) bool {
	if source == nil {
		return false
	}
	if plan.tapSource && !canTapForAbility(s, source) {
		return false
	}
	return additionalCostPlanStillValid(s, player, plan.additional) &&
		paymentPlanStillValid(s, player, plan.mana)
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

// hasTapCost reports whether the ability has a tap additional cost.
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

// costHasVariableMana reports whether the cost contains an X (variable) symbol.
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

// abilityAdditionalCosts returns a copy of the ability's additional costs.
func abilityAdditionalCosts(ability *game.AbilityDef) []game.AdditionalCost {
	if ability == nil {
		return nil
	}
	return append([]game.AdditionalCost(nil), ability.AdditionalCosts...)
}

// manaCostPtr returns a pointer to the mana cost value, or nil if it does not exist.
func manaCostPtr(cost opt.V[mana.Cost]) *mana.Cost {
	if !cost.Exists {
		return nil
	}
	return &cost.Val
}

// canTapForAbility reports whether the permanent can be tapped as an ability cost.
func canTapForAbility(s State, p *game.Permanent) bool {
	if p.Tapped {
		return false
	}
	return !s.PermanentHasType(p, game.TypeCreature) || !p.SummoningSick
}

// tapForAbility taps a permanent as an ability cost.
func tapForAbility(s State, p *game.Permanent) bool {
	if !canTapForAbility(s, p) {
		return false
	}
	s.SetTapped(p, true)
	return true
}

// canConvokeWith reports whether the permanent can be used for convoke.
func canConvokeWith(s State, playerID game.PlayerID, p *game.Permanent, exclude map[id.ID]bool) bool {
	if exclude[p.ObjectID] || p.Tapped || p.PhasedOut || s.EffectiveController(p) != playerID {
		return false
	}
	return s.PermanentHasType(p, game.TypeCreature)
}
