package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

// referenceResolver is the internal module that binds a game and a resolving
// stack object and owns every runtime reference lookup: target-slot objects and
// players, source and attached permanents, linked objects, event/last-known
// information, alive-player checks, and group enumeration with Selection
// matching and exclusions. It is created per resolution; the resolveObjectReference,
// resolvePlayerReference, permanentAt, and playerAt entry points are thin
// adapters over it.
//
// By default the resolver derives the source permanent from the stack object
// (obj.SourceID). A caller whose source basis differs from the stack source,
// such as Damage resolving a group relative to its own DamageSource, builds the
// resolver with newReferenceResolverWithSource so source-dependent references
// (the source permanent and the EquippedCreature/Other... selectors) bind to the
// caller's permanent, including an explicit nil that resolves to no source.
type referenceResolver struct {
	g   *game.Game
	obj *game.StackObject

	// source overrides the stack-derived source permanent when sourceFixed is
	// set. A nil source with sourceFixed true means "no source permanent".
	source      *game.Permanent
	sourceFixed bool
}

func newReferenceResolver(g *game.Game, obj *game.StackObject) referenceResolver {
	return referenceResolver{g: g, obj: obj}
}

// newReferenceResolverWithSource builds a resolver whose source permanent is
// fixed to source rather than derived from obj.SourceID. Passing a nil source
// fixes the basis to "no source permanent", preserving the legacy behavior of
// source-dependent selectors when the calling context has no source object.
func newReferenceResolverWithSource(g *game.Game, obj *game.StackObject, source *game.Permanent) referenceResolver {
	return referenceResolver{g: g, obj: obj, source: source, sourceFixed: true}
}

type resolvedObjectReference struct {
	permanent            *game.Permanent
	snapshot             game.ObjectSnapshot
	stack                *game.StackObject
	stackController      game.PlayerID
	stackControllerKnown bool
}

func (r *resolvedObjectReference) controller(g *game.Game) (game.PlayerID, bool) {
	if r.permanent != nil {
		return effectiveController(g, r.permanent), true
	}
	if r.snapshot.ObjectID != 0 {
		return r.snapshot.Controller, true
	}
	if r.stack != nil {
		return r.stack.Controller, true
	}
	if r.stackControllerKnown {
		return r.stackController, true
	}
	return 0, false
}

func (r *resolvedObjectReference) owner(g *game.Game) (game.PlayerID, bool) {
	if r.permanent != nil {
		return r.permanent.Owner, true
	}
	if r.snapshot.ObjectID != 0 {
		return r.snapshot.Owner, true
	}
	if r.stack != nil {
		if r.stack.Kind != game.StackSpell || r.stack.Copy {
			return r.stack.Controller, true
		}
		if r.stack.SourceCardID != 0 {
			if card, ok := g.GetCardInstance(r.stack.SourceCardID); ok {
				return card.Owner, true
			}
		}
		if card, ok := g.GetCardInstance(r.stack.SourceID); ok {
			return card.Owner, true
		}
		return r.stack.Controller, true
	}
	return 0, false
}

