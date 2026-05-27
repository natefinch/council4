package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/opt"
)

func (e *Engine) resolveSpellEffects(g *game.Game, obj *game.StackObject, card *game.CardInstance, log *TurnLog) {
	e.resolveSpellEffectsWithChoices(g, obj, card, [game.NumPlayers]PlayerAgent{}, log)
}

func (e *Engine) resolveSpellEffectsWithChoices(g *game.Game, obj *game.StackObject, card *game.CardInstance, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	if e.resolveCardImplementationSpell(g, obj, card, log) {
		return
	}
	ability, ok := firstSpellAbility(card.Def)
	if !ok {
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
	ability, ok := firstSpellAbility(card)
	return ok && ability.KickerCost.Exists
}

func firstSpellAbility(card *game.CardDef) (*game.AbilityDef, bool) {
	for i := range card.Abilities {
		if card.Abilities[i].Kind == game.SpellAbility {
			return &card.Abilities[i], true
		}
	}
	return nil, false
}

func (e *Engine) resolveEffect(g *game.Game, obj *game.StackObject, effect game.Effect, log *TurnLog) {
	e.resolveEffectWithChoices(g, obj, effect, [game.NumPlayers]PlayerAgent{}, log)
}

func (e *Engine) resolveEffectWithChoices(g *game.Game, obj *game.StackObject, effect game.Effect, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	if !effectConditionSatisfied(g, obj, effect.Condition) {
		return
	}
	if !effectResultConditionSatisfied(obj, effect.ResultCondition) {
		return
	}
	amount := effectAmount(g, obj, effect)
	accepted := true
	succeeded := false
	// Record linked resolution state after the instruction is attempted so
	// follow-up "if you do" / "that much" effects see what actually happened
	// during sequential resolution (CR 608.2c; impossible actions CR 101.3).
	defer func() {
		if accepted && succeeded {
			rememberEffectAmount(obj, effect, amount)
		}
		rememberEffectResolutionResult(obj, effect, accepted, succeeded, amount)
	}()
	if effect.Optional && !e.chooseMay(g, agents, stackObjectController(obj), "Apply optional effect?", log) {
		accepted = false
		return
	}
	if effect.Choice.Exists {
		if !e.resolveResolutionChoice(g, obj, effect, agents, log) {
			return
		}
		succeeded = true
		if effect.Type == game.EffectChoose {
			return
		}
	}
	if effect.Payment.Exists {
		accepted, succeeded = e.resolveResolutionPayment(g, obj, effect, agents, log)
		if !succeeded || effect.Type == game.EffectPay {
			return
		}
	}
	if !IsEffectTypeExecuted(effect.Type) {
		logUnsupportedEffect(log, obj, effect)
		return
	}
	if effect.Selector != game.EffectSelectorNone {
		succeeded = resolveMassPermanentEffect(g, obj, effect, amount)
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
		succeeded = e.drawCards(g, playerID, amount, log)

	case game.EffectGainLife:
		if amount <= 0 {
			return
		}
		playerID, ok := effectPlayer(g, obj, effect)
		if !ok {
			return
		}
		succeeded = gainLife(g, playerID, amount) > 0
	case game.EffectLoseLife:
		if amount <= 0 {
			return
		}
		playerID, ok := effectPlayer(g, obj, effect)
		if !ok {
			return
		}
		succeeded = loseLife(g, playerID, amount) > 0
	case game.EffectAddMana:
		if amount <= 0 {
			amount = 1
		}
		player, ok := playerByID(g, obj.Controller)
		if !ok || player.Eliminated {
			return
		}
		if stackObjectSourceIsSnow(g, obj) {
			player.ManaPool.AddSnow(effectManaColor(obj, effect), amount)
		} else {
			player.ManaPool.Add(effectManaColor(obj, effect), amount)
		}
		succeeded = true
	case game.EffectDamage:
		if amount <= 0 {
			return
		}
		if playerID, ok := effectPlayer(g, obj, effect); ok {
			sourceID, sourceObjectID := damageSourceIDs(g, obj)
			dealPlayerDamage(g, sourceID, sourceObjectID, obj.Controller, playerID, amount, false)
			succeeded = true
			return
		}
		permanent, ok := effectPermanent(g, obj, effect)
		if !ok {
			return
		}
		sourceID, sourceObjectID := damageSourceIDs(g, obj)
		dealPermanentDamage(g, sourceID, sourceObjectID, obj.Controller, permanent, amount, false)
		succeeded = true
	case game.EffectDestroy:
		permanent, ok := effectPermanent(g, obj, effect)
		if !ok {
			return
		}
		_, succeeded = destroyPermanent(g, permanent.ObjectID)
	case game.EffectExile:
		permanent, ok := effectPermanent(g, obj, effect)
		if !ok {
			return
		}
		linkedObjectRef := permanentLinkedObjectRef(permanent)
		succeeded = movePermanentToZone(g, permanent, game.ZoneExile)
		if effect.LinkID != "" {
			rememberLinkedObject(g, linkedObjectSourceKey(g, obj, effect.LinkID), linkedObjectRef)
		}
	case game.EffectBounce:
		permanent, ok := effectPermanent(g, obj, effect)
		if !ok {
			return
		}
		succeeded = movePermanentToZone(g, permanent, game.ZoneHand)
	case game.EffectSacrifice:
		permanent, ok := effectPermanent(g, obj, effect)
		if !ok {
			permanent, ok = firstPermanentControlledBy(g, obj.Controller)
		}
		if !ok || effectiveController(g, permanent) != obj.Controller {
			return
		}
		succeeded = movePermanentToZone(g, permanent, game.ZoneGraveyard)
	case game.EffectTap:
		if permanent, ok := effectPermanent(g, obj, effect); ok {
			setPermanentTapped(g, permanent, true)
			succeeded = true
		}
	case game.EffectUntap:
		if permanent, ok := effectPermanent(g, obj, effect); ok {
			setPermanentTapped(g, permanent, false)
			succeeded = true
		}
	case game.EffectModifyPT:
		if permanent, ok := effectPermanent(g, obj, effect); ok && effect.UntilEndOfTurn {
			g.ContinuousEffects = append(g.ContinuousEffects, untilEndOfTurnContinuousEffect(g, obj, permanent, effect))
			succeeded = true
		}
	case game.EffectAddCounter:
		if permanent, ok := effectPermanent(g, obj, effect); ok && amount > 0 {
			permanent.Counters.Add(effect.CounterKind, amount)
			succeeded = true
		}
	case game.EffectRemoveCounter:
		if permanent, ok := effectPermanent(g, obj, effect); ok && amount > 0 {
			permanent.Counters.Remove(effect.CounterKind, amount)
			succeeded = true
		}
	case game.EffectMoveCounters:
		succeeded = moveCounters(g, obj, effect)
	case game.EffectApplyContinuous:
		permanent, _ := effectPermanent(g, obj, effect)
		succeeded = applyContinuousEffectTemplates(g, obj, permanent, effect)
	case game.EffectCreateToken:
		if amount <= 0 {
			amount = 1
		}
		if !effect.Token.Exists {
			return
		}
		for range amount {
			if _, ok := createTokenPermanent(g, obj.Controller, effect.Token.Val); !ok {
				return
			}
		}
		succeeded = amount > 0
	case game.EffectCreateDelayedTrigger:
		succeeded = effect.DelayedTrigger.Exists && scheduleDelayedTrigger(g, obj, &effect.DelayedTrigger.Val)
	case game.EffectPutOnBattlefield:
		if effect.LinkID != "" {
			succeeded = returnLinkedExiledObjects(g, obj, effect.LinkID)
		}
	case game.EffectPrevent:
		succeeded = createPreventionShield(g, obj, effect)
	case game.EffectRegenerate:
		if permanent, ok := effectPermanent(g, obj, effect); ok {
			permanent.RegenerationShields++
			succeeded = true
		}
	case game.EffectSkipStep:
		playerID, ok := effectPlayer(g, obj, effect)
		if ok {
			scheduleSkipStep(g, playerID, effect.Step)
			succeeded = true
		}
	case game.EffectTransform:
		if permanent, ok := effectPermanent(g, obj, effect); ok {
			// TODO: apply back-face copiable values when double-faced card data exists.
			permanent.Transformed = !permanent.Transformed
			succeeded = true
		}
	case game.EffectPhaseOut:
		if permanent, ok := effectPermanent(g, obj, effect); ok {
			permanent.PhasedOut = true
			removePermanentFromCombat(g, permanent.ObjectID)
			succeeded = true
		}
	case game.EffectCreateEmblem:
		g.Emblems = append(g.Emblems, game.Emblem{Owner: obj.Controller, Abilities: append([]game.AbilityDef(nil), effect.EmblemAbilities...)})
		succeeded = true
	case game.EffectMill:
		playerID, ok := effectPlayer(g, obj, effect)
		if ok {
			millCards(g, playerID, amount)
			succeeded = amount > 0
		}
	case game.EffectScry:
		playerID, ok := effectPlayer(g, obj, effect)
		if ok {
			e.scryCards(g, agents, log, playerID, amount)
			succeeded = amount > 0
		}
	case game.EffectSurveil:
		playerID, ok := effectPlayer(g, obj, effect)
		if ok {
			e.surveilCards(g, agents, log, playerID, amount)
			succeeded = amount > 0
		}
	case game.EffectFight:
		resolveFight(g, obj, effect)
		succeeded = true
	case game.EffectReplace:
		succeeded = createReplacementEffect(g, obj, effect)
	case game.EffectChoose, game.EffectPay:
		succeeded = true
	case game.EffectApplyRule:
		succeeded = createRuleEffects(g, obj, effect)
	case game.EffectProliferate:
		succeeded = e.resolveProliferate(g, obj, agents, log)
	case game.EffectGoad:
		if permanent, ok := effectPermanent(g, obj, effect); ok && permanentHasType(g, permanent, game.TypeCreature) {
			goadPermanent(g, permanent, obj.Controller)
			succeeded = true
		}
	}
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
		game.EffectFight,
		game.EffectReplace,
		game.EffectChoose,
		game.EffectPay,
		game.EffectApplyRule,
		game.EffectProliferate,
		game.EffectGoad:
		return true
	default:
		return false
	}
}

