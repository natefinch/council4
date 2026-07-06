package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/opt"
)

// handlePlayerMayPayGenericOrRule offers the referenced player the option to pay
// a generic mana amount. When the amount is zero the payment is trivially made
// and no consequence follows. Otherwise the player may decline or fail to pay,
// which installs the primitive's rule effects on that player's permanents for
// the given duration (Champions of Minas Tirith: "that opponent may pay {X},
// where X is the number of cards in their hand. If they don't, they can't attack
// you this combat.").
func handlePlayerMayPayGenericOrRule(r *effectResolver, prim game.PlayerMayPayGenericOrRule) effectResolved {
	amount := r.quantity(prim.Amount)
	if amount <= 0 {
		return effectResolved{accepted: true, succeeded: true}
	}
	payment := game.ResolutionPayment{
		Payer:    opt.Val(prim.Player),
		ManaCost: opt.Val(cost.Mana{cost.O(amount)}),
	}
	accepted, succeeded := r.engine.resolveResolutionPaymentValue(r.game, r.obj, &payment, r.agents, r.log)
	if succeeded {
		return effectResolved{accepted: accepted, succeeded: true}
	}
	installed := createRuleEffectTemplates(r.game, r.obj, opt.V[game.ObjectReference]{}, prim.RuleEffects, prim.Duration)
	return effectResolved{accepted: accepted, succeeded: installed}
}
