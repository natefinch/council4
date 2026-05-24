package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
)

func (e *Engine) resolveSpellEffects(g *game.Game, obj *game.StackObject, card *game.CardInstance, log *TurnLog) {
	e.resolveSpellEffectsWithChoices(g, obj, card, [game.NumPlayers]PlayerAgent{}, log)
}

func (e *Engine) resolveSpellEffectsWithChoices(g *game.Game, obj *game.StackObject, card *game.CardInstance, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	if e.resolveCardImplementationSpell(g, obj, card, log) {
		return
	}
	ability := firstSpellAbility(card.Def)
	if ability == nil {
		return
	}
	if len(ability.Modes) > 0 {
		for _, modeIndex := range obj.ChosenModes {
			if modeIndex < 0 || modeIndex >= len(ability.Modes) {
				continue
			}
			for _, effect := range ability.Modes[modeIndex].Effects {
				e.resolveEffectWithChoices(g, obj, effect, agents, log)
			}
		}
		return
	}
	for _, effect := range ability.Effects {
		e.resolveEffectWithChoices(g, obj, effect, agents, log)
	}
	if obj.KickerPaid {
		for _, effect := range ability.KickerEffects {
			e.resolveEffectWithChoices(g, obj, effect, agents, log)
		}
	}
}

func spellHasKicker(card *game.CardDef) bool {
	ability := firstSpellAbility(card)
	return ability != nil && ability.KickerCost != nil
}

func firstSpellAbility(card *game.CardDef) *game.AbilityDef {
	for i := range card.Abilities {
		if card.Abilities[i].Kind == game.SpellAbility {
			return &card.Abilities[i]
		}
	}
	return nil
}

func (e *Engine) resolveEffect(g *game.Game, obj *game.StackObject, effect game.Effect, log *TurnLog) {
	e.resolveEffectWithChoices(g, obj, effect, [game.NumPlayers]PlayerAgent{}, log)
}

