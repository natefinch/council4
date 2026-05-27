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
	EventBeginningOfStep
	EventLifeGained
	EventLifeLost
	EventPermanentTapped
	EventPermanentUntapped
	EventObjectBecameTarget
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

	// Face records the visible/chosen face at the time this event happened.
	// If a card moved to a non-stack, non-battlefield zone, that destination
	// card still uses front-face characteristics even when this records the
	// face it had while leaving the stack or battlefield.
	Face FaceIndex

	// CardTypes records the relevant card types at event time for spell-cast
	// filters such as "noncreature spell" or "artifact spell"; cast triggers
	// look at the spell as cast on the stack (CR 601.2, CR 603.2).
	CardTypes []CardType

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

	// Step identifies the turn step for EventBeginningOfStep triggers
	// (CR 603.6c).
	Step Step

	// Target records the object or player that became a target.
	Target Target
}

// EventsForTurn returns the rules events emitted during the requested turn.
func (g *Game) EventsForTurn(turnNumber int) []GameEvent {
	if turnNumber <= 0 {
		return nil
	}
	index := turnNumber - 1
	if index < 0 || index >= len(g.EventTurnStarts) {
		return nil
	}
	start := g.EventTurnStarts[index]
	end := len(g.Events)
	if index+1 < len(g.EventTurnStarts) {
		end = g.EventTurnStarts[index+1]
	}
	if start < 0 || start > end || end > len(g.Events) {
		return nil
	}
	return g.Events[start:end]
}

// EventsThisTurn returns the rules events emitted during the current turn.
func (g *Game) EventsThisTurn() []GameEvent {
	return g.EventsForTurn(g.Turn.TurnNumber)
}

// EventsPreviousTurn returns the rules events emitted during the previous turn.
func (g *Game) EventsPreviousTurn() []GameEvent {
	return g.EventsForTurn(g.Turn.TurnNumber - 1)
}
