package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
)

// GameResult is the structured output of a completed game.
type GameResult struct {
	Winner           game.PlayerID
	HasWinner        bool
	EliminationOrder []game.PlayerID
	Losses           []LossLog
	TurnCount        int
	Turns            []TurnLog
}

// TurnLog records the decisions and outcomes from a single turn.
type TurnLog struct {
	TurnNumber   int
	ActivePlayer game.PlayerID
	Draws        []DrawLog
	Losses       []LossLog
	Actions      []ActionLog
	Resolves     []ResolveLog
}

// DrawLog records a player draw during a game.
type DrawLog struct {
	Player game.PlayerID
	CardID id.ID
	Failed bool
}

// LossReason describes why a player lost the game.
type LossReason string

const (
	LossReasonEmptyLibraryDraw    LossReason = "draw from empty library"
	LossReasonZeroLife            LossReason = "0 life"
	LossReasonPoisonCounters      LossReason = "10 poison counters"
	LossReasonCommanderDamage     LossReason = "21 commander damage"
	LossReasonStateBasedEliminate LossReason = "state-based elimination"
)

// LossLog records a player losing the game.
type LossLog struct {
	Player game.PlayerID
	Reason LossReason
}

// ActionLog records a player action that occurred during a game.
type ActionLog struct {
	Player game.PlayerID
	Action action.Action
}

// ResolveLog records a stack object resolving.
type ResolveLog struct {
	StackObjectID id.ID
	SourceID      id.ID
	Controller    game.PlayerID
	Kind          game.StackObjectKind
	Result        string
}

func (r *GameResult) addLosses(losses []LossLog) {
	r.Losses = append(r.Losses, losses...)
	for _, loss := range losses {
		r.EliminationOrder = append(r.EliminationOrder, loss.Player)
	}
}
