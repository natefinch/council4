package game

import (
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
)

// AdditionalCostSelection records the concrete objects or cards chosen to pay
// one cost.Additional.
type AdditionalCostSelection struct {
	Cost cost.Additional

	PermanentIDs []id.ID
	CardIDs      []id.ID
	LifePaid     int
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
	AdditionalCost  cost.Additional
	AlternativeCost cost.Alternative
}