func (e *Engine) resolveEffectWithChoices(g *game.Game, obj *game.StackObject, effect game.Effect, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	if !effectConditionSatisfied(g, obj, effect.Condition) {
		return
	}
	amount := effectAmount(g, obj, effect)
	if !IsEffectTypeExecuted(effect.Type) {
		logUnsupportedEffect(log, obj, effect)
		rememberEffectAmount(obj, effect, amount)
		return
	}
	if effect.Selector != game.EffectSelectorNone {
		resolveMassPermanentEffect(g, obj, effect, amount)
		rememberEffectAmount(obj, effect, amount)
		return
	}
	switch effect.Type {
	case game.EffectDraw:
		if amount <= 0 {
			return
		}
		playerID, ok := effectPlayer(g, obj, effect)
		if !ok {
			return
		}
		e.drawCards(g, playerID, amount, log)

	case game.EffectGainLife:
		if amount <= 0 {
			return
		}
		playerID, ok := effectPlayer(g, obj, effect)
		if !ok {
			return
		}
		player := g.Players[playerID]
		player.Life += amount
	case game.EffectLoseLife:
		if amount <= 0 {
			return
		}
		playerID, ok := effectPlayer(g, obj, effect)
		if !ok {
			return
		}
		player := g.Players[playerID]
		player.Life -= amount
	case game.EffectAddMana:
		if amount <= 0 {
			amount = 1
		}
		player := playerByID(g, obj.Controller)
		if player == nil || player.Eliminated {
			return
		}
		if stackObjectSourceIsSnow(g, obj) {
			player.ManaPool.AddSnow(effect.ManaColor, amount)
		} else {
			player.ManaPool.Add(effect.ManaColor, amount)
		}
	case game.EffectDamage:
		if amount <= 0 {
			return
		}
		if playerID, ok := effectPlayer(g, obj, effect); ok {
			sourceID, sourceObjectID := damageSourceIDs(g, obj)
			dealPlayerDamage(g, sourceID, sourceObjectID, obj.Controller, playerID, amount, false)
			return
		}
		permanent := effectPermanent(g, obj, effect)
		if permanent == nil {
			return
		}
		sourceID, sourceObjectID := damageSourceIDs(g, obj)
		dealPermanentDamage(g, sourceID, sourceObjectID, obj.Controller, permanent, amount, false)
	case game.EffectDestroy:
		permanent := effectPermanent(g, obj, effect)
		if permanent == nil {
			return
		}
		destroyPermanent(g, permanent.ObjectID)
	case game.EffectExile:
		permanent := effectPermanent(g, obj, effect)
		if permanent == nil {
			return
		}
		linkedObjectRef := permanentLinkedObjectRef(permanent)
		movePermanentToZone(g, permanent, game.ZoneExile)
		if effect.LinkID != "" {
			rememberLinkedObject(g, linkedObjectSourceKey(g, obj, effect.LinkID), linkedObjectRef)
		}
	case game.EffectBounce:
		permanent := effectPermanent(g, obj, effect)
		if permanent == nil {
			return
		}
		movePermanentToZone(g, permanent, game.ZoneHand)
	case game.EffectSacrifice:
		permanent := effectPermanent(g, obj, effect)
		if permanent == nil {
			permanent = firstPermanentControlledBy(g, obj.Controller)
		}
		if permanent == nil || effectiveController(g, permanent) != obj.Controller {
			return
		}
		movePermanentToZone(g, permanent, game.ZoneGraveyard)
	case game.EffectTap:
		permanent := effectPermanent(g, obj, effect)
		if permanent != nil {
			permanent.Tapped = true
		}
	case game.EffectUntap:
		permanent := effectPermanent(g, obj, effect)
		if permanent != nil {
			permanent.Tapped = false
		}
	case game.EffectModifyPT:
		permanent := effectPermanent(g, obj, effect)
		if permanent != nil && effect.UntilEndOfTurn {
			g.ContinuousEffects = append(g.ContinuousEffects, untilEndOfTurnContinuousEffect(g, obj, permanent, effect))
		}
	case game.EffectAddCounter:
		permanent := effectPermanent(g, obj, effect)
		if permanent != nil && amount > 0 {
			permanent.Counters.Add(effect.CounterKind, amount)
		}
	case game.EffectRemoveCounter:
		permanent := effectPermanent(g, obj, effect)
		if permanent != nil && amount > 0 {
			permanent.Counters.Remove(effect.CounterKind, amount)
		}
	case game.EffectMoveCounters:
		moveCounters(g, obj, effect)
	case game.EffectApplyContinuous:
		applyContinuousEffectTemplates(g, obj, effectPermanent(g, obj, effect), effect)
	case game.EffectCreateToken:
		if amount <= 0 {
			amount = 1
		}
		for range amount {
			createTokenPermanent(g, obj.Controller, effect.Token)
		}
	case game.EffectCreateDelayedTrigger:
		scheduleDelayedTrigger(g, obj, effect.DelayedTrigger)
	case game.EffectPutOnBattlefield:
		if effect.LinkID != "" {
			returnLinkedExiledObjects(g, obj, effect.LinkID)
		}
	case game.EffectPrevent:
		createPreventionShield(g, obj, effect)
	case game.EffectRegenerate:
		permanent := effectPermanent(g, obj, effect)
		if permanent != nil {
			permanent.RegenerationShields++
		}
	case game.EffectSkipStep:
		playerID, ok := effectPlayer(g, obj, effect)
		if ok {
			scheduleSkipStep(g, playerID, effect.Step)
		}
	case game.EffectTransform:
		permanent := effectPermanent(g, obj, effect)
		if permanent != nil {
			// TODO: apply back-face copiable values when double-faced card data exists.
			permanent.Transformed = !permanent.Transformed
		}
	case game.EffectPhaseOut:
		permanent := effectPermanent(g, obj, effect)
		if permanent != nil {
			permanent.PhasedOut = true
			removePermanentFromCombat(g, permanent.ObjectID)
		}
	case game.EffectCreateEmblem:
		g.Emblems = append(g.Emblems, game.Emblem{Owner: obj.Controller, Abilities: append([]game.AbilityDef(nil), effect.EmblemAbilities...)})
	case game.EffectMill:
		playerID, ok := effectPlayer(g, obj, effect)
		if ok {
			millCards(g, playerID, amount)
		}
	case game.EffectScry:
		playerID, ok := effectPlayer(g, obj, effect)
		if ok {
			e.scryCards(g, agents, log, playerID, amount)
		}
	case game.EffectSurveil:
		playerID, ok := effectPlayer(g, obj, effect)
		if ok {
			e.surveilCards(g, agents, log, playerID, amount)
		}
	case game.EffectFight:
		resolveFight(g, obj, effect)
	}
	rememberEffectAmount(obj, effect, amount)
}

