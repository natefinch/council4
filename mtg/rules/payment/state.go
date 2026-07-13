// Package payment contains the payment planner — the rules logic for
// building, validating, and applying spell, ability, and generic mana costs.
// It is a deep module under mtg/rules that holds all payment behavior while
// keeping mana-source selection, additional-cost matching, and plan application
// private. Callers in mtg/rules access it through mtg/rules/payment_orchestrator.go.
package payment

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
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

//nolint:interfacebloat // The payment planner needs one read-only adapter surface for all rules-derived game-state queries.
type stateQueries interface {
	// Player returns the payment-eligible player, or false if the player is
	// invalid or eliminated.
	Player(playerID game.PlayerID) (*game.Player, bool)

	// CanPayLife reports whether the player may currently pay life.
	CanPayLife(playerID game.PlayerID) bool

	// PayLifeForManaColor reports whether an active rule effect lets the player
	// pay 2 life rather than a mana of color c for each such colored symbol in a
	// cost ("For each {B} in a cost, you may pay 2 life rather than pay that
	// mana.", K'rrik). It makes matching colored symbols payable like Phyrexian
	// symbols of that color.
	PayLifeForManaColor(playerID game.PlayerID, c mana.Color) bool

	// ManaProductionMultiplier returns the factor by which mana produced when
	// playerID taps a permanent for mana is scaled, the product of all active
	// RuleEffectManaProductionMultiplier effects the player controls ("If you tap
	// a permanent for mana, it produces twice as much of that mana instead.", Mana
	// Reflection; Nyxbloom Ancient). It returns 1 when no such effect applies.
	ManaProductionMultiplier(playerID game.PlayerID) int

	// ActivateAbilitiesAsThoughHaste reports whether an active rule effect lets
	// playerID activate abilities of creatures they control as though those
	// creatures had haste ("You may activate abilities of creatures you control as
	// though those creatures had haste.", Thousand-Year Elixir). When it does, a
	// summoning-sick creature the player controls may still pay a {T} or {Q} cost
	// in one of its own activated abilities (CR 302.6, 702.10c).
	ActivateAbilitiesAsThoughHaste(playerID game.PlayerID) bool

	// ActivePlayer returns the player whose turn it currently is.
	ActivePlayer() game.PlayerID

	// OpponentLostLifeThisTurn reports whether any opponent of playerID has lost
	// life so far this turn, backing the Spectacle alternative-cost condition.
	OpponentLostLifeThisTurn(playerID game.PlayerID) bool

	// OpponentGainedLifeThisTurn reports whether any opponent of playerID has
	// gained life so far this turn, backing the "If an opponent gained life this
	// turn," mana-only alternative-cost condition (Needlebite Trap).
	OpponentGainedLifeThisTurn(playerID game.PlayerID) bool

	// AttackingCreatureCount returns the number of creatures currently declared
	// as attackers, backing the "If N or more creatures are attacking," mana-only
	// alternative-cost condition (Lethargy Trap, Arrow Volley Trap, Pitfall Trap).
	AttackingCreatureCount() int

	// AdditionalDynamicAmountValue resolves a rules-derived additional-cost
	// amount against live game state.
	AdditionalDynamicAmountValue(playerID game.PlayerID, kind cost.AdditionalDynamicAmount) int

	// Battlefield returns all permanents in deterministic iteration order.
	Battlefield() []*game.Permanent

	// EffectiveController returns the player who controls the permanent,
	// accounting for continuous-effect control changes.
	EffectiveController(p *game.Permanent) game.PlayerID

	// PermanentCardDef returns the card definition for a permanent.
	PermanentCardDef(p *game.Permanent) (*game.CardDef, bool)

	// PermanentByObjectID looks up a permanent by its object ID.
	PermanentByObjectID(objectID id.ID) (*game.Permanent, bool)

	// PermanentPower returns the permanent's effective power, accounting for
	// counters and continuous power-modifying effects.
	PermanentPower(p *game.Permanent) int

	// IsCommanderPermanent reports whether a permanent contains a modeled
	// commander card.
	IsCommanderPermanent(p *game.Permanent) bool

	// CardInstance returns the card instance for a card ID.
	CardInstance(cardID id.ID) (*game.CardInstance, bool)

	// CardFace returns the requested face of a card instance, falling back to
	// the base definition when the face does not exist.
	CardFace(card *game.CardInstance, face game.FaceIndex) *game.CardDef

	// CardMatchesSelection reports whether the card face satisfies the
	// selection's printed-characteristic predicates. The planner uses it so an
	// additional card cost (discard/exile/reveal) tests the same eligibility
	// predicate as the choice layer.
	CardMatchesSelection(card *game.CardDef, sel game.Selection) bool
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

	// PermanentMatchesSelection reports whether the permanent satisfies the
	// selection's characteristic predicates, evaluated against live continuous
	// effects. The planner uses it so additional-cost candidate filtering matches
	// the choice layer's eligible set exactly.
	PermanentMatchesSelection(p *game.Permanent, sel game.Selection) bool
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
	// game modifiers, commander tax, and static rule-effect modifiers. targets
	// carries the spell's chosen targets so target-dependent modifiers ("Spells
	// that target this creature cost {N} more to cast.") can match.
	CostModifiersForSpell(playerID game.PlayerID, card *game.CardDef, cardID id.ID, sourceZone zone.Type, targets []game.Target) []game.CostModifier

	// SpellHasGrantedKeyword reports whether an active rule effect grants keyword
	// to the spell playerID is casting from sourceZone, in addition to any
	// keyword the card carries natively ("Nonartifact spells you cast have
	// improvise.", Inspiring Statuary; "The next spell you cast this turn has
	// improvise.", Archway of Innovation). It lets the payment planner honor a
	// granted cost-affecting keyword (Improvise) while building cost options,
	// before costs are paid. The one-shot next-spell grant stays active here and
	// is consumed only when a matching spell is actually cast.
	SpellHasGrantedKeyword(playerID game.PlayerID, card *game.CardDef, cardID id.ID, sourceZone zone.Type, keyword game.Keyword) bool
}

