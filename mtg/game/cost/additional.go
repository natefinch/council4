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

	// SourceSelf requires the ability or spell's own source card rather than a
	// freely chosen matching card.
	SourceSelf bool

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
}

// Alternative describes an optional cost that replaces a spell or ability's
// normal mana cost when selected.
type Alternative struct {
	Label           string
	ManaCost        opt.V[Mana]
	AdditionalCosts []Additional
}
