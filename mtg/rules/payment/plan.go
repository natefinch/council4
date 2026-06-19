package payment

import (
	"maps"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
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
	permanent    *game.Permanent
	color        mana.Color
	amount       int
	snow         bool
	untap        bool
	abilityIndex int
	timing       game.TimingRestriction
}

// manaSource is a candidate mana-producing permanent used during plan building.
type manaSource struct {
	permanent    *game.Permanent
	color        mana.Color
	amount       int
	snow         bool
	untap        bool
	abilityIndex int
	timing       game.TimingRestriction
}

// paymentColors is the deterministic ordering used when spending mana. Callers
// must consume mana sources through this slice rather than ranging over maps.
var paymentColors = []mana.Color{
	mana.W,
	mana.U,
	mana.B,
	mana.R,
	mana.G,
	mana.C,
}

func canPayCostWithX(s State, playerID game.PlayerID, manaCost *cost.Mana, xValue int) bool {
	_, ok := buildPaymentPlan(s, playerID, manaCost, xValue, nil)
	return ok
}

func canPaySpellCosts(s State, req SpellRequest) bool {
	for _, option := range spellCostOptionsForRequest(req) {
		if _, ok := buildSpellCostPlanForOption(s, req.PlayerID, req.CardID, req.SourceZone, option, req.XValue, nil); ok {
			return true
		}
	}
	return false
}

func paySpellCosts(s State, req SpellRequest) (additionalPaid []string, poolSpend map[mana.Color]int, ok bool) {
	plan, ok := buildSpellCostPlan(s, req)
	if !ok {
		return nil, nil, false
	}
	player, ok := s.Player(req.PlayerID)
	if !ok || !additionalCostPlanStillValid(s, player, plan.additional) || !paymentPlanStillValid(s, player, plan.mana) {
		return nil, nil, false
	}
	if !applyPaymentPlan(s, req.PlayerID, plan.mana) {
		return nil, nil, false
	}
	if !applyAdditionalCostPlan(s, plan.additional) {
		panic("spell cost plan became invalid while paying additional costs")
	}
	return plan.additional.paid, coloredPoolSpend(plan.mana.poolSpend), true
}

// coloredPoolSpend sums a plan's per-unit pool spend into per-color totals
// (folding snow and nonsnow of a color together). It reports exactly how much
// pool mana of each color the plan consumed, which the rules engine uses to
// resolve mana-spend riders without inferring spend from gross pool deltas that
// mid-payment mana production could mask.
func coloredPoolSpend(poolSpend map[mana.Unit]int) map[mana.Color]int {
	if len(poolSpend) == 0 {
		return nil
	}
	colored := make(map[mana.Color]int, len(poolSpend))
	for unit, amount := range poolSpend {
		if amount > 0 {
			colored[unit.Color] += amount
		}
	}
	return colored
}

func buildSpellCostPlan(s State, req SpellRequest) (spellCostPlan, bool) {
	options := spellCostOptionsForRequest(req)
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
	if req.Source == nil && req.SourceCardID == 0 {
		return plan, false
	}
	if req.XValue != 0 && !costHasVariableMana(manaCostPtr(req.ManaCost)) && !additionalCostsUseX(req.AdditionalCosts) {
		return plan, false
	}
	tapSource := hasTapCostOf(req.AdditionalCosts)
	if tapSource && (req.Source == nil || !canTapForAbility(s, req.Source)) {
		return plan, false
	}
	sourceCardID := req.SourceCardID
	sourceZone := req.SourceZone
	if req.Source != nil && sourceCardID == 0 {
		sourceCardID = req.Source.CardInstanceID
		sourceZone = zone.Battlefield
	}
	additional, ok := buildAdditionalCostPlanForCosts(s, req.PlayerID, req.AdditionalCosts, req.XValue, clonePreferences(req.Prefs), req.Source, sourceCardID, sourceZone)
	if !ok {
		return plan, false
	}
	manaPlan, ok := buildPaymentPlanWithPreferences(s, req.PlayerID, manaCostPtr(req.ManaCost), req.XValue, abilityManaExclusions(additional, tapSource, req.Source, true), clonePreferences(req.Prefs))
	if !ok {
		additional, manaPlan, ok = retryAbilityCostPlanAvoidingManaTapConflict(s, req, sourceCardID, sourceZone, tapSource, additional)
		if !ok {
			return plan, false
		}
	}
	plan.mana = manaPlan
	plan.additional = additional
	plan.tapSource = tapSource
	return plan, true
}