// object resolves an ObjectReference to a live permanent or its last-known
// snapshot.
func (r referenceResolver) object(ref game.ObjectReference) (resolvedObjectReference, bool) {
	switch ref.Kind() {
	case game.ObjectReferenceTargetPermanent:
		objectID, ok := targetPermanentObjectID(r.obj, ref.TargetIndex())
		if !ok {
			return resolvedObjectReference{}, false
		}
		return resolvePermanentOrLastKnown(r.g, objectID)
	case game.ObjectReferenceTargetStackObject:
		objectID, ok := effectStackObjectID(r.obj, ref.TargetIndex())
		if !ok {
			controller, known := r.obj.TargetControllerLKI[ref.TargetIndex()]
			return resolvedObjectReference{
				stackController:      controller,
				stackControllerKnown: known,
			}, known
		}
		stackObject, ok := stackObjectByID(r.g, objectID)
		if ok {
			return resolvedObjectReference{stack: stackObject}, true
		}
		controller, ok := r.obj.TargetControllerLKI[ref.TargetIndex()]
		return resolvedObjectReference{
			stackController:      controller,
			stackControllerKnown: ok,
		}, ok
	case game.ObjectReferenceSourcePermanent:
		if r.sourceFixed {
			if r.source == nil || r.source.PhasedOut {
				return resolvedObjectReference{}, false
			}
			return resolvedObjectReference{permanent: r.source}, true
		}
		if r.obj == nil {
			return resolvedObjectReference{}, false
		}
		return resolveSourcePermanentOrLastKnown(r.g, r.obj.SourceID)
	case game.ObjectReferenceSourceCard:
		if r.obj == nil || r.obj.SourceCardID == 0 {
			return resolvedObjectReference{}, false
		}
		for _, permanent := range r.g.Battlefield {
			if permanent.CardInstanceID == r.obj.SourceCardID {
				return resolvedObjectReference{permanent: permanent}, true
			}
		}
		return resolvedObjectReference{}, false
	case game.ObjectReferenceSourceAttachedPermanent:
		permanent, ok := r.sourcePermanent()
		if !ok || !permanent.AttachedTo.Exists {
			return resolvedObjectReference{}, false
		}
		return resolvePermanentOrLastKnown(r.g, permanent.AttachedTo.Val)
	case game.ObjectReferenceTargetAttachedPermanent:
		objectID, ok := targetPermanentObjectID(r.obj, ref.TargetIndex())
		if !ok {
			return resolvedObjectReference{}, false
		}
		permanent, ok := permanentByObjectID(r.g, objectID)
		if !ok || !permanent.AttachedTo.Exists {
			return resolvedObjectReference{}, false
		}
		return resolvePermanentOrLastKnown(r.g, permanent.AttachedTo.Val)
	case game.ObjectReferenceLinkedObject:
		for _, linked := range linkedObjects(r.g, linkedObjectSourceKey(r.g, r.obj, ref.LinkID())) {
			if resolved, ok := resolvePermanentOrLastKnown(r.g, linked.ObjectID); ok {
				return resolved, true
			}
			// A card-only linked reference (ObjectID zero) names a card that was
			// never a battlefield permanent, such as a card exiled straight from a
			// graveyard (The Aesir Escape Valhalla). It carries no object snapshot,
			// so resolve it to a card snapshot the printed-characteristic readers
			// (mana value) consult through the card instance.
			if linked.ObjectID == 0 && linked.CardID != 0 {
				if _, ok := r.g.GetCardInstance(linked.CardID); ok {
					return resolvedObjectReference{snapshot: game.ObjectSnapshot{CardID: linked.CardID}}, true
				}
			}
			if linked.CardID != 0 {
				return resolvedObjectReference{snapshot: game.ObjectSnapshot{
					CardID: linked.CardID,
					Face:   game.FaceFront,
				}}, true
			}
		}
		return resolvedObjectReference{}, false
	case game.ObjectReferenceEventPermanent:
		if r.obj != nil && r.obj.HasTriggerEvent && r.obj.TriggerEvent.PermanentID != 0 {
			return resolvePermanentOrLastKnown(r.g, r.obj.TriggerEvent.PermanentID)
		}
		return resolvedObjectReference{}, false
	case game.ObjectReferenceEventRelatedPermanent:
		if r.obj != nil && r.obj.HasTriggerEvent && r.obj.TriggerEvent.RelatedPermanentID != 0 {
			return resolvePermanentOrLastKnown(r.g, r.obj.TriggerEvent.RelatedPermanentID)
		}
		return resolvedObjectReference{}, false
	case game.ObjectReferenceTargetObject:
		return r.targetObject(ref.TargetIndex())
	case game.ObjectReferenceSacrificedCost:
		if r.obj == nil || len(r.obj.SacrificedAsCostIDs) == 0 {
			return resolvedObjectReference{}, false
		}
		return resolvePermanentOrLastKnown(r.g, r.obj.SacrificedAsCostIDs[0])
	default:
		return resolvedObjectReference{}, false
	}
}

