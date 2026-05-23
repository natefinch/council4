// Package agent contains AI player implementations.
package agent

import (
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/rules"
)

// FirstLegal chooses the first legal action offered by the rules engine.
//
// This is intentionally simple: legal actions are ordered by the engine so
// productive actions appear before Pass. In the current minimal game loop,
// FirstLegal plays a land when possible and otherwise passes.
type FirstLegal struct{}

// ChooseAction implements rules.PlayerAgent.
func (FirstLegal) ChooseAction(obs rules.PlayerObservation, legal []action.Action) action.Action {
	if len(legal) == 0 {
		return action.Pass()
	}
	return legal[0]
}
