package game

import (
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ChoiceKind classifies an engine-mediated decision that is not a normal
// priority action.
type ChoiceKind int

// Choice kind values classify engine-mediated decisions.
const (
	ChoiceMay ChoiceKind = iota
	ChoiceTarget
	ChoiceOrder
	ChoicePayment
	ChoiceScry
	ChoiceSurveil
	ChoiceZoneSelection
	ChoiceSearch
	ChoiceModal
	ChoiceResolution
	ChoiceProliferate
	ChoicePlayer
	ChoiceExplore
	ChoiceManifest
	// ChoiceDamageAllocation asks the controller of a divided-damage effect to
	// split a fixed total among the chosen targets. Each option corresponds to
	// one chosen target; the returned selection lists option indices with
	// repetition, so the number of times an option appears equals the damage
	// allocated to that target. MinChoices and MaxChoices both equal the total
	// damage, and every target must receive at least one (CR 601.2d).
	ChoiceDamageAllocation
	// ChoiceDig asks the controller of a Dig effect which of the cards revealed
	// from the top of their library to put into their hand. Each option
	// corresponds to one looked-at card; the returned selection lists the option
	// indices of the cards taken. MinChoices and MaxChoices both equal the number
	// of cards taken, bounded by the number of cards actually seen.
	ChoiceDig
)

// ChoiceCardInfo carries the public characteristics of a card or permanent that
// a ChoiceOption or ChoiceRequest refers to, so an agent can make a card-aware
// decision (which card to discard, where to scry a card) without looking the
// card up itself. Only public characteristics are included.
type ChoiceCardInfo struct {
	CardID    id.ID
	Name      string
	Types     []types.Card
	ManaValue int
	Colors    []color.Color
}

// ChoiceOption is one legal option in a ChoiceRequest.
type ChoiceOption struct {
	Index int
	Label string

	// Card carries the public characteristics of the card or permanent this
	// option selects, when the option represents one. It is unset for options
	// that are not a specific card (e.g. "top"/"bottom" or "Pay 2 life").
	Card opt.V[ChoiceCardInfo]
}

// ChoiceRequest describes a bounded decision the rules engine needs from a
// player while resolving rules procedures such as triggered abilities.
type ChoiceRequest struct {
	ID         int
	Kind       ChoiceKind
	Player     PlayerID
	Prompt     string
	Options    []ChoiceOption
	MinChoices int
	MaxChoices int

	// DefaultSelection is used by the rules engine when no agent supplies a
	// valid answer. It must contain option indices valid for this request.
	DefaultSelection []int

	// MaxTotalManaValue, when set, caps the combined mana value of the selected
	// options. A selection is valid only when the sum of its options' card mana
	// values does not exceed this cap. Options without card info contribute zero.
	MaxTotalManaValue opt.V[int]

	// Subject carries the public characteristics of the single card a decision
	// concerns, when there is one — for example the card being placed by a scry
	// or surveil prompt. It is unset for choices that are not about one specific
	// card.
	Subject opt.V[ChoiceCardInfo]
}

// ChoiceDecision records the selected option indices for a ChoiceRequest.
type ChoiceDecision struct {
	Request      ChoiceRequest
	Selected     []int
	UsedFallback bool
}
