package rules

import (
	"github.com/natefinch/council4/mtg/game"
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
	if effect.Selector != game.EffectSelectorNone {
		resolveMassPermanentEffect(g, obj, effect)
		return
	}
	switch effect.Type {
	case game.EffectDraw:
		if effect.Amount <= 0 {
			return
		}
		playerID, ok := effectPlayer(g, obj, effect)
		if !ok {
			return
		}
		e.drawCards(g, playerID, effect.Amount, log)

	case game.EffectGainLife:
		if effect.Amount <= 0 {
			return
		}
		playerID, ok := effectPlayer(g, obj, effect)
		if !ok {
			return
		}
		player := g.Players[playerID]
		player.Life += effect.Amount
	case game.EffectLoseLife:
		if effect.Amount <= 0 {
			return
		}
		playerID, ok := effectPlayer(g, obj, effect)
		if !ok {
			return
		}
		player := g.Players[playerID]
		player.Life -= effect.Amount
	case game.EffectAddMana:
		amount := effect.Amount
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
		if effect.Amount <= 0 {
			return
		}
		if playerID, ok := effectPlayer(g, obj, effect); ok {
			sourceID, sourceObjectID := damageSourceIDs(g, obj)
			dealPlayerDamage(g, sourceID, sourceObjectID, obj.Controller, playerID, effect.Amount, false)
			return
		}
		permanent := effectPermanent(g, obj, effect)
		if permanent == nil {
			return
		}
		sourceID, sourceObjectID := damageSourceIDs(g, obj)
		dealPermanentDamage(g, sourceID, sourceObjectID, obj.Controller, permanent, effect.Amount, false)
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
	case game.EffectCreateToken:
		amount := effect.Amount
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
			millCards(g, playerID, effect.Amount)
		}
	case game.EffectScry:
		playerID, ok := effectPlayer(g, obj, effect)
		if ok {
			e.scryCards(g, agents, log, playerID, effect.Amount)
		}
	case game.EffectSurveil:
		playerID, ok := effectPlayer(g, obj, effect)
		if ok {
			e.surveilCards(g, agents, log, playerID, effect.Amount)
		}
	case game.EffectFight:
		resolveFight(g, obj, effect)
	}
}

func (e *Engine) drawCards(g *game.Game, playerID game.PlayerID, amount int, log *TurnLog) {
	if amount <= 0 {
		return
	}
	for range amount {
		cardID, ok := e.drawCard(g, playerID)
		if log != nil {
			log.Draws = append(log.Draws, DrawLog{
				Player: playerID,
				CardID: cardID,
				Failed: !ok,
			})
		}
	}
}

func stackObjectSourceIsSnow(g *game.Game, obj *game.StackObject) bool {
	if g == nil || obj == nil {
		return false
	}
	permanent := permanentByObjectID(g, obj.SourceID)
	return permanentIsSnow(g, permanent)
}

func resolveMassPermanentEffect(g *game.Game, obj *game.StackObject, effect game.Effect) {
	permanentIDs := selectedPermanentIDs(g, obj.Controller, nil, effect.Selector)
	sourceID, sourceObjectID := damageSourceIDs(g, obj)
	for _, permanentID := range permanentIDs {
		permanent := permanentByObjectID(g, permanentID)
		if permanent == nil {
			continue
		}
		switch effect.Type {
		case game.EffectDamage:
			if effect.Amount > 0 {
				dealPermanentDamage(g, sourceID, sourceObjectID, obj.Controller, permanent, effect.Amount, false)
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
		}
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
