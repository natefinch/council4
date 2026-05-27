package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

func createRuleEffects(g *game.Game, obj *game.StackObject, effect game.Effect) bool {
	if len(effect.RuleEffects) == 0 {
		return false
	}
	sourceID, sourceObjectID := damageSourceIDs(g, obj)
	for _, ruleEffect := range effect.RuleEffects {
		ruleEffect.ID = g.IDGen.Next()
		ruleEffect.Controller = obj.Controller
		ruleEffect.SourceCardID = sourceID
		ruleEffect.SourceObjectID = sourceObjectID
		ruleEffect.CreatedTurn = g.Turn.TurnNumber
		if effect.Duration != game.DurationPermanent {
			ruleEffect.Duration = effect.Duration
		}
		if ruleEffect.Duration == game.DurationPermanent && effect.UntilEndOfTurn {
			ruleEffect.Duration = game.DurationUntilEndOfTurn
		}
		g.RuleEffects = append(g.RuleEffects, ruleEffect)
	}
	return true
}

func activeRuleEffects(g *game.Game) []game.RuleEffect {
	effects := make([]game.RuleEffect, 0, len(g.RuleEffects))
	for _, effect := range g.RuleEffects {
		if ruleEffectSourceStillApplies(g, effect) {
			effects = append(effects, effect)
		}
	}
	effects = append(effects, staticRuleEffects(g)...)
	return effects
}

func staticRuleEffects(g *game.Game) []game.RuleEffect {
	var effects []game.RuleEffect
	for _, source := range g.Battlefield {
		if source.PhasedOut {
			continue
		}
		sourceDef, ok := permanentCardDef(g, source)
		if !ok {
			continue
		}
		for i := range sourceDef.Abilities {
			ability := &sourceDef.Abilities[i]
			if ability.Kind != game.StaticAbility || !abilityFunctionsOnBattlefield(ability) {
				continue
			}
			for _, effect := range ability.Effects {
				if effect.Type != game.EffectApplyRule {
					continue
				}
				for _, ruleEffect := range effect.RuleEffects {
					ruleEffect.Controller = effectiveController(g, source)
					ruleEffect.SourceObjectID = source.ObjectID
					ruleEffect.SourceCardID = source.CardInstanceID
					effects = append(effects, ruleEffect)
				}
			}
		}
	}
	return effects
}

func ruleEffectSourceStillApplies(g *game.Game, effect game.RuleEffect) bool {
	if effect.Duration != game.DurationPermanent || effect.SourceObjectID == 0 {
		return true
	}
	_, ok := permanentByObjectID(g, effect.SourceObjectID)
	return ok
}

func expireRuleEffects(g *game.Game) {
	if len(g.RuleEffects) == 0 {
		return
	}
	kept := g.RuleEffects[:0]
	for _, effect := range g.RuleEffects {
		if effect.Duration == game.DurationUntilEndOfTurn || effect.Duration == game.DurationThisTurn {
			continue
		}
		if !ruleEffectSourceStillApplies(g, effect) {
			continue
		}
		kept = append(kept, effect)
	}
	g.RuleEffects = kept
}

func canGainLife(g *game.Game, playerID game.PlayerID) bool {
	for _, effect := range activeRuleEffects(g) {
		if effect.Kind != game.RuleEffectCantGainLife {
			continue
		}
		if playerRelationMatches(effect.Controller, playerID, effect.AffectedPlayer) {
			return false
		}
	}
	return true
}

func gainLife(g *game.Game, playerID game.PlayerID, amount int) int {
	if amount <= 0 || !canGainLife(g, playerID) {
		return 0
	}
	player, ok := playerByID(g, playerID)
	if !ok || player.Eliminated {
		return 0
	}
	player.Life += amount
	emitEvent(g, game.GameEvent{
		Kind:   game.EventLifeGained,
		Player: playerID,
		Amount: amount,
	})
	return amount
}

func loseLife(g *game.Game, playerID game.PlayerID, amount int) int {
	if amount <= 0 {
		return 0
	}
	player, ok := playerByID(g, playerID)
	if !ok || player.Eliminated {
		return 0
	}
	player.Life -= amount
	emitEvent(g, game.GameEvent{
		Kind:   game.EventLifeLost,
		Player: playerID,
		Amount: amount,
	})
	return amount
}

