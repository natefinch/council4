package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// handleRemoveTargetsForToken resolves the variable-target removal-token family
// "Destroy any number of target creatures." (Descent of the Dragons) and "Exile X
// target creatures." (Curse of the Swine). It removes every permanent chosen for
// the spell's lone variable-count target spec, destroying them (or exiling them
// when prim.Exile is set) as one simultaneous event, and remembers each removed
// permanent under prim.LinkedKey keyed by the source so the paired
// CreateTokenForEachDestroyed clause mints exactly one token per removed creature
// under its last-known controller. The destroy form honors indestructibility and
// regeneration replacements just like an ordinary group destroy; the exile form
// removes every chosen permanent. The link is not cleared here; the token clause
// consumes it.
func handleRemoveTargetsForToken(r *effectResolver, prim game.RemoveTargetsForToken) effectResolved {
	res := effectResolved{accepted: true}
	key := linkedObjectSourceKey(r.game, r.obj, string(prim.LinkedKey))
	removing := make([]*game.Permanent, 0, len(r.obj.Targets))
	var candidates []*game.Permanent
	for i := range r.obj.Targets {
		permanent, ok := r.resolveObject(game.TargetPermanentReference(i))
		if !ok {
			continue
		}
		if prim.Exile {
			removing = append(removing, permanent)
		} else {
			candidates = append(candidates, permanent)
		}
	}
	var batch *destroyBatch
	var replacements []plannedDestroyReplacement
	if !prim.Exile {
		batch = &destroyBatch{game: r.game, simultaneousID: r.game.IDGen.Next()}
		removing, replacements = planDestroyPermanents(r.game, candidates, prim.PreventRegeneration, batch.simultaneousID)
	}
	destination := zone.Graveyard
	if prim.Exile {
		destination = zone.Exile
	}
	var refs []game.LinkedObjectRef
	movedCount := 0
	if prim.Exile {
		results := movePermanentsToZoneSimultaneouslyWithResults(r.game, removing, destination)
		for _, result := range results {
			if !result.moved || result.destination != zone.Exile {
				continue
			}
			refs = append(refs, permanentObjectBindingRef(result.permanent))
			movedCount++
		}
	} else {
		for _, permanent := range removing {
			// permanentObjectBindingRef preserves the ObjectID even for a token
			// (CardInstanceID == 0) so the paired CreateTokenForEachDestroyed still
			// mints a token for a removed token permanent's controller.
			refs = append(refs, permanentObjectBindingRef(permanent))
		}
		if applyPlannedDestroyBatch(r.game, removing, replacements, batch) {
			movedCount = len(removing)
		}
	}
	if movedCount > 0 {
		res.succeeded = true
		res.amount = movedCount
	}
	for _, ref := range refs {
		rememberLinkedObject(r.game, key, ref)
	}
	return res
}