// targetObject resolves a kind-agnostic target reference to whichever permanent
// or stack object was chosen for the slot, delegating to the kind-specific
// resolution so a combined "spell or permanent" target binds either choice.
func (r referenceResolver) targetObject(index int) (resolvedObjectReference, bool) {
	if r.obj == nil || index < 0 || index >= len(r.obj.Targets) {
		return resolvedObjectReference{}, false
	}
	switch r.obj.Targets[index].Kind {
	case game.TargetPermanent:
		return r.object(game.TargetPermanentReference(index))
	case game.TargetStackObject:
		return r.object(game.TargetStackObjectReference(index))
	default:
		return resolvedObjectReference{}, false
	}
}

// player resolves a PlayerReference to a live player.
func (r referenceResolver) player(ref game.PlayerReference) (game.PlayerID, bool) {
	var playerID game.PlayerID
	var ok bool
	switch ref.Kind() {
	case game.PlayerReferenceController:
		playerID, ok = r.obj.Controller, true
	case game.PlayerReferenceTargetPlayer:
		playerID, ok = r.targetPlayer(ref.TargetIndex())
	case game.PlayerReferenceObjectController:
		playerID, ok = r.referencedObjectController(ref)
	case game.PlayerReferenceObjectOwner:
		playerID, ok = r.referencedObjectOwner(ref)
	case game.PlayerReferenceEventPlayer:
		if r.obj.HasTriggerEvent {
			playerID, ok = triggeringEventPlayer(r.obj.TriggerEvent)
		}
	case game.PlayerReferenceCapturedTargetController:
		playerID, ok = r.obj.CapturedTargetControllerLKI[ref.TargetIndex()]
	case game.PlayerReferenceDefendingPlayer:
		if r.obj.HasTriggerEvent && defendingPlayerEvent(r.obj.TriggerEvent.Kind) {
			playerID, ok = r.obj.TriggerEvent.Player, true
		}
	default:
		return 0, false
	}
	if !ok || !isPlayerAlive(r.g, playerID) {
		return 0, false
	}
	return playerID, true
}

// defendingPlayerEvent reports whether an event kind carries the defending
// player of an attack in Event.Player. The attacker-declared event sets it at
// declare-attackers; the became-blocked and became-unblocked events set it from
// the attacker's declared target at declare-blockers, so a "defending player ..."
// effect on any of these combat triggers resolves to the attacked player.
func defendingPlayerEvent(kind game.EventKind) bool {
	switch kind {
	case game.EventAttackerDeclared,
		game.EventAttackerBecameBlocked,
		game.EventAttackerBecameUnblocked:
		return true
	default:
		return false
	}
}

func triggeringEventPlayer(event game.Event) (game.PlayerID, bool) {
	switch event.Kind {
	case game.EventSpellCast, game.EventSpellCopied, game.EventPermanentTapped:
		return event.Controller, true
	case game.EventCardDrawn,
		game.EventCardDiscarded,
		game.EventCycled,
		game.EventPermanentSacrificed,
		game.EventScry,
		game.EventSurveil,
		game.EventAbilityActivated,
		game.EventBeginningOfStep,
		game.EventLifeGained,
		game.EventLifeLost,
		game.EventLibrarySearched:
		return event.Player, true
	case game.EventDamageDealt:
		return event.Player, event.DamageRecipient == game.DamageRecipientPlayer
	default:
		return 0, false
	}
}

// playerGroup resolves a PlayerGroupReference to alive players in stable player order.
func (r referenceResolver) playerGroup(ref game.PlayerGroupReference) []game.PlayerID {
	switch ref.Kind {
	case game.PlayerGroupReferenceOpponents:
		return aliveOpponents(r.g, r.obj.Controller)
	case game.PlayerGroupReferenceAllPlayers:
		players := make([]game.PlayerID, 0, game.NumPlayers)
		for _, player := range r.g.Players {
			if !player.Eliminated {
				players = append(players, player.ID)
			}
		}
		return players
	default:
		return nil
	}
}

// permanentAt resolves a target slot to a permanent.
func (r referenceResolver) permanentAt(targetIndex int) (*game.Permanent, bool) {
	return effectPermanentTarget(r.g, r.obj, targetIndex)
}

