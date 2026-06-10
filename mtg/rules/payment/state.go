// Package payment contains the payment planner — the rules logic for
// building, validating, and applying spell, ability, and generic mana costs.
// It is a deep module under mtg/rules that holds all payment behavior while
// keeping mana-source selection, additional-cost matching, and plan application
// private. Callers in mtg/rules access it through mtg/rules/payment_orchestrator.go.
package payment

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// State is the adapter interface that the payment planner requires from the
// rules engine. It exposes the game queries and mutations that involve
// rules-specific logic (continuous effects, event emission, zone transitions).
// Implementations in mtg/rules wrap *game.Game and the relevant rules helpers.
type State interface {
	stateQueries
	statePermanentQueries
	stateAbilityQueries
	stateMutations
}

type stateQueries interface {
	// Player returns the payment-eligible player, or false if the player is
	// invalid or eliminated.
	Player(playerID game.PlayerID) (*game.Player, bool)

	// Battlefield returns all permanents in deterministic iteration order.
	Battlefield() []*game.Permanent

	// EffectiveController returns the player who controls the permanent,
	// accounting for continuous-effect control changes.
	EffectiveController(p *game.Permanent) game.PlayerID

	// PermanentCardDef returns the card definition for a permanent.
	PermanentCardDef(p *game.Permanent) (*game.CardDef, bool)

	// PermanentByObjectID looks up a permanent by its object ID.
	PermanentByObjectID(objectID id.ID) (*game.Permanent, bool)

	// CardInstance returns the card instance for a card ID.
	CardInstance(cardID id.ID) (*game.CardInstance, bool)

	// CardFace returns the requested face of a card instance, falling back to
	// the base definition when the face does not exist.
	CardFace(card *game.CardInstance, face game.FaceIndex) *game.CardDef
}

type statePermanentQueries interface {
	// PermanentHasType reports whether the permanent currently has the given
	// card type, accounting for continuous type-changing effects.
	PermanentHasType(p *game.Permanent, t types.Card) bool

	// PermanentHasSupertype reports whether the permanent has the given supertype.
	PermanentHasSupertype(p *game.Permanent, s types.Super) bool

	// PermanentHasSubtype reports whether the permanent currently has the given subtype.
	PermanentHasSubtype(p *game.Permanent, s types.Sub) bool

	// PermanentEffectiveColors returns the effective colors of the permanent.
	PermanentEffectiveColors(p *game.Permanent) []color.Color
}

type stateAbilityQueries interface {
	// PermanentEffectiveAbilities returns the permanent's abilities in canonical
	// index order, including abilities from merged components.
	PermanentEffectiveAbilities(p *game.Permanent) []game.Ability

	// ActivationConditionSatisfied reports whether an activated ability's
	// non-timing activation restriction is satisfied.
	ActivationConditionSatisfied(playerID game.PlayerID, permanent *game.Permanent, condition opt.V[game.Condition]) bool

	// ManaAbilityTimingAllowed reports whether timing and per-turn restrictions
	// allow the mana ability to be activated.
	ManaAbilityTimingAllowed(playerID game.PlayerID, permanent *game.Permanent, abilityIndex int, timing game.TimingRestriction) bool

	// CostModifiersForSpell returns all applicable cost modifiers for a spell
	// being cast by the given player from the given zone. This includes global
	// game modifiers, commander tax, and static rule-effect modifiers.
	CostModifiersForSpell(playerID game.PlayerID, card *game.CardDef, cardID id.ID, sourceZone zone.Type) []game.CostModifier
}

type stateMutations interface {
	// SetTapped sets the tapped state of a permanent and emits the appropriate
	// tapped/untapped event.
	SetTapped(p *game.Permanent, tapped bool)

	// RecordManaAbilityUse records a restricted mana ability activation.
	RecordManaAbilityUse(p *game.Permanent, abilityIndex int, timing game.TimingRestriction)

	// RemoveCounters removes exactly amount counters of kind from a permanent.
	RemoveCounters(p *game.Permanent, kind counter.Kind, amount int) bool

	// LoseLife applies life loss to a player, including any applicable
	// replacement effects.
	LoseLife(playerID game.PlayerID, amount int)

	// EmitZoneChange emits a zone-change game event.
	EmitZoneChange(event game.Event)

	// EmitCardReveal records that a card was revealed from a zone while paying a cost.
	EmitCardReveal(playerID game.PlayerID, sourceCardID, cardID id.ID, from zone.Type)

	// MovePermanentToZone moves a permanent to the destination zone,
	// handling detach, zone-change events, and token cleanup.
	MovePermanentToZone(p *game.Permanent, dest zone.Type) bool

	// DiscardFromHand discards a card from the player's hand, emitting the
	// appropriate discard and zone-change events.
	DiscardFromHand(playerID game.PlayerID, cardID id.ID) bool

	// MoveCard moves a non-battlefield card between zones and emits a zone-change event.
	MoveCard(playerID game.PlayerID, cardID id.ID, from zone.Type, to zone.Type) bool
}
