package game

import (
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
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

	// Owner is the player who owns the underlying card. For tokens, this is
	// the player who created the token (CR 111.2).
	Owner PlayerID

	// Controller is the player who currently controls this permanent.
	// Defaults to Owner but can change via control-changing effects.
	Controller PlayerID

	// --- Status flags ---

	// Tapped is true if the permanent is tapped (turned sideways).
	Tapped bool

	// SummoningSick is true if this permanent has not been under its
	// controller's control since the start of that player's most recent turn.
	// It only restricts creatures from attacking and activating abilities with
	// {T} or {Q} in the cost (CR 302.6).
	SummoningSick bool

	// PhasedOut is true if this permanent is phased out. Phased-out
	// permanents are treated as though they don't exist (CR 702.26).
	PhasedOut bool

	// FaceDown is true if this permanent is face-down (e.g., via Morph
	// or Disguise). Face-down permanents are 2/2 creatures with no name,
	// no type, no abilities, and no mana cost (CR 708.2).
	FaceDown bool

	// Flipped is true for flip cards that have been flipped (CR 710).
	Flipped bool

	// Transformed is true for double-faced cards showing their back face
	// (CR 712).
	Transformed bool

	// --- Counters and damage ---

	// Counters tracks all counters on this permanent.
	Counters counter.Set

	// MarkedDamage is the amount of damage currently marked on this
	// permanent. Cleared during the cleanup step (CR 120.3).
	MarkedDamage int

	// MarkedDeathtouchDamage records whether any damage currently marked on
	// this permanent came from a source with deathtouch. Cleared during the
	// cleanup step alongside MarkedDamage.
	MarkedDeathtouchDamage bool

	// TemporaryPowerModifier and TemporaryToughnessModifier are additive
	// until-end-of-turn P/T changes. They are cleared during cleanup.
	TemporaryPowerModifier     int
	TemporaryToughnessModifier int

	// RegenerationShields replace future destruction events for this permanent.
	RegenerationShields int

	// --- Attachments ---

	// Attachments lists the ObjectIDs of permanents attached to this one
	// (Equipment, Auras, Fortifications).
	Attachments []id.ID

	// AttachedTo is the ObjectID of the permanent this is attached to,
	// if this is an Aura, Equipment, or Fortification. Nil if not attached.
	AttachedTo *id.ID

	// --- Timestamps and layer ordering ---

	// Timestamp records when this permanent entered the battlefield or
	// when its most recent control-change occurred, for continuous effect
	// ordering in the layer system (CR 613.7).
	Timestamp int64

	// --- Combat modifiers ---

	// Goaded tracks which players have goaded this creature. A goaded
	// creature must attack each combat if able, and must attack a player
	// other than the one who goaded it if able (CR 701.15).
	Goaded map[PlayerID]bool

	// --- Token support ---

	// Token is true if this permanent is a token rather than a card.
	Token bool

	// TokenDef holds the card definition for tokens. Nil for non-tokens.
	// Tokens use this instead of CardInstanceID.
	TokenDef *CardDef
}
