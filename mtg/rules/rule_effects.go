package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/opt"
)

func createRuleEffectTemplates(g *game.Game, obj *game.StackObject, object opt.V[game.ObjectReference], templates []game.RuleEffect, duration game.EffectDuration) bool {
	if len(templates) == 0 {
		return false
	}
	sourceID, sourceObjectID := damageSourceIDs(g, obj)
	for _, ruleEffect := range templates {
		ruleEffect.ID = g.IDGen.Next()
		ruleEffect.Controller = obj.Controller
		ruleEffect.SourceCardID = sourceID
		ruleEffect.SourceObjectID = sourceObjectID
		if ruleEffect.AffectedSource {
			ruleEffect.AffectedObjectID = sourceObjectID
		} else if ruleEffect.AffectedObjectID == 0 {
			if object.Exists {
				if resolved, ok := resolveObjectReference(g, obj, object.Val); ok && resolved.permanent != nil {
					ruleEffect.AffectedObjectID = resolved.permanent.ObjectID
				}
			}
		}
		ruleEffect.CreatedTurn = g.Turn.TurnNumber
		if duration != game.DurationPermanent {
			ruleEffect.Duration = duration
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
	effects = append(effects, stackStaticRuleEffects(g)...)
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
		for i := range sourceDef.StaticAbilities {
			body := sourceDef.StaticAbilities[i]
			if !bodyFunctionsOnBattlefield(body) {
				continue
			}
			if !conditionSatisfied(g, conditionContext{
				controller: effectiveController(g, source),
				source:     source,
			}, body.Condition) {
				continue
			}
			for _, ruleEffect := range body.RuleEffects {
				ruleEffect.Controller = effectiveController(g, source)
				ruleEffect.SourceObjectID = source.ObjectID
				ruleEffect.SourceCardID = source.CardInstanceID
				if ruleEffect.AffectedSource {
					ruleEffect.AffectedObjectID = source.ObjectID
				}
				effects = append(effects, ruleEffect)
			}
		}
	}
	return effects
}

func stackStaticRuleEffects(g *game.Game) []game.RuleEffect {
	var effects []game.RuleEffect
	for _, source := range g.Stack.Objects() {
		if source.Kind != game.StackSpell {
			continue
		}
		_, sourceDef, ok := cardInstanceFaceDef(g, source.SourceID, source.Face)
		if !ok {
			continue
		}
		for i := range sourceDef.StaticAbilities {
			body := &sourceDef.StaticAbilities[i]
			if body.ZoneOfFunction != zone.Stack {
				continue
			}
			if !conditionSatisfied(g, conditionContext{
				controller: source.Controller,
			}, body.Condition) {
				continue
			}
			for _, ruleEffect := range body.RuleEffects {
				ruleEffect.Controller = source.Controller
				ruleEffect.SourceObjectID = source.ID
				ruleEffect.SourceCardID = source.SourceID
				if ruleEffect.AffectedSource {
					ruleEffect.AffectedObjectID = source.ID
				}
				effects = append(effects, ruleEffect)
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
	emitEvent(g, game.Event{
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
	increaseActivePlayerSpeedForOpponentLifeLoss(g, playerID)
	emitEvent(g, game.Event{
		Kind:   game.EventLifeLost,
		Player: playerID,
		Amount: amount,
	})
	return amount
}

func startEngines(g *game.Game, playerID game.PlayerID) bool {
	player, ok := playerByID(g, playerID)
	if !ok || player.Eliminated {
		return false
	}
	if player.Speed == 0 {
		player.Speed = 1
	}
	return true
}

func increaseActivePlayerSpeedForOpponentLifeLoss(g *game.Game, losingPlayer game.PlayerID) {
	active := g.Turn.ActivePlayer
	if active == losingPlayer || active < 0 || active >= game.NumPlayers {
		return
	}
	player, ok := playerByID(g, active)
	if !ok || player.Eliminated || player.Speed <= 0 || player.Speed >= 4 || player.SpeedIncreasedTurn == g.Turn.TurnNumber {
		return
	}
	player.Speed++
	player.SpeedIncreasedTurn = g.Turn.TurnNumber
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

func ruleEffectRequiresAttack(g *game.Game, attacker *game.Permanent) bool {
	for _, effect := range activeRuleEffects(g) {
		if effect.Kind == game.RuleEffectMustAttack && ruleEffectMatchesPermanent(g, effect, attacker) {
			return true
		}
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

func ruleEffectProhibitsBeingBlocked(g *game.Game, attacker *game.Permanent) bool {
	for _, effect := range activeRuleEffects(g) {
		if effect.Kind == game.RuleEffectCantBeBlocked && ruleEffectMatchesPermanent(g, effect, attacker) {
			return true
		}
	}
	return false
}

func ruleEffectRequiresBeingBlocked(g *game.Game, attacker *game.Permanent) bool {
	for _, effect := range activeRuleEffects(g) {
		if effect.Kind == game.RuleEffectMustBeBlocked && ruleEffectMatchesPermanent(g, effect, attacker) {
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
	if effect.AffectedObjectID != 0 && effect.AffectedObjectID != permanent.ObjectID {
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

func staticCostModifiersForContext(g *game.Game, card *game.CardDef) []game.CostModifier {
	var modifiers []game.CostModifier
	for _, effect := range activeRuleEffects(g) {
		if effect.Kind != game.RuleEffectCostModifier {
			continue
		}
		modifier := effect.CostModifier
		if modifier.Kind != game.CostModifierSpell {
			continue
		}
		if modifier.MatchCardType && (card == nil || !card.HasType(modifier.CardType)) {
			continue
		}
		modifiers = append(modifiers, modifier)
	}
	return modifiers
}

func canCastFromZoneByRuleEffect(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type) bool {
	card, cardOK := g.GetCardInstance(cardID)
	if sourceZone == zone.Graveyard && cardOK && cardHasFlashbackAlternative(card) {
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

func castableZonesForPlayer(g *game.Game, playerID game.PlayerID) []zone.Type {
	zones := []zone.Type{zone.Hand}
	if player, ok := playerByID(g, playerID); ok {
		for _, cardID := range player.Graveyard.All() {
			if canCastFromZoneByRuleEffect(g, playerID, cardID, zone.Graveyard) {
				zones = append(zones, zone.Graveyard)
				break
			}
		}
		for _, cardID := range player.Exile.All() {
			if g.AdventureCards[cardID] {
				zones = append(zones, zone.Exile)
				break
			}
		}
	}
	return slices.Compact(zones)
}

func cardHasFlashbackAlternative(card *game.CardInstance) bool {
	frontDef := cardFaceOrDefault(card, game.FaceFront)
	if !frontDef.HasKeyword(game.Flashback) {
		return false
	}
	return slices.ContainsFunc(frontDef.AlternativeCosts, isFlashbackAlternative)
}
