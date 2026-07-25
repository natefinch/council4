package rules

import (
	"github.com/natefinch/council4/mtg/game"
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
	batch := &destroyBatch{game: r.game, simultaneousID: r.game.IDGen.Next()}
	destroyed, replacements := planDestroyPermanents(r.game, chosen, prim.PreventRegeneration, batch.simultaneousID)
	res.succeeded = applyPlannedDestroyBatch(r.game, destroyed, replacements, batch)
	res.amount = len(destroyed)
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
