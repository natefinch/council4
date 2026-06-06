package payment

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

func applyPaymentPlan(s State, playerID game.PlayerID, plan paymentPlan) bool {
	player, ok := s.Player(playerID)
	if !ok || !paymentPlanStillValid(s, player, plan) {
		return false
	}
	for _, tap := range plan.manaTaps {
		if !tapForMana(s, tap.permanent, tap.color, tap.amount, tap.snow) {
			panic("payment plan became invalid while tapping mana sources")
		}
	}
	for _, permanent := range plan.convokeTaps {
		if !canConvokeWith(s, playerID, permanent, nil) {
			panic("payment plan became invalid while tapping convoke creatures")
		}
		s.SetTapped(permanent, true)
	}
	for _, cardID := range plan.delveExiles {
		if !player.Graveyard.Remove(cardID) {
			panic("payment plan became invalid while exiling delve cards")
		}
		player.Exile.Add(cardID)
		s.EmitZoneChange(game.GameEvent{
			Player:   playerID,
			CardID:   cardID,
			FromZone: zone.Graveyard,
			ToZone:   zone.Exile,
		})
	}
	for _, color := range paymentColors {
		for _, snow := range []bool{false, true} {
			unit := mana.Unit{Color: color, Snow: snow}
			amount := plan.poolSpend[unit]
			if amount > 0 && !player.ManaPool.SpendMatching(amount, func(candidate mana.Unit) bool { return candidate == unit }) {
				panic("payment plan became invalid while spending mana")
			}
		}
	}
	if plan.lifePayment > 0 {
		if player.Life < plan.lifePayment {
			return false
		}
		s.LoseLife(playerID, plan.lifePayment)
	}
	return true
}

// tapForMana taps a permanent to produce mana and adds it to the controller's pool.
func tapForMana(s State, permanent *game.Permanent, color mana.Color, amount int, snow bool) bool {
	if permanent.Tapped {
		return false
	}
	controllerID := s.EffectiveController(permanent)
	player, ok := s.Player(controllerID)
	if !ok {
		return false
	}
	output, ok := permanentManaOutput(s, permanent)
	if !ok || output.color != color || output.amount != amount || output.snow != snow {
		return false
	}
	s.SetTapped(permanent, true)
	if output.snow {
		player.ManaPool.AddSnow(color, amount)
	} else {
		player.ManaPool.Add(color, amount)
	}
	return true
}

func payColoredSymbol(plan *paymentPlan, pool map[mana.Unit]int, sources map[mana.Color][]manaSource, symbol cost.Symbol, color mana.Color, method game.SymbolPaymentMethod) bool {
	if !paySpecificMana(plan, pool, sources, color) {
		return false
	}
	plan.symbolPayments = append(plan.symbolPayments, game.SymbolPayment{
		Symbol: symbol,
		Method: method,
		Color:  color,
	})
	return true
}

func paySpecificMana(plan *paymentPlan, pool map[mana.Unit]int, sources map[mana.Color][]manaSource, color mana.Color) bool {
	if spendUnitFromSnapshot(plan, pool, mana.Unit{Color: color}, 1) {
		return true
	}
	if source, ok := takeNonSnowManaSource(sources, color); ok {
		plan.manaTaps = append(plan.manaTaps, manaTap{permanent: source.permanent, color: source.color, amount: source.amount, snow: source.snow})
		pool[mana.Unit{Color: source.color, Snow: source.snow}] += source.amount
		return paySpecificMana(plan, pool, sources, color)
	}
	if spendUnitFromSnapshot(plan, pool, mana.Unit{Color: color, Snow: true}, 1) {
		return true
	}
	source, ok := takeManaSource(sources, color)
	if !ok {
		return false
	}
	plan.manaTaps = append(plan.manaTaps, manaTap{permanent: source.permanent, color: source.color, amount: source.amount, snow: source.snow})
	pool[mana.Unit{Color: source.color, Snow: source.snow}] += source.amount
	return paySpecificMana(plan, pool, sources, color)
}

