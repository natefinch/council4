package payment

import (
	"maps"
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
)

// permanentCostSlot is a fixed-count permanent cost still to be selected:
// amount distinct permanents drawn from candidates. The sacrifice, tap and
// return costs of one payment each take their own permanents (CR 601.2h), so
// their selections must be planned together rather than greedily in cost
// order.
type permanentCostSlot struct {
	candidates []*game.Permanent
	amount     int
}

// additionalCostAmountIn is the number of objects or units an additional cost
// takes in this payment, honoring its announced X and dynamic amount.
func additionalCostAmountIn(s State, playerID game.PlayerID, additional cost.Additional, xValue int) int {
	if additional.AmountDynamic != cost.AdditionalDynamicAmountNone {
		return s.AdditionalDynamicAmountValue(playerID, additional.AmountDynamic)
	}
	return AdditionalCostAmountFor(additional, xValue)
}

// returnCostSource is the permanent a return cost's "another" exclusion refers
// to. Resolution payments carry only the source card ID, so it is resolved to
// the source's battlefield permanent when the cost excludes its source.
func returnCostSource(s State, additional cost.Additional, source *game.Permanent, sourceCardID id.ID) *game.Permanent {
	if source == nil && additional.ExcludeSource {
		return costSourcePermanentByCardID(s, sourceCardID)
	}
	return source
}

// permanentCostSlots returns a slot for every mandatory fixed-count sacrifice,
// tap and return cost among costs. Members of an unresolved "or" choice group
// are skipped because they may not be paid, and total-power tap costs are not
// fixed-count selections. tapReserved holds permanents no tap cost may choose,
// such as an ability's own {T} source.
func permanentCostSlots(s State, playerID game.PlayerID, costs []cost.Additional, xValue int, source *game.Permanent, sourceCardID id.ID, tapReserved map[id.ID]bool) []permanentCostSlot {
	var slots []permanentCostSlot
	for _, additional := range costs {
		if additional.ChoiceGroup != 0 {
			continue
		}
		amount := additionalCostAmountIn(s, playerID, additional, xValue)
		if amount <= 0 {
			continue
		}
		var costSource *game.Permanent
		var reserved map[id.ID]bool
		switch additional.Kind {
		case cost.AdditionalSacrifice:
			costSource = source
		case cost.AdditionalTapPermanents:
			if additional.TotalPowerAtLeast > 0 {
				continue
			}
			reserved = tapReserved
		case cost.AdditionalReturnToHand, cost.AdditionalReturnUnblockedAttacker:
			costSource = returnCostSource(s, additional, source, sourceCardID)
		default:
			continue
		}
		choice, ok := objectCostChoiceForCost(additional)
		if !ok {
			continue
		}
		slots = append(slots, permanentCostSlot{
			candidates: candidatePermanentsForObjectCost(s, playerID, choice, costSource, reserved),
			amount:     amount,
		})
	}
	return slots
}

// permanentSlotsAssignable reports whether every slot can take its amount of
// distinct unreserved permanents. It is a bipartite matching of slot units to
// permanents, found with augmenting paths, so it is exact and polynomial.
func permanentSlotsAssignable(slots []permanentCostSlot, reserved map[id.ID]bool) bool {
	var units []int
	for i, slot := range slots {
		for range slot.amount {
			units = append(units, i)
		}
	}
	owner := make(map[id.ID]int, len(units))
	var augment func(unit int, visited map[id.ID]bool) bool
	augment = func(unit int, visited map[id.ID]bool) bool {
		for _, permanent := range slots[units[unit]].candidates {
			objectID := permanent.ObjectID
			if reserved[objectID] || visited[objectID] {
				continue
			}
			visited[objectID] = true
			other, taken := owner[objectID]
			if !taken || augment(other, visited) {
				owner[objectID] = unit
				return true
			}
		}
		return false
	}
	for unit := range units {
		if !augment(unit, map[id.ID]bool{}) {
			return false
		}
	}
	return true
}

// assignablePermanents returns, in candidate order, each unreserved candidate
// that may be chosen next for a cost still needing need permanents: choosing
// it leaves the cost's other need-1 selections and every later slot payable
// with distinct permanents. With firstOnly it stops at the first one.
func assignablePermanents(candidates []*game.Permanent, need int, reserved map[id.ID]bool, later []permanentCostSlot, firstOnly bool) []*game.Permanent {
	if need <= 0 {
		return nil
	}
	reserved = maps.Clone(reserved)
	if reserved == nil {
		reserved = map[id.ID]bool{}
	}
	slots := append([]permanentCostSlot{{candidates: candidates, amount: need - 1}}, later...)
	var assignable []*game.Permanent
	for _, permanent := range candidates {
		if reserved[permanent.ObjectID] {
			continue
		}
		reserved[permanent.ObjectID] = true
		ok := permanentSlotsAssignable(slots, reserved)
		delete(reserved, permanent.ObjectID)
		if !ok {
			continue
		}
		assignable = append(assignable, permanent)
		if firstOnly {
			break
		}
	}
	return assignable
}

// chooseAssignablePermanents deterministically selects amount candidates for a
// cost, taking in battlefield order each candidate that keeps the rest of the
// payment's permanent costs (later) payable with distinct permanents, instead
// of committing to the first candidates a later cost may need. It returns
// fewer than amount when no such assignment exists.
func chooseAssignablePermanents(candidates []*game.Permanent, amount int, reserved map[id.ID]bool, later []permanentCostSlot) []*game.Permanent {
	if len(later) == 0 {
		var chosen []*game.Permanent
		for _, permanent := range candidates {
			if len(chosen) == amount {
				break
			}
			if !reserved[permanent.ObjectID] {
				chosen = append(chosen, permanent)
			}
		}
		return chosen
	}
	reserved = maps.Clone(reserved)
	if reserved == nil {
		reserved = map[id.ID]bool{}
	}
	var chosen []*game.Permanent
	for len(chosen) < amount {
		next := assignablePermanents(candidates, amount-len(chosen), reserved, later, true)
		if len(next) == 0 {
			return chosen
		}
		chosen = append(chosen, next[0])
		reserved[next[0].ObjectID] = true
	}
	return chosen
}

// AssignablePermanentsForCost returns the candidates a player may choose next
// for costs[index], a sacrifice, tap or return cost that still needs need
// permanents. Each returned candidate leaves the cost's other selections and
// every later mandatory sacrifice, tap and return cost in costs payable with
// distinct permanents not in reservedIDs (CR 601.2h). tapReservedIDs are
// permanents no later tap cost may choose. shared reports whether a later such
// cost may choose one of the candidates, so that a multi-permanent choice must
// be made one permanent at a time to stay compatible.
func AssignablePermanentsForCost(s State, playerID game.PlayerID, costs []cost.Additional, index, xValue int, candidates []*game.Permanent, need int, sourceCardID id.ID, reservedIDs, tapReservedIDs []id.ID) (assignable []*game.Permanent, shared bool) {
	if index < 0 || index >= len(costs) {
		return nil, false
	}
	later := permanentCostSlots(s, playerID, costs[index+1:], xValue, nil, sourceCardID, reservedCardIDSet(tapReservedIDs))
	candidateIDs := reservedPermanentIDs(candidates)
	for _, slot := range later {
		if slices.ContainsFunc(slot.candidates, func(p *game.Permanent) bool { return candidateIDs[p.ObjectID] }) {
			shared = true
			break
		}
	}
	return assignablePermanents(candidates, need, reservedCardIDSet(reservedIDs), later, false), shared
}
