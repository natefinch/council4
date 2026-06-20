package game

import (
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// EventKind identifies a rules-relevant fact that occurred during a game.
type EventKind int

// Event kind values enumerate rules-relevant game facts.
const (
	EventUnknown EventKind = iota
	EventSpellCast
	EventSpellResolved
	EventPermanentEnteredBattlefield
	EventPermanentDied
	EventCountersAdded
	EventDamageDealt
	EventCardDrawn
	EventZoneChanged
	EventAttackerDeclared
	EventBlockerDeclared
	EventCardDiscarded
	EventCycled
	EventDamagePrevented
	EventDestroyReplaced
	EventBeginningOfStep
	EventLifeGained
	EventLifeLost
	EventPermanentTapped
	EventPermanentUntapped
	EventObjectBecameTarget
	EventCardRevealed
	EventPermanentTurnedFaceUp
	EventPermanentSacrificed
	EventScry
	EventSurveil
	EventAbilityActivated
	EventFight
	EventPermanentMutated
	EventAttackerBecameBlocked
	EventTokenCreated
	// EventSpellCopied marks a spell copy created on the stack (CR 707). It is
	// distinct from EventSpellCast so cast-only triggers and cast counts ignore
	// copies, while "cast or copy" (magecraft) triggers can opt in.
	EventSpellCopied
)

// DamageRecipientKind identifies what received damage. Values are flags so a
// trigger pattern can match either kind.
type DamageRecipientKind int

// Damage recipient values identify what received damage.
const (
	DamageRecipientNone DamageRecipientKind = iota
	DamageRecipientPlayer
	DamageRecipientPermanent
)

// EventTriggeredAbility preserves a triggered ability and its source/controller
// identity at the moment an event occurs. Trigger processing may be deferred
// until a player would receive priority, after the source has left the
// battlefield or changed controller.
type EventTriggeredAbility struct {
	Controller     PlayerID
	SourceID       id.ID
	SourceCardID   id.ID
	SourceTokenDef *CardDef
	Face           FaceIndex
	AbilityIndex   int
	Ability        *TriggeredAbility
}

// Event records a rules-relevant fact emitted by rules helpers as state
// changes happen. Events are data, not behavior: card definitions and reports
// may refer to this vocabulary, while mtg/rules owns emission and consumers.
type Event struct {
	Kind EventKind

	// SourceID is the source card instance ID when there is one.
	SourceID id.ID

	// SourceObjectID is the source permanent object ID when there is one.
	SourceObjectID id.ID

	// StackObjectID is set for spell or ability cast/resolution events.
	StackObjectID id.ID

	// SimultaneousID groups events that happened simultaneously.
	SimultaneousID id.ID

	// AbilityIndex identifies an activated or triggered ability on its source.
	AbilityIndex int

	// ManaAbility is true when EventAbilityActivated describes a mana ability.
	ManaAbility bool

	// Controller is the player who controlled the source spell, ability, or permanent.
	Controller PlayerID

	// Player is the affected player for draw, discard, sacrifice, player-damage,
	// and player-counter events, or the player in whose direction an attacker
	// was declared.
	Player PlayerID

	// CardID identifies the card that moved, was drawn, discarded, or became a permanent.
	CardID id.ID

	// CardZoneVersion identifies the incarnation of CardID involved in this
	// zone-change event.
	CardZoneVersion uint64

	// Face records the visible/chosen face at the time this event happened.
	// If a card moved to a non-stack, non-battlefield zone, that destination
	// card still uses front-face characteristics even when this records the
	// face it had while leaving the stack or battlefield.
	Face FaceIndex

	// FaceDown records whether the moving object had hidden face-down
	// characteristics at event time. Printed abilities do not apply while this
	// is true.
	FaceDown bool

	// KickerPaid records whether a spell-cast or entering-permanent spell's
	// kicker cost was paid. It is false for objects that were not kicked.
	KickerPaid bool

	// EnterWasCast records whether a permanent entered from resolving a cast,
	// non-copy spell.
	EnterWasCast bool

	// CardTypes records the relevant card types at event time for spell-cast
	// filters such as "noncreature spell" or "artifact spell"; cast triggers
	// look at the spell as cast on the stack (CR 601.2, CR 603.2).
	CardTypes []types.Card

	// CardSupertypes and CardSubtypes record spell characteristics at cast time
	// for spell-cast filters such as "historic spell" or "Spirit or Arcane spell".
	CardSupertypes []types.Super
	CardSubtypes   []types.Sub

	// Colors records the colors of the spell as cast on the stack for
	// color-filtered cast triggers such as "a blue spell". Populated at every
	// EventSpellCast emission site from the effective face being cast.
	Colors []color.Color

	// ManaValue records the mana value of the spell as cast on the stack for
	// mana-value-filtered cast triggers.
	ManaValue opt.V[int]

	// PermanentID identifies the permanent that entered, left, was damaged, attacked, or blocked.
	PermanentID id.ID

	// RelatedPermanentID identifies a secondary permanent for paired events such
	// as fights, or the other combatant for block declarations.
	RelatedPermanentID id.ID

	// TokenName gives token events a stable human-readable identity when CardID is zero.
	TokenName string

	// TokenDef preserves last-known card definition data for token events.
	// CardDef values are shared immutable definitions; event consumers must not
	// mutate data reachable through this pointer.
	TokenDef *CardDef

	// FromZone and ToZone describe a zone transition. zone.Battlefield is the
	// battlefield side of permanent enter/leave events.
	FromZone zone.Type
	ToZone   zone.Type

	// Amount is the number of damage dealt, cards drawn, cards discarded,
	// counters added, or cards instructed to be scried or surveilled.
	Amount int

	// PlayerEventOrdinalThisTurn is this player's ordinal occurrence of Kind
	// during the current turn. It is populated only for events with supported
	// ordinal trigger semantics.
	PlayerEventOrdinalThisTurn int

	// CounterKind and PreviousCounterAmount describe EventCountersAdded for
	// either PermanentID or Player.
	CounterKind           counter.Kind
	PreviousCounterAmount int

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

	// TriggeredAbilitiesCaptured distinguishes an event whose battlefield
	// triggers were checked at event time, including when none matched.
	TriggeredAbilitiesCaptured bool
	TriggeredAbilities         []EventTriggeredAbility
}

// AppendEvent records a rules event. Unknown events are ignored.
func (g *Game) AppendEvent(event Event) {
	if event.Kind == EventUnknown {
		return
	}
	g.Events = append(g.Events, cloneEvent(event))
}

// EventsForTurn returns the rules events emitted during the requested turn.
func (g *Game) EventsForTurn(turnNumber int) []Event {
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
	return cloneEvents(g.Events[start:end])
}

// EventsThisTurn returns the rules events emitted during the current turn.
func (g *Game) EventsThisTurn() []Event {
	return g.EventsForTurn(g.Turn.TurnNumber)
}

// EventsPreviousTurn returns the rules events emitted during the previous turn.
func (g *Game) EventsPreviousTurn() []Event {
	return g.EventsForTurn(g.Turn.TurnNumber - 1)
}

func cloneEvents(events []Event) []Event {
	if len(events) == 0 {
		return nil
	}
	cloned := make([]Event, len(events))
	for i, event := range events {
		cloned[i] = cloneEvent(event)
	}
	return cloned
}

func cloneEvent(event Event) Event {
	event.CardTypes = append([]types.Card(nil), event.CardTypes...)
	event.CardSupertypes = append([]types.Super(nil), event.CardSupertypes...)
	event.CardSubtypes = append([]types.Sub(nil), event.CardSubtypes...)
	event.Colors = append([]color.Color(nil), event.Colors...)
	event.TriggeredAbilities = append([]EventTriggeredAbility(nil), event.TriggeredAbilities...)
	return event
}