func retryAbilityCostPlanAvoidingManaTapConflict(s State, req AbilityRequest, sourceCardID id.ID, sourceZone zone.Type, tapSource bool, previous additionalCostPlan) (additionalCostPlan, paymentPlan, bool) {
	if len(previous.permanentsToTap) == 0 {
		return additionalCostPlan{}, paymentPlan{}, false
	}
	manaPlan, ok := buildPaymentPlanWithPreferences(s, req.PlayerID, manaCostPtr(req.ManaCost), req.XValue, abilityManaExclusions(previous, tapSource, req.Source, false), clonePreferences(req.Prefs))
	if !ok {
		return additionalCostPlan{}, paymentPlan{}, false
	}
	additional, ok := buildAdditionalCostPlanForCosts(s, req.PlayerID, req.AdditionalCosts, req.XValue, tapRetryPreferences(req.Prefs), req.Source, sourceCardID, sourceZone, paymentPlanTappedPermanents(manaPlan)...)
	if !ok {
		return additionalCostPlan{}, paymentPlan{}, false
	}
	manaPlan, ok = buildPaymentPlanWithPreferences(s, req.PlayerID, manaCostPtr(req.ManaCost), req.XValue, abilityManaExclusions(additional, tapSource, req.Source, true), clonePreferences(req.Prefs))
	if !ok {
		return additionalCostPlan{}, paymentPlan{}, false
	}
	return additional, manaPlan, true
}

func additionalCostsUseX(costs []cost.Additional) bool {
	for _, additional := range costs {
		if additional.AmountFromX {
			return true
		}
	}
	return false
}

func abilityManaExclusions(additional additionalCostPlan, tapSource bool, source *game.Permanent, includeTapPermanents bool) map[id.ID]bool {
	excluded := additionalManaExclusions(nil, additional, includeTapPermanents)
	if tapSource && source != nil {
		excluded[source.ObjectID] = true
	}
	return excluded
}

func additionalManaExclusions(base map[id.ID]bool, additional additionalCostPlan, includeTapPermanents bool) map[id.ID]bool {
	excluded := make(map[id.ID]bool)
	maps.Copy(excluded, base)
	for _, sacrifice := range additional.sacrifices {
		excluded[sacrifice.ObjectID] = true
	}
	for _, permanent := range additional.exilePermanents {
		excluded[permanent.ObjectID] = true
	}
	if includeTapPermanents {
		for _, permanent := range additional.permanentsToTap {
			excluded[permanent.ObjectID] = true
		}
	}
	return excluded
}

func tapRetryPreferences(prefs *Preferences) *Preferences {
	cloned := clonePreferences(prefs)
	if cloned != nil {
		cloned.TapChoices = nil
	}
	return cloned
}

func paymentPlanTappedPermanents(plan paymentPlan) []*game.Permanent {
	permanents := make([]*game.Permanent, 0, len(plan.manaTaps)+len(plan.convokeTaps))
	for _, tap := range plan.manaTaps {
		permanents = append(permanents, tap.permanent)
	}
	permanents = append(permanents, plan.convokeTaps...)
	return permanents
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
	additional, ok := buildAdditionalCostPlanForCosts(s, req.PlayerID, req.AdditionalCosts, req.XValue, clonePreferences(req.Prefs), nil, req.SourceCardID, zone.None)
	if !ok {
		return plan, false
	}
	manaPlan, ok := buildPaymentPlanWithPreferences(s, req.PlayerID, req.Cost, req.XValue, additionalManaExclusions(req.Exclude, additional, true), clonePreferences(req.Prefs))
	if !ok {
		additional, manaPlan, ok = retryGenericCostPlanAvoidingManaTapConflict(s, req, additional)
		if !ok {
			return plan, false
		}
	}
	plan.additional = additional
	plan.mana = manaPlan
	return plan, true
}

func retryGenericCostPlanAvoidingManaTapConflict(s State, req GenericRequest, previous additionalCostPlan) (additionalCostPlan, paymentPlan, bool) {
	if len(previous.permanentsToTap) == 0 {
		return additionalCostPlan{}, paymentPlan{}, false
	}
	manaPlan, ok := buildPaymentPlanWithPreferences(s, req.PlayerID, req.Cost, req.XValue, additionalManaExclusions(req.Exclude, previous, false), clonePreferences(req.Prefs))
	if !ok {
		return additionalCostPlan{}, paymentPlan{}, false
	}
	additional, ok := buildAdditionalCostPlanForCosts(s, req.PlayerID, req.AdditionalCosts, req.XValue, tapRetryPreferences(req.Prefs), nil, req.SourceCardID, zone.None, paymentPlanTappedPermanents(manaPlan)...)
	if !ok {
		return additionalCostPlan{}, paymentPlan{}, false
	}
	manaPlan, ok = buildPaymentPlanWithPreferences(s, req.PlayerID, req.Cost, req.XValue, additionalManaExclusions(req.Exclude, additional, true), clonePreferences(req.Prefs))
	if !ok {
		return additionalCostPlan{}, paymentPlan{}, false
	}
	return additional, manaPlan, true
}