// IsEffectTypeExecuted reports whether the generic rules resolver currently
// implements the given effect primitive.
func IsEffectTypeExecuted(effectType game.EffectType) bool {
	switch effectType {
	case game.EffectDraw,
		game.EffectGainLife,
		game.EffectLoseLife,
		game.EffectAddMana,
		game.EffectDamage,
		game.EffectDestroy,
		game.EffectExile,
		game.EffectBounce,
		game.EffectSacrifice,
		game.EffectTap,
		game.EffectUntap,
		game.EffectModifyPT,
		game.EffectAddCounter,
		game.EffectRemoveCounter,
		game.EffectMoveCounters,
		game.EffectApplyContinuous,
		game.EffectCreateToken,
		game.EffectCreateDelayedTrigger,
		game.EffectPutOnBattlefield,
		game.EffectPrevent,
		game.EffectRegenerate,
		game.EffectSkipStep,
		game.EffectTransform,
		game.EffectPhaseOut,
		game.EffectCreateEmblem,
		game.EffectMill,
		game.EffectScry,
		game.EffectSurveil,
		game.EffectFight:
		return true
	default:
		return false
	}
}

func (e *Engine) drawCards(g *game.Game, playerID game.PlayerID, amount int, log *TurnLog) {
	if amount <= 0 {
		return
	}
	for range amount {
		cardID, ok := e.drawCard(g, playerID)
		log.addDraw(DrawLog{
			Player: playerID,
			CardID: cardID,
			Failed: !ok,
		})
	}
}

func stackObjectSourceIsSnow(g *game.Game, obj *game.StackObject) bool {
	if g == nil || obj == nil {
		return false
	}
	permanent := permanentByObjectID(g, obj.SourceID)
	return permanentIsSnow(g, permanent)
}

func resolveMassPermanentEffect(g *game.Game, obj *game.StackObject, effect game.Effect, amount int) {
	permanentIDs := selectedPermanentIDs(g, obj.Controller, nil, effect.Selector)
	sourceID, sourceObjectID := damageSourceIDs(g, obj)
	for _, permanentID := range permanentIDs {
		permanent := permanentByObjectID(g, permanentID)
		if permanent == nil {
			continue
		}
		switch effect.Type {
		case game.EffectDamage:
			if amount > 0 {
				dealPermanentDamage(g, sourceID, sourceObjectID, obj.Controller, permanent, amount, false)
			}
		case game.EffectDestroy:
			destroyPermanent(g, permanent.ObjectID)
		case game.EffectExile:
			movePermanentToZone(g, permanent, game.ZoneExile)
		case game.EffectBounce:
			movePermanentToZone(g, permanent, game.ZoneHand)
		case game.EffectTap:
			permanent.Tapped = true
		case game.EffectUntap:
			permanent.Tapped = false
		case game.EffectAddCounter:
			if amount > 0 {
				permanent.Counters.Add(effect.CounterKind, amount)
			}
		case game.EffectRemoveCounter:
			if amount > 0 {
				permanent.Counters.Remove(effect.CounterKind, amount)
			}
		case game.EffectApplyContinuous:
			applyContinuousEffectTemplates(g, obj, permanent, effect)
		}
	}
}

func logUnsupportedEffect(log *TurnLog, obj *game.StackObject, effect game.Effect) {
	log.addUnsupportedEffect(UnsupportedEffectLog{
		StackObjectID: stackObjectID(obj),
		SourceID:      stackObjectSourceID(obj),
		Controller:    stackObjectController(obj),
		EffectType:    effect.Type,
		Description:   effect.Description,
	})
}

