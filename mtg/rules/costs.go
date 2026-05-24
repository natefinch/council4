package rules

import (
	"strings"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
)

var paymentColors = []mana.Color{
	mana.White,
	mana.Blue,
	mana.Black,
	mana.Red,
	mana.Green,
	mana.Colorless,
}

func canPayCost(g *game.Game, playerID game.PlayerID, cost *mana.Cost) bool {
	_, ok := buildPaymentPlan(g, playerID, cost)
	return ok
}

func payCost(g *game.Game, playerID game.PlayerID, cost *mana.Cost) bool {
	plan, ok := buildPaymentPlan(g, playerID, cost)
	if !ok {
		return false
	}
	return applyPaymentPlan(g, playerID, plan)
}

type paymentPlan struct {
	poolSpend map[mana.Color]int
	landTaps  []landTap
}

type landTap struct {
	permanent *game.Permanent
	color     mana.Color
}

func buildPaymentPlan(g *game.Game, playerID game.PlayerID, cost *mana.Cost) (paymentPlan, bool) {
	plan := paymentPlan{poolSpend: make(map[mana.Color]int)}
	player := playerForCostPayment(g, playerID)
	if player == nil {
		return plan, false
	}
	colored, generic, ok := costRequirements(cost)
	if !ok {
		return plan, false
	}

	pool := snapshotPool(player)
	lands := availableBasicLandMana(g, playerID)

	for _, color := range paymentColors {
		need := colored[color]
		if need == 0 {
			continue
		}
		spent := spendSnapshot(pool, color, need)
		need -= spent
		if spent > 0 {
			plan.poolSpend[color] += spent
		}
		for need > 0 {
			land := takeLand(lands, color)
			if land == nil {
				return plan, false
			}
			plan.landTaps = append(plan.landTaps, landTap{permanent: land, color: color})
			pool[color]++
			spendSnapshot(pool, color, 1)
			plan.poolSpend[color]++
			need--
		}
	}

	remainingGeneric := generic
	for _, color := range paymentColors {
		if remainingGeneric == 0 {
			break
		}
		spent := spendSnapshot(pool, color, remainingGeneric)
		remainingGeneric -= spent
		if spent > 0 {
			plan.poolSpend[color] += spent
		}
	}

	for remainingGeneric > 0 {
		land, color := takeAnyLand(lands)
		if land == nil {
			return plan, false
		}
		plan.landTaps = append(plan.landTaps, landTap{permanent: land, color: color})
		pool[color]++
		spendSnapshot(pool, color, 1)
		plan.poolSpend[color]++
		remainingGeneric--
	}

	return plan, true
}

func applyPaymentPlan(g *game.Game, playerID game.PlayerID, plan paymentPlan) bool {
	player := playerForCostPayment(g, playerID)
	if player == nil || !paymentPlanStillValid(g, player, plan) {
		return false
	}
	for _, tap := range plan.landTaps {
		if !tapLandForMana(g, tap.permanent, tap.color) {
			panic("payment plan became invalid while tapping lands")
		}
	}
	for _, color := range paymentColors {
		amount := plan.poolSpend[color]
		if amount > 0 && !player.ManaPool.Spend(color, amount) {
			panic("payment plan became invalid while spending mana")
		}
	}
	return true
}

func paymentPlanStillValid(g *game.Game, player *game.Player, plan paymentPlan) bool {
	tappedMana := make(map[mana.Color]int)
	for _, tap := range plan.landTaps {
		if tap.permanent == nil || tap.permanent.Tapped || tap.permanent.Controller != player.ID {
			return false
		}
		color, ok := basicLandManaColor(g, tap.permanent)
		if !ok || color != tap.color {
			return false
		}
		tappedMana[tap.color]++
	}
	for _, color := range paymentColors {
		if player.ManaPool.Amount(color)+tappedMana[color] < plan.poolSpend[color] {
			return false
		}
	}
	return true
}

func costRequirements(cost *mana.Cost) (map[mana.Color]int, int, bool) {
	colored := make(map[mana.Color]int)
	if cost == nil {
		return colored, 0, true
	}

	generic := 0
	for _, symbol := range *cost {
		switch symbol.Kind {
		case mana.ColoredSymbol:
			colored[symbol.Color]++
		case mana.GenericSymbol:
			generic += symbol.Generic
		default:
			return nil, 0, false
		}
	}
	return colored, generic, true
}

func snapshotPool(player *game.Player) map[mana.Color]int {
	pool := make(map[mana.Color]int)
	for _, color := range paymentColors {
		pool[color] = player.ManaPool.Amount(color)
	}
	return pool
}

func spendSnapshot(pool map[mana.Color]int, color mana.Color, amount int) int {
	if amount <= 0 {
		return 0
	}
	spent := min(pool[color], amount)
	pool[color] -= spent
	return spent
}

func availableBasicLandMana(g *game.Game, playerID game.PlayerID) map[mana.Color][]*game.Permanent {
	available := make(map[mana.Color][]*game.Permanent)
	if g == nil {
		return available
	}
	for _, permanent := range g.Battlefield {
		if permanent == nil || permanent.Controller != playerID || permanent.Tapped {
			continue
		}
		color, ok := basicLandManaColor(g, permanent)
		if !ok {
			continue
		}
		available[color] = append(available[color], permanent)
	}
	return available
}

func tapLandForMana(g *game.Game, permanent *game.Permanent, color mana.Color) bool {
	if g == nil || permanent == nil || permanent.Tapped {
		return false
	}
	player := playerForCostPayment(g, permanent.Controller)
	if player == nil {
		return false
	}
	landColor, ok := basicLandManaColor(g, permanent)
	if !ok || landColor != color {
		return false
	}
	permanent.Tapped = true
	player.ManaPool.Add(color, 1)
	return true
}

func basicLandManaColor(g *game.Game, permanent *game.Permanent) (mana.Color, bool) {
	card := g.GetCardInstance(permanent.CardInstanceID)
	if card == nil || card.Def == nil || !card.Def.HasType(game.TypeLand) {
		return 0, false
	}
	for _, landType := range basicLandTypes {
		if card.Def.HasSubtype(landType.subtype) || strings.EqualFold(card.Def.Name, landType.subtype) {
			return landType.color, true
		}
	}
	return 0, false
}

var basicLandTypes = []struct {
	subtype string
	color   mana.Color
}{
	{subtype: "Plains", color: mana.White},
	{subtype: "Island", color: mana.Blue},
	{subtype: "Swamp", color: mana.Black},
	{subtype: "Mountain", color: mana.Red},
	{subtype: "Forest", color: mana.Green},
}

func takeLand(lands map[mana.Color][]*game.Permanent, color mana.Color) *game.Permanent {
	if len(lands[color]) == 0 {
		return nil
	}
	land := lands[color][0]
	lands[color] = lands[color][1:]
	return land
}

func takeAnyLand(lands map[mana.Color][]*game.Permanent) (*game.Permanent, mana.Color) {
	for _, color := range paymentColors {
		if land := takeLand(lands, color); land != nil {
			return land, color
		}
	}
	return nil, 0
}

func playerForCostPayment(g *game.Game, playerID game.PlayerID) *game.Player {
	if g == nil || playerID < 0 || int(playerID) >= len(g.Players) {
		return nil
	}
	player := g.Players[playerID]
	if player == nil || player.Eliminated || g.TurnOrder.IsEliminated(playerID) {
		return nil
	}
	return player
}
