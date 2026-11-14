package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

// sneakCastFor returns the Sneak permission of a cast of castFace, spellDef's
// face as cast, from sourceZone by playerID (CR 702.190a). Sneak is "you may
// cast this spell by paying [cost] and returning an unblocked attacker you
// control to its owner's hand rather than paying its mana cost", and it may be
// used only from hand, any time playerID could cast an instant during their
// own declare blockers step. When ordinary timing also lets the spell be cast
// now the Sneak cost is one more option; when only Sneak's timing does, the
// cast is limited to its Sneak cost.
//
// A card a player may cast from the command zone (their commander, CR 903.8)
// may use its own Sneak there, paying the commander tax on top. A permanent
// spell cast for its sneak cost enters tapped and attacking what the returned
// creature was attacking (CR 702.190b); see putSneakedPermanentAttacking.
func sneakCastFor(g *game.Game, playerID game.PlayerID, sourceZone zone.Type, spellDef, castFace *game.CardDef) payment.SneakCast {
	if (sourceZone != zone.Hand && sourceZone != zone.Command) || spellDef == nil || castFace == nil ||
		!slices.ContainsFunc(spellDef.AlternativeCosts, isSneakAlternative) ||
		!sneakCastWindow(g, playerID) {
		return payment.SneakUnavailable
	}
	if canCastAtCurrentTiming(g, playerID, castFace) {
		return payment.SneakAvailable
	}
	if cardCastRestrictionsSatisfied(g, playerID, castFace) {
		return payment.SneakRequired
	}
	return payment.SneakUnavailable
}

// sneakCastWindow reports whether it is playerID's declare blockers step and
// they control an unblocked attacker to return (CR 702.190a, 509.1h). The
// caller still requires priority, as for any cast.
func sneakCastWindow(g *game.Game, playerID game.PlayerID) bool {
	return g.Turn.ActivePlayer == playerID &&
		g.Turn.Phase == game.PhaseCombat &&
		g.Turn.Step == game.StepDeclareBlockers &&
		len(unblockedAttackers(g, playerID)) > 0
}

// sneakAttackTarget returns what the unblocked attacker returned to pay a
// sneak cost was attacking. It reads the combat declarations captured before
// payment removed that creature from combat (CR 506.4).
func sneakAttackTarget(attackers []game.AttackDeclaration, returned []id.ID) opt.V[game.AttackTarget] {
	if len(returned) == 0 {
		return opt.V[game.AttackTarget]{}
	}
	for _, attack := range attackers {
		if attack.Attacker == returned[0] {
			return opt.Val(attack.Target)
		}
	}
	return opt.V[game.AttackTarget]{}
}

// putSneakedPermanentAttacking puts the permanent a sneaked spell became into
// combat attacking target, what the returned creature was attacking (CR
// 702.190b). It was never declared as an attacker, so no attack triggers fire
// and attack restrictions and requirements do not apply (CR 508.4, 508.4c);
// entering during the declare blockers step it is unblocked (CR 509.1h). It
// stays on the battlefield without attacking when it is not a creature or is
// also a battle (CR 506.3a, 506.3f), when its controller is not the attacking
// player (CR 506.3b), or when target can no longer be attacked: its player has
// left the game, or its planeswalker or battle has left the battlefield, is no
// longer a planeswalker or battle, or is no longer controlled or protected by
// an opponent of the attacking player (CR 506.3c, 508.4a). That includes a
// returned creature that was attacking nothing because its planeswalker or
// battle was removed from combat before the cast (CR 506.4c): there is nothing
// left for the new creature to attack.
func putSneakedPermanentAttacking(g *game.Game, permanent *game.Permanent, target game.AttackTarget) {
	if g.Combat == nil || permanent == nil || target.NoTarget ||
		!permanentHasType(g, permanent, types.Creature) || permanentHasType(g, permanent, types.Battle) {
		return
	}
	controller := effectiveController(g, permanent)
	if controller != g.Turn.ActivePlayer || !isPlayerAlive(g, target.Player) {
		return
	}
	if target.PlaneswalkerID != 0 || target.BattleID != 0 {
		index := slices.IndexFunc(legalAttackTargets(g, controller), func(legal game.AttackTarget) bool {
			return legal.PlaneswalkerID == target.PlaneswalkerID && legal.BattleID == target.BattleID
		})
		if index < 0 {
			return
		}
		target = legalAttackTargets(g, controller)[index]
	} else if target.Player == controller {
		return
	}
	declareTokenAttackingTarget(g, controller, permanent, target)
}

func isSneakAlternative(alternative cost.Alternative) bool {
	return alternative.Mechanic == cost.AlternativeMechanicSneak
}
