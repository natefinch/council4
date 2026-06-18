package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
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
	return dynamicAmountValueBeforeLayer(g, obj, controller, dynamic, 0)
}

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
	case game.DynamicAmountControllerLife:
		if player, ok := playerByID(g, controller); ok {
			amount = player.Life
		}
	case game.DynamicAmountControllerHandSize:
		if player, ok := playerByID(g, controller); ok {
			amount = cardInstanceCount(g, player.Hand.All())
		}
	case game.DynamicAmountControllerGraveyardSize:
		if player, ok := playerByID(g, controller); ok {
			amount = cardInstanceCount(g, player.Graveyard.All())
		}
	case game.DynamicAmountControllerBasicLandTypeCount:
		amount = controllerBasicLandTypeCount(g, conditionContext{
			controller:            controller,
			characteristicsBefore: before,
		})
	case game.DynamicAmountCountSelector:
		amount = countPermanentsMatchingGroup(g, obj, controller, dynamic.Group)
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
	case game.DynamicAmountOpponentCount:
		amount = len(aliveOpponents(g, controller))
	case game.DynamicAmountEventDamage:
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
	default:
	}
	multiplier := dynamic.Multiplier
	if multiplier == 0 {
		multiplier = 1
	}
	return amount * multiplier
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
		if ok && cardMatchesCondition(card.Def, condition) {
			return true
		}
	}
	return false
}

func cardMatchesCondition(card *game.CardDef, condition opt.V[game.CardCondition]) bool {
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
