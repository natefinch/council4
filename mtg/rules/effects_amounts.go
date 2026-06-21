package rules

import (
	"slices"

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

//nolint:gocritic // Value semantics keep dynamic expressions immutable during evaluation.
func dynamicAmountValue(g *game.Game, obj *game.StackObject, controller game.PlayerID, dynamic game.DynamicAmount) int {
	return dynamicAmountValueBeforeLayer(g, obj, controller, dynamic, 0)
}

//nolint:gocritic // Value semantics keep dynamic expressions immutable during evaluation.
func dynamicAmountValueBeforeLayer(g *game.Game, obj *game.StackObject, controller game.PlayerID, dynamic game.DynamicAmount, before game.ContinuousLayer) int {
	amount := 0
	switch dynamic.Kind {
	case game.DynamicAmountConstant:
		amount = dynamic.Constant
	case game.DynamicAmountX:
		if obj != nil {
			amount = obj.XValue
		}
	case game.DynamicAmountTargetPower:
		if obj == nil {
			break
		}
		if resolved, ok := resolveObjectReference(g, obj, dynamic.Object); ok && resolved.permanent != nil {
			permanent := resolved.permanent
			amount = effectivePower(g, permanent)
		}
	case game.DynamicAmountTargetToughness:
		if obj == nil {
			break
		}
		if resolved, ok := resolveObjectReference(g, obj, dynamic.Object); ok && resolved.permanent != nil {
			permanent := resolved.permanent
			if toughness, ok := effectiveToughness(g, permanent); ok {
				amount = toughness
			}
		}
	case game.DynamicAmountTargetManaValue:
		if obj == nil {
			break
		}
		if resolved, ok := resolveObjectReference(g, obj, dynamic.Object); ok && resolved.permanent != nil {
			permanent := resolved.permanent
			if def, ok := permanentCardDef(g, permanent); ok {
				amount = def.ManaValue()
			}
		}
	case game.DynamicAmountTargetCounters:
		if obj == nil {
			break
		}
		if resolved, ok := resolveObjectReference(g, obj, dynamic.Object); ok && resolved.permanent != nil {
			permanent := resolved.permanent
			amount = permanent.Counters.Get(dynamic.CounterKind)
		}
	case game.DynamicAmountControllerLife, game.DynamicAmountControllerHandSize,
		game.DynamicAmountControllerGraveyardSize, game.DynamicAmountControllerBasicLandTypeCount,
		game.DynamicAmountOpponentCount:
		amount = controllerAggregateAmount(g, controller, dynamic, before)
	case game.DynamicAmountDevotion:
		// ColorFrom binds devotion to the color chosen as the ability resolves
		// (Nykthos, Shrine to Nyx's "devotion to that color"); otherwise the
		// amount's fixed Colors apply. A missing or unreadable choice yields no
		// colors, so devotion is zero.
		colors := dynamic.Colors
		if dynamic.ColorFrom != "" {
			colors = nil
			if result, ok := linkedResolutionChoice(obj, string(dynamic.ColorFrom)); ok {
				if chosen, ok := manaColor(result.Color); ok {
					colors = []color.Color{chosen}
				}
			}
		}
		amount = controllerDevotion(g, controller, colors)
	case game.DynamicAmountCountSelector, game.DynamicAmountGreatestPowerInGroup,
		game.DynamicAmountGreatestToughnessInGroup, game.DynamicAmountGreatestManaValueInGroup,
		game.DynamicAmountTotalPowerInGroup, game.DynamicAmountTotalToughnessInGroup,
		game.DynamicAmountColorCountInGroup:
		amount = groupDynamicAmount(g, obj, controller, &dynamic)
	case game.DynamicAmountCountCardsInZone:
		if dynamic.Player != nil && dynamic.Selection != nil {
			amount = countCardsInZoneMatchingSelection(g, obj, controller, *dynamic.Player, dynamic.CardZone, *dynamic.Selection)
		}
	case game.DynamicAmountPreviousEffectResult:
		key := dynamicResultKey(dynamic)
		if obj != nil && key != "" {
			amount = obj.ResolvedAmounts[key]
		}
	case game.DynamicAmountPreviousEffectExcessDamage:
		key := dynamicResultKey(dynamic)
		if obj != nil && key != "" {
			amount = obj.ResolvedExcessDamage[key]
		}
	case game.DynamicAmountEventDamage, game.DynamicAmountEventLifeChange:
		if obj != nil && obj.HasTriggerEvent {
			amount = obj.TriggerEvent.Amount
		}
	case game.DynamicAmountEventCardCount:
		if obj != nil && obj.HasTriggerEvent {
			amount = triggerEventCardCount(g, obj.TriggerEvent)
		}
	case game.DynamicAmountObjectPower:
		if obj == nil {
			break
		}
		if resolved, ok := resolveObjectReference(g, obj, dynamic.Object); ok {
			amount = resolvedObjectPower(g, &resolved)
		}
	case game.DynamicAmountObjectToughness:
		if obj == nil {
			break
		}
		if resolved, ok := resolveObjectReference(g, obj, dynamic.Object); ok {
			amount = resolvedObjectToughness(g, &resolved)
		}
	case game.DynamicAmountObjectManaValue, game.DynamicAmountCapturedTargetManaValue:
		amount = dynamicObjectManaValue(g, obj, &dynamic)
	case game.DynamicAmountObjectCounters:
		if obj == nil {
			break
		}
		if resolved, ok := resolveObjectReference(g, obj, dynamic.Object); ok {
			if resolved.permanent != nil {
				amount = resolved.permanent.Counters.Get(dynamic.CounterKind)
			} else {
				amount = resolved.snapshot.Counters.Get(dynamic.CounterKind)
			}
		}
	case game.DynamicAmountChosenNumber:
		if choice, ok := linkedResolutionChoice(obj, string(dynamic.ResultKey)); ok &&
			choice.Kind == game.ResolutionChoiceNumber {
			amount = choice.Number
		}
	case game.DynamicAmountSpellsCastThisTurn:
		amount = spellsCastThisTurn(g, controller)
	default:
	}
	multiplier := dynamic.Multiplier
	if multiplier == 0 {
		multiplier = 1
	}
	return amount * multiplier
}

// spellsCastThisTurn counts the spells the controller has cast so far this turn
// from the turn's recorded spell-cast events (CR 608.2c). A triggered ability's
// own triggering spell counts, because its cast event precedes the ability's
// resolution.
func spellsCastThisTurn(g *game.Game, controller game.PlayerID) int {
	count := 0
	for _, event := range g.EventsThisTurn() {
		if event.Kind == game.EventSpellCast && event.Controller == controller {
			count++
		}
	}
	return count
}

// controllerAggregateAmount computes the player-relative dynamic amounts that
// depend only on the controller's own board and zones (life total, hand and
// graveyard sizes, basic land type and opponent counts, and devotion). It is
// split out of dynamicAmountValueBeforeLayer so that large switch stays within
// the maintainability budget; behavior is identical to the inlined cases.
//
//nolint:gocritic // Value semantics keep dynamic expressions immutable during evaluation.
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
	default:
	}
	return 0
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

