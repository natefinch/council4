package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

// handleExileTargetSpells exiles the spell chosen for a single stack-object
// target or every spell chosen for a variable-count stack-object target group
// ("Exile any number of target spells.", Mindbreak Trap).
//
// It is not a counter (CR 701.6): each targeted spell's physical card is moved
// from the stack to exile through the normal zone-change replacements, so spells
// that can't be countered are still exiled. All targeted stack objects are
// resolved to their IDs before any spell is removed, so removing one spell can
// neither reorder the stack nor cause a still-listed target to resolve. Targets
// that are no longer legal (deferred at resolution) or already gone are skipped,
// and a group with zero chosen targets resolves as a legal no-op.
func handleExileTargetSpells(r *effectResolver, prim game.ExileTargetSpells) effectResolved {
	res := effectResolved{accepted: true}
	for _, stackObjectID := range r.exileSpellTargetStackObjectIDs(prim.Object) {
		if exileStackSpell(r.game, stackObjectID) {
			res.succeeded = true
		}
	}
	return res
}

// exileSpellTargetStackObjectIDs resolves the stack-object IDs an
// ExileTargetSpells effect acts on: the single chosen stack-object target
// (ObjectReferenceTargetStackObject) or every stack object chosen for the
// variable-count target group (ObjectReferenceAllTargetStackObjects). Every ID
// is gathered from the current targets up front so the caller can remove each
// spell without invalidating another target's lookup.
func (r *effectResolver) exileSpellTargetStackObjectIDs(ref game.ObjectReference) []id.ID {
	switch ref.Kind() {
	case game.ObjectReferenceTargetStackObject:
		if stackObjectID, ok := effectStackObjectID(r.game, r.obj, ref.TargetIndex()); ok {
			return []id.ID{stackObjectID}
		}
		return nil
	case game.ObjectReferenceAllTargetStackObjects:
		return r.targetSpecStackObjectIDs(ref.TargetIndex())
	default:
		return nil
	}
}

// targetSpecStackObjectIDs gathers the stack-object ID chosen for each target of
// the target spec at specIndex, slicing the flat chosen-target list by the
// per-spec counts exactly as targetSpecPermanents does for permanents. It backs
// the all-target-stack-objects group reference so a group exile can act on every
// chosen spell under one effect. Targets that are not (or are no longer) stack
// objects — including markers deferred at resolution for becoming illegal — are
// skipped, so a partially illegal group exiles only its still-legal spells.
func (r *effectResolver) targetSpecStackObjectIDs(specIndex int) []id.ID {
	if r.obj == nil {
		return nil
	}
	all := r.obj.Targets
	start, end := 0, len(all)
	if counts := r.obj.TargetCounts; specIndex >= 0 && specIndex < len(counts) {
		start = 0
		for i := range specIndex {
			start += counts[i]
		}
		end = start + counts[specIndex]
	}
	if start < 0 || end > len(all) || start > end {
		return nil
	}
	ids := make([]id.ID, 0, end-start)
	for i := start; i < end; i++ {
		if all[i].Kind != game.TargetStackObject || all[i].StackObjectID == 0 {
			continue
		}
		ids = append(ids, all[i].StackObjectID)
	}
	return ids
}

// exileStackSpell removes one spell from the stack and exiles its card, honoring
// the normal stack zone-change replacements (flashback/exile-on-resolution and
// commander) applied by moveStackCardToZone. It is not a counter, so it never
// consults whether the spell can be countered: an uncounterable spell is still
// exiled. A spell copy has no card and simply ceases to exist, the same way a
// countered or bounced copy disappears. Non-spell stack objects and objects no
// longer on the stack are left untouched.
func exileStackSpell(g *game.Game, stackObjectID id.ID) bool {
	obj, ok := stackObjectByID(g, stackObjectID)
	if !ok || obj.Kind != game.StackSpell {
		return false
	}
	removed, ok := g.Stack.RemoveByID(stackObjectID)
	if !ok {
		return false
	}
	if removed.Copy {
		return true
	}
	card, ok := g.GetCardInstance(removed.SourceID)
	if !ok {
		return false
	}
	return moveStackCardToZone(g, removed, card, zone.Exile, false)
}
