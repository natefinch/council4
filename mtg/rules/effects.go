package rules

import "github.com/natefinch/council4/mtg/game"

func (e *Engine) resolveSpellEffects(g *game.Game, obj *game.StackObject, card *game.CardInstance, log *TurnLog) {
	ability := firstSpellAbility(card.Def)
	if ability == nil {
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
	if effect.Amount <= 0 {
		return
	}
	playerID, ok := effectPlayer(g, obj, effect)
	if !ok {
		return
	}
	player := g.Players[playerID]

	switch effect.Type {
	case game.EffectDraw:
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
		player.Life += effect.Amount
	case game.EffectLoseLife:
		player.Life -= effect.Amount
	case game.EffectDamage:
		// Damage to players is life loss for now; prevention and damage events come later.
		player.Life -= effect.Amount
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
