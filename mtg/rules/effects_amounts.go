package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func stackObjectSourceIsSnow(g *game.Game, obj *game.StackObject) bool {
	permanent, ok := permanentByObjectID(g, obj.SourceID)
	return ok && permanentIsSnow(g, permanent)
}

func permanentIsSnow(g *game.Game, permanent *game.Permanent) bool {
	return permanentHasSupertype(g, permanent, types.Snow)
}

func dynamicAmountValue(g *game.Game, obj *game.StackObject, controller game.PlayerID, dynamic game.DynamicAmount) int {
	return dynamicAmountValueBeforeLayer(g, opt.Val(obj), controller, dynamic, 0)
}

func dynamicAmountValueBeforeLayer(g *game.Game, obj opt.V[*game.StackObject], controller game.PlayerID, dynamic game.DynamicAmount, before game.ContinuousLayer) int {
	amount := 0
	switch dynamic.Kind {
	case game.DynamicAmountConstant:
		amount = dynamic.Constant
	case game.DynamicAmountX:
		if obj.Exists {
			amount = obj.Val.XValue
		}
	case game.DynamicAmountTargetPower:
		if !obj.Exists {
			break
		}
		if resolved, ok := resolveObjectReference(g, obj.Val, dynamic.Object); ok && resolved.permanent != nil {
			permanent := resolved.permanent
			amount = effectivePower(g, permanent)
		}
	case game.DynamicAmountTargetToughness:
		if !obj.Exists {
			break
		}
		if resolved, ok := resolveObjectReference(g, obj.Val, dynamic.Object); ok && resolved.permanent != nil {
			permanent := resolved.permanent
			if toughness, ok := effectiveToughness(g, permanent); ok {
				amount = toughness
			}
		}
	case game.DynamicAmountTargetManaValue:
		if !obj.Exists {
			break
		}
		if resolved, ok := resolveObjectReference(g, obj.Val, dynamic.Object); ok && resolved.permanent != nil {
			permanent := resolved.permanent
			if def, ok := permanentCardDef(g, permanent); ok {
				amount = def.ManaValue()
			}
		}
	case game.DynamicAmountTargetCounters:
		if !obj.Exists {
			break
		}
		if resolved, ok := resolveObjectReference(g, obj.Val, dynamic.Object); ok && resolved.permanent != nil {
			permanent := resolved.permanent
			amount = permanent.Counters.Get(dynamic.CounterKind)
		}
	case game.DynamicAmountControllerLife, game.DynamicAmountControllerHandSize,
		game.DynamicAmountControllerGraveyardSize, game.DynamicAmountControllerBasicLandTypeCount,
		game.DynamicAmountOpponentCount, game.DynamicAmountOpponentsAttackedThisCombat,
		game.DynamicAmountControllerSpeed, game.DynamicAmountCommanderCastCount,
		game.DynamicAmountPartySize:
		amount = controllerAggregateAmount(g, controller, dynamic, before)
	case game.DynamicAmountOpponentControllingCount:
		for _, opponent := range aliveOpponents(g, controller) {
			if countPermanentsMatchingGroup(g, nil, opponent, dynamic.Group) > 0 {
				amount++
			}
		}
	case game.DynamicAmountDevotion:
		// ColorFrom binds devotion to the color chosen as the ability resolves
		// (Nykthos, Shrine to Nyx's "devotion to that color"); otherwise the
		// amount's fixed Colors apply. A missing or unreadable choice yields no
		// colors, so devotion is zero.
		colors := dynamic.Colors
		if dynamic.ColorFrom != "" {
			colors = nil
			if result, ok := linkedResolutionChoice(obj.Val, string(dynamic.ColorFrom)); ok {
				if chosen, ok := manaColor(result.Color); ok {
					colors = []color.Color{chosen}
				}
			}
		}
		amount = controllerDevotion(g, controller, colors)
	case game.DynamicAmountCountSelector, game.DynamicAmountGreatestPowerInGroup,
		game.DynamicAmountGreatestToughnessInGroup, game.DynamicAmountGreatestManaValueInGroup,
		game.DynamicAmountTotalPowerInGroup, game.DynamicAmountTotalToughnessInGroup,
		game.DynamicAmountTotalManaValueInGroup,
		game.DynamicAmountColorCountInGroup:
		amount = groupDynamicAmount(g, obj, controller, &dynamic)
	case game.DynamicAmountCountCardsInZone:
		if dynamic.Player != nil && dynamic.Selection != nil {
			amount = countCardsInZoneMatchingSelection(g, obj, controller, *dynamic.Player, dynamic.CardZone, *dynamic.Selection)
		}
	case game.DynamicAmountPlayerLife:
		if dynamic.Player != nil && obj.Exists {
			if playerID, ok := resolvePlayerReference(g, obj.Val, *dynamic.Player); ok {
				if player, ok := playerByID(g, playerID); ok {
					amount = player.Life
				}
			}
		}
	case game.DynamicAmountPreviousEffectResult:
		key := dynamicResultKey(dynamic)
		if obj.Exists && key != "" {
			amount = obj.Val.ResolvedAmounts[key]
		}
	case game.DynamicAmountPreviousEffectExcessDamage:
		key := dynamicResultKey(dynamic)
		if obj.Exists && key != "" {
			amount = obj.Val.ResolvedExcessDamage[key]
		}
	case game.DynamicAmountEventDamage, game.DynamicAmountEventLifeChange, game.DynamicAmountEventCounterCount:
		if obj.Exists && obj.Val.HasTriggerEvent {
			amount = obj.Val.TriggerEvent.Amount
		}
	case game.DynamicAmountEventCardCount:
		if obj.Exists && obj.Val.HasTriggerEvent {
			amount = triggerEventCardCount(g, obj.Val.TriggerEvent)
		}
	case game.DynamicAmountSpellTargetCount:
		if obj.Exists && obj.Val.HasTriggerEvent && dynamic.Selection != nil {
			amount = countSpellTargetsMatching(g, controller, game.TargetAllowPermanent, *dynamic.Selection, obj.Val.TriggerEvent)
		}
	case game.DynamicAmountObjectPower:
		if !obj.Exists {
			break
		}
		if resolved, ok := resolveObjectReference(g, obj.Val, dynamic.Object); ok {
			amount = resolvedObjectPower(g, &resolved)
		}
	case game.DynamicAmountObjectToughness:
		if !obj.Exists {
			break
		}
		if resolved, ok := resolveObjectReference(g, obj.Val, dynamic.Object); ok {
			amount = resolvedObjectToughness(g, &resolved)
		}
	case game.DynamicAmountSourceCardPower, game.DynamicAmountBlockingCreaturesBeyondFirst,
		game.DynamicAmountObjectManaValue, game.DynamicAmountCapturedTargetManaValue:
		amount = sourceDerivedDynamicAmount(g, obj, dynamic)
	case game.DynamicAmountObjectCounters:
		if !obj.Exists {
			break
		}
		if resolved, ok := resolveObjectReference(g, obj.Val, dynamic.Object); ok {
			if resolved.permanent != nil {
				amount = resolved.permanent.Counters.Get(dynamic.CounterKind)
			} else {
				amount = resolved.snapshot.Counters.Get(dynamic.CounterKind)
			}
		}
	case game.DynamicAmountChosenNumber:
		if choice, ok := linkedResolutionChoice(obj.Val, string(dynamic.ResultKey)); ok &&
			choice.Kind == game.ResolutionChoiceNumber {
			amount = choice.Number
		}
	case game.DynamicAmountBlockingCreatures:
		if !obj.Exists {
			break
		}
		if resolved, ok := resolveObjectReference(g, obj.Val, dynamic.Object); ok && resolved.permanent != nil {
			amount = blockingCreaturesOf(g, resolved.permanent.ObjectID)
		}
	case game.DynamicAmountSpellsCastThisTurn, game.DynamicAmountLifeLostThisTurn,
		game.DynamicAmountLifeGainedThisTurn, game.DynamicAmountCardsDrawnThisTurn:
		if player, ok := resolveTurnEventPlayer(g, obj, controller, dynamic.Player); ok {
			amount = turnEventDynamicAmount(g, player, dynamic.Kind)
		}
	case game.DynamicAmountCardsNamedSourceInGraveyards:
		amount = countCardsNamedSourceInAllGraveyards(g, obj.Val)
	case game.DynamicAmountCardsNamedSourceInControllerGraveyard:
		amount = countCardsNamedSourceInControllerGraveyard(g, obj.Val, controller)
	case game.DynamicAmountColorsOfManaSpentToCast:
		if obj.Exists {
			amount = obj.Val.ColorsOfManaSpentToCast
		}
	case game.DynamicAmountTimesKicked:
		if obj.Exists {
			amount = obj.Val.KickerCount
		}
	case game.DynamicAmountMaxOf:
		amount = maxOfDynamicAmounts(g, obj, controller, dynamic.Operands, before)
	default:
	}
	multiplier := dynamic.Multiplier
	if multiplier == 0 {
		multiplier = 1
	}
	return applyDynamicAmountDivisor(amount*multiplier+dynamic.Addend, dynamic)
}

