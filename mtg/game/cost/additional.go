package cost

import (
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// AdditionalKind classifies a non-mana cost component.
type AdditionalKind int

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

	// MatchPermanentType constrains battlefield costs such as "sacrifice a
	// creature." When false, any permanent is allowed for permanent costs.
	MatchPermanentType bool
	PermanentType      types.Card

	// MatchCardType constrains card costs such as "discard a creature card."
	// When false, any card in the relevant zone is allowed for card costs.
	MatchCardType bool
	CardType      types.Card

	// Source identifies the zone cards are chosen from for card costs.
	// zone.None delegates to the rules-defined default for the cost kind.
	Source zone.Type
}

// Alternative describes an optional cost that replaces a spell or ability's
// normal mana cost when selected.
type Alternative struct {
	Label           string
	ManaCost        opt.V[Mana]
	AdditionalCosts []Additional
}
