package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
)

// PlayerAgent chooses one action from the legal actions available to a player.
// The engine does not call ChooseAction with an empty legal action list, and it
// validates the returned action before applying it.
type PlayerAgent interface {
	ChooseAction(obs PlayerObservation, legal []action.Action) action.Action
}

// PlayerObservation is the fog-of-war filtered game state visible to a player.
type PlayerObservation struct {
	Player game.PlayerID
	Turn   TurnObservation
}

// TurnObservation describes the public turn state relevant to action choice.
type TurnObservation struct {
	TurnNumber     int
	ActivePlayer   game.PlayerID
	PriorityPlayer game.PlayerID
	Phase          game.Phase
	Step           game.Step
}