// applyDynamicAmountDivisor divides a dynamic amount's value by its Divisor when
// one is set, rounding down unless RoundUp is set ("half their library, rounded
// up/down" — Traumatize, Fleet Swallower; CR 107.4). A Divisor of zero or one
// leaves the value unchanged. Library sizes are non-negative, so truncating
// integer division yields the floor and the (value+Divisor-1) form yields the
// ceiling.
func applyDynamicAmountDivisor(value int, dynamic game.DynamicAmount) int {
	if dynamic.Divisor <= 1 {
		return value
	}
	if dynamic.RoundUp {
		return (value + dynamic.Divisor - 1) / dynamic.Divisor
	}
	return value / dynamic.Divisor
}

// sourceDerivedDynamicAmount evaluates the dynamic amounts that read from the
// resolving ability's source object or its just-exiled card, split out of
// dynamicAmountValueBeforeLayer so that large switch stays within the
// maintainability budget.
func sourceDerivedDynamicAmount(g *game.Game, obj opt.V[*game.StackObject], dynamic game.DynamicAmount) int {
	switch dynamic.Kind {
	case game.DynamicAmountSourceCardPower:
		return sourceCardPrintedPower(g, obj.Val)
	case game.DynamicAmountBlockingCreaturesBeyondFirst:
		return blockingCreaturesBeyondFirst(g, obj.Val)
	case game.DynamicAmountObjectManaValue, game.DynamicAmountCapturedTargetManaValue:
		return dynamicObjectManaValue(g, obj.Val, &dynamic)
	default:
		return 0
	}
}

// blockingCreaturesBeyondFirst counts the creatures blocking the resolving
// ability's source permanent beyond the first, read from the current combat's
// block declarations (CR 509.1, CR 702.23). It is zero when combat is not active
// or the source is blocked by one or no creatures, so a Rampage trigger that
// somehow resolves outside combat contributes nothing.
func blockingCreaturesBeyondFirst(g *game.Game, obj *game.StackObject) int {
	if g.Combat == nil {
		return 0
	}
	permanent, ok := sourcePermanent(g, obj)
	if !ok {
		return 0
	}
	blockers := 0
	for _, block := range g.Combat.Blockers {
		if block.Blocking == permanent.ObjectID {
			blockers++
		}
	}
	if blockers <= 1 {
		return 0
	}
	return blockers - 1
}

