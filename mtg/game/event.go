package game

import (
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
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
	EventPermanentPhasedOut
	EventPermanentPhasedIn
	// EventLibrarySearched marks a player searching their library (CR 701.19).
	// It is emitted once per search instruction regardless of whether a card is
	// found, so "whenever a player/an opponent/you searches their library"
	// triggers fire. Event.Player is the searching player.
	EventLibrarySearched
	// EventAttackerBecameUnblocked marks an attacker that remained unblocked
	// after the declare-blockers step completed (CR 509.1h). It is emitted once
	// per unblocked attacker so "whenever this creature attacks and isn't
	// blocked" triggers fire. SourceObjectID/PermanentID identify the attacker.
	EventAttackerBecameUnblocked
	// EventClassLevelGained marks a Class permanent's level rising to a new
	// value (CR 716). It is emitted once for each level passed through so
	// "When this Class becomes level N" triggers fire. SourceObjectID/PermanentID
	// identify the Class, Controller is its controller, and Amount is the new
	// level reached.
	EventClassLevelGained
	// EventCrimeCommitted marks a player committing a crime (CR 700.15): they
	// put a spell or ability on the stack that targets one or more opponents,
	// objects an opponent controls, or cards in an opponent's graveyard. It is
	// emitted once per spell or ability put on the stack, regardless of how many
	// such targets it has. Controller and Player identify the player who
	// committed the crime, and StackObjectID identifies the spell or ability.
	EventCrimeCommitted
	// EventBecameMonarch marks a player becoming the monarch (CR 720.2). It is
	// emitted once each time a player who was not already the monarch takes the
	// crown, whether by a "become the monarch" effect or by dealing combat damage
	// to the monarch, so "whenever you/an opponent become(s) the monarch"
	// triggers fire. Controller and Player identify the new monarch.
	EventBecameMonarch
	// EventCardPlayedFromExile marks a player playing a card (casting it as a
	// spell or playing it as a land) from exile. It is emitted once per card
	// played from exile so "whenever a player plays a card exiled with this"
	// triggers can fire when the played card belongs to their source-keyed
	// linked-exile pool (Prowl, Stoic Strategist). Player identifies the player
	// who played the card and CardID identifies the card.
	EventCardPlayedFromExile
	// EventLandPlayed marks a player playing a land as the land-play special
	// action (CR 305, CR 505.5b), whether the normal land drop for the turn or
	// an additional land play granted by another effect, and regardless of which
	// zone the land was played from. It is emitted once each time a land is
	// played so "whenever a player/an opponent/you play(s) a land" triggers
	// (Burgeoning, Dirtcowl Wurm, Horn of Greed) fire. It is deliberately not
	// emitted when an effect puts a land onto the battlefield without playing it.
	// Controller and Player identify the player who played the land and CardID
	// identifies the land card.
	EventLandPlayed
	// EventGotCityBlessing marks a player getting the city's blessing (CR
	// 702.131 ascend). It is emitted once each time a player who did not already
	// have the city's blessing gets it, whether from a permanent's ascend static
	// ability reaching ten permanents or from resolving an ascend spell, so
	// "whenever you get the city's blessing" triggers could fire. The blessing is
	// never removed, so this event is emitted at most once per player per game.
	// Controller and Player identify the player who got the city's blessing.
	EventGotCityBlessing
)

// EventKindCount is the number of EventKind values, including EventUnknown. It
// is appended at the end of the const block so existing wire values are
// preserved; new kinds must be added immediately before this sentinel.
const EventKindCount = int(EventGotCityBlessing) + 1

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
	Controller                PlayerID
	SourceID                  id.ID
	SourceCardID              id.ID
	SourceTokenDef            *CardDef
	Face                      FaceIndex
	AbilityIndex              int
	Ability                   *TriggeredAbility
	AdditionalTriggers        int
	TriggerMultiplierCaptured bool
}

// ChosenTypeTriggerDoubler is the event-time snapshot of one active
// chosen-creature-type trigger doubler (CR 603.3; Roaming Throne). It records
// the doubler source, its controller, and the chosen subtype as they were when
// an event was emitted, so chosen-type trigger multiplication reflects the
// doubler set, controller, and chosen type at the moment a triggered ability
// triggers rather than at resolution time, even if the doubler later changes
// controller or chosen type or leaves the battlefield.
type ChosenTypeTriggerDoubler struct {
	SourceID   id.ID
	Controller PlayerID
	Subtype    types.Sub
}