func effectAmount(g *game.Game, obj *game.StackObject, effect game.Effect) int {
	if effect.DynamicAmount == nil || effect.DynamicAmount.Kind == game.DynamicAmountNone {
		return effect.Amount
	}
	dynamic := effect.DynamicAmount
	amount := 0
	switch dynamic.Kind {
	case game.DynamicAmountConstant:
		amount = dynamic.Constant
	case game.DynamicAmountX:
		if obj != nil {
			amount = obj.XValue
		}
	case game.DynamicAmountTargetPower:
		if permanent := effectPermanent(g, obj, game.Effect{TargetIndex: dynamic.TargetIndex}); permanent != nil {
			amount = effectivePower(g, permanent)
		}
	case game.DynamicAmountTargetToughness:
		if permanent := effectPermanent(g, obj, game.Effect{TargetIndex: dynamic.TargetIndex}); permanent != nil {
			if toughness, ok := effectiveToughness(g, permanent); ok {
				amount = toughness
			}
		}
	case game.DynamicAmountTargetManaValue:
		if permanent := effectPermanent(g, obj, game.Effect{TargetIndex: dynamic.TargetIndex}); permanent != nil {
			if def := permanentCardDef(g, permanent); def != nil {
				amount = def.ManaValue
			}
		}
	case game.DynamicAmountTargetCounters:
		if permanent := effectPermanent(g, obj, game.Effect{TargetIndex: dynamic.TargetIndex}); permanent != nil {
			amount = permanent.Counters.Get(dynamic.CounterKind)
		}
	case game.DynamicAmountControllerLife:
		if player := playerByID(g, stackObjectController(obj)); player != nil {
			amount = player.Life
		}
	case game.DynamicAmountControllerHandSize:
		if player := playerByID(g, stackObjectController(obj)); player != nil {
			amount = player.Hand.Size()
		}
	case game.DynamicAmountControllerGraveyardSize:
		if player := playerByID(g, stackObjectController(obj)); player != nil {
			amount = player.Graveyard.Size()
		}
	case game.DynamicAmountCountSelector:
		amount = len(selectedPermanentIDs(g, stackObjectController(obj), nil, dynamic.Selector))
	case game.DynamicAmountPreviousEffectResult:
		if obj != nil && dynamic.LinkID != "" {
			amount = obj.ResolvedAmounts[dynamic.LinkID]
		}
	}
	multiplier := dynamic.Multiplier
	if multiplier == 0 {
		multiplier = 1
	}
	return amount * multiplier
}

func rememberEffectAmount(obj *game.StackObject, effect game.Effect, amount int) {
	if obj == nil || effect.LinkID == "" {
		return
	}
	if obj.ResolvedAmounts == nil {
		obj.ResolvedAmounts = make(map[string]int)
	}
	obj.ResolvedAmounts[effect.LinkID] = amount
}

func moveCounters(g *game.Game, obj *game.StackObject, effect game.Effect) {
	destination := effectPermanent(g, obj, effect)
	if destination == nil {
		return
	}
	counters, source := effectCounterSource(g, obj, effect.CounterSource)
	if counters.IsEmpty() {
		return
	}
	if source != nil && source.ObjectID == destination.ObjectID {
		return
	}
	for kind, amount := range counters.All() {
		destination.Counters.Add(kind, amount)
	}
	if source == nil {
		return
	}
	for kind, amount := range counters.All() {
		source.Counters.Remove(kind, amount)
	}
}

func effectCounterSource(g *game.Game, obj *game.StackObject, source game.CounterSourceSpec) (counter.Set, *game.Permanent) {
	switch source.Kind {
	case game.CounterSourceTarget:
		permanent := effectPermanent(g, obj, game.Effect{TargetIndex: source.TargetIndex})
		if permanent == nil {
			return counter.Set{}, nil
		}
		return cloneCounters(permanent.Counters), permanent
	case game.CounterSourceEventPermanent:
		if obj == nil || !obj.HasTriggerEvent || obj.TriggerEvent.PermanentID == 0 {
			return counter.Set{}, nil
		}
		if permanent := permanentByObjectID(g, obj.TriggerEvent.PermanentID); permanent != nil {
			return cloneCounters(permanent.Counters), permanent
		}
		if snapshot, ok := lastKnownObject(g, obj.TriggerEvent.PermanentID); ok {
			return cloneCounters(snapshot.Counters), nil
		}
	}
	return counter.Set{}, nil
}

