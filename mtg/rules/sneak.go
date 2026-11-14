package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules/payment"
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
// A permanent spell cast for its sneak cost also enters the battlefield
// attacking (CR 702.190b), which is not supported, so Sneak on a permanent
// spell is never offered.
func sneakCastFor(g *game.Game, playerID game.PlayerID, sourceZone zone.Type, spellDef, castFace *game.CardDef) payment.SneakCast {
	if sourceZone != zone.Hand || spellDef == nil || castFace == nil || castFace.IsPermanent() ||
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

func isSneakAlternative(alternative cost.Alternative) bool {
	return alternative.Mechanic == cost.AlternativeMechanicSneak
}