// blockingCreaturesOf counts every creature blocking the given permanent, read
// from the current combat's block declarations (CR 509.1, CR 608.2c). It is zero
// when combat is not active. Unlike blockingCreaturesBeyondFirst, which drops the
// first blocker for Rampage, it counts all blockers, backing the "+N/+N until end
// of turn for each creature blocking it" pump (Rabid Elephant, Gang of Elk).
func blockingCreaturesOf(g *game.Game, permanentID game.ObjectID) int {
	if g.Combat == nil {
		return 0
	}
	blockers := 0
	for _, block := range g.Combat.Blockers {
		if block.Blocking == permanentID {
			blockers++
		}
	}
	return blockers
}

// resolveTurnEventPlayer resolves the player whose turn-event totals a
// referenced-player amount reads. A nil reference (the common controller-scoped
// case) resolves to controller. A non-nil reference — set only by the "the life
// that player lost/gained this turn" family — resolves against the resolving
// stack object, matching the player the effect targets. It returns ok=false when
// a non-controller reference cannot be resolved without a stack object.
func resolveTurnEventPlayer(g *game.Game, obj opt.V[*game.StackObject], controller game.PlayerID, ref *game.PlayerReference) (game.PlayerID, bool) {
	if ref == nil {
		return controller, true
	}
	if obj.Exists {
		return resolvePlayerReference(g, obj.Val, *ref)
	}
	if ref.Kind() == game.PlayerReferenceController {
		return controller, true
	}
	return 0, false
}

// turnEventDynamicAmount dispatches the controller-scoped amounts derived from
// the current turn's event log (CR 608.2c): the number of spells cast, the
// number of cards drawn, and the total life gained or lost so far this turn. It
// is split out of dynamicAmountValueBeforeLayer so that large switch stays
// within the maintainability budget.
func turnEventDynamicAmount(g *game.Game, controller game.PlayerID, kind game.DynamicAmountKind) int {
	switch kind {
	case game.DynamicAmountLifeLostThisTurn:
		return lifeChangedThisTurn(g, controller, game.EventLifeLost)
	case game.DynamicAmountLifeGainedThisTurn:
		return lifeChangedThisTurn(g, controller, game.EventLifeGained)
	case game.DynamicAmountCardsDrawnThisTurn:
		return cardsDrawnThisTurn(g, controller)
	default:
		return spellsCastThisTurn(g, controller)
	}
}

// maxOfDynamicAmounts evaluates each operand of a DynamicAmountMaxOf combinator
// against the same resolution context and returns the greatest value
// (CR 608.2c). It backs the "whichever is greater" wording; an empty operand
// list yields zero.
func maxOfDynamicAmounts(g *game.Game, obj opt.V[*game.StackObject], controller game.PlayerID, operands []game.DynamicAmount, before game.ContinuousLayer) int {
	best := 0
	for i := range operands {
		value := dynamicAmountValueBeforeLayer(g, obj, controller, operands[i], before)
		if i == 0 || value > best {
			best = value
		}
	}
	return best
}

// spellsCastThisTurn counts the spells the controller has cast so far this turn
// from the turn's recorded spell-cast events (CR 608.2c). A triggered ability's
// own triggering spell counts, because its cast event precedes the ability's
// resolution.
func spellsCastThisTurn(g *game.Game, controller game.PlayerID) int {
	return eventsThisTurnWindow(g).count(eventKindController(game.EventSpellCast, controller))
}

// lifeChangedThisTurn sums the life a player gained or lost so far this turn from
// the turn's recorded life-change events (CR 608.2c). Pass game.EventLifeGained
// for "the life you gained this turn" or game.EventLifeLost for "the life you've
// lost this turn"; damage to the player contributes to the life-lost total
// because dealing damage to a player causes that much life loss (CR 120.3),
// emitted as an EventLifeLost.
func lifeChangedThisTurn(g *game.Game, player game.PlayerID, kind game.EventKind) int {
	return eventsThisTurnWindow(g).sumAmount(eventKindPlayer(kind, player))
}

// opponentLostLifeThisTurn reports whether any opponent of playerID has lost
// life so far this turn (CR 702.107b, Spectacle). Damage to a player is life
// loss (CR 120.3), so both combat and noncombat damage to an opponent qualify.
func opponentLostLifeThisTurn(g *game.Game, playerID game.PlayerID) bool {
	return eventsThisTurnWindow(g).any(func(event game.Event) bool {
		return event.Kind == game.EventLifeLost && event.Player != playerID && event.Amount > 0
	})
}

// controllerAggregateAmount computes the player-relative dynamic amounts that
// depend only on the controller's own board and zones (life total, hand and
// graveyard sizes, basic land type and opponent counts, and devotion). It is
// split out of dynamicAmountValueBeforeLayer so that large switch stays within
// the maintainability budget; behavior is identical to the inlined cases.
func controllerAggregateAmount(g *game.Game, controller game.PlayerID, dynamic game.DynamicAmount, before game.ContinuousLayer) int {
	switch dynamic.Kind {
	case game.DynamicAmountControllerLife:
		if player, ok := playerByID(g, controller); ok {
			return player.Life
		}
	case game.DynamicAmountControllerHandSize:
		if player, ok := playerByID(g, controller); ok {
			return cardInstanceCount(g, player.Hand.All())
		}
	case game.DynamicAmountControllerGraveyardSize:
		if player, ok := playerByID(g, controller); ok {
			return cardInstanceCount(g, player.Graveyard.All())
		}
	case game.DynamicAmountControllerBasicLandTypeCount:
		return controllerBasicLandTypeCount(g, conditionContext{
			controller:            controller,
			characteristicsBefore: before,
		})
	case game.DynamicAmountOpponentCount:
		return len(aliveOpponents(g, controller))
	case game.DynamicAmountCommanderCastCount:
		if player, ok := playerByID(g, controller); ok {
			return player.CommanderCastCount
		}
	case game.DynamicAmountOpponentsAttackedThisCombat:
		return opponentsAttackedThisCombat(g, controller)
	case game.DynamicAmountControllerSpeed:
		if player, ok := playerByID(g, controller); ok {
			return player.Speed
		}
	case game.DynamicAmountPartySize:
		return controllerPartySize(g, controller)
	default:
	}
	return 0
}