func effectConditionSatisfied(g *game.Game, obj *game.StackObject, condition *game.EffectCondition) bool {
	if condition == nil {
		return true
	}
	if condition.MatchPermanentType {
		permanent := effectPermanent(g, obj, game.Effect{TargetIndex: condition.TargetIndex})
		if permanent == nil {
			return false
		}
		matches := permanentHasType(g, permanent, condition.PermanentType)
		if condition.Negate {
			matches = !matches
		}
		if !matches {
			return false
		}
	}
	return true
}

func applyContinuousEffectTemplates(g *game.Game, obj *game.StackObject, permanent *game.Permanent, effect game.Effect) {
	if g == nil || obj == nil || len(effect.ContinuousEffects) == 0 {
		return
	}
	sourceID, sourceObjectID := damageSourceIDs(g, obj)
	timestamp := int64(g.IDGen.Next())
	for _, template := range effect.ContinuousEffects {
		runtimeEffect := template
		runtimeEffect.ID = g.IDGen.Next()
		runtimeEffect.SourceCardID = sourceID
		runtimeEffect.SourceObjectID = sourceObjectID
		runtimeEffect.Controller = obj.Controller
		runtimeEffect.Timestamp = timestamp
		runtimeEffect.CreatedTurn = g.Turn.TurnNumber
		if effect.UntilEndOfTurn {
			runtimeEffect.Duration = game.DurationUntilEndOfTurn
		} else if effect.Duration != game.DurationPermanent {
			runtimeEffect.Duration = effect.Duration
		}
		if runtimeEffect.Duration == game.DurationUntilYourNextTurn && runtimeEffect.ExpiresFor == game.Player1 {
			runtimeEffect.ExpiresFor = obj.Controller
		}
		if runtimeEffect.AffectedObjectID == 0 && runtimeEffect.Selector == game.EffectSelectorNone {
			if permanent == nil {
				continue
			}
			runtimeEffect.AffectedObjectID = permanent.ObjectID
		}
		g.ContinuousEffects = append(g.ContinuousEffects, runtimeEffect)
	}
}

func damageSourceIDs(g *game.Game, obj *game.StackObject) (id.ID, id.ID) {
	if obj == nil {
		return 0, 0
	}
	switch obj.Kind {
	case game.StackActivatedAbility, game.StackTriggeredAbility:
		if obj.SourceCardID != 0 {
			return obj.SourceCardID, obj.SourceID
		}
		permanent := permanentByObjectID(g, obj.SourceID)
		if permanent == nil {
			return 0, obj.SourceID
		}
		return permanent.CardInstanceID, permanent.ObjectID
	default:
		return obj.SourceID, 0
	}
}

func selectedPermanentIDs(g *game.Game, controller game.PlayerID, source *game.Permanent, selector game.EffectSelector) []id.ID {
	if g == nil {
		return nil
	}
	permanentIDs := make([]id.ID, 0, len(g.Battlefield))
	for _, permanent := range g.Battlefield {
		if permanent == nil || !permanentMatchesSelectorForSource(g, source, controller, permanent, selector) {
			continue
		}
		permanentIDs = append(permanentIDs, permanent.ObjectID)
	}
	return permanentIDs
}

func permanentMatchesSelector(g *game.Game, permanent *game.Permanent, selector game.EffectSelector) bool {
	return permanentMatchesSelectorForSource(g, nil, 0, permanent, selector)
}

