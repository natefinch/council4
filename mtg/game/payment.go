package game

import (
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AdditionalCostKind classifies a non-mana cost component.
type AdditionalCostKind int

// Additional cost kind values identify supported non-mana costs.
const (
	AdditionalCostUnknown AdditionalCostKind = iota
	AdditionalCostSacrifice
	AdditionalCostSacrificeSource
	AdditionalCostDiscard
	AdditionalCostPayLife
	AdditionalCostExile
	AdditionalCostReveal
	AdditionalCostTap
	AdditionalCostExileSource
)

// AdditionalCost describes a typed non-mana cost printed on a spell, ability,
// or alternative cost. It is data only; mtg/rules chooses and pays it.
type AdditionalCost struct {
	Kind AdditionalCostKind

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

	// Zone is the zone cards are chosen from for card costs. Zero values are
	// interpreted by rules for each cost kind, usually hand.
	Zone ZoneType
}

// AdditionalCostSelection records the concrete objects or cards chosen to pay
// one AdditionalCost.
type AdditionalCostSelection struct {
	Cost AdditionalCost

	PermanentIDs []id.ID
	CardIDs      []id.ID
	LifePaid     int
}

// AlternativeCost describes an optional cost that replaces a spell or ability's
// normal mana cost when selected.
type AlternativeCost struct {
	Label           string
	ManaCost        opt.V[cost.Mana]
	AdditionalCosts []AdditionalCost
}

// SymbolPaymentMethod classifies how one mana symbol was paid.
type SymbolPaymentMethod int

// Symbol payment method values identify how a mana symbol was paid.
const (
	SymbolPaymentUnknown SymbolPaymentMethod = iota
	SymbolPaymentMana
	SymbolPaymentGeneric
	SymbolPaymentHybridFirst
	SymbolPaymentHybridSecond
	SymbolPaymentMonoHybridColor
	SymbolPaymentMonoHybridGeneric
	SymbolPaymentPhyrexianMana
	SymbolPaymentPhyrexianLife
	SymbolPaymentSnow
	SymbolPaymentX
)

// SymbolPayment records how a particular printed or expanded mana symbol was
// satisfied by a payment plan.
type SymbolPayment struct {
	Symbol cost.Symbol
	Method SymbolPaymentMethod

	Color         mana.Color
	GenericAmount int
	LifePaid      int
	Snow          bool
}

// PaymentChoiceKind classifies a non-action choice needed to pay a legal cost.
type PaymentChoiceKind int

// Payment choice kind values classify bounded payment decisions.
const (
	PaymentChoiceUnknown PaymentChoiceKind = iota
	PaymentChoiceSymbol
	PaymentChoiceAdditionalCost
	PaymentChoiceAlternativeCost
)

// PaymentChoice describes a bounded payment decision. Rules code can translate
// these into ChoiceRequests when an agent needs to choose a payment branch.
type PaymentChoice struct {
	Kind   PaymentChoiceKind
	Player PlayerID
	Prompt string

	Symbol          cost.Symbol
	AdditionalCost  AdditionalCost
	AlternativeCost AlternativeCost
}