func (e *Engine) drawCards(g *game.Game, playerID game.PlayerID, amount int, log *TurnLog) bool {
	if amount <= 0 {
		return false
	}
	drew := false
	for range amount {
		cardID, ok := e.drawCard(g, playerID)
		drew = drew || ok
		log.addDraw(DrawLog{
			Player: playerID,
			CardID: cardID,
			Failed: !ok,
		})
	}
	return drew
}

func stackObjectSourceIsSnow(g *game.Game, obj *game.StackObject) bool {
	permanent, ok := permanentByObjectID(g, obj.SourceID)
	return ok && permanentIsSnow(g, permanent)
}

func resolveMassPermanentEffect(g *game.Game, obj *game.StackObject, effect game.Effect, amount int) bool {
	permanentIDs := selectedPermanentIDs(g, obj.Controller, nil, effect.Selector)
	sourceID, sourceObjectID := damageSourceIDs(g, obj)
	succeeded := false
	for _, permanentID := range permanentIDs {
		permanent, ok := permanentByObjectID(g, permanentID)
		if !ok {
			continue
		}
		switch effect.Type {
		case game.EffectDamage:
			if amount > 0 {
				dealPermanentDamage(g, sourceID, sourceObjectID, obj.Controller, permanent, amount, false)
				succeeded = true
			}
		case game.EffectDestroy:
			_, ok := destroyPermanent(g, permanent.ObjectID)
			succeeded = succeeded || ok
		case game.EffectExile:
			succeeded = movePermanentToZone(g, permanent, game.ZoneExile) || succeeded
		case game.EffectBounce:
			succeeded = movePermanentToZone(g, permanent, game.ZoneHand) || succeeded
		case game.EffectTap:
			setPermanentTapped(g, permanent, true)
			succeeded = true
		case game.EffectUntap:
			setPermanentTapped(g, permanent, false)
			succeeded = true
		case game.EffectAddCounter:
			if amount > 0 {
				permanent.Counters.Add(effect.CounterKind, amount)
				succeeded = true
			}
		case game.EffectRemoveCounter:
			if amount > 0 {
				permanent.Counters.Remove(effect.CounterKind, amount)
				succeeded = true
			}
		case game.EffectApplyContinuous:
			succeeded = applyContinuousEffectTemplates(g, obj, permanent, effect) || succeeded
		}
	}
	return succeeded
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
	if !effect.DynamicAmount.Exists || effect.DynamicAmount.Val.Kind == game.DynamicAmountNone {
		return effect.Amount
	}
	dynamic := effect.DynamicAmount.Val
	amount := 0
	switch dynamic.Kind {
	case game.DynamicAmountConstant:
		amount = dynamic.Constant
	case game.DynamicAmountX:
		amount = obj.XValue
	case game.DynamicAmountTargetPower:
		if permanent, ok := effectPermanent(g, obj, game.Effect{TargetIndex: dynamic.TargetIndex}); ok {
			amount = effectivePower(g, permanent)
		}
	case game.DynamicAmountTargetToughness:
		if permanent, ok := effectPermanent(g, obj, game.Effect{TargetIndex: dynamic.TargetIndex}); ok {
			if toughness, ok := effectiveToughness(g, permanent); ok {
				amount = toughness
			}
		}
	case game.DynamicAmountTargetManaValue:
		if permanent, ok := effectPermanent(g, obj, game.Effect{TargetIndex: dynamic.TargetIndex}); ok {
			if def, ok := permanentCardDef(g, permanent); ok {
				amount = def.ManaValue
			}
		}
	case game.DynamicAmountTargetCounters:
		if permanent, ok := effectPermanent(g, obj, game.Effect{TargetIndex: dynamic.TargetIndex}); ok {
			amount = permanent.Counters.Get(dynamic.CounterKind)
		}
	case game.DynamicAmountControllerLife:
		if player, ok := playerByID(g, stackObjectController(obj)); ok {
			amount = player.Life
		}
	case game.DynamicAmountControllerHandSize:
		if player, ok := playerByID(g, stackObjectController(obj)); ok {
			amount = player.Hand.Size()
		}
	case game.DynamicAmountControllerGraveyardSize:
		if player, ok := playerByID(g, stackObjectController(obj)); ok {
			amount = player.Graveyard.Size()
		}
	case game.DynamicAmountCountSelector:
		amount = len(selectedPermanentIDs(g, stackObjectController(obj), nil, dynamic.Selector))
	case game.DynamicAmountPreviousEffectResult:
		if dynamic.LinkID != "" {
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
	if effect.LinkID == "" {
		return
	}
	if obj.ResolvedAmounts == nil {
		obj.ResolvedAmounts = make(map[string]int)
	}
	obj.ResolvedAmounts[effect.LinkID] = amount
}

func moveCounters(g *game.Game, obj *game.StackObject, effect game.Effect) bool {
	destination, ok := effectPermanent(g, obj, effect)
	if !ok {
		return false
	}
	counters, source, ok := effectCounterSource(g, obj, effect.CounterSource)
	if !ok || counters.IsEmpty() {
		return false
	}
	if source != nil && source.ObjectID == destination.ObjectID {
		return false
	}
	for kind, amount := range counters.All() {
		destination.Counters.Add(kind, amount)
	}
	if source == nil {
		return true
	}
	for kind, amount := range counters.All() {
		source.Counters.Remove(kind, amount)
	}
	return true
}

func effectCounterSource(g *game.Game, obj *game.StackObject, source game.CounterSourceSpec) (counter.Set, *game.Permanent, bool) {
	switch source.Kind {
	case game.CounterSourceTarget:
		permanent, ok := effectPermanent(g, obj, game.Effect{TargetIndex: source.TargetIndex})
		if !ok {
			return counter.Set{}, nil, false
		}
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
	}
	return counter.Set{}, nil, false
}

func effectConditionSatisfied(g *game.Game, obj *game.StackObject, condition opt.V[game.EffectCondition]) bool {
	if !condition.Exists {
		return true
	}
	cond := condition.Val
	if cond.MatchPermanentType {
		permanent, ok := effectPermanent(g, obj, game.Effect{TargetIndex: cond.TargetIndex})
		if !ok {
			return false
		}
		matches := permanentHasType(g, permanent, cond.PermanentType)
		if cond.Negate {
			matches = !matches
		}
		if !matches {
			return false
		}
	}
	return true
}

func effectResultConditionSatisfied(obj *game.StackObject, condition opt.V[game.EffectResultCondition]) bool {
	if !condition.Exists || condition.Val.LinkID == "" {
		return true
	}
	if obj == nil || obj.ResolutionResults == nil {
		return false
	}
	cond := condition.Val
	result, ok := obj.ResolutionResults[cond.LinkID]
	if !ok {
		return false
	}
	if cond.Accepted != game.TriAny && (cond.Accepted == game.TriTrue) != result.Accepted {
		return false
	}
	if cond.Succeeded != game.TriAny && (cond.Succeeded == game.TriTrue) != result.Succeeded {
		return false
	}
	return true
}

func rememberEffectResolutionResult(obj *game.StackObject, effect game.Effect, accepted bool, succeeded bool, amount int) {
	if obj == nil || effect.LinkID == "" {
		return
	}
	if obj.ResolutionResults == nil {
		obj.ResolutionResults = make(map[string]game.EffectResolutionResult)
	}
	obj.ResolutionResults[effect.LinkID] = game.EffectResolutionResult{
		Accepted:  accepted,
		Succeeded: succeeded,
		Amount:    amount,
	}
}

func applyContinuousEffectTemplates(g *game.Game, obj *game.StackObject, permanent *game.Permanent, effect game.Effect) bool {
	if len(effect.ContinuousEffects) == 0 {
		return false
	}
	sourceID, sourceObjectID := damageSourceIDs(g, obj)
	timestamp := int64(g.IDGen.Next())
	applied := false
	for _, template := range effect.ContinuousEffects {
		// Runtime continuous effects are applied by the layer system; animation
		// effects such as "becomes a 0/0 creature" use type and P/T layers
		// (CR 611, CR 613).
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
		applied = true
	}
	return applied
}

func damageSourceIDs(g *game.Game, obj *game.StackObject) (id.ID, id.ID) {
	switch obj.Kind {
	case game.StackActivatedAbility, game.StackTriggeredAbility:
		if obj.SourceCardID != 0 {
			return obj.SourceCardID, obj.SourceID
		}
		permanent, ok := permanentByObjectID(g, obj.SourceID)
		if !ok {
			return 0, obj.SourceID
		}
		return permanent.CardInstanceID, permanent.ObjectID
	default:
		return obj.SourceID, 0
	}
}

func selectedPermanentIDs(g *game.Game, controller game.PlayerID, source *game.Permanent, selector game.EffectSelector) []id.ID {
	permanentIDs := make([]id.ID, 0, len(g.Battlefield))
	for _, permanent := range g.Battlefield {
		if !permanentMatchesSelectorForSource(g, source, controller, permanent, selector) {
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
	if choice, ok := linkedResolutionChoice(obj, effect.ChoiceLinkID); ok && choice.Kind == game.ResolutionChoicePlayer {
		if !isPlayerAlive(g, choice.Player) {
			return 0, false
		}
		return choice.Player, true
	}
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

func effectManaColor(obj *game.StackObject, effect game.Effect) mana.Color {
	if choice, ok := linkedResolutionChoice(obj, effect.ChoiceLinkID); ok && choice.Kind == game.ResolutionChoiceColor {
		return choice.Color
	}
	return effect.ManaColor
}

func effectPermanent(g *game.Game, obj *game.StackObject, effect game.Effect) (*game.Permanent, bool) {
	if effect.TargetIndex == -2 {
		return sourcePermanent(g, obj)
	}
	if effect.TargetIndex < 0 || effect.TargetIndex >= len(obj.Targets) {
		return nil, false
	}
	target := obj.Targets[effect.TargetIndex]
	if target.Kind != game.TargetPermanent {
		return nil, false
	}
	return permanentByObjectID(g, target.PermanentID)
}

func sourcePermanent(g *game.Game, obj *game.StackObject) (*game.Permanent, bool) {
	return permanentByObjectID(g, obj.SourceID)
}

func firstPermanentControlledBy(g *game.Game, controller game.PlayerID) (*game.Permanent, bool) {
	for _, permanent := range g.Battlefield {
		if effectiveController(g, permanent) == controller {
			return permanent, true
		}
	}
	return nil, false
}

func permanentLinkedObjectRef(permanent *game.Permanent) game.LinkedObjectRef {
	if permanent.CardInstanceID == 0 {
		return game.LinkedObjectRef{}
	}
	return game.LinkedObjectRef{ObjectID: permanent.ObjectID, CardID: permanent.CardInstanceID}
}

func returnLinkedExiledObjects(g *game.Game, obj *game.StackObject, linkID string) bool {
	key := linkedObjectSourceKey(g, obj, linkID)
	returned := false
	for _, ref := range linkedObjects(g, key) {
		if snapshot, ok := lastKnownObject(g, ref.ObjectID); !ok || snapshot.CardID != ref.CardID {
			continue
		}
		card, ok := g.GetCardInstance(ref.CardID)
		if !ok {
			continue
		}
		owner, ok := playerByID(g, card.Owner)
		if !ok || !owner.Exile.Remove(ref.CardID) {
			continue
		}
		if _, ok := createCardPermanent(g, card, obj.Controller, game.ZoneExile); ok {
			returned = true
		}
	}
	clearLinkedObjects(g, key)
	return returned
}

func createTokenPermanent(g *game.Game, controller game.PlayerID, token *game.CardDef) (*game.Permanent, bool) {
	if token == nil {
		return nil, false
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
	return permanent, true
}
