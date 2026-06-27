package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// distinctManaColorsSpent counts the distinct colors of mana among a payment's
// per-unit pool spend, the Converge count "for each color of mana spent to cast
// it" (CR 202.2, CR 702.76). Colorless mana contributes no color, and snow
// provenance is ignored so the same color produced by a snow and a non-snow
// source counts once. It reads the exact units the payment consumed, so a
// generic cost paid with colored mana still counts those colors.
func distinctManaColorsSpent(poolSpend map[mana.Unit]int) int {
	seen := make(map[color.Color]bool, len(poolSpend))
	for unit, count := range poolSpend {
		if count <= 0 {
			continue
		}
		c, ok := manaColor(unit.Color)
		if !ok {
			continue
		}
		seen[c] = true
	}
	return len(seen)
}

// manaSpentByColor records, per color, how much colored mana a payment's
// per-unit pool spend consumed (CR 202.2), backing the Adamant ability word's
// "at least three <color> mana was spent to cast this spell" and "mana of the
// same color" predicates (CR 702.132). Colorless mana contributes no entry and
// snow provenance is ignored, so the same color from a snow and a non-snow
// source accrues to one tally. It returns nil when no colored mana was spent.
func manaSpentByColor(poolSpend map[mana.Unit]int) map[color.Color]int {
	var byColor map[color.Color]int
	for unit, count := range poolSpend {
		if count <= 0 {
			continue
		}
		c, ok := manaColor(unit.Color)
		if !ok {
			continue
		}
		if byColor == nil {
			byColor = make(map[color.Color]int, len(poolSpend))
		}
		byColor[c] += count
	}
	return byColor
}

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
	indicesByUnit := make(map[mana.Unit][]int, len(player.ManaRiders))
	qualified := make([]bool, len(player.ManaRiders))
	for i, instance := range player.ManaRiders {
		indicesByUnit[instance.Unit] = append(indicesByUnit[instance.Unit], i)
		qualified[i] = qualifies(instance)
	}
	consume := make([]bool, len(player.ManaRiders))
	for unit, indices := range indicesByUnit {
		// The planner spends existing pool mana before tapping new sources, so
		// the pre-existing pool consumed (which alone can include tagged mana) is
		// the lesser of what was in the pool and what the payment drew from it.
		preExistingSpent := min(before[unit], spent[unit])
		if preExistingSpent <= 0 {
			continue
		}
		remaining := preExistingSpent
		remaining = consumeQualifiedRiders(
			player, indices, qualified, consume,
			game.ManaSpendRestrictedToCondition, remaining,
		)
		remaining = consumeQualifiedRiders(
			player, indices, qualified, consume,
			game.ManaSpendUnrestricted, remaining,
		)
		plain := max(before[unit]-len(indices), 0)
		remaining -= min(remaining, plain)
		for _, index := range indices {
			if remaining == 0 {
				break
			}
			if !qualified[index] &&
				player.ManaRiders[index].Rider.Restriction == game.ManaSpendUnrestricted {
				consume[index] = true
				remaining--
			}
		}
	}
	remaining := player.ManaRiders[:0]
	for i, instance := range player.ManaRiders {
		if consume[i] {
			if qualified[i] {
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

func consumeQualifiedRiders(
	player *game.Player,
	indices []int,
	qualified, consume []bool,
	restriction game.ManaSpendRestrictionKind,
	remaining int,
) int {
	for _, index := range indices {
		if remaining == 0 {
			break
		}
		if qualified[index] &&
			player.ManaRiders[index].Rider.Restriction == restriction {
			consume[index] = true
			remaining--
		}
	}
	return remaining
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
	spellObjects ...*game.StackObject,
) {
	player, ok := playerByID(g, playerID)
	if !ok || len(player.ManaRiders) == 0 {
		return
	}
	qualifies := func(rider game.ManaRiderInstance) bool {
		return manaSpendConditionSatisfied(g, rider, spellDef)
	}
	processManaSpendRiders(player, before, spent, qualifies, func(rider game.ManaRiderInstance) {
		if rider.Rider.SpellRuleEffect != game.RuleEffectNone && len(spellObjects) > 0 {
			applyManaSpendSpellRuleEffect(spellObjects[0], rider)
			return
		}
		if len(rider.Rider.SpellGainsKeywords) > 0 && len(spellObjects) > 0 {
			tagSpellGainsKeywords(spellObjects[0], rider)
			return
		}
		if !rider.Rider.FiresOnSpend() {
			return
		}
		fireManaSpendRider(g, rider)
	})
}

func manaSpendConditionSatisfied(g *game.Game, rider game.ManaRiderInstance, spellDef *game.CardDef) bool {
	switch rider.Rider.Condition {
	case game.ManaSpendCastCommanderCreatureType:
		return spellSatisfiesCommanderCreatureTypeRider(g, rider.Controller, spellDef)
	case game.ManaSpendCastChosenCreatureType, game.ManaSpendCastOrActivateChosenCreatureType:
		return rider.MatchesChosenCreatureType(spellDef)
	case game.ManaSpendCastLegendarySpell:
		return spellDef != nil && spellDef.HasSupertype(types.Legendary)
	case game.ManaSpendCastCreatureSpell:
		return spellDef != nil && spellDef.HasType(types.Creature)
	case game.ManaSpendCastArtifactSpell,
		game.ManaSpendCastArtifactSpellOnly,
		game.ManaSpendCastOrActivateArtifact,
		game.ManaSpendCastArtifactOrActivateAbility:
		return spellDef != nil && spellDef.HasType(types.Artifact)
	default:
		// ManaSpendActivateArtifactAbility and other activation-only conditions
		// are never satisfied by a spell cast.
		return false
	}
}

// tagSpellGainsKeywords records on a qualifying creature spell the keyword
// abilities its mana-spend rider grants (Arena of Glory, Generator Servant: "it
// gains haste until end of turn"). The keywords are applied as an until-end-of-
// turn continuous effect to the resolved permanent in resolvePermanentSpell.
func tagSpellGainsKeywords(obj *game.StackObject, rider game.ManaRiderInstance) {
	if obj == nil {
		return
	}
	obj.GainsKeywordsUntilEndOfTurn = append(obj.GainsKeywordsUntilEndOfTurn, rider.Rider.SpellGainsKeywords...)
}

func applyManaSpendSpellRuleEffect(obj *game.StackObject, rider game.ManaRiderInstance) {
	if obj == nil || rider.Rider.SpellRuleEffect != game.RuleEffectCantBeCountered {
		return
	}
	obj.RuleEffects = append(obj.RuleEffects, game.RuleEffect{
		Kind:             game.RuleEffectCantBeCountered,
		Controller:       rider.Controller,
		SourceObjectID:   rider.SourceObjectID,
		SourceCardID:     rider.SourceID,
		AffectedObjectID: obj.ID,
	})
}

// consumeManaSpendRidersForPayment drops the tagged mana-spend rider units whose
// mana a just-completed non-spell payment (an activated ability, a ward or other
// additional cost, and similar) spent. Most such payments never satisfy a
// rider's condition, so no rider fires; consuming the units keeps rider
// provenance exact so later same-color mana cannot inherit a stale rider.
//
// The exceptions are the activation-admitting restrictions: the cast-or-activate
// chosen-type restriction (Secluded Courtyard), whose tagged mana may pay to
// activate an ability of a creature source of the chosen type, and the artifact
// restrictions (Power Depot, Soldevi Machinist, Guidelight Optimizer), whose
// tagged mana may pay to activate an ability of an artifact source or, for the
// any-ability form, any activated ability. When source satisfies the rider's
// condition its (effectless) unit is consumed as a qualifying spend. It is a
// no-op when the player holds no riders.
func consumeManaSpendRidersForPayment(
	g *game.Game,
	playerID game.PlayerID,
	source *game.Permanent,
	before map[mana.Unit]int,
	spent map[mana.Unit]int,
) {
	player, ok := playerByID(g, playerID)
	if !ok || len(player.ManaRiders) == 0 {
		return
	}
	qualifies := func(rider game.ManaRiderInstance) bool {
		return abilityActivationSatisfiesManaSpendRider(g, source, rider)
	}
	processManaSpendRiders(player, before, spent, qualifies, func(rider game.ManaRiderInstance) {
		if rider.Rider.FiresOnSpend() {
			fireManaSpendRider(g, rider)
		}
	})
}

// abilityActivationSatisfiesManaSpendRider reports whether activating an ability
// of source satisfies the rider's spend condition. Only the cast-or-activate
// chosen-type restriction is satisfied by an activation, and only when source is
// a creature permanent of the rider's captured chosen subtype.
func abilityActivationSatisfiesManaSpendRider(g *game.Game, source *game.Permanent, rider game.ManaRiderInstance) bool {
	switch rider.Rider.Condition {
	case game.ManaSpendCastOrActivateChosenCreatureType:
		return source != nil &&
			types.KnownSubtypeForType(types.Creature, rider.ChosenSubtype) &&
			permanentHasType(g, source, types.Creature) &&
			permanentHasSubtype(g, source, rider.ChosenSubtype)
	case game.ManaSpendCastOrActivateArtifact, game.ManaSpendActivateArtifactAbility:
		return source != nil && permanentHasType(g, source, types.Artifact)
	case game.ManaSpendCastArtifactOrActivateAbility:
		return source != nil
	default:
		return false
	}
}

// drainFiredManaSpendRiders converts mana-spend riders that fired since the last
// trigger pass into pending triggered abilities and clears the queue, so they
// are ordered with that turn's other triggered abilities under APNAP and
// same-controller ordering (CR 603.3b). Each rider's stack object mirrors a
// battlefield-sourced ability: its source is the producing permanent, so the
// rider resolves with the controller even if that permanent has since left the
// battlefield. The rider effect is validated to have no targets, so the pending
// entry carries no targets.
func (*Engine) drainFiredManaSpendRiders(g *game.Game) []pendingTriggeredAbility {
	if len(g.FiredManaSpendRiders) == 0 {
		return nil
	}
	pending := make([]pendingTriggeredAbility, 0, len(g.FiredManaSpendRiders))
	for _, instance := range g.FiredManaSpendRiders {
		ability := instance.Rider.Ability()
		pending = append(pending, pendingTriggeredAbility{
			controller:   instance.Controller,
			sourceID:     instance.SourceObjectID,
			sourceCardID: instance.SourceID,
			face:         game.FaceFront,
			inline:       &ability,
		})
	}
	g.FiredManaSpendRiders = nil
	return pending
}

// stack with that turn's other triggered abilities, ordered under APNAP and
// same-controller ordering (CR 603.3b). Draining the queue (see
// drainFiredManaSpendRiders) builds the stack object, mirroring a
// battlefield-sourced ability whose source is the producing permanent, so the
// rider resolves with the controller even if that permanent has since left the
// battlefield.
func fireManaSpendRider(g *game.Game, instance game.ManaRiderInstance) {
	g.FiredManaSpendRiders = append(g.FiredManaSpendRiders, instance)
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
// commander instance or its current characteristics cannot be faithfully
// resolved.
//
// The commander's current creature types come from its current object:
//   - On the battlefield, whether the commander is a standalone permanent or a
//     component merged under another card (Mutate), the object is that single
//     permanent, so its effective subtypes (reflecting transform, face-down, and
//     type-changing effects, and a Mutate pile's chosen top card) are the current
//     characteristics. A face-down commander therefore has no creature subtypes
//     and correctly matches nothing.
//   - On the stack (being cast), the spell's selected face determines its current
//     characteristics, so a commander cast as its back face uses that face rather
//     than the printed front face.
//   - Anywhere else (command zone, hand, graveyard, exile, or library) no effects
//     alter its types and a double-faced or modal card uses its front face by
//     default (CR 711.2, 712.4a), which the card definition's front face
//     represents.
func commanderCreatureSubtypeMatcher(g *game.Game, player *game.Player) (func(types.Sub) bool, bool) {
	commander, ok := g.GetCardInstance(player.CommanderInstanceID)
	if !ok || commander.Def == nil {
		return nil, false
	}
	if permanent, ok := commanderPermanent(g, player.CommanderInstanceID); ok {
		return func(subtype types.Sub) bool {
			return permanentHasSubtype(g, permanent, subtype)
		}, true
	}
	if faceDef, ok := commanderStackFaceDef(g, commander); ok {
		if faceDef == nil {
			return nil, false
		}
		return func(subtype types.Sub) bool {
			return faceDef.HasSubtype(subtype)
		}, true
	}
	return func(subtype types.Sub) bool {
		return commander.Def.HasSubtype(subtype)
	}, true
}

// commanderStackFaceDef returns the face definition of the commander while it is
// a spell on the stack, using the spell's selected face. The bool reports
// whether the commander is currently on the stack as a spell; when it is and the
// selected face cannot be resolved the returned face is nil so the caller fails
// closed rather than falling back to the printed front face.
func commanderStackFaceDef(g *game.Game, commander *game.CardInstance) (*game.CardDef, bool) {
	for _, obj := range g.Stack.Objects() {
		if obj.Kind != game.StackSpell || obj.SourceID != commander.ID {
			continue
		}
		faceDef, ok := commander.Def.FaceDef(obj.Face)
		if !ok {
			return nil, true
		}
		return faceDef, true
	}
	return nil, false
}

// greatestSameColorManaSpent returns the largest amount of a single color of
// mana recorded in a per-color spend tally, backing the Adamant "at least three
// mana of the same color was spent to cast this spell" predicate (CR 702.132).
// It returns zero for an empty tally.
func greatestSameColorManaSpent(byColor map[color.Color]int) int {
	greatest := 0
	for _, count := range byColor {
		if count > greatest {
			greatest = count
		}
	}
	return greatest
}