// playerAt resolves a target slot to a player.
func (r referenceResolver) playerAt(targetIndex int) (game.PlayerID, bool) {
	return r.targetPlayer(targetIndex)
}

func (r referenceResolver) sourcePermanent() (*game.Permanent, bool) {
	if r.sourceFixed {
		return r.source, r.source != nil
	}
	return sourcePermanent(r.g, r.obj)
}

func (r referenceResolver) targetPlayer(targetIndex int) (game.PlayerID, bool) {
	if targetIndex < 0 || targetIndex >= len(r.obj.Targets) {
		return 0, false
	}
	target := r.obj.Targets[targetIndex]
	if target.Kind != game.TargetPlayer || !isPlayerAlive(r.g, target.PlayerID) {
		return 0, false
	}
	return target.PlayerID, true
}

func (r referenceResolver) referencedObjectController(ref game.PlayerReference) (game.PlayerID, bool) {
	object, ok := ref.Object()
	if !ok {
		return 0, false
	}
	resolved, ok := r.object(object)
	if !ok {
		return 0, false
	}
	return resolved.controller(r.g)
}

func (r referenceResolver) referencedObjectOwner(ref game.PlayerReference) (game.PlayerID, bool) {
	object, ok := ref.Object()
	if !ok {
		return 0, false
	}
	resolved, ok := r.object(object)
	if !ok {
		return 0, false
	}
	return resolved.owner(r.g)
}

// groupMembers enumerates the object IDs of the permanents that belong to a
// resolved GroupReference, preserving battlefield iteration order. It owns the
// candidate-domain enumeration and the object-reference exclusions that
// Selection deliberately keeps outside itself.
func (r referenceResolver) groupMembers(ref game.GroupReference) []id.ID {
	switch ref.Domain() {
	case game.GroupDomainAttachedObject:
		return r.attachedObjectGroupMembers(ref)
	case game.GroupDomainObjectControlled:
		return r.objectControlledGroupMembers(ref)
	case game.GroupDomainPlayerControlled:
		return r.playerControlledGroupMembers(ref)
	case game.GroupDomainBattlefield:
		return r.battlefieldGroupMembers(ref)
	case game.GroupDomainSameName:
		return r.sameNameGroupMembers(ref)
	default:
		return []id.ID{}
	}
}

func (r referenceResolver) attachedObjectGroupMembers(ref game.GroupReference) []id.ID {
	anchor, ok := ref.Anchor()
	if !ok {
		return []id.ID{}
	}
	resolved, ok := r.object(anchor)
	if !ok || resolved.permanent == nil || !resolved.permanent.AttachedTo.Exists {
		return []id.ID{}
	}
	attachedID := resolved.permanent.AttachedTo.Val
	if _, ok := permanentByObjectID(r.g, attachedID); !ok {
		return []id.ID{}
	}
	return []id.ID{attachedID}
}

func (r referenceResolver) battlefieldGroupMembers(ref game.GroupReference) []id.ID {
	sel := ref.Selection()
	source, _ := r.sourcePermanent()
	if sel.ExcludeSource && source == nil {
		return []id.ID{}
	}
	excludedID := r.exclusionObjectID(ref)
	members := make([]id.ID, 0, len(r.g.Battlefield))
	for _, permanent := range r.g.Battlefield {
		if excludedID != 0 && permanent.ObjectID == excludedID {
			continue
		}
		if !r.permanentMatchesGroupSelection(&sel, source, permanent) {
			continue
		}
		members = append(members, permanent.ObjectID)
	}
	return members
}

