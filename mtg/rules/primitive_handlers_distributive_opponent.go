package rules

import (
	"github.com/natefinch/council4/mtg/game"
)

// handleDrawForEachExiled resolves the per-controller payoff "For each permanent
// exiled this way, its controller draws a card." (King Solomon's Frogs). For
// every permanent a sibling ExileForEachOpponent recorded under prim.LinkedKey,
// the exiled permanent's last-known controller draws one card. It clears the
// link afterward so the payoff fires exactly once for the linked set.
func handleDrawForEachExiled(r *effectResolver, prim game.DrawForEachExiled) effectResolved {
	res := effectResolved{accepted: true}
	key := linkedObjectSourceKey(r.game, r.obj, string(prim.LinkedKey))
	for _, ref := range linkedObjects(r.game, key) {
		snapshot, ok := lastKnownObject(r.game, ref.ObjectID)
		if !ok {
			continue
		}
		if r.engine.drawCards(r.game, snapshot.Controller, 1, r.agents, r.log) {
			res.succeeded = true
		}
	}
	clearLinkedObjects(r.game, key)
	return res
}

// handleManifestForEachLinked resolves a per-controller face-down payoff. For
// every permanent a sibling removal recorded under prim.LinkedKey, that
// permanent's last-known controller manifests or cloaks one card.
func handleManifestForEachLinked(r *effectResolver, prim game.ManifestForEachLinked) effectResolved {
	res := effectResolved{accepted: true}
	key := linkedObjectSourceKey(r.game, r.obj, string(prim.LinkedKey))
	for _, ref := range linkedObjects(r.game, key) {
		snapshot, ok := lastKnownObject(r.game, ref.ObjectID)
		if !ok {
			continue
		}
		if _, manifested := manifestForPlayer(r, snapshot.Controller, prim.Dread, prim.Cloak); manifested {
			res.succeeded = true
		}
	}
	clearLinkedObjects(r.game, key)
	return res
}
