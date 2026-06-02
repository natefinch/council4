package game

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
)

// ChoiceOption is one legal option in a ChoiceRequest.
type ChoiceOption struct {
	Index int
	Label string
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
}

// ChoiceDecision records the selected option indices for a ChoiceRequest.
type ChoiceDecision struct {
	Request      ChoiceRequest
	Selected     []int
	UsedFallback bool
}
