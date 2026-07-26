package game

import "github.com/natefinch/council4/mtg/game/zone"

// MovePermanent moves referenced permanents off the battlefield to Destination.
// It is the single "move a permanent to a zone" primitive: the destination is a
// parameter rather than a distinct primitive type, so a new destination, or a
// new combination of destination and selection, costs no new runtime code.
//
// It replaces three primitives that differed only in where the permanent landed:
// Bounce (hand), Exile (exile), and PutPermanentOnLibrary (library). Each had
// picked up a rider the others could not express — Bounce alone could let the
// controller choose which permanents move, Exile alone could publish a linked
// key, PutPermanentOnLibrary alone could reach the bottom of the library — so
// "exile a creature you control" and "put target creature on the bottom of its
// owner's library" each needed a fourth primitive. Here they are combinations of
// existing fields.
//
// Destroy and Sacrifice deliberately remain separate. They are distinct game
// actions rather than distinct destinations: a destroy consults indestructible,
// destroy replacements, and commander redirection, and a sacrifice cannot be
// replaced or prevented.
//
// Exactly one referent applies: Object for a single permanent, Group for every
// permanent in the group at once, or Group as a candidate pool when
// ControlledChoice is set. A single-Object move falls back to the resolving
// spell on the stack when the reference resolves to a stack object rather than a
// permanent, which backs "return target spell to its owner's hand".
type MovePermanent struct {
	// Object references a single permanent to move.
	Object ObjectReference
	// Group references permanents to move. Without ControlledChoice every
	// permanent in the group moves simultaneously; with it, the group is the
	// candidate pool the chooser picks Amount permanents from.
	Group GroupReference
	// Destination is the zone the permanents move to. zone.None is rejected by
	// validation, so a caller cannot silently move a permanent nowhere.
	Destination zone.Type
	// LibraryBottom places each moved card on the bottom of its owner's library
	// rather than the top. It applies only when Destination is zone.Library and
	// is ignored for tokens, which cease to exist.
	LibraryBottom bool
	// ControlledChoice has the resolving controller choose Amount permanents
	// from among those matched by Group ("Return a creature you control to its
	// owner's hand."). Object must be unset and Group set when it is true. When
	// the pool holds no more permanents than Amount, every candidate moves with
	// no prompt.
	ControlledChoice bool
	// Amount is the number of permanents ControlledChoice picks. It is
	// meaningful only with ControlledChoice.
	Amount Quantity
	// PublishLinked remembers each permanent that actually moves under a
	// source-scoped linked key, so a paired return instruction can bring back
	// exactly that set (blink and exile-until-leaves). The link is captured
	// before the move so it survives the permanent leaving the battlefield.
	PublishLinked LinkedKey
}

// MoveResolvingSpell moves the spell currently resolving to Destination instead
// of putting it into its owner's graveyard, backing resolution tails such as
// "Exile this card." (Eldritch Evolution) and "Shuffle this card into its
// owner's library." (Green Sun's Zenith).
//
// It has no referent of its own: it always acts on the object being resolved. It
// replaces Exile's SourceSpell flag and the separate ShuffleSpellIntoLibrary
// primitive, which were the same instruction with two destinations, lowered by
// two near-identical functions from two parser recognizers.
type MoveResolvingSpell struct {
	// Destination is the zone the resolving spell moves to. zone.None is
	// rejected by validation.
	Destination zone.Type
	// Shuffle shuffles the owner's library after the spell arrives. Every
	// printed library wording shuffles, so validation requires it with a library
	// destination; the field exists so an unshuffled "put this card on top of
	// its owner's library" tail needs only a validation change.
	Shuffle bool
}
