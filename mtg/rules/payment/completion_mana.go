package payment

import (
	"fmt"
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

type manaRequirement struct {
	symbol    cost.Symbol
	phyrexian int
}

type manaPaymentBranch struct {
	color   mana.Color
	generic bool
	snow    bool
	amount  int
	life    int
	method  game.SymbolPaymentMethod
}

// completeMana solves the whole remaining cost, backtracking through alternate
// symbol payments and resource allocations. Failed resource states are memoized
// within this search so interchangeable sources do not cause factorial retries.
// No game is cloned and no speculative mana ability is executed.
func completeMana(s State, playerID game.PlayerID, manaCost *cost.Mana, x int, restrictions manaSourceRestrictions, ctx spendContext, prefs *Preferences, lifeReserved int, spell *spellCostOption, cardID id.ID, sourceZone zone.Type, partial Partial) (paymentPlan, bool) {
	if partial.joint && !partial.poolOnly {
		if state, ok := s.(ManaState); ok {
			return completeManaActivations(state, playerID, manaCost, x, restrictions, ctx, prefs, lifeReserved, spell, cardID, sourceZone, partial)
		}
	}
	player, ok := s.Player(playerID)
	if !ok || x < 0 || player.Life < lifeReserved {
		return paymentPlan{}, false
	}
	if spell != nil {
		if plan, ok := completeMana(s, playerID, manaCost, x, restrictions, ctx, prefs, lifeReserved, nil, cardID, sourceZone, partial); ok {
			return plan, true
		}
	}
	var symbols []cost.Symbol
	if manaCost != nil {
		symbols = spendPermissionSymbols(effectiveManaSymbols(s, playerID, *manaCost), ctx.spendAs)
	}
	var requirements []manaRequirement
	phyrexian := 0
	for _, symbol := range symbols {
		index := -1
		if symbol.Kind == cost.PhyrexianSymbol || symbol.Kind == cost.PhyrexianGenericSymbol {
			index = phyrexian
			phyrexian++
		}
		requirements = append(requirements, manaRequirement{symbol: symbol, phyrexian: index})
	}
	for index := range partial.Phyrexian {
		if index < 0 || index >= phyrexian {
			return paymentPlan{}, false
		}
	}
	// Leave flexible and generic requirements until after constrained mana.
	slices.SortStableFunc(requirements, func(a, b manaRequirement) int {
		return manaRequirementRank(a.symbol) - manaRequirementRank(b.symbol)
	})
	var sources map[mana.Color][]manaSource
	if !partial.poolOnly {
		sources = availableManaSources(s, playerID, restrictions)
	}
	var convoke, improvise []*game.Permanent
	var delve []id.ID
	if spell != nil && spell.card != nil {
		if spellHasCostKeyword(s, playerID, spell.card, cardID, sourceZone, game.Convoke) {
			convoke = convokeCandidates(s, playerID, restrictions.excluded)
		}
		if spellHasCostKeyword(s, playerID, spell.card, cardID, sourceZone, game.Improvise) {
			improvise = improviseCandidates(s, playerID, restrictions.excluded)
		}
		if spellHasCostKeyword(s, playerID, spell.card, cardID, sourceZone, game.Delve) {
			delve = delveCandidates(s, playerID, player.Graveyard.Size(), cardID, sourceZone, restrictions.reservedGraveyardCardIDs)
		}
	}
	used := make(map[id.ID]bool)
	exiled := make(map[id.ID]bool)
	failed := make(map[string]bool)
	var result paymentPlan
	var visit func(int, paymentPlan, map[mana.Unit]int) bool
	var spend func(int, manaPaymentBranch, int, paymentPlan, map[mana.Unit]int) bool
	visit = func(index int, plan paymentPlan, pool map[mana.Unit]int) bool {
		if index == len(requirements) {
			if !partial.acceptsMana(plan) {
				return false
			}
			result = plan
			return true
		}
		req := requirements[index]
		for _, branch := range manaBranches(req, x, prefs, partial.Phyrexian) {
			if branch.amount < 0 || plan.lifePayment+branch.life+lifeReserved > player.Life || branch.life > 0 && !s.CanPayLife(playerID) {
				continue
			}
			trial := clonePaymentPlan(plan)
			trial.lifePayment += branch.life
			trial.symbolPayments = append(trial.symbolPayments, game.SymbolPayment{
				Symbol: req.symbol, Method: branch.method, Color: branch.color,
				GenericAmount: func() int {
					if branch.generic {
						return branch.amount
					}
					return 0
				}(), LifePaid: branch.life, Snow: branch.snow,
			})
			if spend(index, branch, branch.amount, trial, pool) {
				return true
			}
		}
		return false
	}
	spend = func(index int, branch manaPaymentBranch, left int, plan paymentPlan, pool map[mana.Unit]int) bool {
		if left == 0 {
			return visit(index+1, plan, pool)
		}
		key := fmt.Sprint(index, branch, left, plan.lifePayment, pool, used, exiled)
		key += fmt.Sprint(permanentIDs(plan.convokeTaps), permanentIDs(plan.improviseTaps))
		if partial.Allocation != nil {
			key += fmt.Sprint(freezeMana(plan))
		}
		if failed[key] {
			return false
		}
		for _, unit := range paymentUnits {
			if pool[unit] <= 0 || !branch.accepts(unit) || !partial.allowsPoolUnit(unit, plan.poolSpend[unit]) {
				continue
			}
			next := clonePaymentPlan(plan)
			next.poolSpend[unit]++
			nextPool := cloneUnitCounts(pool)
			nextPool[unit]--
			if spend(index, branch, left-1, next, nextPool) {
				return true
			}
		}
		// Convoke may pay colored or generic mana; improvise and delve only
		// pay generic mana. These are payments, not mana abilities (CR 702).
		if !branch.snow && (branch.generic || branch.color != mana.C) {
			for _, permanent := range convoke {
				if used[permanent.ObjectID] || !partial.allowsMechanic(permanent.ObjectID, partial.Convoke) ||
					!branch.generic && !permanentHasManaColor(s, permanent, branch.color) {
					continue
				}
				next := clonePaymentPlan(plan)
				next.convokeTaps = append(slices.Clone(plan.convokeTaps), permanent)
				used[permanent.ObjectID] = true
				if spend(index, branch, left-1, next, pool) {
					return true
				}
				delete(used, permanent.ObjectID)
			}
		}
		if branch.generic {
			for _, permanent := range improvise {
				if used[permanent.ObjectID] || !partial.allowsMechanic(permanent.ObjectID, partial.Improvise) {
					continue
				}
				next := clonePaymentPlan(plan)
				next.improviseTaps = append(slices.Clone(plan.improviseTaps), permanent)
				used[permanent.ObjectID] = true
				if spend(index, branch, left-1, next, pool) {
					return true
				}
				delete(used, permanent.ObjectID)
			}
		}
		for _, color := range paymentColors {
			for _, source := range sources[color] {
				unit := mana.Unit{Color: source.color, Snow: source.snow, FromCreature: source.fromCreature}
				if used[source.permanent.ObjectID] || !branch.accepts(unit) || !partial.allowsManaSource(source) {
					continue
				}
				next := clonePaymentPlan(plan)
				next.manaTaps = append(next.manaTaps, manaTap(source))
				nextPool := cloneUnitCounts(pool)
				nextPool[unit] += source.amount
				used[source.permanent.ObjectID] = true
				if spend(index, branch, left, next, nextPool) {
					return true
				}
				delete(used, source.permanent.ObjectID)
			}
		}
		if branch.generic {
			for _, card := range delve {
				if exiled[card] || !partial.allowsMechanic(card, partial.Delve) {
					continue
				}
				next := clonePaymentPlan(plan)
				next.delveExiles = append(slices.Clone(plan.delveExiles), card)
				exiled[card] = true
				if spend(index, branch, left-1, next, pool) {
					return true
				}
				delete(exiled, card)
			}
		}
		failed[key] = true
		return false
	}
	ok = visit(0, paymentPlan{poolSpend: make(map[mana.Unit]int)}, snapshotPool(s, player, ctx))
	return result, ok
}

func permanentIDs(permanents []*game.Permanent) []id.ID {
	var result []id.ID
	for _, permanent := range permanents {
		result = append(result, permanent.ObjectID)
	}
	return result
}

func manaRequirementRank(symbol cost.Symbol) int {
	switch symbol.Kind {
	case cost.ColoredSymbol, cost.ColorlessSymbol:
		return 0
	case cost.SnowSymbol:
		return 1
	case cost.GenericSymbol, cost.VariableSymbol:
		return 3
	default:
		return 2
	}
}

func manaBranches(req manaRequirement, x int, prefs *Preferences, fixed map[int]bool) []manaPaymentBranch {
	symbol := req.symbol
	colored := func(c mana.Color, method game.SymbolPaymentMethod) manaPaymentBranch {
		return manaPaymentBranch{color: c, amount: 1, method: method}
	}
	generic := func(n int, method game.SymbolPaymentMethod) manaPaymentBranch {
		return manaPaymentBranch{generic: true, amount: n, method: method}
	}
	switch symbol.Kind {
	case cost.ColoredSymbol:
		return []manaPaymentBranch{colored(symbol.Color, game.SymbolPaymentMana)}
	case cost.ColorlessSymbol:
		return []manaPaymentBranch{colored(mana.C, game.SymbolPaymentMana)}
	case cost.SnowSymbol:
		return []manaPaymentBranch{{snow: true, amount: 1, method: game.SymbolPaymentSnow}}
	case cost.GenericSymbol:
		return []manaPaymentBranch{generic(symbol.Generic, game.SymbolPaymentGeneric)}
	case cost.VariableSymbol:
		return []manaPaymentBranch{generic(x, game.SymbolPaymentX)}
	case cost.HybridSymbol:
		return []manaPaymentBranch{colored(symbol.Color, game.SymbolPaymentHybridFirst), colored(symbol.AltColor, game.SymbolPaymentHybridSecond)}
	case cost.TwobridSymbol:
		return []manaPaymentBranch{colored(symbol.Color, game.SymbolPaymentMonoHybridColor), generic(2, game.SymbolPaymentMonoHybridGeneric)}
	case cost.PhyrexianSymbol, cost.PhyrexianGenericSymbol:
		payLife, manaOnly := prefs.phyrexianChoiceAt(req.phyrexian)
		if selected, set := fixed[req.phyrexian]; set {
			payLife, manaOnly = selected, !selected
		}
		life := manaPaymentBranch{life: 2, method: game.SymbolPaymentPhyrexianLife}
		if payLife {
			return []manaPaymentBranch{life}
		}
		branch := colored(symbol.Color, game.SymbolPaymentPhyrexianMana)
		if symbol.Kind == cost.PhyrexianGenericSymbol {
			branch = generic(symbol.Generic, game.SymbolPaymentPhyrexianMana)
		}
		if manaOnly {
			return []manaPaymentBranch{branch}
		}
		return []manaPaymentBranch{branch, life}
	default:
		return nil
	}
}

func (b manaPaymentBranch) accepts(unit mana.Unit) bool {
	return b.snow && unit.Snow || !b.snow && (b.generic || unit.Color == b.color)
}

func permanentHasManaColor(s State, permanent *game.Permanent, color mana.Color) bool {
	for _, c := range s.PermanentEffectiveColors(permanent) {
		if cost.ManaForColor(c) == color {
			return true
		}
	}
	return false
}
