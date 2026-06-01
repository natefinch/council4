package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

type resolvedObjectReference struct {
	permanent *game.Permanent
	snapshot  game.ObjectSnapshot
}

func (r resolvedObjectReference) controller(g *game.Game) (game.PlayerID, bool) {
	if r.permanent != nil {
		return effectiveController(g, r.permanent), true
	}
	if r.snapshot.ObjectID != 0 {
		return r.snapshot.Controller, true
	}
	return 0, false
}

func (r resolvedObjectReference) owner() (game.PlayerID, bool) {
	if r.permanent != nil {
		return r.permanent.Owner, true
	}
	if r.snapshot.ObjectID != 0 {
		return r.snapshot.Owner, true
	}
	return 0, false
}

func resolveObjectReference(g *game.Game, obj *game.StackObject, ref game.ObjectReference) (resolvedObjectReference, bool) {
	switch ref.Kind {
	case game.ObjectReferenceTargetPermanent:
		objectID, ok := targetPermanentObjectID(obj, ref.TargetIndex)
		if !ok {
			return resolvedObjectReference{}, false
		}
		return resolvePermanentOrLastKnown(g, objectID)
	case game.ObjectReferenceSourcePermanent:
		if permanent, ok := sourcePermanent(g, obj); ok {
			return resolvedObjectReference{permanent: permanent}, true
		}
		return resolvePermanentOrLastKnown(g, obj.SourceID)
	case game.ObjectReferenceAttachedPermanent:
		objectID, ok := attachedPermanentObjectID(g, obj, ref)
		if !ok {
			return resolvedObjectReference{}, false
		}
		return resolvePermanentOrLastKnown(g, objectID)
	case game.ObjectReferenceLinkedObject:
		for _, linked := range linkedObjects(g, linkedObjectSourceKey(g, obj, ref.LinkID)) {
			if resolved, ok := resolvePermanentOrLastKnown(g, linked.ObjectID); ok {
				return resolved, true
			}
		}
		return resolvedObjectReference{}, false
	default:
		return resolvedObjectReference{}, false
	}
}

func resolvePermanentOrLastKnown(g *game.Game, objectID id.ID) (resolvedObjectReference, bool) {
	if permanent, ok := permanentByObjectID(g, objectID); ok {
		return resolvedObjectReference{permanent: permanent}, true
	}
	if snapshot, ok := lastKnownObject(g, objectID); ok {
		return resolvedObjectReference{snapshot: snapshot}, true
	}
	return resolvedObjectReference{}, false
}

func targetPermanentObjectID(obj *game.StackObject, targetIndex int) (id.ID, bool) {
	if targetIndex < 0 || targetIndex >= len(obj.Targets) {
		return 0, false
	}
	target := obj.Targets[targetIndex]
	if target.Kind != game.TargetPermanent {
		return 0, false
	}
	return target.PermanentID, true
}

func attachedPermanentObjectID(g *game.Game, obj *game.StackObject, ref game.ObjectReference) (id.ID, bool) {
	var permanent *game.Permanent
	var ok bool
	if ref.TargetIndex >= 0 {
		objectID, targetOK := targetPermanentObjectID(obj, ref.TargetIndex)
		if !targetOK {
			return 0, false
		}
		permanent, ok = permanentByObjectID(g, objectID)
	} else {
		permanent, ok = sourcePermanent(g, obj)
	}
	if !ok || !permanent.AttachedTo.Exists {
		if ref.TargetIndex < 0 && obj.HasTriggerEvent && obj.TriggerEvent.PermanentID != 0 {
			return obj.TriggerEvent.PermanentID, true
		}
		return 0, false
	}
	return permanent.AttachedTo.Val, true
}

func resolvePlayerReference(g *game.Game, obj *game.StackObject, ref game.PlayerReference) (game.PlayerID, bool) {
	var playerID game.PlayerID
	var ok bool
	switch ref.Kind {
	case game.PlayerReferenceController:
		playerID, ok = obj.Controller, true
	case game.PlayerReferenceTargetPlayer:
		playerID, ok = targetPlayer(g, obj, ref.TargetIndex)
	case game.PlayerReferenceObjectController:
		playerID, ok = referencedObjectController(g, obj, ref)
	case game.PlayerReferenceObjectOwner:
		playerID, ok = referencedObjectOwner(g, obj, ref)
	default:
		return 0, false
	}
	if !ok || !isPlayerAlive(g, playerID) {
		return 0, false
	}
	return playerID, true
}

func targetPlayer(g *game.Game, obj *game.StackObject, targetIndex int) (game.PlayerID, bool) {
	if targetIndex < 0 || targetIndex >= len(obj.Targets) {
		return 0, false
	}
	target := obj.Targets[targetIndex]
	if target.Kind != game.TargetPlayer || !isPlayerAlive(g, target.PlayerID) {
		return 0, false
	}
	return target.PlayerID, true
}

func referencedObjectController(g *game.Game, obj *game.StackObject, ref game.PlayerReference) (game.PlayerID, bool) {
	if !ref.Object.Exists {
		return 0, false
	}
	resolved, ok := resolveObjectReference(g, obj, ref.Object.Val)
	if !ok {
		return 0, false
	}
	return resolved.controller(g)
}

func referencedObjectOwner(g *game.Game, obj *game.StackObject, ref game.PlayerReference) (game.PlayerID, bool) {
	if !ref.Object.Exists {
		return 0, false
	}
	resolved, ok := resolveObjectReference(g, obj, ref.Object.Val)
	if !ok {
		return 0, false
	}
	return resolved.owner()
}
