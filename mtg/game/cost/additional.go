package cost

import (
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// AdditionalKind classifies a non-mana cost component.
type AdditionalKind int

// AdditionalDynamicAmount identifies a rules-derived amount for an additional
// cost whose value is not a fixed integer or the announced X. The rules engine
// resolves it against live game state while building the payment plan, so the
// payment vocabulary stays independent of the effect-resolution dynamic-amount
// machinery (which lives in package game and cannot be imported here).
type AdditionalDynamicAmount uint8

// Additional dynamic amount kinds recognized by the payment planner.
const (
	AdditionalDynamicAmountNone AdditionalDynamicAmount = iota
	// AdditionalDynamicCommanderColorIdentityCount is the number of colors in
	// the paying player's commander's color identity (CR 903.4), backing
	// "Pay life equal to the number of colors in your commanders' color
	// identity" (War Room).
	AdditionalDynamicCommanderColorIdentityCount
	// AdditionalDynamicHandSize is the number of cards in the paying player's
	// hand, backing a "discard your hand" cost (Lion's Eye Diamond).
	AdditionalDynamicHandSize
)

// SubtypeSet holds the one or two alternative subtypes supported by a card
// cost. Empty entries are ignored.
type SubtypeSet [2]types.Sub

// Additional cost kinds identify supported non-mana costs.
const (
	AdditionalUnknown AdditionalKind = iota
	AdditionalSacrifice
	AdditionalSacrificeSource
	AdditionalDiscard
	AdditionalPayLife
	AdditionalExile
	AdditionalReveal
	AdditionalTap
	AdditionalExileSource
	AdditionalUntap
	AdditionalRemoveCounter
	AdditionalReturnUnblockedAttacker
	AdditionalTapPermanents
	AdditionalEnergy
	AdditionalReturnToHand
	AdditionalExert
	AdditionalMill
	AdditionalPutCounter
	AdditionalCollectEvidence
)

// Additional describes a typed non-mana cost printed on a spell, ability, or
// alternative cost. It is data only; mtg/rules chooses and pays it.
type Additional struct {
	Kind AdditionalKind

	// Text preserves the human-readable cost text for logs and diagnostics.
	Text string

	// Amount is the number of matching objects/cards or life points required.
	// Zero means one for object/card costs.
	Amount int

	// AmountFromX uses the announced X value as the required amount.
	AmountFromX bool

	// AmountDynamic, when not AdditionalDynamicAmountNone, names a rules-derived
	// amount the payment planner resolves against live game state. It takes
	// precedence over Amount and AmountFromX for the cost's required count.
	AmountDynamic AdditionalDynamicAmount

	// MatchPermanentType constrains battlefield costs such as "sacrifice a
	// creature." When false, any permanent is allowed for permanent costs.
	MatchPermanentType bool
	PermanentType      types.Card

	// PermanentTypeAlt is an optional second permanent type accepted by a
	// battlefield cost printed as a two-type union, such as "sacrifice an
	// artifact or creature." It is honored only when MatchPermanentType is
	// true; an empty value constrains the cost to PermanentType alone.
	PermanentTypeAlt types.Card

	// MatchCardType constrains card costs such as "discard a creature card."
	// When false, any card in the relevant zone is allowed for card costs.
	MatchCardType bool
	CardType      types.Card

	// MatchCardColor constrains card costs to cards with the listed color.
	MatchCardColor bool
	CardColor      color.Color

	// SubtypesAny constrains card costs to cards with at least one listed
	// subtype. It is independent of MatchCardType and remains bounded so
	// Additional values are comparable.
	SubtypesAny SubtypeSet

	// Source identifies the zone cards are chosen from for card costs.
	// zone.None delegates to the rules-defined default for the cost kind.
	Source zone.Type

	// CounterKind identifies the counter removed from the source permanent by
	// an AdditionalRemoveCounter cost.
	CounterKind counter.Kind

	// RequireTapped constrains battlefield costs to tapped permanents.
	RequireTapped bool

	// RequireSupertype constrains battlefield costs to permanents with a
	// particular supertype, such as Snow.
	RequireSupertype types.Super

	// ExcludeSource constrains a battlefield cost to permanents other than the
	// paying ability's own source, as required by "another" (e.g. "Sacrifice
	// another creature").
	ExcludeSource bool

	// ChoiceGroup tags this cost as one alternative within a numbered choice
	// group printed as "<cost> or <cost>" (e.g. "sacrifice an artifact or
	// discard a card"). Zero means a mandatory standalone cost. Costs sharing a
	// nonzero ChoiceGroup are alternatives; the payer pays exactly one member of
	// each group. It stays a scalar so Additional values remain comparable.
	ChoiceGroup uint8
}

// Alternative describes an optional cost that replaces a spell or ability's
// normal mana cost when selected.
type Alternative struct {
	Label           string
	ManaCost        opt.V[Mana]
	AdditionalCosts []Additional
	Condition       AlternativeCondition
}

// AlternativeCondition identifies a condition that must be true to select an
// alternative cost.
type AlternativeCondition uint8

// Supported alternative-cost conditions.
const (
	AlternativeConditionNone AlternativeCondition = iota
	AlternativeConditionControlsCommander
	// AlternativeConditionNotYourTurn requires that it is not the casting
	// player's turn, backing the Force of Negation pitch family.
	AlternativeConditionNotYourTurn
)