func controllerPartySize(g *game.Game, controller game.PlayerID) int {
	const (
		cleric = 1 << iota
		rogue
		warrior
		wizard
	)
	reachable := [16]bool{0: true}
	for _, permanent := range g.Battlefield {
		if permanent.PhasedOut ||
			effectiveController(g, permanent) != controller ||
			!permanentHasType(g, permanent, types.Creature) {
			continue
		}
		roles := 0
		if permanentHasSubtype(g, permanent, types.Cleric) {
			roles |= cleric
		}
		if permanentHasSubtype(g, permanent, types.Rogue) {
			roles |= rogue
		}
		if permanentHasSubtype(g, permanent, types.Warrior) {
			roles |= warrior
		}
		if permanentHasSubtype(g, permanent, types.Wizard) {
			roles |= wizard
		}
		next := reachable
		for mask, ok := range reachable {
			if !ok {
				continue
			}
			for role := 1; role <= wizard; role <<= 1 {
				if roles&role != 0 && mask&role == 0 {
					next[mask|role] = true
				}
			}
		}
		reachable = next
	}
	maximum := 0
	for mask, ok := range reachable {
		if !ok {
			continue
		}
		size := 0
		for value := mask; value != 0; value &= value - 1 {
			size++
		}
		if size > maximum {
			maximum = size
		}
	}
	return maximum
}

// opponentsAttackedThisCombat counts the distinct opponents of controller being
// attacked this combat by creatures controller controls, read from the current
// combat's attack declarations as the ability resolves (CR 506.2, CR 702.72).
// It backs the Melee count "for each opponent you attacked this combat" and is
// zero outside combat.
func opponentsAttackedThisCombat(g *game.Game, controller game.PlayerID) int {
	if g.Combat == nil {
		return 0
	}
	opponents := make(map[game.PlayerID]bool, game.NumPlayers)
	for _, opponent := range aliveOpponents(g, controller) {
		opponents[opponent] = false
	}
	for _, declaration := range g.Combat.Attackers {
		attacker, ok := permanentByObjectID(g, declaration.Attacker)
		if !ok || effectiveController(g, attacker) != controller {
			continue
		}
		if _, ok := opponents[declaration.Target.Player]; ok {
			opponents[declaration.Target.Player] = true
		}
	}
	attacked := 0
	for _, did := range opponents {
		if did {
			attacked++
		}
	}
	return attacked
}

// controllerDevotion returns the controller's devotion to colors: the number of
// mana symbols of those colors among the mana costs of the permanents the
// controller controls (CR 700.5). A hybrid or Phyrexian symbol counts once when
// it matches any listed color, so multi-color devotion counts each qualifying
// symbol a single time.
func controllerDevotion(g *game.Game, controller game.PlayerID, colors []color.Color) int {
	if len(colors) == 0 {
		return 0
	}
	targets := make(map[mana.Color]bool, len(colors))
	for _, c := range colors {
		targets[mana.Color(c.Abbreviation())] = true
	}
	devotion := 0
	for _, permanent := range g.Battlefield {
		if permanent.PhasedOut || permanent.Controller != controller {
			continue
		}
		def, ok := permanentCardDef(g, permanent)
		if !ok || !def.ManaCost.Exists {
			continue
		}
		for _, symbol := range def.ManaCost.Val {
			for _, symbolColor := range symbol.Colors() {
				if targets[symbolColor] {
					devotion++
					break
				}
			}
		}
	}
	return devotion
}

// characteristic identifies a numeric permanent characteristic compared when
// taking the greatest value across a group of permanents.
type characteristic int

const (
	characteristicPower characteristic = iota
	characteristicToughness
	characteristicManaValue
)

func greatestGroupCharacteristic(kind game.DynamicAmountKind) characteristic {
	switch kind {
	case game.DynamicAmountGreatestToughnessInGroup:
		return characteristicToughness
	case game.DynamicAmountGreatestManaValueInGroup:
		return characteristicManaValue
	default:
		return characteristicPower
	}
}

// greatestCharacteristicInGroup returns the greatest value of the
// characteristic named by kind among the permanents of group, evaluated as the
// effect resolves (CR 608.2c). An empty group yields zero, matching the "draw
// cards equal to the greatest power among <group>" amounts whose group is empty.
func greatestCharacteristicInGroup(g *game.Game, obj *game.StackObject, controller game.PlayerID, group game.GroupReference, kind game.DynamicAmountKind) int {
	resolverObj := obj
	if resolverObj == nil {
		resolverObj = &game.StackObject{Controller: controller}
	}
	which := greatestGroupCharacteristic(kind)
	greatest := 0
	for _, objectID := range newReferenceResolver(g, resolverObj).groupMembers(group) {
		permanent, ok := permanentByObjectID(g, objectID)
		if !ok {
			continue
		}
		value, ok := permanentCharacteristicValue(g, permanent, which)
		if ok && value > greatest {
			greatest = value
		}
	}
	return greatest
}

// totalGroupCharacteristic maps a total-characteristic dynamic amount kind to
// the permanent characteristic summed across the group.
func totalGroupCharacteristic(kind game.DynamicAmountKind) characteristic {
	switch kind {
	case game.DynamicAmountTotalToughnessInGroup:
		return characteristicToughness
	case game.DynamicAmountTotalManaValueInGroup:
		return characteristicManaValue
	default:
		return characteristicPower
	}
}