// groupDynamicAmount dispatches the battlefield-group amounts, each derived from
// the permanents of dynamic.Group as the effect resolves (CR 608.2c): the member
// count, the greatest or total power/toughness/mana value, and the distinct
// color count.
func groupDynamicAmount(g *game.Game, obj *game.StackObject, controller game.PlayerID, dynamic *game.DynamicAmount) int {
	switch dynamic.Kind {
	case game.DynamicAmountGreatestPowerInGroup, game.DynamicAmountGreatestToughnessInGroup,
		game.DynamicAmountGreatestManaValueInGroup:
		return greatestCharacteristicInGroup(g, obj, controller, dynamic.Group, dynamic.Kind)
	case game.DynamicAmountTotalPowerInGroup, game.DynamicAmountTotalToughnessInGroup:
		return totalCharacteristicInGroup(g, obj, controller, dynamic.Group, dynamic.Kind)
	case game.DynamicAmountColorCountInGroup:
		return colorCountInGroup(g, obj, controller, dynamic.Group)
	default:
		return countPermanentsMatchingGroup(g, obj, controller, dynamic.Group)
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

func countCardsInZoneMatchingSelection(g *game.Game, obj *game.StackObject, controller game.PlayerID, playerRef game.PlayerReference, cardZone zone.Type, selection game.Selection) int {
	playerID, ok := resolvePlayerReference(g, obj, playerRef)
	if !ok {
		return 0
	}
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
			viewer:     controller,
		}
		if matchSelection(&subject, &selection) {
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

//nolint:gocritic // Kept by value to match the evaluator's immutable expression semantics.
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

func cardConditionSatisfied(g *game.Game, obj *game.StackObject, condition opt.V[game.CardCondition]) bool {
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
		if ok && cardMatchesCondition(card.Def, condition, obj) {
			return true
		}
	}
	return false
}

func cardMatchesCondition(card *game.CardDef, condition opt.V[game.CardCondition], obj *game.StackObject) bool {
	if !condition.Exists {
		return true
	}
	if card == nil {
		return false
	}
	cond := condition.Val
	if cond.RequirePermanentCard && !card.IsPermanent() {
		return false
	}
	face := card.DefaultFace()
	for _, cardType := range cond.Types {
		if !face.HasType(cardType) {
			return false
		}
	}
	for _, supertype := range cond.Supertypes {
		if !face.HasSupertype(supertype) {
			return false
		}
	}
	if len(cond.SubtypesAny) > 0 && !slices.ContainsFunc(cond.SubtypesAny, face.HasSubtype) {
		return false
	}
	if cond.ChosenSubtypeFrom != "" {
		choice, ok := linkedResolutionChoice(obj, string(cond.ChosenSubtypeFrom))
		if !ok ||
			choice.Kind != game.ResolutionChoiceSubtype ||
			!types.KnownSubtypeForType(types.Creature, choice.Subtype) ||
			!face.HasSubtype(choice.Subtype) {
			return false
		}
	}
	return true
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
