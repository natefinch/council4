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
	ChoiceExplore
	ChoiceManifest
	// ChoiceDamageAllocation asks the controller of a divided-damage effect to
	// split a fixed total among the chosen targets. Each option corresponds to
	// one chosen target; the returned selection lists option indices with
	// repetition, so the number of times an option appears equals the damage
	// allocated to that target. MinChoices and MaxChoices both equal the total
	// damage, and every target must receive at least one (CR 601.2d).
	ChoiceDamageAllocation
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
