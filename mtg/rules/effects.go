package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

func (e *Engine) resolveSpellEffects(g *game.Game, obj *game.StackObject, card *game.CardInstance, log *TurnLog) {
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
		resolveMassPermanentEffect(g, effect)
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
		for range effect.Amount {
			cardID, ok := e.drawCard(g, playerID)
			if log != nil {
				log.Draws = append(log.Draws, DrawLog{
					Player: playerID,
					CardID: cardID,
					Failed: !ok,
				})
			}
		}

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
		player.ManaPool.Add(effect.ManaColor, amount)
	case game.EffectDamage:
		if effect.Amount <= 0 {
			return
		}
		if playerID, ok := effectPlayer(g, obj, effect); ok {
			// Damage to players is life loss for now; prevention and damage events come later.
			g.Players[playerID].Life -= effect.Amount
			return
		}
		permanent := effectPermanent(g, obj, effect)
		if permanent == nil {
			return
		}
		markPermanentDamage(g, permanent, effect.Amount)
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

func resolveMassPermanentEffect(g *game.Game, effect game.Effect) {
	permanentIDs := selectedPermanentIDs(g, effect.Selector)
	for _, permanentID := range permanentIDs {
		permanent := permanentByObjectID(g, permanentID)
		if permanent == nil {
			continue
		}
		switch effect.Type {
		case game.EffectDamage:
			if effect.Amount > 0 {
				markPermanentDamage(g, permanent, effect.Amount)
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

func selectedPermanentIDs(g *game.Game, selector game.EffectSelector) []id.ID {
	if g == nil {
		return nil
	}
	permanentIDs := make([]id.ID, 0, len(g.Battlefield))
	for _, permanent := range g.Battlefield {
		if permanent == nil || !permanentMatchesSelector(g, permanent, selector) {
			continue
		}
		permanentIDs = append(permanentIDs, permanent.ObjectID)
	}
	return permanentIDs
}

func permanentMatchesSelector(g *game.Game, permanent *game.Permanent, selector game.EffectSelector) bool {
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
	return permanent
}
