package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// poolUnitsSnapshot records a player's per-unit mana pool counts. The rules
// engine captures it immediately before paying a cost so it can measure, after
// the payment, how much pre-existing mana of each exact unit (color and snow
// provenance) was spent and thereby which tagged mana-spend rider units (Path of
// Ancestry) were consumed. mana.Pool.Units already returns an independent copy.
func poolUnitsSnapshot(player *game.Player) map[mana.Unit]int {
	return player.ManaPool.Units()
}

// manaSpendRiderSnapshot captures a player's per-unit pool counts before a
// payment, but only when the player currently holds mana-spend riders. hasRiders
// reports whether any rider was present, so callers skip rider processing (and
// the snapshot allocation) entirely for the common case of a player with no
// tagged mana.
func manaSpendRiderSnapshot(g *game.Game, playerID game.PlayerID) (before map[mana.Unit]int, hasRiders bool) {
	player, ok := playerByID(g, playerID)
	if !ok || len(player.ManaRiders) == 0 {
		return nil, false
	}
	return poolUnitsSnapshot(player), true
}

// processManaSpendRiders consumes the tagged mana-spend rider units whose mana a
// just-completed payment spent, firing each consumed rider whose condition the
// payment satisfied. It tracks provenance on individual mana units: each rider
// instance is an independent unit of tagged mana that is consumed and fired or
// dropped on the exact payment that spends it, never reattached to a later unit
// of the same color.
//
// before is the per-unit pool snapshot captured immediately before the payment
// and spent is the exact per-unit pool mana the payment consumed (reported by
// the payment planner). The pre-existing pool mana consumed for a unit is
// min(before[unit], spent[unit]): the planner spends existing pool mana before
// tapping new sources, so taking the minimum keeps the accounting exact even
// when a source produces extra mana of that unit mid-payment, which a gross
// before/after pool delta would otherwise mask (the missed-spend case).
//
// It models the fungibility of identical mana units (CR 106.6, 106.12): on a
// payment that satisfies a unit's rider condition the player keeps the most
// value by spending tagged mana, so tagged units are consumed first and each
// consumed rider fires; on any other payment the player keeps the most value by
// preserving tagged mana for a later qualifying spell, so plain units are
// consumed first and only forced tagged consumption (when plain mana of that
// unit ran out) removes riders without firing. Running on every payment path,
// not lazily reconciling against the pool, is what prevents a stale rider from
// reattaching to later same-color mana (the false-trigger case).
//
// qualifies reports, for a rider instance, whether this payment satisfied its
// condition. fire resolves a fired rider (putting its effect on the stack). It
// is a no-op when the player holds no riders, so cost payment carries no
// overhead for ordinary mana, and a free function because it needs no engine
// state.
func processManaSpendRiders(
	player *game.Player,
	before map[mana.Unit]int,
	spent map[mana.Unit]int,
	qualifies func(rider game.ManaRiderInstance) bool,
	fire func(rider game.ManaRiderInstance),
) {
	if len(player.ManaRiders) == 0 {
		return
	}
	riderCount := make(map[mana.Unit]int, len(player.ManaRiders))
	unitQualifies := make(map[mana.Unit]bool, len(player.ManaRiders))
	for _, instance := range player.ManaRiders {
		if _, seen := riderCount[instance.Unit]; !seen {
			unitQualifies[instance.Unit] = qualifies(instance)
		}
		riderCount[instance.Unit]++
	}
	consume := make(map[mana.Unit]int, len(riderCount))
	for unit, riders := range riderCount {
		// The planner spends existing pool mana before tapping new sources, so
		// the pre-existing pool consumed (which alone can include tagged mana) is
		// the lesser of what was in the pool and what the payment drew from it.
		preExistingSpent := min(before[unit], spent[unit])
		if preExistingSpent <= 0 {
			continue
		}
		plain := max(before[unit]-riders, 0)
		var take int
		if unitQualifies[unit] {
			take = min(riders, preExistingSpent)
		} else {
			take = preExistingSpent - plain
		}
		if take > 0 {
			consume[unit] = min(take, riders)
		}
	}
	if len(consume) == 0 {
		return
	}
	remaining := player.ManaRiders[:0]
	for _, instance := range player.ManaRiders {
		if consume[instance.Unit] > 0 {
			consume[instance.Unit]--
			if qualifies(instance) {
				fire(instance)
			}
			continue
		}
		remaining = append(remaining, instance)
	}
	if len(remaining) == 0 {
		player.ManaRiders = nil
		return
	}
	player.ManaRiders = remaining
}