// totalCharacteristicInGroup returns the sum of the characteristic named by kind
// across the permanents of group, evaluated as the effect resolves (CR 608.2c).
// An empty group yields zero, matching "the total power of <group>" amounts
// (Ghalta, Primal Hunger's cost reduction) over an empty battlefield.
func totalCharacteristicInGroup(g *game.Game, obj *game.StackObject, controller game.PlayerID, group game.GroupReference, kind game.DynamicAmountKind) int {
	resolverObj := obj
	if resolverObj == nil {
		resolverObj = &game.StackObject{Controller: controller}
	}
	which := totalGroupCharacteristic(kind)
	total := 0
	for _, objectID := range newReferenceResolver(g, resolverObj).groupMembers(group) {
		permanent, ok := permanentByObjectID(g, objectID)
		if !ok {
			continue
		}
		value, ok := permanentCharacteristicValue(g, permanent, which)
		if ok {
			total += value
		}
	}
	return total
}

// dynamicAmountValueForPermanent evaluates a dynamic amount for a continuous
// modification applied to permanent. Most dynamic amounts do not depend on the
// affected permanent and delegate to dynamicAmountValueBeforeLayer; the
// shared-creature-type count yields a different value per affected permanent, so
// it counts the other creatures in its group that share a creature type with it.
func dynamicAmountValueForPermanent(g *game.Game, permanent *game.Permanent, controller game.PlayerID, dynamic game.DynamicAmount, before game.ContinuousLayer) int {
	if dynamic.Kind != game.DynamicAmountSharedCreatureTypeCountInGroup {
		return dynamicAmountValueBeforeLayer(g, opt.V[*game.StackObject]{}, controller, dynamic, before)
	}
	multiplier := dynamic.Multiplier
	if multiplier == 0 {
		multiplier = 1
	}
	return sharedCreatureTypeCountInGroup(g, permanent, controller, dynamic.Group) * multiplier
}

// sharedCreatureTypeCountInGroup returns the number of permanents in the group's
// scope, other than permanent, that share at least one creature type with it
// (CR 700.4, CR 608.2c). The battlefield is scanned directly rather than through
// the group resolver because membership resolution evaluates each permanent's
// full continuous values, which would re-enter the power/toughness modification
// this count feeds and recurse without bound; only the group's controller and
// combat-state scopes are honored here. Creature subtypes are read before the
// power/toughness layers for the same reason. A Changeling, which has every
// creature type, shares with any other creature that has at least one; a
// permanent with no creature types shares with nothing.
func sharedCreatureTypeCountInGroup(g *game.Game, permanent *game.Permanent, controller game.PlayerID, group game.GroupReference) int {
	own := creatureSubtypesBeforePowerToughness(g, permanent)
	if len(own) == 0 {
		return 0
	}
	selection := group.Selection()
	count := 0
	for _, other := range g.Battlefield {
		if other.ObjectID == permanent.ObjectID {
			continue
		}
		if selection.Controller == game.ControllerYou && effectiveController(g, other) != controller {
			continue
		}
		if !combatStateMatches(g, other, selection.CombatState) {
			continue
		}
		if shareCreatureSubtype(own, creatureSubtypesBeforePowerToughness(g, other)) {
			count++
		}
	}
	return count
}

// creatureSubtypesBeforePowerToughness returns the set of creature subtypes a
// permanent has after the type-changing layers but before the power/toughness
// layers, so reading it from inside a power/toughness modification cannot
// recurse. Non-creature subtypes are dropped because only creature types are
// shared (CR 700.4).
func creatureSubtypesBeforePowerToughness(g *game.Game, permanent *game.Permanent) map[types.Sub]struct{} {
	values := permanentValuesBeforeLayer(g, permanent, game.LayerPowerToughnessSet)
	subtypes := make(map[types.Sub]struct{})
	for _, subtype := range values.subtypes {
		if types.KnownSubtypeForType(types.Creature, subtype) {
			subtypes[subtype] = struct{}{}
		}
	}
	return subtypes
}

// shareCreatureSubtype reports whether two creature-subtype sets intersect.
func shareCreatureSubtype(a, b map[types.Sub]struct{}) bool {
	for subtype := range a {
		if _, ok := b[subtype]; ok {
			return true
		}
	}
	return false
}

// groupDynamicAmount dispatches the battlefield-group amounts, each derived from
// the permanents of dynamic.Group as the effect resolves (CR 608.2c): the member
// count, the greatest or total power/toughness/mana value, and the distinct
// color count.
func groupDynamicAmount(g *game.Game, obj opt.V[*game.StackObject], controller game.PlayerID, dynamic *game.DynamicAmount) int {
	switch dynamic.Kind {
	case game.DynamicAmountGreatestPowerInGroup, game.DynamicAmountGreatestToughnessInGroup,
		game.DynamicAmountGreatestManaValueInGroup:
		return greatestCharacteristicInGroup(g, obj.Val, controller, dynamic.Group, dynamic.Kind)
	case game.DynamicAmountTotalPowerInGroup, game.DynamicAmountTotalToughnessInGroup,
		game.DynamicAmountTotalManaValueInGroup:
		return totalCharacteristicInGroup(g, obj.Val, controller, dynamic.Group, dynamic.Kind)
	case game.DynamicAmountColorCountInGroup:
		return colorCountInGroup(g, obj.Val, controller, dynamic.Group)
	default:
		return countPermanentsMatchingGroup(g, obj.Val, controller, dynamic.Group)
	}
}