//nolint:interfacebloat // Payment plans need one adapter surface for all atomic game-state mutations.
type stateMutations interface {
	// SetTapped sets the tapped state of a permanent and emits the appropriate
	// tapped/untapped event.
	SetTapped(p *game.Permanent, tapped bool)

	// SetTappedForMana taps a permanent to pay a mana ability's cost, recording
	// tapped-for-mana provenance on the emitted event.
	SetTappedForMana(p *game.Permanent)

	// RecordManaAbilityUse records a restricted mana ability activation.
	RecordManaAbilityUse(p *game.Permanent, abilityIndex int, timing game.TimingRestriction)

	// RemoveCounters removes exactly amount counters of kind from a permanent.
	RemoveCounters(p *game.Permanent, kind counter.Kind, amount int) bool

	// AddCounters adds amount counters of kind to a permanent.
	AddCounters(playerID game.PlayerID, p *game.Permanent, kind counter.Kind, amount int) bool

	// ExertPermanent marks a permanent to skip its controller's next untap step.
	ExertPermanent(p *game.Permanent) bool

	// MillCards moves up to amount cards from the player's library to their graveyard.
	MillCards(playerID game.PlayerID, amount int)

	// LoseLife applies life loss to a player, including any applicable
	// replacement effects.
	LoseLife(playerID game.PlayerID, amount int)

	// SetPlayerEnergyCounters sets a player's energy counter total.
	SetPlayerEnergyCounters(playerID game.PlayerID, amount int) bool

	// EmitZoneChange emits a zone-change game event.
	EmitZoneChange(event game.Event)

	// EmitCardReveal records that a card was revealed from a zone while paying a cost.
	EmitCardReveal(playerID game.PlayerID, sourceCardID, cardID id.ID, from zone.Type)

	// MovePermanentToZone moves a permanent to the destination zone,
	// handling detach, zone-change events, and token cleanup.
	MovePermanentToZone(p *game.Permanent, dest zone.Type) bool

	// SacrificePermanent moves a permanent to its graveyard as a sacrifice and
	// emits the corresponding sacrifice and zone-change events.
	SacrificePermanent(p *game.Permanent) bool

	// DiscardFromHand discards a card from the player's hand, emitting the
	// appropriate discard and zone-change events.
	DiscardFromHand(playerID game.PlayerID, cardID id.ID) bool

	// DiscardAtRandom discards exactly amount cards chosen uniformly at random
	// from the player's hand as a single simultaneous batch (CR 701.9a),
	// emitting the appropriate discard and zone-change events. It returns false
	// when the player's hand holds fewer than amount cards.
	DiscardAtRandom(playerID game.PlayerID, amount int) bool

	// MoveCard moves a non-battlefield card between zones and emits a zone-change event.
	MoveCard(playerID game.PlayerID, cardID id.ID, from zone.Type, to zone.Type) bool
}