func buildSpellCostPlanForOption(s State, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, option spellCostOption, xValue int, prefs *Preferences) (spellCostPlan, bool) {
	option = applyCostModifiers(s, costModificationContext{player: playerID, card: option.card, cardID: cardID, sourceZone: sourceZone, option: option})
	plan := spellCostPlan{option: option}
	additional, ok := buildAdditionalCostPlanForCosts(s, playerID, option.additionalCosts, xValue, clonePreferences(prefs), nil, cardID, sourceZone)
	if !ok {
		return plan, false
	}
	manaPlan, ok := buildSpellManaPlanForOption(s, playerID, cardID, sourceZone, option, xValue, additionalManaExclusions(nil, additional, true), clonePreferences(prefs))
	if !ok {
		additional, manaPlan, ok = retrySpellCostPlanAvoidingManaTapConflict(s, playerID, cardID, sourceZone, option, xValue, prefs, additional)
		if !ok {
			return plan, false
		}
	}
	plan.additional = additional
	plan.mana = manaPlan
	return plan, true
}

func retrySpellCostPlanAvoidingManaTapConflict(s State, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, option spellCostOption, xValue int, prefs *Preferences, previous additionalCostPlan) (additionalCostPlan, paymentPlan, bool) {
	if len(previous.permanentsToTap) == 0 {
		return additionalCostPlan{}, paymentPlan{}, false
	}
	manaPlan, ok := buildSpellManaPlanForOption(s, playerID, cardID, sourceZone, option, xValue, additionalManaExclusions(nil, previous, false), clonePreferences(prefs))
	if !ok {
		return additionalCostPlan{}, paymentPlan{}, false
	}
	additional, ok := buildAdditionalCostPlanForCosts(s, playerID, option.additionalCosts, xValue, tapRetryPreferences(prefs), nil, cardID, sourceZone, paymentPlanTappedPermanents(manaPlan)...)
	if !ok {
		return additionalCostPlan{}, paymentPlan{}, false
	}
	manaPlan, ok = buildSpellManaPlanForOption(s, playerID, cardID, sourceZone, option, xValue, additionalManaExclusions(nil, additional, true), clonePreferences(prefs))
	if !ok {
		return additionalCostPlan{}, paymentPlan{}, false
	}
	return additional, manaPlan, true
}

func buildSpellManaPlanForOption(s State, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, option spellCostOption, xValue int, excluded map[id.ID]bool, prefs *Preferences) (paymentPlan, bool) {
	manaPlan, ok := buildPaymentPlanWithPreferences(s, playerID, option.manaCost, xValue, excluded, prefs)
	if ok {
		return manaPlan, true
	}
	convokeTaps, convokedCost, convokeOK := convokePayment(s, playerID, option.manaCost, xValue, excluded)
	if option.card.HasKeyword(game.Convoke) && convokeOK {
		convokeExcluded := maps.Clone(excluded)
		for _, permanent := range convokeTaps {
			convokeExcluded[permanent.ObjectID] = true
		}
		manaPlan, ok = buildPaymentPlanWithPreferences(s, playerID, convokedCost, xValue, convokeExcluded, prefs)
		if ok {
			manaPlan.convokeTaps = convokeTaps
			return manaPlan, true
		}
	}
	if option.card.HasKeyword(game.Delve) {
		delveExiles, generic, delveOK := delveCandidates(s, playerID, option.manaCost, xValue, cardID, sourceZone)
		for exiledCount := 1; delveOK && exiledCount <= min(generic, len(delveExiles)); exiledCount++ {
			delvedCost := costWithGenericRequirement(option.manaCost, generic-exiledCount)
			manaPlan, ok = buildPaymentPlanWithPreferences(s, playerID, delvedCost, 0, excluded, prefs)
			if ok {
				manaPlan.delveExiles = append([]id.ID(nil), delveExiles[:exiledCount]...)
				return manaPlan, true
			}
		}
	}
	return paymentPlan{}, false
}

func buildPaymentPlan(s State, playerID game.PlayerID, manaCost *cost.Mana, xValue int, exclude map[id.ID]bool) (paymentPlan, bool) {
	return buildPaymentPlanWithPreferences(s, playerID, manaCost, xValue, exclude, nil)
}

