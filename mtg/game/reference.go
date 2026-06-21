package game

import (
	"fmt"

	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ObjectReferenceKind identifies a runtime object binding used by declarative
// effects.
type ObjectReferenceKind int

// Object reference kind values identify supported object bindings.
const (
	ObjectReferenceNone ObjectReferenceKind = iota
	ObjectReferenceTargetPermanent
	ObjectReferenceTargetStackObject
	ObjectReferenceSourcePermanent
	ObjectReferenceSourceAttachedPermanent
	ObjectReferenceTargetAttachedPermanent
	ObjectReferenceLinkedObject
	ObjectReferenceEventPermanent
	ObjectReferenceSourceCard
	// ObjectReferenceCapturedTargetStackObject identifies a stack-object target
	// captured by an enclosing effect for use inside a delayed trigger.
	ObjectReferenceCapturedTargetStackObject
	// ObjectReferenceTargetObject references the object chosen for a target slot
	// without committing to its kind, resolving to whichever permanent or stack
	// object was selected. It backs combined targets that accept either a spell
	// on the stack or a permanent ("target spell or nonland permanent").
	ObjectReferenceTargetObject
	// ObjectReferenceSacrificedCost references the permanent sacrificed to pay
	// the resolving activated ability's cost, read from last-known information
	// once it has left the battlefield. It backs effects scaled by the
	// sacrificed permanent ("the sacrificed creature's power") on a
	// sacrifice-cost ability (Altar of Dementia).
	ObjectReferenceSacrificedCost
)

// ObjectReference describes how a rules effect finds an object at resolution.
type ObjectReference struct {
	kind ObjectReferenceKind

	// targetIndex indexes the stack object's selected targets for target-derived
	// references.
	targetIndex int

	// linkID identifies a linked object recorded by an earlier effect.
	linkID string
}

// Kind reports the reference kind.
func (r ObjectReference) Kind() ObjectReferenceKind { return r.kind }

// TargetIndex reports the target slot index for target-derived references.
func (r ObjectReference) TargetIndex() int { return r.targetIndex }

// LinkID reports the linked-object identifier for linked references.
func (r ObjectReference) LinkID() string { return r.linkID }

// TargetPermanentReference references the permanent chosen for the target slot at
// targetIndex.
func TargetPermanentReference(targetIndex int) ObjectReference {
	return ObjectReference{kind: ObjectReferenceTargetPermanent, targetIndex: targetIndex}
}

// TargetStackObjectReference references the stack object chosen for the target
// slot at targetIndex.
func TargetStackObjectReference(targetIndex int) ObjectReference {
	return ObjectReference{kind: ObjectReferenceTargetStackObject, targetIndex: targetIndex}
}

// CapturedTargetStackObjectReference references the enclosing effect's
// stack-object target at targetIndex from inside a delayed trigger.
func CapturedTargetStackObjectReference(targetIndex int) ObjectReference {
	return ObjectReference{kind: ObjectReferenceCapturedTargetStackObject, targetIndex: targetIndex}
}

// TargetObjectReference references the object chosen for the target slot at
// targetIndex regardless of whether it is a permanent or a stack object.
func TargetObjectReference(targetIndex int) ObjectReference {
	return ObjectReference{kind: ObjectReferenceTargetObject, targetIndex: targetIndex}
}

// SourcePermanentReference references the source permanent of the resolving stack
// object.
func SourcePermanentReference() ObjectReference {
	return ObjectReference{kind: ObjectReferenceSourcePermanent}
}

// SourceCardPermanentReference references the battlefield permanent represented
// by the resolving stack object's source card.
func SourceCardPermanentReference() ObjectReference {
	return ObjectReference{kind: ObjectReferenceSourceCard}
}

// SacrificedCostReference references the permanent sacrificed to pay the
// resolving activated ability's cost.
func SacrificedCostReference() ObjectReference {
	return ObjectReference{kind: ObjectReferenceSacrificedCost}
}

// SourceAttachedPermanentReference references the permanent that the source permanent
// is attached to, such as the creature an Aura or Equipment is attached to.
func SourceAttachedPermanentReference() ObjectReference {
	return ObjectReference{kind: ObjectReferenceSourceAttachedPermanent}
}

// TargetAttachedPermanentReference references the permanent that the targeted
// permanent at targetIndex is attached to.
func TargetAttachedPermanentReference(targetIndex int) ObjectReference {
	return ObjectReference{kind: ObjectReferenceTargetAttachedPermanent, targetIndex: targetIndex}
}

// LinkedObjectReference references a linked object recorded by an earlier effect
// under linkID.
func LinkedObjectReference(linkID string) ObjectReference {
	return ObjectReference{kind: ObjectReferenceLinkedObject, linkID: linkID}
}

// EventPermanentReference references the permanent named by the triggering event of
// the resolving stack object.
func EventPermanentReference() ObjectReference {
	return ObjectReference{kind: ObjectReferenceEventPermanent}
}

// Validate reports structural problems with an ObjectReference that represent
// card-definition bugs. It checks kind/field consistency only; target-index
// bounds depend on the surrounding TargetSpec list and are checked by
// ValidateCardDef.
func (r ObjectReference) Validate() []string {
	switch r.kind {
	case ObjectReferenceTargetPermanent:
		if r.linkID != "" {
			return []string{"target permanent reference must not set LinkID"}
		}
		if r.targetIndex < 0 {
			return []string{"target permanent reference must not use a negative TargetIndex"}
		}
	case ObjectReferenceTargetStackObject:
		if r.linkID != "" {
			return []string{"target stack object reference must not set LinkID"}
		}
		if r.targetIndex < 0 {
			return []string{"target stack object reference must not use a negative TargetIndex"}
		}
	case ObjectReferenceSourcePermanent:
		if r.targetIndex != 0 || r.linkID != "" {
			return []string{"source permanent reference must not set TargetIndex or LinkID"}
		}
	case ObjectReferenceSourceAttachedPermanent:
		if r.targetIndex != 0 || r.linkID != "" {
			return []string{"source-attached permanent reference must not set TargetIndex or LinkID"}
		}
	case ObjectReferenceTargetAttachedPermanent:
		if r.linkID != "" {
			return []string{"target-attached permanent reference must not set LinkID"}
		}
		if r.targetIndex < 0 {
			return []string{"target-attached permanent reference must not use a negative TargetIndex"}
		}
	case ObjectReferenceLinkedObject:
		if r.linkID == "" {
			return []string{"linked object reference requires LinkID"}
		}
		if r.targetIndex != 0 {
			return []string{"linked object reference must not set TargetIndex"}
		}
	case ObjectReferenceEventPermanent:
		if r.targetIndex != 0 || r.linkID != "" {
			return []string{"event permanent reference must not set TargetIndex or LinkID"}
		}
	case ObjectReferenceSourceCard:
		if r.targetIndex != 0 || r.linkID != "" {
			return []string{"source card permanent reference must not set TargetIndex or LinkID"}
		}
	case ObjectReferenceCapturedTargetStackObject:
		if r.linkID != "" {
			return []string{"captured target stack object reference must not set LinkID"}
		}
		if r.targetIndex < 0 {
			return []string{"captured target stack object reference must not use a negative TargetIndex"}
		}
	case ObjectReferenceTargetObject:
		if r.linkID != "" {
			return []string{"target object reference must not set LinkID"}
		}
		if r.targetIndex < 0 {
			return []string{"target object reference must not use a negative TargetIndex"}
		}
	case ObjectReferenceSacrificedCost:
		if r.targetIndex != 0 || r.linkID != "" {
			return []string{"sacrificed cost reference must not set TargetIndex or LinkID"}
		}
	case ObjectReferenceNone:
		return []string{"object reference has no kind"}
	default:
		return []string{fmt.Sprintf("unknown object reference kind %d", r.kind)}
	}
	return nil
}

// PlayerReferenceKind identifies a runtime player binding used by declarative
// effects.
type PlayerReferenceKind int

// Player reference kind values identify supported player bindings.
const (
	PlayerReferenceNone PlayerReferenceKind = iota
	PlayerReferenceController
	PlayerReferenceTargetPlayer
	PlayerReferenceObjectController
	PlayerReferenceObjectOwner
	// PlayerReferenceEventPlayer references the player identified by the
	// triggering event, such as the player who drew, discarded, or cast a card.
	// It is only valid for event kinds with a well-defined player subject.
	PlayerReferenceEventPlayer
	// PlayerReferenceCapturedTargetController reads a target stack object's
	// controller captured by the effect that created a delayed trigger.
	PlayerReferenceCapturedTargetController
	// PlayerReferenceDefendingPlayer references the defending player of the
	// triggering attack event ("defending player sacrifices N permanents" in the
	// Annihilator keyword). It is valid only inside triggered abilities whose
	// event is an attacker declaration.
	PlayerReferenceDefendingPlayer
)

// PlayerReference describes how a rules effect finds a player at resolution.
type PlayerReference struct {
	kind PlayerReferenceKind

	// targetIndex indexes the stack object's selected targets for target-player
	// references.
	targetIndex int

	// object binds controller/owner lookups to a reusable object reference.
	object opt.V[ObjectReference]
}

// Kind reports the reference kind.
func (r PlayerReference) Kind() PlayerReferenceKind { return r.kind }

// TargetIndex reports the target slot index for target-player references.
func (r PlayerReference) TargetIndex() int { return r.targetIndex }

// Object reports the nested object reference for controller/owner lookups.
func (r PlayerReference) Object() (ObjectReference, bool) {
	return r.object.Val, r.object.Exists
}

// ControllerReference references the controller of the resolving stack object.
func ControllerReference() PlayerReference {
	return PlayerReference{kind: PlayerReferenceController}
}

// TargetPlayerReference references the player chosen for the target slot at
// targetIndex.
func TargetPlayerReference(targetIndex int) PlayerReference {
	return PlayerReference{kind: PlayerReferenceTargetPlayer, targetIndex: targetIndex}
}

// ObjectControllerReference references the controller of the object identified by
// object.
func ObjectControllerReference(object ObjectReference) PlayerReference {
	return PlayerReference{kind: PlayerReferenceObjectController, object: opt.Val(object)}
}

// ObjectOwnerReference references the owner of the object identified by object.
func ObjectOwnerReference(object ObjectReference) PlayerReference {
	return PlayerReference{kind: PlayerReferenceObjectOwner, object: opt.Val(object)}
}

// EventPlayerReference references the player identified by the triggering event.
// It is valid only inside triggered abilities whose event kind has a
// well-defined player subject.
func EventPlayerReference() PlayerReference {
	return PlayerReference{kind: PlayerReferenceEventPlayer}
}

// CapturedTargetControllerReference references the controller captured for the
// targeted stack object at targetIndex when a delayed trigger was created.
func CapturedTargetControllerReference(targetIndex int) PlayerReference {
	return PlayerReference{kind: PlayerReferenceCapturedTargetController, targetIndex: targetIndex}
}

// DefendingPlayerReference references the defending player of the triggering
// attack event. It is valid only inside triggered abilities whose event is an
// attacker declaration (the Annihilator keyword's combat trigger).
func DefendingPlayerReference() PlayerReference {
	return PlayerReference{kind: PlayerReferenceDefendingPlayer}
}

// Validate reports structural problems with a PlayerReference that represent
// card-definition bugs. It checks player-level kind/field consistency and the
// structure of any nested object reference; target-index bounds depend on the
// surrounding TargetSpec list and are checked by ValidateCardDef.
func (r PlayerReference) Validate() []string {
	switch r.kind {
	case PlayerReferenceController:
		if r.targetIndex != 0 || r.object.Exists {
			return []string{"controller reference must not set TargetIndex or Object"}
		}
	case PlayerReferenceTargetPlayer:
		if r.object.Exists {
			return []string{"target player reference must not set Object"}
		}
		if r.targetIndex < 0 {
			return []string{"target player reference must not use a negative TargetIndex"}
		}
	case PlayerReferenceObjectController, PlayerReferenceObjectOwner:
		if !r.object.Exists {
			return []string{"object controller/owner reference requires Object"}
		}
		if r.targetIndex != 0 {
			return []string{"object controller/owner reference must not set TargetIndex"}
		}
		if problems := appendPrefixed(nil, "object", r.object.Val.Validate()); len(problems) > 0 {
			return problems
		}
	case PlayerReferenceNone:
		return []string{"player reference has no kind"}
	case PlayerReferenceEventPlayer:
		if r.targetIndex != 0 || r.object.Exists {
			return []string{"event player reference must not set TargetIndex or Object"}
		}
	case PlayerReferenceCapturedTargetController:
		if r.object.Exists {
			return []string{"captured target controller reference must not set Object"}
		}
		if r.targetIndex < 0 {
			return []string{"captured target controller reference must not use a negative TargetIndex"}
		}
	case PlayerReferenceDefendingPlayer:
		if r.targetIndex != 0 || r.object.Exists {
			return []string{"defending player reference must not set TargetIndex or Object"}
		}
	default:
		return []string{fmt.Sprintf("unknown player reference kind %d", r.kind)}
	}
	return nil
}

// CardReferenceKind identifies a runtime card binding used by declarative
// effects.
type CardReferenceKind int

// Card reference kind values identify supported card bindings.
const (
	CardReferenceNone CardReferenceKind = iota
	CardReferenceLinked
	CardReferenceSource
	CardReferenceEvent
	CardReferenceTarget
)

// CardReference describes how a rules effect finds a card at resolution.
type CardReference struct {
	Kind CardReferenceKind

	// TargetIndex identifies which card target slot to read when Kind is
	// CardReferenceTarget. The zero value preserves the first target.
	TargetIndex int

	// LinkID identifies a linked card recorded by an earlier effect.
	LinkID string
}

// CardCondition describes characteristics a referenced card must have for an
// effect to apply.
type CardCondition struct {
	Card CardReference

	RequirePermanentCard bool
	Types                []types.Card
	Supertypes           []types.Super
	SubtypesAny          []types.Sub
	ChosenSubtypeFrom    ChoiceKey
}
