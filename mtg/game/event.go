package game

import "github.com/natefinch/council4/mtg/game/id"

// EventKind identifies a rules-relevant fact that occurred during a game.
type EventKind int

const (
	EventUnknown EventKind = iota
	EventSpellCast
	EventSpellResolved
	EventPermanentEnteredBattlefield
	EventPermanentDied
	EventDamageDealt
	EventCardDrawn
	EventZoneChanged
	EventAttackerDeclared
	EventBlockerDeclared
	EventCardDiscarded
	EventDamagePrevented
	EventDestroyReplaced
)

// DamageRecipientKind identifies what received damage.
type DamageRecipientKind int

const (
	DamageRecipientNone DamageRecipientKind = iota
	DamageRecipientPlayer
	DamageRecipientPermanent
)

// GameEvent records a rules-relevant fact emitted by rules helpers as state
// changes happen. Events are data, not behavior: card definitions and reports
// may refer to this vocabulary, while mtg/rules owns emission and consumers.
type GameEvent struct {
	Kind EventKind

	// SourceID is the source card instance ID when there is one.
	SourceID id.ID

	// SourceObjectID is the source permanent object ID when there is one.
	SourceObjectID id.ID

	// StackObjectID is set for spell or ability cast/resolution events.
	StackObjectID id.ID

	// AbilityIndex identifies an activated or triggered ability on its source.
	AbilityIndex int

	// Controller is the player who controlled the source spell, ability, or permanent.
	Controller PlayerID

	// Player is the affected player for draw, discard, and player-damage events.
	Player PlayerID

	// CardID identifies the card that moved, was drawn, discarded, or became a permanent.
	CardID id.ID

	// PermanentID identifies the permanent that entered, left, was damaged, attacked, or blocked.
	PermanentID id.ID

	// TokenName gives token events a stable human-readable identity when CardID is zero.
	TokenName string

	// TokenDef preserves last-known card definition data for token events.
	TokenDef *CardDef

	// FromZone and ToZone describe a zone transition. ZoneBattlefield is the
	// battlefield side of permanent enter/leave events.
	FromZone ZoneType
	ToZone   ZoneType

	// Amount is the number of damage dealt, cards drawn, or cards discarded.
	Amount int

	// DamageRecipient describes whether damage was dealt to a player or permanent.
	DamageRecipient DamageRecipientKind

	// CombatDamage is true when EventDamageDealt came from combat damage.
	CombatDamage bool

	// AttackTarget is set for EventAttackerDeclared.
	AttackTarget AttackTarget

	// BlockedAttackerID is set for EventBlockerDeclared.
	BlockedAttackerID id.ID
}