// colorCountInGroup returns the number of distinct colors among the permanents
// of group, evaluated as the effect resolves (CR 608.2c). Each permanent
// contributes each of its colors (CR 105.2, CR 202.2); a colorless permanent
// contributes none, so an empty or fully colorless group yields zero. It backs
// Faeburrow Elder's "+1/+1 for each color among permanents you control" and the
// "number of colors among <group>" amount family.
func colorCountInGroup(g *game.Game, obj *game.StackObject, controller game.PlayerID, group game.GroupReference) int {
	resolverObj := obj
	if resolverObj == nil {
		resolverObj = &game.StackObject{Controller: controller}
	}
	var found colorSet
	for _, objectID := range newReferenceResolver(g, resolverObj).groupMembers(group) {
		permanent, ok := permanentByObjectID(g, objectID)
		if !ok {
			continue
		}
		values := effectivePermanentValues(g, permanent)
		for _, c := range values.colors {
			found.add(c)
		}
	}
	return len(found.ordered())
}

func permanentCharacteristicValue(g *game.Game, permanent *game.Permanent, which characteristic) (int, bool) {
	switch which {
	case characteristicPower:
		return effectivePower(g, permanent), true
	case characteristicToughness:
		return effectiveToughness(g, permanent)
	case characteristicManaValue:
		if def, ok := permanentCardDef(g, permanent); ok {
			return def.ManaValue(), true
		}
		return 0, false
	default:
		return 0, false
	}
}

func dynamicObjectManaValue(g *game.Game, obj *game.StackObject, dynamic *game.DynamicAmount) int {
	if obj == nil {
		return 0
	}
	if dynamic.Kind == game.DynamicAmountCapturedTargetManaValue {
		if dynamic.Object.Kind() != game.ObjectReferenceCapturedTargetStackObject {
			return 0
		}
		return obj.CapturedTargetManaValueLKI[dynamic.Object.TargetIndex()]
	}
	if dynamic.Object.Kind() == game.ObjectReferenceTargetStackObject {
		targetIndex := dynamic.Object.TargetIndex()
		targetID, ok := effectStackObjectID(obj, targetIndex)
		if ok {
			if _, live := stackObjectByID(g, targetID); !live {
				return obj.TargetManaValueLKI[targetIndex]
			}
		}
	}
	resolved, ok := resolveObjectReference(g, obj, dynamic.Object)
	if !ok {
		return 0
	}
	return resolvedObjectManaValue(g, &resolved)
}

// triggerEventCardCount reports the number of cards drawn or discarded in the
// triggering event. A "one or more" draw or discard trigger coalesces the
// simultaneous batch into a single trigger and retains the first matching event
// (game.TriggerPattern.OneOrMore), so the batch size is the count of events that
// share the trigger event's SimultaneousID, Kind, and affected player. A
// trigger with no batch (SimultaneousID zero) counts the triggering event alone.
func triggerEventCardCount(g *game.Game, trigger game.Event) int {
	if trigger.SimultaneousID == 0 {
		return 1
	}
	count := 0
	for _, event := range g.Events {
		if event.SimultaneousID == trigger.SimultaneousID &&
			event.Kind == trigger.Kind &&
			event.Player == trigger.Player {
			count++
		}
	}
	if count == 0 {
		return 1
	}
	return count
}

func countCardsInZoneMatchingSelection(g *game.Game, obj opt.V[*game.StackObject], controller game.PlayerID, playerRef game.PlayerReference, cardZone zone.Type, selection game.Selection) int {
	var playerID game.PlayerID
	var ok bool
	if obj.Exists {
		playerID, ok = resolvePlayerReference(g, obj.Val, playerRef)
	} else if playerRef.Kind() == game.PlayerReferenceController {
		// Continuous-layer evaluation can occur without a stack object (for
		// example while checking a resolving permanent's types). A Controller
		// reference still resolves to the permanent's controller passed here.
		playerID, ok = controller, true
	}
	if !ok {
		return 0
	}
	return countCardsInZoneForPlayer(g, playerID, controller, cardZone, selection)
}

// countCardsInZoneForPlayer counts the cards a player owns in a card zone that
// match selection, viewed by controller. It is the player-id core shared by the
// reference-resolving dynamic-amount path and the cost-time self cost reductions
// that count cards in the caster's own zone without a resolving stack object.
func countCardsInZoneForPlayer(g *game.Game, playerID game.PlayerID, viewer game.PlayerID, cardZone zone.Type, selection game.Selection) int {
	player, ok := playerByID(g, playerID)
	if !ok {
		return 0
	}
	collection, ok := playerCardsInZone(player, cardZone)
	if !ok {
		return 0
	}
	count := 0
	for _, cardID := range collection.All() {
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			continue
		}
		subject := selectionSubject{
			kind:       subjectCard,
			g:          g,
			card:       card,
			controller: card.Owner,
			viewer:     viewer,
		}
		if matchSelection(&subject, &selection) {
			count++
		}
	}
	return count
}

// countCardsNamedSourceInAllGraveyards counts the cards in every player's
// graveyard whose name equals the resolving stack object's source card name
// (CR 201.2). It backs the "for each card named <this card> in each graveyard"
// dynamic amount (Rite of Flame). A missing or unnamed source counts nothing.
func countCardsNamedSourceInAllGraveyards(g *game.Game, obj *game.StackObject) int {
	if obj == nil {
		return 0
	}
	name := stackObjectSourceName(g, obj)
	if name == "" {
		return 0
	}
	count := 0
	for _, player := range g.Players {
		for _, cardID := range player.Graveyard.All() {
			card, ok := g.GetCardInstance(cardID)
			if !ok || card.Def == nil {
				continue
			}
			if card.Def.Name == name {
				count++
			}
		}
	}
	return count
}

