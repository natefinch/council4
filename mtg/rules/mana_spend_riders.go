package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// coloredPoolSnapshot records a player's per-color mana totals. The rules engine
// captures it immediately before paying a cost so it can measure, after the
// payment, how much mana of each color was spent and thereby which tagged
// mana-spend rider units (Path of Ancestry) were consumed.
func coloredPoolSnapshot(player *game.Player) map[mana.Color]int {
	snapshot := make(map[mana.Color]int, len(allManaColors))
	for _, color := range allManaColors {
		amount := player.ManaPool.Amount(color)
		if amount > 0 {
			snapshot[color] = amount
		}
	}
	return snapshot
}

var allManaColors = []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G, mana.C}

// prepareManaSpendRiderSnapshot reconciles a player's tagged mana against their
// current pool, then snapshots the colored pool ahead of a payment. The
// reconciliation drops rider instances whose mana was already spent on a
// non-spell (an activated ability, a ward cost, and similar), which do not fire
// the rider; this keeps the tagged-mana bookkeeping consistent with the pool so
// the upcoming spell payment fires riders only for mana actually spent on it.
func prepareManaSpendRiderSnapshot(player *game.Player) map[mana.Color]int {
	reconcileManaRidersToPool(player)
	return coloredPoolSnapshot(player)
}

// reconcileManaRidersToPool drops rider instances whose tagged mana is no longer
// in the pool, keeping at most pool.Amount(color) riders of each color. Riders
// are interchangeable within a color, so the trailing (most recently produced)
// excess instances are dropped.
func reconcileManaRidersToPool(player *game.Player) {
	if len(player.ManaRiders) == 0 {
		return
	}
	keptByColor := make(map[mana.Color]int, len(player.ManaRiders))
	kept := player.ManaRiders[:0]
	for _, instance := range player.ManaRiders {
		if keptByColor[instance.Color] < player.ManaPool.Amount(instance.Color) {
			keptByColor[instance.Color]++
			kept = append(kept, instance)
		}
	}
	if len(kept) == 0 {
		player.ManaRiders = nil
		return
	}
	player.ManaRiders = kept
}

// resolveManaSpendRiders reconciles a player's tagged mana against the colored
// pool mana actually spent during a just-completed payment, removing the rider
// instances whose mana was spent. before is the colored pool snapshot captured
// immediately before the payment and spent is the exact per-color pool mana the
// payment consumed (reported by the payment planner). The pre-existing pool mana
// consumed for a color is min(before[color], spent[color]), because the planner
// spends existing pool mana before tapping new sources; taking the minimum keeps
// the accounting exact even when a source produces extra mana of that color
// mid-payment, which a gross before/after pool delta would otherwise mask.
//
// It models the fungibility of equal-color mana (CR 106.6, 106.12): on a payment
// that satisfies a rider's condition the player keeps the most value by spending
// tagged mana, so tagged units of a color are consumed first and each consumed
// rider fires; on any other payment the player keeps the most value by
// preserving tagged mana for a later qualifying spell, so plain units of a color
// are consumed first and only forced tagged consumption (when plain mana of that
// color ran out) removes riders without firing.
//
// qualifies reports, for a rider whose tagged mana was spent on this payment,
// whether the payment satisfied that rider's condition. fire resolves a fired
// rider (putting its effect on the stack). It is a no-op when the player holds
// no riders, so cost payment carries no overhead for ordinary mana, and a free
// function because it needs no engine state.
func resolveManaSpendRiders(
	player *game.Player,
	before map[mana.Color]int,
	spent map[mana.Color]int,
	qualifies func(rider game.ManaRiderInstance) bool,
	fire func(rider game.ManaRiderInstance),
) {
	if len(player.ManaRiders) == 0 {
		return
	}
	riderCount := make(map[mana.Color]int, len(player.ManaRiders))
	for _, instance := range player.ManaRiders {
		riderCount[instance.Color]++
	}
	consume := make(map[mana.Color]int, len(riderCount))
	for color, riders := range riderCount {
		// The planner spends existing pool mana before tapping new sources, so
		// the pre-existing pool consumed (which alone can include tagged mana) is
		// the lesser of what was in the pool and what the payment drew from it.
		preExistingSpent := min(before[color], spent[color])
		if preExistingSpent <= 0 {
			continue
		}
		// A rider fires only when spending its tagged mana benefits the player,
		// so determine the per-color fungibility split from the first rider of
		// the color; all riders of one color share the same condition here.
		plain := before[color] - riders
		var take int
		if colorRiderQualifies(player, color, qualifies) {
			take = min(riders, preExistingSpent)
		} else {
			take = preExistingSpent - plain
		}
		if take > 0 {
			consume[color] = min(take, riders)
		}
	}
	if len(consume) == 0 {
		return
	}
	remaining := player.ManaRiders[:0]
	for _, instance := range player.ManaRiders {
		if consume[instance.Color] > 0 {
			consume[instance.Color]--
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

// colorRiderQualifies reports whether the riders of the given color would fire
// if their tagged mana were spent on this payment. All riders of one color share
// the same condition during a single payment, so the first such rider decides.
func colorRiderQualifies(
	player *game.Player,
	color mana.Color,
	qualifies func(rider game.ManaRiderInstance) bool,
) bool {
	for _, instance := range player.ManaRiders {
		if instance.Color == color {
			return qualifies(instance)
		}
	}
	return false
}

// resolveSpellCastManaSpendRiders consumes the casting player's tagged mana that
// was spent paying for a just-cast spell, firing each rider whose condition the
// spell satisfies. before is the colored pool snapshot captured immediately
// before the spell's costs were paid and spent is the exact per-color pool mana
// the payment consumed. It is a no-op when the player holds no riders, so
// ordinary spell casts carry no overhead.
func resolveSpellCastManaSpendRiders(
	g *game.Game,
	playerID game.PlayerID,
	before map[mana.Color]int,
	spent map[mana.Color]int,
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
	resolveManaSpendRiders(player, before, spent, qualifies, func(rider game.ManaRiderInstance) {
		fireManaSpendRider(g, rider)
	})
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

// spell that shares a creature type with the rider controller's commander
// (Path of Ancestry's spend condition). It fails closed when the controller has
// no single modeled commander, so partner or Background commanders (not modeled
// as one commander instance) never spuriously satisfy the condition.
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
	commander, ok := g.GetCardInstance(player.CommanderInstanceID)
	if !ok || commander.Def == nil {
		return false
	}
	for _, subtype := range spellDef.Subtypes {
		if !types.KnownSubtypeForType(types.Creature, subtype) {
			continue
		}
		if commander.Def.HasSubtype(subtype) {
			return true
		}
	}
	return false
}