// sameNameGroupMembers enumerates every battlefield permanent whose name equals
// the anchor object's name and satisfies the group's Selection, the "<target>
// and all other <group> with the same name as that permanent" wording (Maelstrom
// Pulse, the Echoing cycle). The anchor permanent is included because it shares
// its own name. A missing anchor, an anchor with no available name (a face-down
// permanent or a token without a card def), or a permanent with no available
// name fails closed, so an unnamed object never matches.
func (r referenceResolver) sameNameGroupMembers(ref game.GroupReference) []id.ID {
	anchor, ok := ref.Anchor()
	if !ok {
		return []id.ID{}
	}
	resolved, ok := r.object(anchor)
	if !ok || resolved.permanent == nil {
		return []id.ID{}
	}
	anchorName, ok := permanentNameForSameNameGroup(r.g, resolved.permanent)
	if !ok {
		return []id.ID{}
	}
	sel := ref.Selection()
	source, _ := r.sourcePermanent()
	if sel.ExcludeSource && source == nil {
		return []id.ID{}
	}
	members := make([]id.ID, 0, len(r.g.Battlefield))
	for _, permanent := range r.g.Battlefield {
		name, ok := permanentNameForSameNameGroup(r.g, permanent)
		if !ok || name != anchorName {
			continue
		}
		if !r.permanentMatchesGroupSelection(&sel, source, permanent) {
			continue
		}
		members = append(members, permanent.ObjectID)
	}
	return members
}

// permanentNameForSameNameGroup resolves a permanent's effective name for
// same-name group matching, reporting false when no name is available (a
// face-down permanent or a permanent without a card def) so an unnamed object
// neither anchors nor joins the group.
func permanentNameForSameNameGroup(g *game.Game, permanent *game.Permanent) (string, bool) {
	if permanent == nil || permanent.FaceDown {
		return "", false
	}
	if _, ok := permanentCardDef(g, permanent); !ok {
		return "", false
	}
	return permanentEffectiveName(g, permanent), true
}

func (r referenceResolver) objectControlledGroupMembers(ref game.GroupReference) []id.ID {
	anchor, ok := ref.Anchor()
	if !ok {
		return []id.ID{}
	}
	resolved, ok := r.object(anchor)
	if !ok {
		return []id.ID{}
	}
	controller, ok := resolved.controller(r.g)
	if !ok {
		return []id.ID{}
	}
	sel := ref.Selection()
	source, _ := r.sourcePermanent()
	if sel.ExcludeSource && source == nil {
		return []id.ID{}
	}
	excludedID := r.exclusionObjectID(ref)
	members := make([]id.ID, 0, len(r.g.Battlefield))
	for _, permanent := range r.g.Battlefield {
		if excludedID != 0 && permanent.ObjectID == excludedID {
			continue
		}
		if effectiveController(r.g, permanent) != controller {
			continue
		}
		if !r.permanentMatchesGroupSelection(&sel, source, permanent) {
			continue
		}
		members = append(members, permanent.ObjectID)
	}
	return members
}

// playerControlledGroupMembers enumerates the battlefield permanents controlled
// by the player named by the group's anchor player reference and satisfying its
// Selection, such as every creature a targeted player controls. It mirrors
// objectControlledGroupMembers but resolves the controlling player directly from
// the anchor player reference rather than from an anchor object's controller.
func (r referenceResolver) playerControlledGroupMembers(ref game.GroupReference) []id.ID {
	anchor, ok := ref.PlayerAnchor()
	if !ok {
		return []id.ID{}
	}
	controller, ok := r.player(anchor)
	if !ok {
		return []id.ID{}
	}
	sel := ref.Selection()
	source, _ := r.sourcePermanent()
	if sel.ExcludeSource && source == nil {
		return []id.ID{}
	}
	excludedID := r.exclusionObjectID(ref)
	members := make([]id.ID, 0, len(r.g.Battlefield))
	for _, permanent := range r.g.Battlefield {
		if excludedID != 0 && permanent.ObjectID == excludedID {
			continue
		}
		if effectiveController(r.g, permanent) != controller {
			continue
		}
		if !r.permanentMatchesGroupSelection(&sel, source, permanent) {
			continue
		}
		members = append(members, permanent.ObjectID)
	}
	return members
}

