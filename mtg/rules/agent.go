package rules

import (
	"github.com/natefinch/council4/mtg/game/action"
)

// PlayerAgent chooses a sampled action or proposes a complete typed spell or
// activation announcement. Numeric policy samples are not rules limits.
// The engine always supplies a nonempty candidate list and validates proposals.
type PlayerAgent interface {
	ChooseAction(obs PlayerObservation, legal []action.Action) action.Action
}