// resolveSpellCastManaSpendRiders consumes the casting player's tagged mana that
// was spent paying for a just-cast spell, firing each rider whose condition the
// spell satisfies. before is the per-unit pool snapshot captured immediately
// before the spell's costs were paid and spent is the exact per-unit pool mana
// the payment consumed. It is a no-op when the player holds no riders, so
// ordinary spell casts carry no overhead.
func resolveSpellCastManaSpendRiders(
	g *game.Game,
	playerID game.PlayerID,
	before map[mana.Unit]int,
	spent map[mana.Unit]int,
	spellDef *game.CardDef,
) {
	player, ok := playerByID(g, playerID)
	if !ok || len(player.ManaRiders) == 0 {
		return
	}
	qualifies := func(rider game.ManaRiderInstance) bool {
		if rider.Rider.Condition != game.ManaSpendCastCommanderCreatureType {
			return false
		}
		return spellSatisfiesCommanderCreatureTypeRider(g, rider.Controller, spellDef)
	}
	processManaSpendRiders(player, before, spent, qualifies, func(rider game.ManaRiderInstance) {
		fireManaSpendRider(g, rider)
	})
}

// consumeManaSpendRidersForPayment drops the tagged mana-spend rider units whose
// mana a just-completed non-spell payment (an activated ability, a ward or other
// additional cost, and similar) spent. Such a payment never satisfies a rider's
// condition, so no rider fires; consuming the units keeps rider provenance exact
// so later same-color mana cannot inherit a stale rider. It is a no-op when the
// player holds no riders.
func consumeManaSpendRidersForPayment(
	g *game.Game,
	playerID game.PlayerID,
	before map[mana.Unit]int,
	spent map[mana.Unit]int,
) {
	player, ok := playerByID(g, playerID)
	if !ok || len(player.ManaRiders) == 0 {
		return
	}
	processManaSpendRiders(
		player,
		before,
		spent,
		func(game.ManaRiderInstance) bool { return false },
		func(game.ManaRiderInstance) {},
	)
}

// fireManaSpendRider puts a fired rider's effect on the stack as a triggered
// ability controlled by the mana's controller (CR 603.2c). It mirrors a
// battlefield-sourced ability: the stack object's source is the producing
// permanent, so the rider resolves with the controller even if that permanent
// has since left the battlefield.
func fireManaSpendRider(g *game.Game, instance game.ManaRiderInstance) {
	ability := instance.Rider.Ability()
	g.Stack.Push(&game.StackObject{
		ID:            g.IDGen.Next(),
		Kind:          game.StackTriggeredAbility,
		SourceID:      instance.SourceObjectID,
		SourceCardID:  instance.SourceID,
		Controller:    instance.Controller,
		InlineTrigger: &ability,
	})
}

// spellSatisfiesCommanderCreatureTypeRider reports whether spellDef is a creature
// spell that shares a creature type with the rider controller's commander (Path
// of Ancestry's spend condition). It resolves the commander's current
// characteristics from its current zone and face rather than its printed card
// definition, and fails closed when the controller has no single modeled
// commander or the commander's characteristics cannot be faithfully resolved, so
// partner or Background commanders (not modeled as one commander instance) never
// spuriously satisfy the condition.
func spellSatisfiesCommanderCreatureTypeRider(
	g *game.Game,
	controller game.PlayerID,
	spellDef *game.CardDef,
) bool {
	if spellDef == nil || !spellDef.HasType(types.Creature) {
		return false
	}
	player, ok := playerByID(g, controller)
	if !ok || player.CommanderInstanceID == 0 {
		return false
	}
	commanderHasSubtype, ok := commanderCreatureSubtypeMatcher(g, player)
	if !ok {
		return false
	}
	for _, subtype := range spellDef.Subtypes {
		if !types.KnownSubtypeForType(types.Creature, subtype) {
			continue
		}
		if commanderHasSubtype(subtype) {
			return true
		}
	}
	return false
}

// commanderCreatureSubtypeMatcher returns a predicate reporting whether the
// player's commander currently has a given subtype, resolving the commander's
// current characteristics from its current zone, object, and face rather than
// its printed card definition. It fails closed (returns false) when the
// commander instance cannot be resolved.
//
// When the commander is currently a battlefield permanent its effective subtypes
// reflect transform, face-down, and type-changing effects, so those are the
// current characteristics. A face-down commander therefore has no creature
// subtypes and correctly matches nothing. Anywhere else (command zone, hand,
// graveyard, exile, library, or on the stack) no continuous effects alter its
// types and a double-faced or modal card uses its front face by default
// (CR 711.2, 712.4a), which the card definition represents.
func commanderCreatureSubtypeMatcher(g *game.Game, player *game.Player) (func(types.Sub) bool, bool) {
	commander, ok := g.GetCardInstance(player.CommanderInstanceID)
	if !ok || commander.Def == nil {
		return nil, false
	}
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == player.CommanderInstanceID {
			return func(subtype types.Sub) bool {
				return permanentHasSubtype(g, permanent, subtype)
			}, true
		}
	}
	return func(subtype types.Sub) bool {
		return commander.Def.HasSubtype(subtype)
	}, true
}
