package rules

import (
	"github.com/natefinch/council4/mtg/game/action"
)

// PlayerAgent chooses one action from the legal actions available to a player.
// The engine does not call ChooseAction with an empty legal action list, and it
// validates the returned action before applying it.
type PlayerAgent interface {
	ChooseAction(obs PlayerObservation, legal []action.Action) action.Action
}
