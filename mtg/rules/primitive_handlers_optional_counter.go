package rules

import (
	"github.com/natefinch/council4/mtg/game"
)

// handleOptionalCounterForEachPlayer walks Players in APNAP order. Each member
// with at least one matching permanent may participate, chooses their own
// permanent, and is the controller of that counter-placement event. Only
// permanents that actually receive counters after replacements are published.
func handleOptionalCounterForEachPlayer(r *effectResolver, prim game.OptionalCounterForEachPlayer) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	key := linkedObjectSourceKey(r.game, r.obj, string(prim.PublishLinked))
	clearLinkedObjects(r.game, key)
	if res.amount <= 0 {
		return res
	}

	source, _ := sourcePermanent(r.game, r.obj)
	resolver := newReferenceResolver(r.game, r.obj)
	members := playersInAPNAPOrder(r.game, r.playerGroupMembers(prim.Players))
	for _, member := range members {
		candidates := playerControlledSelectionCandidates(r.game, resolver, source, member, prim.Selection)
		if len(candidates) == 0 ||
			!r.engine.chooseMay(r.game, r.agents, member, "Put counters on a permanent you control?", r.log) {
			continue
		}
		chosen, ok := r.engine.chooseOnePermanent(
			r.game,
			candidates,
			member,
			"Choose a permanent to receive counters",
			r.agents,
			r.log,
		)
		if !ok {
			continue
		}

		permanent, ok := permanentByObjectID(r.game, chosen.ObjectID)
		if !ok || !permanentStillEligibleForPlayer(
			r.game,
			resolver,
			source,
			member,
			prim.Selection,
			permanent,
		) {
			continue
		}
		before := permanent.Counters.Get(prim.CounterKind)
		addCountersToPermanentControlledBy(r.game, member, permanent, prim.CounterKind, res.amount)
		placed := permanent.Counters.Get(prim.CounterKind) - before
		if placed <= 0 {
			continue
		}
		rememberLinkedObject(r.game, key, permanentObjectBindingRef(permanent))
		res.succeeded = true
	}
	return res
}

func permanentStillEligibleForPlayer(
	g *game.Game,
	resolver referenceResolver,
	source *game.Permanent,
	player game.PlayerID,
	selection game.Selection,
	permanent *game.Permanent,
) bool {
	if permanent == nil || !activeBattlefieldPermanent(permanent) || effectiveController(g, permanent) != player {
		return false
	}
	return resolver.permanentMatchesGroupSelection(&selection, source, permanent)
}
