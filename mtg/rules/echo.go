package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// echoObligationPending reports whether the Echo triggered ability (CR 702.29)
// should fire for the given controller: the permanent's echo obligation is
// pending when its recorded resolved controller is unset (it just entered the
// battlefield) or differs from the current controller (a different player gained
// control since the obligation was last resolved). This is the state-based
// evaluation of "it came under your control since the beginning of your most
// recent upkeep" (CR 702.29e) without a discrete control-change event.
//
// Scope of the approximation (tracked in #3014): because only the last resolved
// controller is recorded, rather than a per-player control-since-last-upkeep
// history, two cases diverge from the rules and are knowingly not handled here:
//
//   - Temporary control that returns to the same player before the intervening
//     controller takes an upkeep. For a P->Q->P change (a Threaten-style "gain
//     control until end of turn" or an instant steal-and-return) where Q never
//     reaches an upkeep, the recorded controller is still P, so P's next upkeep
//     sees no pending obligation even though the permanent did come back under
//     P's control since P's last upkeep. This misses a trigger.
//
//   - A countered/removed Echo trigger. The obligation is recorded only when the
//     trigger resolves (handleRecordEchoObligation), so if the trigger is
//     countered (Stifle/Disallow) or otherwise leaves the stack without
//     resolving, the marker stays pending and Echo incorrectly triggers again at
//     the next upkeep, even though the rules obligation was already checked.
//
// A faithful fix needs an authoritative control-change signal plus an obligation
// consumed when the trigger is put on the stack (not at resolution); that is a
// deeper engine capability deliberately deferred to #3014 rather than
// approximated further here. Ordinary entry, normal steals with an intervening
// upkeep, and blink-to-new-object all remain correct.
func echoObligationPending(g *game.Game, source *game.Permanent, controller game.PlayerID) bool {
	if source == nil {
		return false
	}
	if !source.EchoResolvedController.Exists {
		return true
	}
	return source.EchoResolvedController.Val != controller
}

// handleRecordEchoObligation records the resolving controller as the player for
// whom the source permanent's echo obligation has been resolved, so later
// upkeeps of that same controller do not re-trigger the pay-or-sacrifice. It is
// a no-op when the source has already left the battlefield.
func handleRecordEchoObligation(r *effectResolver, prim game.RecordEchoObligation) effectResolved {
	res := effectResolved{accepted: true}
	permanent, ok := r.resolveObject(prim.Object)
	if ok {
		permanent.EchoResolvedController = opt.Val(stackObjectController(r.obj))
		res.succeeded = true
	}
	return res
}
