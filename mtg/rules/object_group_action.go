package rules

import "github.com/natefinch/council4/mtg/game"

// objectGroupTargets holds the permanents an object or group effect acts on. It
// is produced by resolveObjectGroup so the destroy, exile, bounce, tap, and
// phase-out handlers share one single-vs-group resolution path instead of each
// re-deriving it inline. The single and group forms keep distinct
// simultaneous-event batching (CR 603.3b), so a handler that batches a group
// move under one SimultaneousID but a single move under none branches on single.
type objectGroupTargets struct {
	// permanents are the permanents the rules action applies to: one element for
	// a resolved single ObjectReference, or the full group membership for a
	// GroupReference (possibly empty).
	permanents []*game.Permanent
	// single reports that the effect named a single ObjectReference (the group
	// reference was not valid), so the single-object terminal/batching applies.
	single bool
	// resolved reports that a single ObjectReference resolved to a permanent. It
	// is meaningful only when single is true; a group always counts as resolved.
	resolved bool
}

// resolveObjectGroup resolves the object or group an effect acts on into the
// permanents its rules action applies to. A valid GroupReference always takes
// the group form, even when it resolves to no permanents, matching the legacy
// prim.Group.Valid() branch; otherwise the ObjectReference is resolved to a
// single permanent.
//
// Resolution is shared, but the explicit rules action — destroy, exile, bounce,
// tap, phase out — stays with the calling handler so each cause's terminal
// events, replacements, and simultaneous batching remain authoritative and
// distinct (the handlers must not collapse into one opaque zone move).
func (r *effectResolver) resolveObjectGroup(object game.ObjectReference, group game.GroupReference) objectGroupTargets {
	if group.Valid() {
		return objectGroupTargets{permanents: r.groupPermanents(group)}
	}
	if object.Kind() == game.ObjectReferenceAllTargetPermanents {
		return objectGroupTargets{permanents: r.targetSpecPermanents(object.TargetIndex())}
	}
	if permanent, ok := r.resolveObject(object); ok {
		return objectGroupTargets{permanents: []*game.Permanent{permanent}, single: true, resolved: true}
	}
	return objectGroupTargets{single: true}
}

// targetSpecPermanents gathers every battlefield permanent chosen for the target
// spec at specIndex, slicing the flat chosen-target list by the per-spec counts
// the same way dividedTargets does. It backs the all-target-permanents group
// reference, so the group-blink exile can capture every chosen permanent under
// one linked key. Targets that have left the battlefield since selection are
// skipped.
func (r *effectResolver) targetSpecPermanents(specIndex int) []*game.Permanent {
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
	permanents := make([]*game.Permanent, 0, end-start)
	for i := start; i < end; i++ {
		if all[i].Kind != game.TargetPermanent {
			continue
		}
		if permanent, ok := permanentByObjectID(r.game, all[i].PermanentID); ok {
			permanents = append(permanents, permanent)
		}
	}
	return permanents
}
