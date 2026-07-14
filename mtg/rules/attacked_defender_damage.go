package rules

import (
	"github.com/natefinch/council4/mtg/game"
)

// damageAttackedDefender deals the resolved amount to the player, planeswalker,
// or battle the resolving ability's triggering attacker is attacking ("deals N
// damage to the player or planeswalker it's attacking"). It routes to the
// attacked permanent when the attack was declared against a planeswalker or
// battle rather than always redirecting to the defending player, and deals no
// damage when that permanent has left the battlefield (the effect neither
// redirects to the player nor errors).
func (r *effectResolver) damageAttackedDefender(res effectResolved, source effectDamageSource, resultKind game.EffectResultAmountKind) effectResolved {
	target, ok := r.attackedDefenderTarget()
	if !ok {
		res.succeeded = false
		return res
	}
	if target.IsPlayerAttack() {
		if !isPlayerAlive(r.game, target.Player) {
			res.succeeded = false
			return res
		}
		dealt := dealPlayerDamage(r.game, source.sourceID, source.sourceObjectID, source.controller, target.Player, res.amount, false)
		applyDamageSourceLifelink(r.game, source, dealt)
		res.amount = typedDamageResultAmount(resultKind, dealt, 0)
		res.succeeded = dealt > 0
		return res
	}
	permanent, ok := attackTargetPermanent(r.game, target)
	if !ok {
		res.succeeded = false
		return res
	}
	dealt := dealPermanentDamage(r.game, source.sourceID, source.sourceObjectID, source.controller, permanent, res.amount, false)
	applyDamageSourceKeywordEffects(r.game, source, permanent, dealt)
	res.amount = typedDamageResultAmount(resultKind, dealt, 0)
	res.succeeded = dealt > 0
	return res
}

// attackedDefenderTarget resolves what the resolving ability's triggering
// attacker is attacking. It prefers the attack target captured on the trigger
// event (populated for the attacker-declared, became-blocked, and
// became-unblocked combat events), so a triggering attacker that has since left
// combat — including the ability's own source (Myr Battlesphere) — still deals
// damage to the player or planeswalker it was declared against. When only the
// defending player was captured it recovers the planeswalker or battle from live
// combat, and finally falls back to the recorded defending player.
func (r *effectResolver) attackedDefenderTarget() (game.AttackTarget, bool) {
	if r.obj == nil || !r.obj.HasTriggerEvent {
		return game.AttackTarget{}, false
	}
	event := r.obj.TriggerEvent
	captured := event.AttackTarget
	if captured.Player != 0 || captured.PlaneswalkerID != 0 || captured.BattleID != 0 {
		return captured, true
	}
	if event.PermanentID != 0 {
		if target, ok := attackTargetForAttacker(r.game, event.PermanentID); ok {
			return target, true
		}
	}
	if defendingPlayerEvent(event.Kind) && event.Player != 0 {
		return game.AttackTarget{Player: event.Player}, true
	}
	return game.AttackTarget{}, false
}