// countCardsNamedSourceInControllerGraveyard counts the cards in the resolving
// ability's controller's graveyard whose name equals the resolving stack
// object's source card name (CR 201.2). It backs the "for each card named <this
// card> in your graveyard" dynamic amount (Compound Fracture, Growth Cycle),
// counting only the controller's graveyard rather than every graveyard. A
// missing or unnamed source counts nothing.
func countCardsNamedSourceInControllerGraveyard(g *game.Game, obj *game.StackObject, controller game.PlayerID) int {
	if obj == nil {
		return 0
	}
	name := stackObjectSourceName(g, obj)
	if name == "" {
		return 0
	}
	player, ok := playerByID(g, controller)
	if !ok {
		return 0
	}
	count := 0
	for _, cardID := range player.Graveyard.All() {
		card, ok := g.GetCardInstance(cardID)
		if !ok || card.Def == nil {
			continue
		}
		if card.Def.Name == name {
			count++
		}
	}
	return count
}

func playerCardsInZone(player *game.Player, cardZone zone.Type) (*zone.Zone, bool) {
	switch cardZone {
	case zone.Library:
		return &player.Library, true
	case zone.Hand:
		return &player.Hand, true
	case zone.Graveyard:
		return &player.Graveyard, true
	case zone.Exile:
		return &player.Exile, true
	case zone.Command:
		return &player.CommandZone, true
	default:
		return nil, false
	}
}

func dynamicResultKey(dynamic game.DynamicAmount) string {
	return string(dynamic.ResultKey)
}

func resolvedObjectPower(g *game.Game, resolved *resolvedObjectReference) int {
	if resolved.permanent != nil {
		return effectivePower(g, resolved.permanent)
	}
	if resolved.snapshot.Power.Exists {
		return resolved.snapshot.Power.Val
	}
	return 0
}

// sourceCardPrintedPower reads the printed power of the resolving ability's
// source card from its card instance, which persists in any zone (CR 702.94d).
// Scavenge exiles the card from the graveyard as a cost, so by resolution the
// card is no longer a battlefield permanent; reading the instance's front face
// yields the card's power for the +1/+1 counter count.
func sourceCardPrintedPower(g *game.Game, obj *game.StackObject) int {
	if obj == nil {
		return 0
	}
	card, ok := g.GetCardInstance(stackObjectSourceID(obj))
	if !ok {
		return 0
	}
	face := cardFaceOrDefault(card, game.FaceFront)
	if face.Power.Exists {
		return face.Power.Val.Value
	}
	return 0
}

// resolvedObjectToughness reads a referenced object's toughness from the live
// permanent or, once it has left the battlefield, from its last-known snapshot
// (CR 608.2h). It mirrors resolvedObjectPower so "gain/lose life equal to its
// toughness" riders read the same last-known value as their power siblings.
func resolvedObjectToughness(g *game.Game, resolved *resolvedObjectReference) int {
	if resolved.permanent != nil {
		if toughness, ok := effectiveToughness(g, resolved.permanent); ok {
			return toughness
		}
		return 0
	}
	if resolved.snapshot.Toughness.Exists {
		return resolved.snapshot.Toughness.Val
	}
	return 0
}

// resolvedObjectManaValue reads a referenced object's mana value from its printed
// mana cost, taken from the live permanent or, once it has left the battlefield,
// from its last-known snapshot (CR 202.3, CR 608.2h). A live permanent reads the
// face on the battlefield; a destroyed or otherwise departed permanent reads the
// card the snapshot identified, using the snapshot's face so transformed or
// modal cards report the mana value of the face that was on the battlefield. It
// mirrors resolvedObjectPower/Toughness so "gain/lose life equal to its mana
// value" riders read the same last-known object as their characteristic
// siblings. A token with no backing card and no recorded card identity has mana
// value 0.
func resolvedObjectManaValue(g *game.Game, resolved *resolvedObjectReference) int {
	if resolved.permanent != nil {
		if def, ok := permanentCardDef(g, resolved.permanent); ok {
			return def.ManaValue()
		}
		return 0
	}
	if resolved.stack != nil {
		if manaValue, ok := stackObjectManaValue(g, resolved.stack); ok {
			return manaValue
		}
		return 0
	}
	if resolved.snapshot.CardID != 0 {
		if card, ok := g.GetCardInstance(resolved.snapshot.CardID); ok {
			return cardFaceOrDefault(card, resolved.snapshot.Face).ManaValue()
		}
	}
	if resolved.snapshot.TokenDef != nil {
		return resolved.snapshot.TokenDef.ManaValue()
	}
	return 0
}

func rememberEffectAmount(obj *game.StackObject, linkID string, amount int) {
	if linkID == "" {
		return
	}
	if obj.ResolvedAmounts == nil {
		obj.ResolvedAmounts = make(map[string]int)
	}
	obj.ResolvedAmounts[linkID] = amount
}

func rememberEffectExcessDamage(obj *game.StackObject, linkID string, excessDamage int) {
	if linkID == "" || excessDamage <= 0 {
		return
	}
	if obj.ResolvedExcessDamage == nil {
		obj.ResolvedExcessDamage = make(map[string]int)
	}
	obj.ResolvedExcessDamage[linkID] = excessDamage
}