func payGenericSymbol(plan *paymentPlan, pool map[mana.Unit]int, sources map[mana.Color][]manaSource, symbol cost.Symbol, amount int, method game.SymbolPaymentMethod) bool {
	if amount < 0 {
		return false
	}
	if !payGenericMana(plan, pool, sources, amount) {
		return false
	}
	plan.symbolPayments = append(plan.symbolPayments, game.SymbolPayment{
		Symbol:        symbol,
		Method:        method,
		GenericAmount: amount,
	})
	return true
}

func payGenericMana(plan *paymentPlan, pool map[mana.Unit]int, sources map[mana.Color][]manaSource, amount int) bool {
	remaining := amount
	for remaining > 0 {
		if spendAnyUnitFromSnapshot(plan, pool) {
			remaining--
			continue
		}
		source, ok := takeAnyManaSource(sources)
		if !ok {
			return false
		}
		plan.manaTaps = append(plan.manaTaps, manaTap{permanent: source.permanent, color: source.color, amount: source.amount, snow: source.snow})
		pool[mana.Unit{Color: source.color, Snow: source.snow}] += source.amount
	}
	return true
}

func payHybridSymbol(plan *paymentPlan, pool map[mana.Unit]int, sources map[mana.Color][]manaSource, symbol cost.Symbol) bool {
	if trySymbolPayment(plan, pool, sources, func(trialPlan *paymentPlan, trialPool map[mana.Unit]int, trialSources map[mana.Color][]manaSource) bool {
		return payColoredSymbol(trialPlan, trialPool, trialSources, symbol, symbol.Color, game.SymbolPaymentHybridFirst)
	}) {
		return true
	}
	return trySymbolPayment(plan, pool, sources, func(trialPlan *paymentPlan, trialPool map[mana.Unit]int, trialSources map[mana.Color][]manaSource) bool {
		return payColoredSymbol(trialPlan, trialPool, trialSources, symbol, symbol.AltColor, game.SymbolPaymentHybridSecond)
	})
}

func payMonoHybridSymbol(plan *paymentPlan, pool map[mana.Unit]int, sources map[mana.Color][]manaSource, symbol cost.Symbol) bool {
	if trySymbolPayment(plan, pool, sources, func(trialPlan *paymentPlan, trialPool map[mana.Unit]int, trialSources map[mana.Color][]manaSource) bool {
		return payColoredSymbol(trialPlan, trialPool, trialSources, symbol, symbol.Color, game.SymbolPaymentMonoHybridColor)
	}) {
		return true
	}
	return trySymbolPayment(plan, pool, sources, func(trialPlan *paymentPlan, trialPool map[mana.Unit]int, trialSources map[mana.Color][]manaSource) bool {
		return payGenericSymbol(trialPlan, trialPool, trialSources, symbol, 2, game.SymbolPaymentMonoHybridGeneric)
	})
}

func paySnowSymbol(plan *paymentPlan, pool map[mana.Unit]int, sources map[mana.Color][]manaSource, symbol cost.Symbol) bool {
	if !paySnowMana(plan, pool, sources) {
		return false
	}
	plan.symbolPayments = append(plan.symbolPayments, game.SymbolPayment{
		Symbol: symbol,
		Method: game.SymbolPaymentSnow,
		Snow:   true,
	})
	return true
}

func paySnowMana(plan *paymentPlan, pool map[mana.Unit]int, sources map[mana.Color][]manaSource) bool {
	if spendAnySnowUnitFromSnapshot(plan, pool) {
		return true
	}
	source, ok := takeAnySnowManaSource(sources)
	if !ok {
		return false
	}
	plan.manaTaps = append(plan.manaTaps, manaTap{permanent: source.permanent, color: source.color, amount: source.amount, snow: source.snow})
	pool[mana.Unit{Color: source.color, Snow: source.snow}] += source.amount
	return spendAnySnowUnitFromSnapshot(plan, pool)
}