func permanentMatchesSelectorForSource(g *game.Game, source *game.Permanent, controller game.PlayerID, permanent *game.Permanent, selector game.EffectSelector) bool {
	if permanent == nil {
		return false
	}
	switch selector {
	case game.EffectSelectorAllCreatures:
		return permanentHasType(g, permanent, game.TypeCreature)
	case game.EffectSelectorAllArtifacts:
		return permanentHasType(g, permanent, game.TypeArtifact)
	case game.EffectSelectorAllEnchantments:
		return permanentHasType(g, permanent, game.TypeEnchantment)
	case game.EffectSelectorAllNonlandPermanents:
		return !permanentHasType(g, permanent, game.TypeLand)
	case game.EffectSelectorAllPermanents:
		return true
	case game.EffectSelectorCreaturesYouControl:
		return effectiveController(g, permanent) == controller && permanentHasType(g, permanent, game.TypeCreature)
	case game.EffectSelectorOtherCreaturesYouControl:
		return source != nil && permanent.ObjectID != source.ObjectID && effectiveController(g, permanent) == controller && permanentHasType(g, permanent, game.TypeCreature)
	default:
		return false
	}
}

func effectPlayer(g *game.Game, obj *game.StackObject, effect game.Effect) (game.PlayerID, bool) {
	if effect.TargetIndex == -1 {
		if !isPlayerAlive(g, obj.Controller) {
			return 0, false
		}
		return obj.Controller, true
	}
	if effect.TargetIndex < 0 || effect.TargetIndex >= len(obj.Targets) {
		return 0, false
	}
	target := obj.Targets[effect.TargetIndex]
	if target.Kind != game.TargetPlayer {
		return 0, false
	}
	if !isPlayerAlive(g, target.PlayerID) {
		return 0, false
	}
	return target.PlayerID, true
}

func effectPermanent(g *game.Game, obj *game.StackObject, effect game.Effect) *game.Permanent {
	if effect.TargetIndex < 0 || effect.TargetIndex >= len(obj.Targets) {
		return nil
	}
	target := obj.Targets[effect.TargetIndex]
	if target.Kind != game.TargetPermanent {
		return nil
	}
	return permanentByObjectID(g, target.PermanentID)
}

func firstPermanentControlledBy(g *game.Game, controller game.PlayerID) *game.Permanent {
	if g == nil {
		return nil
	}
	for _, permanent := range g.Battlefield {
		if permanent != nil && effectiveController(g, permanent) == controller {
			return permanent
		}
	}
	return nil
}

func permanentLinkedObjectRef(permanent *game.Permanent) game.LinkedObjectRef {
	if permanent == nil || permanent.CardInstanceID == 0 {
		return game.LinkedObjectRef{}
	}
	return game.LinkedObjectRef{ObjectID: permanent.ObjectID, CardID: permanent.CardInstanceID}
}

func returnLinkedExiledObjects(g *game.Game, obj *game.StackObject, linkID string) {
	key := linkedObjectSourceKey(g, obj, linkID)
	for _, ref := range linkedObjects(g, key) {
		if snapshot, ok := lastKnownObject(g, ref.ObjectID); !ok || snapshot.CardID != ref.CardID {
			continue
		}
		card := g.GetCardInstance(ref.CardID)
		if card == nil {
			continue
		}
		owner := playerByID(g, card.Owner)
		if owner == nil || !owner.Exile.Remove(ref.CardID) {
			continue
		}
		createCardPermanent(g, card, obj.Controller, game.ZoneExile)
	}
	clearLinkedObjects(g, key)
}

func createTokenPermanent(g *game.Game, controller game.PlayerID, token *game.CardDef) *game.Permanent {
	if g == nil || token == nil {
		return nil
	}
	objectID := g.IDGen.Next()
	permanent := &game.Permanent{
		ObjectID:      objectID,
		Owner:         controller,
		Controller:    controller,
		SummoningSick: entersSummoningSick(token),
		Timestamp:     int64(objectID),
		Token:         true,
		TokenDef:      token,
	}
	initializePermanentCounters(permanent, token)
	g.Battlefield = append(g.Battlefield, permanent)
	event := game.GameEvent{
		Controller:  controller,
		Player:      controller,
		PermanentID: objectID,
		TokenName:   token.Name,
		TokenDef:    token,
		FromZone:    game.ZoneNone,
		ToZone:      game.ZoneBattlefield,
	}
	emitZoneChangeEvent(g, event)
	event.Kind = game.EventPermanentEnteredBattlefield
	emitEvent(g, event)
	return permanent
}
