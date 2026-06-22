package game

import (
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ChooseFromZone is the shared, valence-agnostic envelope describing a
// "the resolving player chooses cards from a zone matching a filter, then those
// chosen cards move to a destination" effect. It is the single canonical shape
// that the historically separate zone-choice primitives (ReturnFromGraveyard,
// MassReturnFromGraveyard, ExileFromHand, PutFromHand, CastForFree, the typed
// library Search, and similar) describe one ad-hoc field at a time. The envelope
// keeps the four genuinely independent concerns apart: which cards are eligible
// (SourceZone + Filter), how many and how correlated the chosen set is (Quantity
// + Count + Grouping), where the cards go (Destination), and the optional
// movement modifiers applied as they arrive (Riders).
//
// ChooseFromZone is pure rules data; mtg/rules owns resolution. The zero value
// chooses nothing: an empty SourceZone yields no candidates.
type ChooseFromZone struct {
	// Player names the player who makes the choice. A card that enters the
	// battlefield does so under this player's control unless Riders.UnderOwnerControl
	// routes it to its own owner instead.
	Player PlayerReference

	// SourceZone is the zone whose cards are the candidate pool for the choice.
	SourceZone zone.Type

	// Filter is the canonical predicate every chosen card must satisfy. The zero
	// Selection matches every card in SourceZone.
	Filter Selection

	// Quantity is the numeric bound on the number of cards chosen, interpreted by
	// Count. It is evaluated when the effect resolves, so it may be a dynamic
	// "that many" formula.
	Quantity Quantity

	// Count selects how Quantity bounds the choice: exactly that many, up to that
	// many, or any number from the whole matching pool (Quantity is ignored for
	// ChooseAnyNumber).
	Count ChooseCount

	// Grouping describes any structural correlation the chosen cards must obey
	// beyond the per-card Filter, such as sharing a subtype or splitting across
	// two destinations. ChooseAcrossSet (the zero value) imposes no correlation.
	Grouping ChooseGrouping

	// Destination is where the chosen cards move. For the ChooseSplitDestination
	// grouping it is the primary slot, paired with Riders.EntersTapped.
	Destination ChooseDestination

	// SplitSecondary is the second destination slot, meaningful only with the
	// ChooseSplitDestination grouping (CR 701.19; Cultivate, Kodama's Reach). The
	// primary slot is Destination with Riders.EntersTapped; this slot carries its
	// own tapped flag because the two slots may differ.
	SplitSecondary opt.V[ChooseSplitSlot]

	// Riders carry the optional movement modifiers applied as the chosen cards
	// move to their destination.
	Riders ChooseRiders
}

// ChooseCount selects how a ChooseFromZone's Quantity bounds the number of cards
// the player chooses.
type ChooseCount uint8

// Supported choose-count interpretations.
const (
	// ChooseExactly makes the player choose exactly Quantity cards, or every
	// matching card when fewer than Quantity exist ("choose two cards"). It is
	// the strict form; a Riders.FailToFindPolicy or Riders.MaxTotalManaValue may
	// still relax the minimum.
	ChooseExactly ChooseCount = iota

	// ChooseUpTo makes the player choose up to Quantity cards, including none
	// ("choose up to two cards"). The empty choice is always legal.
	ChooseUpTo

	// ChooseAnyNumber makes the player choose any number of cards from the whole
	// matching pool, from none up to all of them ("put any number of those cards
	// ..."). Quantity is ignored.
	ChooseAnyNumber
)

// ChooseGrouping describes how a ChooseFromZone's chosen cards relate to one
// another and to the destination beyond the per-card Filter.
type ChooseGrouping uint8

// Supported choose groupings.
const (
	// ChooseAcrossSet imposes no correlation: the player chooses the cards as one
	// set across the whole matching pool, and every chosen card moves to
	// Destination. It is the ordinary multi-card choice.
	ChooseAcrossSet ChooseGrouping = iota

	// ChooseOneOfEachNamedType makes the player choose at most one card matching
	// each card type listed in Filter.RequiredTypesAny ("a creature card and a
	// land card"). Each listed type contributes at most one card, and no card is
	// chosen twice. It requires a non-empty Filter.RequiredTypesAny. Every chosen
	// card moves to Destination.
	ChooseOneOfEachNamedType

	// ChooseSplitDestination distributes the chosen cards across the two
	// single-card slots Destination (primary) and SplitSecondary (CR 701.19;
	// Cultivate). At most two cards are chosen. With two chosen, the player
	// assigns one card to each slot; with one chosen, the player chooses which
	// slot it fills.
	ChooseSplitDestination

	// ChooseSharedSubtype requires every chosen card to share at least one subtype
	// with each other chosen card ("two basic land cards that share a land type",
	// Myriad Landscape). The choice is staged so an illegal combination can never
	// be assembled (CR 701.19); choosing zero or one card satisfies it vacuously.
	// Every chosen card moves to Destination.
	ChooseSharedSubtype
)

// ChooseDestination names where chosen cards move: a zone and, for an ordered
// destination zone, the position within it.
type ChooseDestination struct {
	// Zone is the destination zone.
	Zone zone.Type

	// Position selects the position within an ordered destination zone. It is
	// meaningful only for the library; other zones ignore it. The bottom of the
	// library is selected by Riders.DestinationBottom rather than here, matching
	// the codebase convention that top placement is an explicit position while
	// bottom placement is a movement rider.
	Position ChoosePosition
}

// ChoosePosition identifies a position within an ordered destination zone.
type ChoosePosition uint8

// Supported ordered destination positions.
const (
	// ChoosePositionDefault places the card with the zone's default placement
	// (the top of the library).
	ChoosePositionDefault ChoosePosition = iota

	// ChoosePositionTop places the card explicitly on top of the library, the
	// "put it on top of your library" destination of a tutor that does not
	// reshuffle the chosen card away.
	ChoosePositionTop
)

// ChooseSplitSlot is the secondary destination slot of a ChooseSplitDestination
// grouping: the zone a chosen card enters and whether it enters the battlefield
// tapped.
type ChooseSplitSlot struct {
	Destination  ChooseDestination
	EntersTapped bool
}

// ChooseRiders carry the optional modifiers a ChooseFromZone applies as the
// chosen cards move to their destination. The zero value applies no modifier.
type ChooseRiders struct {
	// EntersTapped makes each card entering the battlefield enter tapped. For the
	// ChooseSplitDestination grouping it applies to the primary slot only.
	EntersTapped bool

	// EntryCounters places counters on each card entering the battlefield, the
	// "enters with N counters" rider. It is meaningful only for a battlefield
	// Destination.
	EntryCounters []CounterPlacement

	// FaceDown makes each card entering the battlefield enter face down with
	// FaceDownKind, the "put it onto the battlefield face down" rider (manifest).
	// It is meaningful only for a battlefield Destination and composes with
	// neither EntersTapped nor EntryCounters.
	FaceDown bool

	// FaceDownKind names the face-down kind applied when FaceDown is set. It must
	// be a real kind (not FaceDownNone) when FaceDown is set.
	FaceDownKind FaceDownKind

	// UnderOwnerControl makes each card entering the battlefield enter under its
	// own owner's control rather than the choosing player's ("under their owners'
	// control"). It is meaningful only for a battlefield Destination.
	UnderOwnerControl bool

	// DestinationBottom routes each card to the bottom of an ordered destination
	// zone (the library) instead of its default top placement.
	DestinationBottom bool

	// Reveal reveals each chosen card as it moves, the "reveal them" rider.
	Reveal bool

	// FromLinked, when set, restricts the candidate pool to the cards remembered
	// under this key by a prior instruction (such as a Mill that published the
	// cards it milled), modeling "from among those cards". When empty, the whole
	// SourceZone is scanned.
	FromLinked LinkedKey

	// PublishLinked, when set, remembers every chosen card under this key so a
	// later instruction can act on exactly those cards. The key is cleared before
	// the chosen cards are remembered.
	PublishLinked LinkedKey

	// MaxTotalManaValue, when set, caps the combined mana value of the chosen
	// cards ("with total mana value 4 or less", Lively Dirge). It also makes the
	// choice optional, since the empty choice always satisfies the cap.
	MaxTotalManaValue opt.V[int]

	// MaxManaValueFromX, when set, restricts the candidate pool to cards whose
	// mana value is at most the resolving spell's chosen X ("with mana value X or
	// less", Green Sun's Zenith, Chord of Calling). The bound is resolved from the
	// resolving stack object as the effect runs, so it lives on the envelope
	// rather than the static Filter.
	MaxManaValueFromX bool

	// FailToFindPolicy controls whether a ChooseExactly choice may legally choose
	// fewer than its bound when matching cards exist. The zero value derives the
	// rule from the choice shape: a single unrestricted ChooseExactly must find a
	// card when the pool is nonempty, while qualified or "up to" choices may fail.
	FailToFindPolicy SearchFailToFindPolicy
}
