package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func (r *effectResolver) damageSource(source game.ObjectReference) (effectDamageSource, bool) {
	if source.Kind() == game.ObjectReferenceNone {
		sourceID, sourceObjectID := damageSourceIDs(r.game, r.obj)
		return effectDamageSource{
			sourceID:       sourceID,
			sourceObjectID: sourceObjectID,
			controller:     r.obj.Controller,
		}, true
	}
	resolved, ok := resolveObjectReference(r.game, r.obj, source)
	if !ok {
		return effectDamageSource{}, false
	}
	if resolved.permanent == nil {
		if resolved.snapshot.ObjectID == 0 {
			return effectDamageSource{}, false
		}
		return effectDamageSource{
			sourceID:       resolved.snapshot.CardID,
			sourceObjectID: resolved.snapshot.ObjectID,
			controller:     resolved.snapshot.Controller,
			deathtouch:     slices.Contains(resolved.snapshot.Keywords, game.Deathtouch),
			lifelink:       slices.Contains(resolved.snapshot.Keywords, game.Lifelink),
		}, true
	}
	return effectDamageSource{
		sourceID:       resolved.permanent.CardInstanceID,
		sourceObjectID: resolved.permanent.ObjectID,
		controller:     effectiveController(r.game, resolved.permanent),
		permanent:      resolved.permanent,
		deathtouch:     hasKeyword(r.game, resolved.permanent, game.Deathtouch),
		lifelink:       hasKeyword(r.game, resolved.permanent, game.Lifelink),
	}, true
}

func handleDamage(r *effectResolver, prim game.Damage) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	if res.amount <= 0 {
		return res
	}
	sourceRef := game.ObjectReference{}
	if prim.DamageSource.Exists {
		sourceRef = prim.DamageSource.Val
	}
	source, ok := r.damageSource(sourceRef)
	if !ok {
		return res
	}
	if object, ok := prim.Recipient.ObjectReference(); ok {
		return r.damageReferencedPermanent(res, source, prim.ResultAmountKind, object)
	}
	if player, ok := prim.Recipient.PlayerReference(); ok {
		return r.damageReferencedPlayer(res, source, prim.ResultAmountKind, player)
	}
	if player, ok := prim.Recipient.AnyTargetPlayerReference(); ok {
		if resolvedPlayer, playerOK := r.resolvePlayer(player); playerOK {
			dealt := dealPlayerDamage(r.game, source.sourceID, source.sourceObjectID, source.controller, resolvedPlayer, res.amount, false)
			applyDamageSourceLifelink(r.game, source, dealt)
			res.amount = typedDamageResultAmount(prim.ResultAmountKind, dealt, 0)
			res.succeeded = dealt > 0
			return res
		}
	}
	if object, ok := prim.Recipient.AnyTargetObjectReference(); ok {
		return r.damageReferencedPermanent(res, source, prim.ResultAmountKind, object)
	}
	if group, ok := prim.Recipient.GroupReference(); ok {
		return r.damageSelectedPermanents(res, source, group)
	}
	if group, ok := prim.Recipient.PlayerGroupReference(); ok {
		for _, playerID := range r.playerGroupMembers(group) {
			dealt := dealPlayerDamage(r.game, source.sourceID, source.sourceObjectID, source.controller, playerID, res.amount, false)
			applyDamageSourceLifelink(r.game, source, dealt)
			res.succeeded = dealt > 0 || res.succeeded
		}
	}
	return res
}

func (r *effectResolver) damageReferencedPlayer(res effectResolved, source effectDamageSource, resultKind game.EffectResultAmountKind, player game.PlayerReference) effectResolved {
	playerID, ok := r.resolvePlayer(player)
	if !ok {
		return res
	}
	dealt := dealPlayerDamage(r.game, source.sourceID, source.sourceObjectID, source.controller, playerID, res.amount, false)
	applyDamageSourceLifelink(r.game, source, dealt)
	res.amount = typedDamageResultAmount(resultKind, dealt, 0)
	res.succeeded = dealt > 0
	return res
}

func (r *effectResolver) damageReferencedPermanent(res effectResolved, source effectDamageSource, resultKind game.EffectResultAmountKind, object game.ObjectReference) effectResolved {
	permanent, ok := r.resolveObject(object)
	if !ok {
		return res
	}
	lethalRemaining := lethalDamageRemaining(r.game, permanent)
	if source.deathtouch {
		lethalRemaining = 1
		if permanent.MarkedDeathtouchDamage {
			lethalRemaining = 0
		}
	} else if source.permanent != nil {
		lethalRemaining = lethalDamageRemainingFromSource(r.game, source.permanent, permanent)
	}
	dealt := dealPermanentDamage(r.game, source.sourceID, source.sourceObjectID, source.controller, permanent, res.amount, false)
	applyDamageSourceKeywordEffects(r.game, source, permanent, dealt)
	res.excessDamage = max(0, dealt-lethalRemaining)
	res.amount = typedDamageResultAmount(resultKind, dealt, res.excessDamage)
	res.succeeded = dealt > 0 && (resultKind != game.EffectResultAmountExcessDamage || res.excessDamage > 0)
	return res
}

func (r *effectResolver) damageSelectedPermanents(res effectResolved, source effectDamageSource, group game.GroupReference) effectResolved {
	for _, permanent := range r.groupPermanentsWithSource(group, source.permanent) {
		dealt := dealPermanentDamage(r.game, source.sourceID, source.sourceObjectID, source.controller, permanent, res.amount, false)
		applyDamageSourceKeywordEffects(r.game, source, permanent, dealt)
		res.succeeded = dealt > 0 || res.succeeded
	}
	return res
}

func typedDamageResultAmount(kind game.EffectResultAmountKind, dealt, excess int) int {
	if kind == game.EffectResultAmountExcessDamage {
		return excess
	}
	return dealt
}

func handleFight(r *effectResolver, prim game.Fight) effectResolved {
	first, firstOK := r.resolveObject(prim.Object)
	second, secondOK := r.resolveObject(prim.RelatedObject)
	if !firstOK || !secondOK || first.ObjectID == second.ObjectID ||
		!permanentHasType(r.game, first, types.Creature) || !permanentHasType(r.game, second, types.Creature) {
		return effectResolved{accepted: true}
	}
	resolveFightPermanents(r.game, first, second)
	return effectResolved{accepted: true, succeeded: true}
}

func handlePreventDamage(r *effectResolver, prim game.PreventDamage) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	res.succeeded = createPreventionShield(r.game, r.obj, res.amount, prim.Object, prim.Player, game.DurationUntilEndOfTurn)
	return res
}