func buildPaymentPlanWithPreferences(s State, playerID game.PlayerID, manaCost *cost.Mana, xValue int, exclude map[id.ID]bool, prefs *Preferences) (paymentPlan, bool) {
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
	if manaCost == nil {
		return plan, true
	}

	for _, symbol := range *manaCost {
		switch symbol.Kind {
		case cost.ColoredSymbol:
			if !payColoredSymbol(&plan, pool, manaSources, symbol, symbol.Color, game.SymbolPaymentMana) {
				return plan, false
			}
		case cost.ColorlessSymbol:
			if !payColoredSymbol(&plan, pool, manaSources, symbol, mana.C, game.SymbolPaymentMana) {
				return plan, false
			}
		default:
		}
	}
	for _, symbol := range *manaCost {
		if symbol.Kind == cost.SnowSymbol {
			if !paySnowSymbol(&plan, pool, manaSources, symbol) {
				return plan, false
			}
		}
	}
	for _, symbol := range *manaCost {
		switch symbol.Kind {
		case cost.HybridSymbol:
			if !payHybridSymbol(&plan, pool, manaSources, symbol) {
				return plan, false
			}
		case cost.TwobridSymbol:
			if !payMonoHybridSymbol(&plan, pool, manaSources, symbol) {
				return plan, false
			}
		case cost.PhyrexianSymbol:
			if !payPhyrexianSymbol(player, &plan, pool, manaSources, symbol, prefs) {
				return plan, false
			}
		default:
		}
	}
	for _, symbol := range *manaCost {
		switch symbol.Kind {
		case cost.GenericSymbol:
			if !payGenericSymbol(&plan, pool, manaSources, symbol, symbol.Generic, game.SymbolPaymentGeneric) {
				return plan, false
			}
		case cost.VariableSymbol:
			if !payGenericSymbol(&plan, pool, manaSources, symbol, xValue, game.SymbolPaymentX) {
				return plan, false
			}
		default:
			if symbol.Kind != cost.ColoredSymbol &&
				symbol.Kind != cost.ColorlessSymbol &&
				symbol.Kind != cost.SnowSymbol &&
				symbol.Kind != cost.HybridSymbol &&
				symbol.Kind != cost.TwobridSymbol &&
				symbol.Kind != cost.PhyrexianSymbol {
				return plan, false
			}
		}
	}
	return plan, true
}

func paymentPlanStillValid(s State, player *game.Player, plan paymentPlan) bool {
	tappedMana := make(map[mana.Unit]int)
	for _, tap := range plan.manaTaps {
		if tap.permanent.Tapped != tap.untap || s.EffectiveController(tap.permanent) != player.ID {
			return false
		}
		output, ok := permanentManaOutput(s, tap.permanent)
		if !ok ||
			output.color != tap.color ||
			output.amount != tap.amount ||
			output.snow != tap.snow ||
			output.untap != tap.untap ||
			output.abilityIndex != tap.abilityIndex ||
			output.timing != tap.timing {
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
	return player.Life >= plan.lifePayment
}

func abilityCostPlanStillValid(s State, player *game.Player, source *game.Permanent, plan abilityCostPlan) bool {
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
	maps.Copy(clone, units)
	return clone
}

func replaceUnitCounts(dst, src map[mana.Unit]int) {
	for unit := range dst {
		delete(dst, unit)
	}
	maps.Copy(dst, src)
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

func costRequirements(manaCost *cost.Mana, xValue int) (colored map[mana.Color]int, generic int, ok bool) {
	colored = make(map[mana.Color]int)
	if xValue < 0 {
		return nil, 0, false
	}
	if manaCost == nil {
		return colored, 0, true
	}

	generic = 0
	for _, symbol := range *manaCost {
		switch symbol.Kind {
		case cost.ColoredSymbol:
			colored[symbol.Color]++
		case cost.ColorlessSymbol:
			colored[mana.C]++
		case cost.GenericSymbol:
			generic += symbol.Generic
		case cost.VariableSymbol:
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

// hasTapCostOf reports whether the cost list has a tap additional cost.
func hasTapCostOf(additionalCosts []cost.Additional) bool {
	for _, addCost := range additionalCosts {
		if addCost.Kind == cost.AdditionalTap {
			return true
		}
	}
	return false
}

// costHasVariableMana reports whether the cost contains an X (variable) symbol.
func costHasVariableMana(manaCost *cost.Mana) bool {
	if manaCost == nil {
		return false
	}
	for _, symbol := range *manaCost {
		if symbol.Kind == cost.VariableSymbol {
			return true
		}
	}
	return false
}

// manaCostPtr returns a pointer to the mana cost value, or nil if it does not exist.
func manaCostPtr(manaCost opt.V[cost.Mana]) *cost.Mana {
	if !manaCost.Exists {
		return nil
	}
	return &manaCost.Val
}

// canTapForAbility reports whether the permanent can be tapped as an ability cost.
func canTapForAbility(s State, p *game.Permanent) bool {
	if p.Tapped {
		return false
	}
	return !s.PermanentHasType(p, types.Creature) || !p.SummoningSick
}

func canUntapForAbility(s State, p *game.Permanent) bool {
	if !p.Tapped {
		return false
	}
	return !s.PermanentHasType(p, types.Creature) || !p.SummoningSick
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
	return s.PermanentHasType(p, types.Creature)
}
