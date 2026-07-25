package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// handleForEachPlayer resolves the whole distributive removal template "For each
// <player group>, <verb> up to one target permanent that player controls."
// Walking prim.Scope in APNAP order, prim.Chooser picks from that player's own
// candidate pool and prim.Removal is applied, with every removed permanent
// remembered under prim.LinkedKey keyed by the source so a paired payoff can
// consume the set. The link is not cleared here; the payoff clause consumes it.
//
// It replaced three near-identical handlers that differed only in scope and
// verb. The surviving behaviour is the union: the chooser is resolved once per
// group member with that member bound as the group-offer member, choices are
// applied through the prepared zone-move path so a replacement effect that
// redirects the removal is honoured, and only permanents that actually reached
// the intended destination are linked.
func handleForEachPlayer(r *effectResolver, prim game.ForEachPlayer) effectResolved {
	res := effectResolved{accepted: true}
	source, _ := sourcePermanent(r.game, r.obj)
	resolver := newReferenceResolver(r.game, r.obj)
	key := linkedObjectSourceKey(r.game, r.obj, string(prim.LinkedKey))
	if prim.ReplaceLink {
		clearLinkedObjects(r.game, key)
	}
	prompt := "Choose a permanent to exile"
	if prim.Removal == game.DistributiveRemovalDestroy {
		prompt = "Choose a permanent to destroy"
	}

	prevMember := r.groupOfferMember
	defer func() { r.groupOfferMember = prevMember }()

	var deferred []distributiveChoice
	for _, playerID := range playersInAPNAPOrder(r.game, r.playerGroupMembers(prim.Scope)) {
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
			permanent, chosen = r.engine.chooseOnePermanent(r.game, candidates, chooser, prompt, r.agents, r.log)
		} else {
			permanent, chosen = r.engine.chooseUpToOnePermanent(r.game, candidates, chooser, prompt, r.agents, r.log)
		}
		if !chosen {
			continue
		}
		// permanentObjectBindingRef preserves the ObjectID even for a token
		// (CardInstanceID == 0) so a paired per-controller payoff still fires for a
		// removed token permanent's controller; permanentLinkedObjectRef drops
		// tokens, which would silently skip that player's payoff. Return-style
		// consumers naturally ignore token entries because they have no CardID.
		linkedRef := permanentObjectBindingRef(permanent)
		linkedRef.CorrelatedPlayer = opt.Val(playerID)
		if prim.Simultaneous {
			deferred = append(deferred, distributiveChoice{permanent: permanent, ref: linkedRef})
			continue
		}
		if applyDistributiveRemoval(r, prim.Removal, permanent, key, linkedRef) {
			res.succeeded = true
		}
	}
	if len(deferred) == 0 {
		return res
	}
	if applyDistributiveRemovalBatch(r, prim.Removal, deferred, key) {
		res.succeeded = true
	}
	return res
}

// distributiveChoice pairs a chosen permanent with the linked reference to
// publish when its removal succeeds.
type distributiveChoice struct {
	permanent *game.Permanent
	ref       game.LinkedObjectRef
}

// applyDistributiveRemoval removes one chosen permanent and publishes its linked
// reference only when the permanent actually reached the intended destination,
// so a replacement effect that redirects the removal does not leave a paired
// payoff acting on a permanent that never went anywhere.
func applyDistributiveRemoval(r *effectResolver, removal game.DistributiveRemoval, permanent *game.Permanent, key game.LinkedObjectKey, ref game.LinkedObjectRef) bool {
	if removal == game.DistributiveRemovalDestroy {
		if _, destroyed := destroyPermanentInBatch(r.game, permanent.ObjectID, 0, false); destroyed {
			rememberLinkedObject(r.game, key, ref)
			return true
		}
		return false
	}
	move, ok := preparePermanentZoneMove(r.game, permanent, zone.Exile)
	if !ok || !applyPreparedPermanentZoneMove(r.game, &move) {
		return false
	}
	if move.actualDestination == zone.Exile {
		rememberLinkedObject(r.game, key, ref)
	}
	return true
}

// applyDistributiveRemovalBatch removes every deferred choice in one
// simultaneous batch, so the whole distribution is a single event for
// replacement and trigger purposes.
//
// A destroy is not a graveyard move. Routing one through the plain zone-move
// batch would destroy an indestructible permanent, ignore a regeneration
// shield, and skip destroy-replacement and commander redirection, all of which
// live in the destroy planner rather than in the move. The two removals
// therefore share the apply step but not the plan step.
func applyDistributiveRemovalBatch(r *effectResolver, removal game.DistributiveRemoval, choices []distributiveChoice, key game.LinkedObjectKey) bool {
	permanents := make([]*game.Permanent, len(choices))
	refsByObject := make(map[game.ObjectID]game.LinkedObjectRef, len(choices))
	for i, chosen := range choices {
		permanents[i] = chosen.permanent
		refsByObject[chosen.permanent.ObjectID] = chosen.ref
	}
	var results []permanentZoneMoveResult
	destination := zone.Exile
	if removal == game.DistributiveRemovalDestroy {
		destination = zone.Graveyard
		results = applyDistributiveDestroyBatch(r, permanents)
	} else {
		results = movePermanentsToZoneSimultaneouslyWithResults(r.game, permanents, destination)
	}
	succeeded := false
	for _, result := range results {
		if !result.moved {
			continue
		}
		succeeded = true
		if result.destination == destination {
			rememberLinkedObject(r.game, key, refsByObject[result.permanent.ObjectID])
		}
	}
	return succeeded
}

// applyDistributiveDestroyBatch destroys permanents simultaneously through the
// destroy planner, so indestructible permanents, regeneration shields, and
// destroy replacements are honoured, and reports per-permanent results so the
// caller links only the permanents that actually died.
//
// It plans and applies inline rather than calling applyPlannedDestroyBatch
// because that helper collapses the outcome to a single bool, and the linked
// payoff needs to know which permanents reached the graveyard.
func applyDistributiveDestroyBatch(r *effectResolver, permanents []*game.Permanent) []permanentZoneMoveResult {
	batch := &destroyBatch{game: r.game, simultaneousID: r.game.IDGen.Next()}
	destroyed, replacements := planDestroyPermanents(r.game, permanents, false, batch.simultaneousID)
	moves := preparePermanentZoneMovesInBatch(r.game, destroyed, zone.Graveyard, batch.simultaneousID)
	for _, planned := range replacements {
		batch.changed = applyDestroyReplacement(r.game, planned.permanent, planned.replacement, batch) || batch.changed
	}
	batch.applyQueued()
	return applyPreparedPermanentZoneMoves(r.game, moves)
}
