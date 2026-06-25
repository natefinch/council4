package payment

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

// applyPaymentPlan applies a prevalidated mana payment plan, performing the
// "pay the total cost" step of casting a spell or activating an ability
// (CR 601.2h): it activates the planned mana sources (CR 601.2g: mana abilities
// are activated before costs are paid), taps convoke/improvise permanents, exiles
// delve cards, spends mana from the pool, and pays life. Partial payments are not
// allowed (CR 601.2h), so the plan must be fully payable. The caller MUST have
// confirmed the plan against current state (paymentApplicationReady or
// paymentPlanStillValid) before calling, so every check here guards an
// invariant: a violation panics as an internal error rather than returning
// after partial mutation, guaranteeing a clean payment failure never leaves the
// game state half-applied.
func applyPaymentPlan(s State, playerID game.PlayerID, plan paymentPlan) {
	player, ok := s.Player(playerID)
	if !ok || !paymentPlanStillValid(s, player, plan) {
		panic("payment plan was not prevalidated before application")
	}
	for _, tap := range plan.manaTaps {
		if !activateManaForPayment(s, playerID, tap) {
			panic("payment plan became invalid while activating mana sources")
		}
	}
	for _, permanent := range plan.convokeTaps {
		if !canConvokeWith(s, playerID, permanent, nil) {
			panic("payment plan became invalid while tapping convoke creatures")
		}
		s.SetTapped(permanent, true)
	}
	for _, permanent := range plan.improviseTaps {
		if !canImproviseWith(s, playerID, permanent, nil) {
			panic("payment plan became invalid while tapping improvise artifacts")
		}
		s.SetTapped(permanent, true)
	}
	for _, cardID := range plan.delveExiles {
		if !player.Graveyard.Remove(cardID) {
			panic("payment plan became invalid while exiling delve cards")
		}
		player.Exile.Add(cardID)
		s.EmitZoneChange(game.Event{
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
		if player.Life < plan.lifePayment || !s.CanPayLife(playerID) {
			panic("payment plan became invalid while paying life")
		}
		s.LoseLife(playerID, plan.lifePayment)
	}
}

// activateManaForPayment activates one mana ability to produce mana for a payment.
// An activated mana ability does not go on the stack; it resolves immediately,
// adding its mana to the player's mana pool (CR 605.3b, CR 106.3-106.4). It taps
// or untaps and optionally sacrifices the source as the ability's cost, then adds
// the produced mana. It returns false if the planned source no longer matches the
// live game state.
func activateManaForPayment(s State, playerID game.PlayerID, activation manaTap) bool {
	permanent := activation.permanent
	if permanent.Tapped != activation.untap || s.EffectiveController(permanent) != playerID {
		return false
	}
	player, ok := s.Player(playerID)
	if !ok {
		return false
	}
	output, ok := permanentManaOutputForActivation(s, permanent, activation)
	if !ok ||
		output.color != activation.color ||
		output.amount != activation.amount ||
		output.snow != activation.snow ||
		output.untap != activation.untap ||
		output.sacrifice != activation.sacrifice ||
		output.abilityIndex != activation.abilityIndex ||
		output.timing != activation.timing {
		return false
	}
	if activation.untap {
		s.SetTapped(permanent, false)
	} else {
		s.SetTappedForMana(permanent)
	}
	if activation.abilityIndex >= 0 {
		s.RecordManaAbilityUse(permanent, activation.abilityIndex, activation.timing)
	}
	if activation.sacrifice && !s.SacrificePermanent(permanent) {
		return false
	}
	if output.snow {
		player.ManaPool.AddSnow(activation.color, activation.amount)
	} else {
		player.ManaPool.Add(activation.color, activation.amount)
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
		plan.manaTaps = append(plan.manaTaps, manaTap(source))
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
	plan.manaTaps = append(plan.manaTaps, manaTap(source))
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
		plan.manaTaps = append(plan.manaTaps, manaTap(source))
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
	plan.manaTaps = append(plan.manaTaps, manaTap(source))
	pool[mana.Unit{Color: source.color, Snow: source.snow}] += source.amount
	return spendAnySnowUnitFromSnapshot(plan, pool)
}

func payPhyrexianSymbol(player *game.Player, plan *paymentPlan, pool map[mana.Unit]int, sources map[mana.Color][]manaSource, symbol cost.Symbol, prefs *Preferences, canPayLife bool) bool {
	if prefs != nil && prefs.NextPhyrexianLifeChoice() {
		if !canPayLife || player.Life-plan.lifePayment < 2 {
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
	if !canPayLife || player.Life-plan.lifePayment < 2 {
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

// payPhyrexianGenericSymbol pays a "{N} or 2 life" generic Phyrexian symbol,
// emitted for the command-zone commander tax of a spell whose static lets the
// caster pay 2 life rather than each {2} of that tax (Liesa, Shroud of Dusk). It
// mirrors payPhyrexianSymbol but pays the symbol's generic mana rather than a
// color when not paying life.
func payPhyrexianGenericSymbol(player *game.Player, plan *paymentPlan, pool map[mana.Unit]int, sources map[mana.Color][]manaSource, symbol cost.Symbol, prefs *Preferences, canPayLife bool) bool {
	if prefs != nil && prefs.NextPhyrexianLifeChoice() {
		if !canPayLife || player.Life-plan.lifePayment < 2 {
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
		return payGenericSymbol(trialPlan, trialPool, trialSources, symbol, symbol.Generic, game.SymbolPaymentPhyrexianMana)
	}) {
		return true
	}
	if !canPayLife || player.Life-plan.lifePayment < 2 {
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
		source := leastFlexibleManaSource(sources[color])
		removeManaSource(sources, source)
		return source, true
	}
	return manaSource{}, false
}

func takeNonSnowManaSource(sources map[mana.Color][]manaSource, color mana.Color) (manaSource, bool) {
	var best manaSource
	found := false
	for _, source := range sources[color] {
		if source.snow || found && source.flexibility >= best.flexibility {
			continue
		}
		best = source
		found = true
	}
	if !found {
		return manaSource{}, false
	}
	removeManaSource(sources, best)
	return best, true
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
	var best manaSource
	found := false
	for _, color := range paymentColors {
		for _, source := range sources[color] {
			if !source.snow || found && source.flexibility >= best.flexibility {
				continue
			}
			best = source
			found = true
		}
	}
	if !found {
		return manaSource{}, false
	}
	removeManaSource(sources, best)
	return best, true
}

func removeManaSource(sources map[mana.Color][]manaSource, selected manaSource) {
	for color, candidates := range sources {
		sources[color] = slices.DeleteFunc(candidates, func(candidate manaSource) bool {
			return candidate.permanent.ObjectID == selected.permanent.ObjectID
		})
	}
}

func leastFlexibleManaSource(sources []manaSource) manaSource {
	best := sources[0]
	for _, source := range sources[1:] {
		if source.flexibility < best.flexibility {
			best = source
		}
	}
	return best
}
