package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// handleExileForEachOpponent resolves the distributive enters trigger "for each
// opponent, exile up to one target permanent that player controls with mana
// value 3 or greater." (King Solomon's Frogs). Walking each opponent of the
// resolving controller in APNAP order, prim.Chooser picks up to one permanent
// that opponent controls matching prim.Selection and the runtime exiles it,
// remembering each exiled permanent under prim.LinkedKey keyed by the source so
// the paired DrawForEachExiled draws one card for each exiled permanent's
// last-known controller. Each opponent's permanents are an independent candidate
// pool, so the trigger exiles at most one per opponent. The link is not cleared
// here; the draw clause consumes it.
func handleExileForEachOpponent(r *effectResolver, prim game.ExileForEachOpponent) effectResolved {
	res := effectResolved{accepted: true}
	source, _ := sourcePermanent(r.game, r.obj)
	resolver := newReferenceResolver(r.game, r.obj)
	key := linkedObjectSourceKey(r.game, r.obj, string(prim.LinkedKey))
	clearLinkedObjects(r.game, key)
	type choice struct {
		player    game.PlayerID
		permanent *game.Permanent
		ref       game.LinkedObjectRef
	}
	var choices []choice
	prevMember := r.groupOfferMember
	defer func() { r.groupOfferMember = prevMember }()
	for _, playerID := range playersInAPNAPOrder(r.game, r.playerGroupMembers(game.OpponentsReference())) {
		r.groupOfferMember = opt.Val(playerID)
		chooser, ok := r.resolvePlayer(prim.Chooser)
		if !ok {
			continue
		}
		candidates := playerControlledSelectionCandidates(r.game, resolver, source, playerID, prim.Selection)
		candidates = permanentChoiceExtremumCandidates(r.game, candidates, prim.Extremum)
		var permanent *game.Permanent
		var chosen bool
		if prim.Required {
			permanent, chosen = r.engine.chooseOnePermanent(r.game, candidates, chooser, "Choose a permanent to exile", r.agents, r.log)
		} else {
			permanent, chosen = r.engine.chooseUpToOnePermanent(r.game, candidates, chooser, "Choose a permanent to exile", r.agents, r.log)
		}
		if !chosen {
			continue
		}
		// permanentObjectBindingRef preserves the ObjectID even for a token
		// (CardInstanceID == 0) so the paired DrawForEachExiled still draws for a
		// token permanent's controller. permanentLinkedObjectRef drops tokens,
		// which would silently deny that opponent the guaranteed draw.
		linkedRef := permanentObjectBindingRef(permanent)
		linkedRef.CorrelatedPlayer = opt.Val(playerID)
		if prim.Simultaneous {
			choices = append(choices, choice{player: playerID, permanent: permanent, ref: linkedRef})
			continue
		}
		move, ok := preparePermanentZoneMove(r.game, permanent, zone.Exile)
		if ok && applyPreparedPermanentZoneMove(r.game, &move) {
			if move.actualDestination == zone.Exile {
				rememberLinkedObject(r.game, key, linkedRef)
			}
			res.succeeded = true
		}
	}
	if !prim.Simultaneous || len(choices) == 0 {
		return res
	}
	permanents := make([]*game.Permanent, len(choices))
	refsByObject := make(map[game.ObjectID]game.LinkedObjectRef, len(choices))
	for i, chosen := range choices {
		permanents[i] = chosen.permanent
		refsByObject[chosen.permanent.ObjectID] = chosen.ref
	}
	for _, result := range movePermanentsToZoneSimultaneouslyWithResults(r.game, permanents, zone.Exile) {
		if !result.moved {
			continue
		}
		res.succeeded = true
		if result.destination == zone.Exile {
			rememberLinkedObject(r.game, key, refsByObject[result.permanent.ObjectID])
		}
	}
	return res
}

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
