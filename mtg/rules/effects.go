package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

func (e *Engine) resolveSpellEffects(g *game.Game, obj *game.StackObject, card *game.CardInstance, log *TurnLog) {
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
				e.resolveEffect(g, obj, effect, log)
			}
		}
		return
	}
	for _, effect := range ability.Effects {
		e.resolveEffect(g, obj, effect, log)
	}
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
		movePermanentToZone(g, permanent, game.ZoneExile)
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
		if permanent == nil || permanent.Controller != obj.Controller {
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
			permanent.TemporaryPowerModifier += effect.PowerDelta
			permanent.TemporaryToughnessModifier += effect.ToughnessDelta
		}
	case game.EffectCreateToken:
		amount := effect.Amount
		if amount <= 0 {
			amount = 1
		}
		for range amount {
			createTokenPermanent(g, obj.Controller, effect.Token)
		}
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
	card := permanentCardDef(g, permanent)
	if card == nil {
		return false
	}
	switch selector {
	case game.EffectSelectorAllCreatures:
		return card.HasType(game.TypeCreature)
	case game.EffectSelectorAllArtifacts:
		return card.HasType(game.TypeArtifact)
	case game.EffectSelectorAllEnchantments:
		return card.HasType(game.TypeEnchantment)
	case game.EffectSelectorAllNonlandPermanents:
		return !card.HasType(game.TypeLand)
	case game.EffectSelectorAllPermanents:
		return true
	case game.EffectSelectorCreaturesYouControl:
		return permanent.Controller == controller && card.HasType(game.TypeCreature)
	case game.EffectSelectorOtherCreaturesYouControl:
		return source != nil && permanent.ObjectID != source.ObjectID && permanent.Controller == controller && card.HasType(game.TypeCreature)
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
		if permanent != nil && permanent.Controller == controller {
			return permanent
		}
	}
	return nil
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
