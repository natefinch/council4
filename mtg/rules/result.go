package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
)

// GameResult is the structured output of a completed game.
type GameResult struct {
	Winner           game.PlayerID
	HasWinner        bool
	EliminationOrder []game.PlayerID
	TurnCount        int
	Turns            []TurnLog
}

// TurnLog records the decisions and outcomes from a single turn.
type TurnLog struct {
	TurnNumber   int
	ActivePlayer game.PlayerID
	Actions      []ActionLog
}

// ActionLog records a player action that occurred during a game.
type ActionLog struct {
	Player game.PlayerID
	Action action.Action
}