// permanentMatchesGroupSelection reports whether permanent satisfies the group's
// Selection. Callers hoist the source-exclusion divergence (an ExcludeSource
// group with no source permanent matches nothing) out of the enumeration loop.
func (r referenceResolver) permanentMatchesGroupSelection(sel *game.Selection, source, permanent *game.Permanent) bool {
	values := effectivePermanentValues(r.g, permanent)
	subject := selectionSubject{
		kind:              subjectPermanent,
		g:                 r.g,
		permanent:         permanent,
		values:            &values,
		viewer:            r.obj.Controller,
		resolutionChoices: r.obj.ResolutionChoices,
	}
	if sel.Controller != game.ControllerAny {
		subject.controller = effectiveController(r.g, permanent)
	}
	if source != nil {
		subject.sourceObjectID = source.ObjectID
	}
	return matchSelection(&subject, sel)
}

// exclusionObjectID resolves the group's excluded object to its identity object
// ID, returning 0 when the group has no exclusion or the excluded object cannot
// be identified. It reads the reference's identity even when the permanent has
// left the battlefield, matching the legacy "except target" exclusion.
func (r referenceResolver) exclusionObjectID(ref game.GroupReference) id.ID {
	exclude, ok := ref.Exclusion()
	if !ok {
		return 0
	}
	objectID, _ := r.objectIdentityID(exclude)
	return objectID
}

func (r referenceResolver) objectIdentityID(ref game.ObjectReference) (id.ID, bool) {
	switch ref.Kind() {
	case game.ObjectReferenceTargetPermanent:
		return targetPermanentObjectID(r.obj, ref.TargetIndex())
	case game.ObjectReferenceSourcePermanent:
		if r.obj == nil || r.obj.SourceID == 0 {
			return 0, false
		}
		return r.obj.SourceID, true
	case game.ObjectReferenceEventPermanent:
		if r.obj != nil && r.obj.HasTriggerEvent && r.obj.TriggerEvent.PermanentID != 0 {
			return r.obj.TriggerEvent.PermanentID, true
		}
		return 0, false
	case game.ObjectReferenceEventRelatedPermanent:
		if r.obj != nil && r.obj.HasTriggerEvent && r.obj.TriggerEvent.RelatedPermanentID != 0 {
			return r.obj.TriggerEvent.RelatedPermanentID, true
		}
		return 0, false
	case game.ObjectReferenceSourceAttachedPermanent:
		permanent, ok := r.sourcePermanent()
		if !ok || !permanent.AttachedTo.Exists {
			return 0, false
		}
		return permanent.AttachedTo.Val, true
	case game.ObjectReferenceTargetAttachedPermanent:
		objectID, ok := targetPermanentObjectID(r.obj, ref.TargetIndex())
		if !ok {
			return 0, false
		}
		permanent, ok := permanentByObjectID(r.g, objectID)
		if !ok || !permanent.AttachedTo.Exists {
			return 0, false
		}
		return permanent.AttachedTo.Val, true
	case game.ObjectReferenceLinkedObject:
		for _, linked := range linkedObjects(r.g, linkedObjectSourceKey(r.g, r.obj, ref.LinkID())) {
			if linked.ObjectID != 0 {
				return linked.ObjectID, true
			}
		}
		return 0, false
	default:
		return 0, false
	}
}

// resolveObjectReference is a thin adapter over the referenceResolver module.
func resolveObjectReference(g *game.Game, obj *game.StackObject, ref game.ObjectReference) (resolvedObjectReference, bool) {
	return newReferenceResolver(g, obj).object(ref)
}

// resolvePlayerReference is a thin adapter over the referenceResolver module.
func resolvePlayerReference(g *game.Game, obj *game.StackObject, ref game.PlayerReference) (game.PlayerID, bool) {
	return newReferenceResolver(g, obj).player(ref)
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

func resolveSourcePermanentOrLastKnown(g *game.Game, objectID id.ID) (resolvedObjectReference, bool) {
	if permanent, ok := permanentByObjectID(g, objectID); ok {
		if !permanent.PhasedOut {
			return resolvedObjectReference{permanent: permanent}, true
		}
		if snapshot, ok := lastKnownObject(g, objectID); ok {
			return resolvedObjectReference{snapshot: snapshot}, true
		}
		snapshot := snapshotPermanent(g, permanent, zone.Battlefield)
		return resolvedObjectReference{snapshot: snapshot}, true
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
