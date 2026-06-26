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
	refs := make([]game.LinkedObjectRef, 0, len(r.obj.Targets))
	for i := range r.obj.Targets {
		permanent, ok := r.resolveObject(game.TargetPermanentReference(i))
		if !ok {
			continue
		}
		if !prim.Exile {
			if hasKeyword(r.game, permanent, game.Indestructible) ||
				replaceDestroyPermanent(r.game, permanent, prim.PreventRegeneration) {
				continue
			}
		}
		removing = append(removing, permanent)
		refs = append(refs, permanentLinkedObjectRef(permanent))
	}
	destination := zone.Graveyard
	if prim.Exile {
		destination = zone.Exile
	}
	if movePermanentsToZoneSimultaneously(r.game, removing, destination) {
		res.succeeded = true
		res.amount = len(removing)
	}
	for _, ref := range refs {
		rememberLinkedObject(r.game, key, ref)
	}
	return res
}