func effectCounterSource(g *game.Game, obj *game.StackObject, source game.CounterSourceSpec) (counter.Set, *game.Permanent, bool) {
	switch source.Kind {
	case game.CounterSourceTarget:
		resolved, ok := resolveObjectReference(g, obj, source.Object)
		if !ok || resolved.permanent == nil {
			return counter.Set{}, nil, false
		}
		permanent := resolved.permanent
		return cloneCounters(permanent.Counters), permanent, true
	case game.CounterSourceEventPermanent:
		if !obj.HasTriggerEvent || obj.TriggerEvent.PermanentID == 0 {
			return counter.Set{}, nil, false
		}
		// Zone-change triggers such as "put those counters on..." use the
		// triggering permanent's current state or its last-known information if it
		// has already left the battlefield (CR 603.10, CR 122).
		if permanent, ok := permanentByObjectID(g, obj.TriggerEvent.PermanentID); ok {
			return cloneCounters(permanent.Counters), permanent, true
		}
		if snapshot, ok := lastKnownObject(g, obj.TriggerEvent.PermanentID); ok {
			return cloneCounters(snapshot.Counters), nil, true
		}
	case game.CounterSourceSelf:
		// "Move a +1/+1 counter from this creature onto target creature." reads
		// counters from the ability's own source permanent.
		if permanent, ok := permanentByObjectID(g, obj.SourceID); ok {
			return cloneCounters(permanent.Counters), permanent, true
		}
		// A dies-trigger that moves the source's own counters (CR 702.44 Modular;
		// Power Depot) resolves after the source has already left the battlefield,
		// so it reads the counters from the source's last-known information (CR
		// 603.10, CR 608.2h). This fallback is scoped to that case: it applies only
		// when this ability triggered on its own source's death. Every other
		// CounterSourceSelf effect — graft's "Whenever another creature enters, move
		// a +1/+1 counter from this creature onto it." and the "{T}: move a counter
		// from this permanent" activated abilities (Explorer's Cache, Diamond City) —
		// moves nothing once its source is gone, because those counters ceased to
		// exist (CR 121.5). The nil permanent tells the caller there is no live
		// source to remove the moved counters from.
		if obj.HasTriggerEvent &&
			obj.TriggerEvent.Kind == game.EventPermanentDied &&
			obj.TriggerEvent.PermanentID == obj.SourceID {
			if snapshot, ok := lastKnownObject(g, obj.SourceID); ok {
				return cloneCounters(snapshot.Counters), nil, true
			}
		}
	default:
	}
	return counter.Set{}, nil, false
}

func effectConditionSatisfied(g *game.Game, obj *game.StackObject, condition opt.V[game.EffectCondition]) bool {
	if !condition.Exists {
		return true
	}
	cond := condition.Val
	if cond.PermanentType.Exists {
		resolved, ok := resolveObjectReference(g, obj, cond.Object)
		if !ok || resolved.permanent == nil {
			return false
		}
		permanent := resolved.permanent
		matches := permanentHasType(g, permanent, cond.PermanentType.Val)
		if cond.Negate {
			matches = !matches
		}
		if !matches {
			return false
		}
	}
	if !conditionSatisfied(g, conditionContext{
		controller: stackObjectController(obj),
		obj:        obj,
	}, cond.Condition) {
		return false
	}
	return true
}

func cardConditionSatisfied(g *game.Game, obj *game.StackObject, condition opt.V[game.CardSelection]) bool {
	if !condition.Exists {
		return true
	}
	cond := condition.Val
	if cond.Card.Kind != game.CardReferenceLinked || cond.Card.LinkID == "" {
		return false
	}
	for _, ref := range linkedObjects(g, linkedObjectSourceKey(g, obj, cond.Card.LinkID)) {
		if ref.CardID == 0 {
			continue
		}
		card, ok := g.GetCardInstance(ref.CardID)
		if ok && cardMatchesSelection(g, obj, card, cond.Selection) {
			return true
		}
	}
	return false
}

// cardConditionPredicateSatisfied reports whether card satisfies an optional
// CardSelection's predicate. It ignores the selection's Card reference (the
// caller has already resolved which card to test) and matches anything when no
// condition is present.
func cardConditionPredicateSatisfied(g *game.Game, obj *game.StackObject, card *game.CardInstance, condition opt.V[game.CardSelection]) bool {
	if !condition.Exists {
		return true
	}
	return cardMatchesSelection(g, obj, card, condition.Val.Selection)
}

// cardMatchesSelection matches a card in a non-battlefield zone against a
// Selection, reading any chosen-subtype provenance from the resolving object's
// choices.
func cardMatchesSelection(g *game.Game, obj *game.StackObject, card *game.CardInstance, selection game.Selection) bool {
	if card == nil || card.Def == nil {
		return false
	}
	subject := &selectionSubject{
		kind:       subjectCard,
		g:          g,
		card:       card,
		controller: card.Owner,
	}
	if obj != nil {
		subject.resolutionChoices = obj.ResolutionChoices
		subject.viewer = obj.Controller
		subject.sourceObjectID = obj.SourceID
	}
	return matchSelection(subject, &selection)
}

func instructionResultGateSatisfied(obj *game.StackObject, gate game.InstructionResultGate) bool {
	if gate.Key == "" {
		return true
	}
	if obj == nil || obj.ResolutionResults == nil {
		return false
	}
	result, ok := obj.ResolutionResults[string(gate.Key)]
	if !ok {
		return false
	}
	if gate.Accepted != game.TriAny && (gate.Accepted == game.TriTrue) != result.Accepted {
		return false
	}
	if gate.Succeeded != game.TriAny && (gate.Succeeded == game.TriTrue) != result.Succeeded {
		return false
	}
	if gate.AmountRange.Exists &&
		(result.Amount < gate.AmountRange.Val.Min || result.Amount > gate.AmountRange.Val.Max) {
		return false
	}
	return true
}

func rememberInstructionResolutionResult(obj *game.StackObject, linkID string, accepted, succeeded bool, amount int) {
	if obj == nil || linkID == "" {
		return
	}
	if obj.ResolutionResults == nil {
		obj.ResolutionResults = make(map[string]game.InstructionResolutionResult)
	}
	obj.ResolutionResults[linkID] = game.InstructionResolutionResult{
		Accepted:  accepted,
		Succeeded: succeeded,
		Amount:    amount,
	}
}