func ruleEffectProhibitsAttack(g *game.Game, attacker *game.Permanent, target *game.AttackTarget) bool {
	for _, effect := range activeRuleEffects(g) {
		if effect.Kind != game.RuleEffectCantAttack {
			continue
		}
		if !ruleEffectMatchesPermanent(g, effect, attacker) {
			continue
		}
		if effect.DefendingPlayer != game.PlayerAny {
			if target == nil {
				continue
			}
			if !playerRelationMatches(effect.Controller, target.Player, effect.DefendingPlayer) {
				continue
			}
		}
		return true
	}
	return false
}

func ruleEffectProhibitsBlock(g *game.Game, blocker *game.Permanent) bool {
	for _, effect := range activeRuleEffects(g) {
		if effect.Kind == game.RuleEffectCantBlock && ruleEffectMatchesPermanent(g, effect, blocker) {
			return true
		}
	}
	return false
}

func ruleEffectMatchesPermanent(g *game.Game, effect game.RuleEffect, permanent *game.Permanent) bool {
	if permanent == nil {
		return false
	}
	if !controllerRelationMatches(effect.Controller, effectiveController(g, permanent), effect.AffectedController) {
		return false
	}
	for _, cardType := range effect.PermanentTypes {
		if !permanentHasType(g, permanent, cardType) {
			return false
		}
	}
	return true
}

func controllerRelationMatches(sourceController, candidate game.PlayerID, relation game.ControllerRelation) bool {
	switch relation {
	case game.ControllerYou:
		return candidate == sourceController
	case game.ControllerOpponent, game.ControllerNotYou:
		return candidate != sourceController && candidate >= 0 && candidate < game.NumPlayers
	default:
		return true
	}
}

func playerRelationMatches(sourceController, candidate game.PlayerID, relation game.PlayerRelation) bool {
	switch relation {
	case game.PlayerYou:
		return candidate == sourceController
	case game.PlayerOpponent, game.PlayerNotYou:
		return candidate != sourceController && candidate >= 0 && candidate < game.NumPlayers
	default:
		return true
	}
}

func staticCostModifiersForContext(g *game.Game, context costModificationContext) []game.CostModifier {
	var modifiers []game.CostModifier
	for _, effect := range activeRuleEffects(g) {
		if effect.Kind != game.RuleEffectCostModifier {
			continue
		}
		modifier := effect.CostModifier
		if modifier.Kind != game.CostModifierSpell {
			continue
		}
		if modifier.MatchCardType && (context.card == nil || !context.card.HasType(modifier.CardType)) {
			continue
		}
		modifiers = append(modifiers, modifier)
	}
	return modifiers
}

func canCastFromZoneByRuleEffect(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone game.ZoneType) bool {
	card, cardOK := g.GetCardInstance(cardID)
	if sourceZone == game.ZoneGraveyard && cardOK && cardHasFlashbackAlternative(card) {
		return true
	}
	for _, effect := range activeRuleEffects(g) {
		if effect.Kind != game.RuleEffectCastFromZone || effect.CastFromZone != sourceZone {
			continue
		}
		if !playerRelationMatches(effect.Controller, playerID, effect.AffectedPlayer) {
			continue
		}
		return true
	}
	return false
}

func castableZonesForPlayer(g *game.Game, playerID game.PlayerID) []game.ZoneType {
	zones := []game.ZoneType{game.ZoneHand}
	if player, ok := playerByID(g, playerID); ok {
		for _, cardID := range player.Graveyard.All() {
			if canCastFromZoneByRuleEffect(g, playerID, cardID, game.ZoneGraveyard) {
				zones = append(zones, game.ZoneGraveyard)
				break
			}
		}
	}
	return slices.Compact(zones)
}

func cardHasFlashbackAlternative(card *game.CardInstance) bool {
	if !card.Def.HasKeyword(game.Flashback) {
		return false
	}
	for _, option := range spellCostOptionsForZoneAndKicker(card.Def, game.ZoneGraveyard, false) {
		if option.index > 0 {
			return true
		}
	}
	return false
}
