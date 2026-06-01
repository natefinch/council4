package game

import "github.com/natefinch/council4/opt"

// ObjectReferenceKind identifies a runtime object binding used by declarative
// effects.
type ObjectReferenceKind int

const (
	ObjectReferenceNone ObjectReferenceKind = iota
	ObjectReferenceTargetPermanent
	ObjectReferenceSourcePermanent
	ObjectReferenceAttachedPermanent
	ObjectReferenceLinkedObject
	ObjectReferenceEventPermanent
)

// ObjectReference describes how a rules effect finds an object at resolution.
type ObjectReference struct {
	Kind ObjectReferenceKind

	// TargetIndex indexes the stack object's selected targets for target-derived
	// references.
	TargetIndex int

	// LinkID identifies a linked object recorded by an earlier effect.
	LinkID string
}

// PlayerReferenceKind identifies a runtime player binding used by declarative
// effects.
type PlayerReferenceKind int

const (
	PlayerReferenceNone PlayerReferenceKind = iota
	PlayerReferenceController
	PlayerReferenceTargetPlayer
	PlayerReferenceObjectController
	PlayerReferenceObjectOwner
)

// PlayerReference describes how a rules effect finds a player at resolution.
type PlayerReference struct {
	Kind PlayerReferenceKind

	// TargetIndex indexes the stack object's selected targets for target-player
	// references.
	TargetIndex int

	// Object binds controller/owner lookups to a reusable object reference.
	Object opt.V[ObjectReference]
}

// CardReferenceKind identifies a runtime card binding used by declarative
// effects.
type CardReferenceKind int

const (
	CardReferenceNone CardReferenceKind = iota
	CardReferenceLinked
	CardReferenceSource
	CardReferenceEvent
)

// CardReference describes how a rules effect finds a card at resolution.
type CardReference struct {
	Kind CardReferenceKind

	// LinkID identifies a linked card recorded by an earlier effect.
	LinkID string
}

// CardCondition describes characteristics a referenced card must have for an
// effect to apply.
type CardCondition struct {
	Card CardReference

	RequirePermanentCard bool
	Types                []CardType
	Supertypes           []Supertype
	SubtypesAny          []string
}
