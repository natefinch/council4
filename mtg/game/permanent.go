package game

import (
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/opt"
)

// FaceDownKind identifies how a face-down spell or permanent became face-down.
// The kind determines rules text not visible in the printed characteristics,
// such as Disguise's Ward {2} and shield counter when turned face up.
type FaceDownKind int

// Face-down kind values identify how a card became face-down.
const (
	FaceDownNone FaceDownKind = iota
	FaceDownMorph
	FaceDownDisguise
	FaceDownManifest
	// FaceDownCloak is a card cloaked onto the battlefield (CR 701.56): a
	// face-down 2/2 creature with ward {2} that may be turned face up for its
	// mana cost if it is a creature card. It shares Manifest's turn-face-up rule
	// (mana cost when the hidden card is a creature) and Disguise's ward {2}
	// characteristic.
	FaceDownCloak
)

// Permanent represents a card (or token) on the battlefield with all its
// in-game state. A Permanent is created when a spell resolves as a permanent
// or when a token is created, and is destroyed when it leaves the battlefield.
//
// When a permanent changes zones, it becomes a new game object — the old
// Permanent is removed and a new one would be created if it re-enters.
type Permanent struct {
	// ObjectID is this permanent's unique game object identity.
	ObjectID id.ID

	// CardInstanceID references the CardInstance this permanent is based on.
	// Zero for tokens (use TokenDef instead).
	CardInstanceID id.ID

	// MergedCards lists the card components below CardInstanceID from top to
	// bottom for a permanent created by Mutate.
	MergedCards []MergedCard

	// Owner is the player who owns the underlying card. For tokens, this is
	// the player who created the token (CR 111.2).
	Owner PlayerID

	// Controller is the player who currently controls this permanent.
	// Defaults to Owner but can change via control-changing effects.
	Controller PlayerID

	// --- Status flags ---

	// Tapped is true if the permanent is tapped (turned sideways).
	Tapped bool

	// Exerted is true if this permanent should not untap during its
	// controller's next untap step.
	Exerted bool

	// SummoningSick is true if this permanent has not been under its
	// controller's control since the start of that player's most recent turn.
	// It only restricts creatures from attacking and activating abilities with
	// {T} or {Q} in the cost (CR 302.6).
	SummoningSick bool

	// PhasedOut is true if this permanent is phased out. Phased-out
	// permanents are treated as though they don't exist (CR 702.26).
	PhasedOut bool
	// PhasedOutFor is the player whose next untap step phases this permanent
	// in. It is captured when the permanent phases out, so later control-effect
	// changes do not alter normal phase-in timing.
	PhasedOutFor PlayerID
	// PhaseInScheduled distinguishes a captured Player1 schedule from legacy
	// state that only marks PhasedOut.
	PhaseInScheduled bool

	// FaceDown is true if this permanent is face-down (e.g., via Morph
	// or Disguise). Face-down permanents are 2/2 creatures with no name,
	// no type, no abilities, and no mana cost (CR 708.2).
	FaceDown bool

	// FaceDownFace records the printed face hidden under a face-down permanent.
	// It is ignored unless FaceDown is true.
	FaceDownFace FaceIndex

	// FaceDownKind records whether this was cast or created face-down by Morph,
	// Disguise, or a future face-down mechanic. It is ignored unless FaceDown is true.
	FaceDownKind FaceDownKind

	// Face is the printed face currently visible for face-up double-faced
	// permanents. Single-faced cards use FaceFront.
	Face FaceIndex

	// Flipped is true for flip cards that have been flipped (CR 710).
	Flipped bool

	// Transformed is true for double-faced cards showing their back face
	// (CR 712).
	Transformed bool

	// --- Counters and damage ---

	// Counters tracks all counters on this permanent.
	Counters counter.Set

	// SagaEntryChapter is the chapter chosen as this Saga entered with read
	// ahead. Chapters below this number remain skipped if lore counters are
	// later removed and added again. Zero identifies an ordinary Saga entry.
	SagaEntryChapter int

	// MarkedDamage is the amount of damage currently marked on this
	// permanent. Cleared during the cleanup step (CR 120.3).
	MarkedDamage int

	// MarkedDeathtouchDamage records whether any damage currently marked on
	// this permanent came from a source with deathtouch. Cleared during the
	// cleanup step alongside MarkedDamage.
	MarkedDeathtouchDamage bool

	// DamagePreventionCounterRemovedThisStep records whether a
	// DamagePreventedRemovesPlusOneCounter replacement (the Phantom mechanic)
	// has already removed a +1/+1 counter from this permanent during the
	// current combat damage step. Combat damage from multiple sources is dealt
	// simultaneously (CR 510.2), so a Phantom blocked by several creatures has
	// all of that damage prevented but loses only one counter total. The flag
	// is reset at the start of each combat damage pass; it is never consulted
	// for noncombat damage, where each source is a separate event.
	DamagePreventionCounterRemovedThisStep bool

	// TemporaryPowerModifier and TemporaryToughnessModifier are additive
	// until-end-of-turn P/T changes. They are cleared during cleanup.
	TemporaryPowerModifier     int
	TemporaryToughnessModifier int

	// RegenerationShields replace future destruction events for this permanent.
	RegenerationShields int
	Monstrous           bool
	ClassLevel          int
	Prepared            bool
	// Renowned records whether this permanent has become renowned (CR 702.111d).
	// A renowned permanent's Renown trigger no longer adds counters.
	Renowned bool

	// Saddled records whether this Mount has been saddled this turn (CR 702.166).
	// It is set by the BecomeSaddled keyword action and cleared during cleanup.
	Saddled bool

	// TributePaid records that a chosen opponent paid this creature's Tribute as
	// it entered by putting its +1/+1 counters on it (CR 702.110). It is read by
	// the paired "if tribute wasn't paid" intervening-if on the creature's enters
	// trigger and is false when tribute was declined or the creature has no
	// Tribute.
	TributePaid bool

	// --- Attachments ---

	// Attachments lists the ObjectIDs of permanents attached to this one
	// (Equipment, Auras, Fortifications).
	Attachments []id.ID

	// AttachedTo is the ObjectID of the permanent this is attached to,
	// if this is an Aura, Equipment, or Fortification. Absent if not attached.
	AttachedTo opt.V[id.ID]

	// AttachedToPlayer is the PlayerID this Aura is attached to, when this is an
	// Aura that enchants a player (a Curse or other "Enchant player" Aura, CR
	// 303.4h). It is deliberately distinct from AttachedTo — which references a
	// permanent — so permanent-only effects that dereference AttachedTo never
	// resolve a player as though it were a permanent. An Aura is attached to a
	// permanent or a player but never both, so at most one of AttachedTo and
	// AttachedToPlayer is present. Absent if this is not attached to a player.
	AttachedToPlayer opt.V[PlayerID]

	// Bestowed is true if this permanent entered the battlefield as a bestowed
	// Aura, i.e. its card was cast for its bestow cost (CR 702.103). While
	// bestowed it is an Aura enchantment and not a creature (see the bestow type
	// change in the characteristic layer). It stops being bestowed — and becomes
	// a creature again on the battlefield — the moment it becomes unattached or
	// is attached to an illegal object (CR 702.103e–g), which clears this flag.
	Bestowed bool

	// ReanimationLinkedObject is set on a graveyard-reanimation Aura (Animate
	// Dead, Dance of the Dead) after it resolves: it records the ObjectID of the
	// creature the Aura returned to the battlefield and attached to. It models
	// the enchant-restriction transition on resolution — the Aura loses "enchant
	// creature card in a graveyard" and gains "enchant creature put onto the
	// battlefield with this Aura" — so the Aura is legally attached only to that
	// one linked permanent. It is zero before the Aura reanimates anything.
	ReanimationLinkedObject id.ID

	// --- Layer ordering ---

	// --- Combat modifiers ---

	// Goaded tracks which players have goaded this creature. A goaded creature
	// must attack each combat if able, and must attack a player other than the
	// one who goaded it if able until the goading player's next turn (CR 701.38).
	Goaded map[PlayerID]GoadStatus

	// SuspendHasteController grants haste while that player controls this
	// permanent after it was cast from suspend.
	SuspendHasteController opt.V[PlayerID]

	// EchoResolvedController records the player for whom this permanent's Echo
	// obligation (CR 702.29) was most recently resolved. It is unset when the
	// permanent enters the battlefield, so the first upkeep of whoever controls
	// it triggers Echo; the echo triggered ability sets it to the resolving
	// controller so later upkeeps of that same controller do not re-trigger. When
	// a different player gains control, the stored controller no longer matches
	// the new controller, so that player's next upkeep triggers Echo again,
	// modeling "it came under your control since the beginning of your last
	// upkeep" without a discrete control-change event. This single scalar is an
	// approximation of full control-since-last-upkeep history; the temporary
	// steal-and-return and countered-trigger gaps it cannot represent are scoped
	// in echoObligationPending and tracked in #3014.
	EchoResolvedController opt.V[PlayerID]

	// EntryChoices stores persistent choices made for this permanent. The legacy
	// name reflects its original entry-choice use; attachment-time choices also
	// live here so all source-scoped chosen values share one cloned, LKI-preserved
	// store. Initialized lazily on first write.
	EntryChoices map[ChoiceKey]ResolutionChoiceResult

	// --- Token support ---

	// Token is true if this permanent is a token rather than a card.
	Token bool

	// TokenDef holds the card definition for tokens. Nil for non-tokens.
	// Tokens use this instead of CardInstanceID.
	TokenDef *CardDef
}

// MergedCard identifies one lower card component of a mutated permanent.
type MergedCard struct {
	CardInstanceID id.ID
	Face           FaceIndex
	FaceDown       bool
	FaceDownFace   FaceIndex
	FaceDownKind   FaceDownKind
	TokenDef       *CardDef
	Owner          PlayerID
}

// Timestamp returns the permanent's timestamp for continuous-effect ordering.
// Permanent timestamps are derived from ObjectID because permanents receive
// monotonically increasing object IDs as they enter the battlefield. Control
// changes are modeled as continuous effects with their own timestamps, so a
// permanent does not need separate mutable timestamp state.
func (p *Permanent) Timestamp() Timestamp {
	return Timestamp(p.ObjectID)
}

// GoadStatus records the duration for one player's goad effect.
type GoadStatus struct {
	CreatedTurn int
	ExpiresFor  PlayerID
	// RestOfGame marks a permanent goad ("goaded for the rest of the game", Life
	// of the Party) that persists until the creature leaves the battlefield
	// instead of expiring on the goading player's next turn. The turn-based
	// expiry skips statuses with this flag set.
	RestOfGame bool
}