// ChosenTypeTriggerDoublerSnapshot holds the active chosen-creature-type trigger
// doublers captured when an event was emitted. It is referenced by pointer from
// Event so events without doublers (the common case) add no storage and keep the
// Event value small enough to pass by value cheaply.
type ChosenTypeTriggerDoublerSnapshot struct {
	Doublers []ChosenTypeTriggerDoubler
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

	// Bargained records whether a spell-cast or entering-permanent spell's
	// Bargain additional cost was paid as it was cast (CR 702.166b). It is false
	// for objects that were not bargained. It feeds the "if it was bargained"
	// enter-trigger intervening-if and the entering permanent's captured state.
	Bargained bool

	// OffspringPaid records whether a spell-cast or entering-permanent spell's
	// Offspring additional mana cost was paid as it was cast (CR 702.171b). It is
	// false for objects that were not cast with offspring paid. It feeds the
	// Offspring enter trigger's "if it was offspring" intervening-if.
	OffspringPaid bool

	// EnterEvoked records whether an entering permanent resulted from a spell
	// cast for its Evoke alternative cost (CR 702.74). It feeds the evoke
	// sacrifice trigger's intervening "if its evoke cost was paid" condition.
	EnterEvoked bool

	// EnterDashed records whether an entering permanent resulted from a spell
	// cast for its Dash alternative cost (CR 702.109). It feeds the Dash
	// trigger's intervening "if its dash cost was paid" condition.
	EnterDashed bool

	// EnterWasCast records whether a permanent entered from resolving a cast,
	// non-copy spell.
	EnterWasCast bool
	// EnterCastController is the player who cast the spell that became the
	// entering permanent. EnterHasCastController distinguishes player zero from
	// an entry that did not result from a cast.
	EnterCastController    PlayerID
	EnterHasCastController bool

	// EnterCastFromZone is the zone the spell was cast from when an entering
	// permanent resulted from resolving a cast. It is only meaningful when
	// EnterWasCast is true; it feeds the "was cast from a graveyard" intervening
	// condition (CR 603.4).
	EnterCastFromZone zone.Type

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

	// ManaSpentToCast records the total amount of mana spent to cast the spell
	// for spell-cast triggers whose intervening-if reads the mana actually paid
	// ("if no mana was spent to cast it", "if at least four mana was spent to
	// cast it"). It is set at every EventSpellCast emission site, including free
	// casts (where it is zero), so the condition reads the same value as the
	// event triggers and as the ability resolves (CR 603.4).
	ManaSpentToCast opt.V[int]

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

	// FirstInDrawStep marks an EventCardDrawn as the player's first draw during
	// their own draw step (the turn-based draw). Triggers carrying
	// ExcludeFirstDrawInDrawStep ignore such a draw ("except the first one they
	// draw in each of their draw steps", Orcish Bowmasters, Xyris).
	FirstInDrawStep bool

	// CounterKind and PreviousCounterAmount describe EventCountersAdded for
	// either PermanentID or Player.
	CounterKind           counter.Kind
	PreviousCounterAmount int

	// DamageRecipient describes whether damage was dealt to a player or permanent.
	DamageRecipient DamageRecipientKind

	// CombatDamage is true when EventDamageDealt came from combat damage.
	CombatDamage bool

	// TappedForMana is true when EventPermanentTapped recorded a tap that paid a
	// mana ability's cost ("tapped for mana"), CR 106.11a / 605.
	TappedForMana bool

	// ProducedManaColors lists, in production order, the distinct types of mana
	// the tap recorded by EventPermanentTapped added (its color, with colorless
	// {C} included). It is populated only for tapped-for-mana taps and backs the
	// "add one mana of any type that land produced" mana-doubler trigger
	// (Mirari's Wake), which mirrors one of these types.
	ProducedManaColors []mana.Color

	// AttackTarget is set for the EventAttackerDeclared, EventAttackerBecameBlocked,
	// and EventAttackerBecameUnblocked combat events, recording what the attacker
	// was declared against (player, planeswalker, or battle).
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

	// ChosenTypeTriggerDoublers snapshots the active chosen-creature-type
	// trigger doublers at event emission, so ordinary triggered abilities this
	// event produces are multiplied by the event-time doubler set, controller,
	// and chosen type rather than by state observed later at resolution. It is
	// nil when no doublers were active.
	ChosenTypeTriggerDoublers *ChosenTypeTriggerDoublerSnapshot
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
	event.ProducedManaColors = append([]mana.Color(nil), event.ProducedManaColors...)
	event.TriggeredAbilities = append([]EventTriggeredAbility(nil), event.TriggeredAbilities...)
	if event.ChosenTypeTriggerDoublers != nil {
		snapshot := ChosenTypeTriggerDoublerSnapshot{
			Doublers: append([]ChosenTypeTriggerDoubler(nil), event.ChosenTypeTriggerDoublers.Doublers...),
		}
		event.ChosenTypeTriggerDoublers = &snapshot
	}
	return event
}
