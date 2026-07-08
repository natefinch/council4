package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// handleEachPlayerChooseDestroy resolves "Starting with you, each player may
// choose an artifact or enchantment you don't control. Destroy each permanent
// chosen this way." (Druid of Purification). Walking every player in turn order
// beginning with the resolving controller, each player is their own chooser and
// may pick up to one permanent from the single shared candidate pool — the
// battlefield permanents matching prim.Selection evaluated relative to the
// ability's controller, so "you don't control" offers every chooser the same
// permanents. The permanents chosen this way are destroyed simultaneously; a
// permanent chosen by more than one player is destroyed once, and prim.Optional
// (the "may") lets any chooser decline.
func handleEachPlayerChooseDestroy(r *effectResolver, prim game.EachPlayerChooseDestroy) effectResolved {
	res := effectResolved{accepted: true}
	source, _ := sourcePermanent(r.game, r.obj)
	resolver := newReferenceResolver(r.game, r.obj)
	pool := eachPlayerChooseCandidates(r.game, resolver, source, prim.Selection)
	if len(pool) == 0 {
		return res
	}
	seen := make(map[game.ObjectID]bool, len(pool))
	chosen := make([]*game.Permanent, 0, len(pool))
	for _, playerID := range votersStartingWith(r.game, r.obj.Controller) {
		var permanent *game.Permanent
		var ok bool
		if prim.Optional {
			permanent, ok = r.engine.chooseUpToOnePermanent(r.game, pool, playerID, "Choose a permanent", r.agents, r.log)
		} else {
			permanent, ok = r.engine.chooseOnePermanent(r.game, pool, playerID, "Choose a permanent", r.agents, r.log)
		}
		if !ok || seen[permanent.ObjectID] {
			continue
		}
		seen[permanent.ObjectID] = true
		chosen = append(chosen, permanent)
	}
	destroyed := make([]*game.Permanent, 0, len(chosen))
	for _, permanent := range chosen {
		if hasKeyword(r.game, permanent, game.Indestructible) || replaceDestroyPermanent(r.game, permanent, prim.PreventRegeneration) {
			continue
		}
		destroyed = append(destroyed, permanent)
	}
	res.succeeded = movePermanentsToZoneSimultaneously(r.game, destroyed, zone.Graveyard)
	res.amount = len(destroyed)
	return res
}

// handleDestroyForEachPlayer resolves the distributive Saga chapter "For each
// player, destroy up to one target creature that player controls." (The Curse of
// Fenric, chapter I). Walking every player in APNAP order, prim.Chooser picks up
// to one permanent that player controls matching prim.Selection and the runtime
// destroys it, remembering each destroyed permanent under prim.LinkedKey keyed by
// the source so the paired CreateTokenForEachDestroyed mints exactly one token
// per destroyed creature under its last-known controller. The link is not cleared
// here; the token clause consumes it.
func handleDestroyForEachPlayer(r *effectResolver, prim game.DestroyForEachPlayer) effectResolved {
	res := effectResolved{accepted: true}
	chooser, ok := r.resolvePlayer(prim.Chooser)
	if !ok {
		return res
	}
	source, _ := sourcePermanent(r.game, r.obj)
	resolver := newReferenceResolver(r.game, r.obj)
	key := linkedObjectSourceKey(r.game, r.obj, string(prim.LinkedKey))
	for _, playerID := range playersInAPNAPOrder(r.game, r.playerGroupMembers(game.AllPlayersReference())) {
		candidates := playerControlledSelectionCandidates(r.game, resolver, source, playerID, prim.Selection)
		permanent, chosen := r.engine.chooseUpToOnePermanent(r.game, candidates, chooser, "Choose a permanent to destroy", r.agents, r.log)
		if !chosen {
			continue
		}
		// permanentObjectBindingRef preserves the ObjectID even for a token
		// (CardInstanceID == 0) so the paired CreateTokenForEachDestroyed still
		// mints a token for a destroyed token permanent's controller;
		// permanentLinkedObjectRef would drop tokens and silently skip the payoff.
		linkedRef := permanentObjectBindingRef(permanent)
		if _, destroyed := destroyPermanentInBatch(r.game, permanent.ObjectID, 0, false); destroyed {
			rememberLinkedObject(r.game, key, linkedRef)
			res.succeeded = true
		}
	}
	return res
}

// handleCreateTokenForEachDestroyed resolves the per-controller Saga payoff "For
// each creature destroyed this way, its controller creates a <token>." (The Curse
// of Fenric, chapter I). For every permanent a sibling DestroyForEachPlayer
// recorded under prim.LinkedKey, the destroyed permanent's last-known controller
// creates one token defined by prim.Source. It clears the link afterward so the
// payoff fires exactly once for the linked set.
func handleCreateTokenForEachDestroyed(r *effectResolver, prim game.CreateTokenForEachDestroyed) effectResolved {
	res := effectResolved{accepted: true}
	token, ok := r.typedTokenDefinition(prim.Source)
	if !ok {
		return res
	}
	key := linkedObjectSourceKey(r.game, r.obj, string(prim.LinkedKey))
	for _, ref := range linkedObjects(r.game, key) {
		snapshot, ok := lastKnownObject(r.game, ref.ObjectID)
		if !ok {
			continue
		}
		if _, created := createTokenPermanentsCollectingWithChoices(r.engine, r.game, snapshot.Controller, token, 1, false, r.agents, r.log); created {
			res.succeeded = true
		}
	}
	clearLinkedObjects(r.game, key)
	return res
}