func payPhyrexianSymbol(player *game.Player, plan *paymentPlan, pool map[mana.Unit]int, sources map[mana.Color][]manaSource, symbol cost.Symbol, prefs *Preferences) bool {
	if prefs != nil && prefs.NextPhyrexianLifeChoice() {
		if player.Life-plan.lifePayment < 2 {
			return false
		}
		plan.lifePayment += 2
		plan.symbolPayments = append(plan.symbolPayments, game.SymbolPayment{
			Symbol:   symbol,
			Method:   game.SymbolPaymentPhyrexianLife,
			LifePaid: 2,
		})
		return true
	}
	if trySymbolPayment(plan, pool, sources, func(trialPlan *paymentPlan, trialPool map[mana.Unit]int, trialSources map[mana.Color][]manaSource) bool {
		return payColoredSymbol(trialPlan, trialPool, trialSources, symbol, symbol.Color, game.SymbolPaymentPhyrexianMana)
	}) {
		return true
	}
	if player.Life-plan.lifePayment < 2 {
		return false
	}
	plan.lifePayment += 2
	plan.symbolPayments = append(plan.symbolPayments, game.SymbolPayment{
		Symbol:   symbol,
		Method:   game.SymbolPaymentPhyrexianLife,
		LifePaid: 2,
	})
	return true
}

func spendUnitFromSnapshot(plan *paymentPlan, pool map[mana.Unit]int, unit mana.Unit, amount int) bool {
	if amount <= 0 {
		return true
	}
	if pool[unit] < amount {
		return false
	}
	pool[unit] -= amount
	plan.poolSpend[unit] += amount
	return true
}

func spendAnyUnitFromSnapshot(plan *paymentPlan, pool map[mana.Unit]int) bool {
	for _, unit := range paymentUnitOrder() {
		if spendUnitFromSnapshot(plan, pool, unit, 1) {
			return true
		}
	}
	return false
}

func spendAnySnowUnitFromSnapshot(plan *paymentPlan, pool map[mana.Unit]int) bool {
	for _, unit := range paymentUnitOrder() {
		if !unit.Snow {
			continue
		}
		if spendUnitFromSnapshot(plan, pool, unit, 1) {
			return true
		}
	}
	return false
}

func paymentUnitOrder() []mana.Unit {
	var units []mana.Unit
	for _, color := range paymentColors {
		units = append(units, mana.Unit{Color: color})
	}
	for _, color := range paymentColors {
		units = append(units, mana.Unit{Color: color, Snow: true})
	}
	return units
}

func trySymbolPayment(plan *paymentPlan, pool map[mana.Unit]int, sources map[mana.Color][]manaSource, pay func(*paymentPlan, map[mana.Unit]int, map[mana.Color][]manaSource) bool) bool {
	trialPlan := clonePaymentPlan(*plan)
	trialPool := cloneUnitCounts(pool)
	trialSources := cloneManaSources(sources)
	if !pay(&trialPlan, trialPool, trialSources) {
		return false
	}
	*plan = trialPlan
	replaceUnitCounts(pool, trialPool)
	replaceManaSources(sources, trialSources)
	return true
}

func takeManaSource(sources map[mana.Color][]manaSource, color mana.Color) (manaSource, bool) {
	if source, ok := takeNonSnowManaSource(sources, color); ok {
		return source, true
	}
	if len(sources[color]) > 0 {
		source := sources[color][0]
		sources[color] = sources[color][1:]
		return source, true
	}
	return manaSource{}, false
}

func takeNonSnowManaSource(sources map[mana.Color][]manaSource, color mana.Color) (manaSource, bool) {
	for i, source := range sources[color] {
		if source.snow {
			continue
		}
		sources[color] = append(sources[color][:i], sources[color][i+1:]...)
		return source, true
	}
	return manaSource{}, false
}

func takeAnyManaSource(sources map[mana.Color][]manaSource) (manaSource, bool) {
	for _, color := range paymentColors {
		if source, ok := takeManaSource(sources, color); ok {
			return source, true
		}
	}
	return manaSource{}, false
}

func takeAnySnowManaSource(sources map[mana.Color][]manaSource) (manaSource, bool) {
	for _, color := range paymentColors {
		for i, source := range sources[color] {
			if !source.snow {
				continue
			}
			sources[color] = append(sources[color][:i], sources[color][i+1:]...)
			return source, true
		}
	}
	return manaSource{}, false
}
