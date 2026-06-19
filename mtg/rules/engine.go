// Package rules contains the Magic rules engine.
package rules

import (
	"math/rand/v2"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

const maxGameTurns = 1000

// Engine owns rule execution and game-loop configuration.
type Engine struct {
	// rng is consumed sequentially by setup and runtime rules. Use one Engine
	// per independently reproducible game stream.
	rng *rand.Rand

	cardImplementations map[string]CardImplementation
}

// NewEngine creates an Engine using rng for deterministic game execution.
func NewEngine(rng *rand.Rand) *Engine {
	if rng == nil {
		rng = rand.New(rand.NewPCG(1, 2))
	}
	return &Engine{
		rng:                 rng,
		cardImplementations: map[string]CardImplementation{},
	}
}

// NewGame creates a game using the engine's RNG for deterministic setup.
func (e *Engine) NewGame(configs [game.NumPlayers]game.PlayerConfig) *game.Game {
	return game.NewGameWithRand(configs, e.rng)
}

// RunGame runs a game to completion and returns its structured result.
func (e *Engine) RunGame(g *game.Game, agents [game.NumPlayers]PlayerAgent) *GameResult {
	result := &GameResult{}
	e.drawOpeningHands(g)
	markCurrentTurnEventStart(g)
	result.addLosses(e.applyStateBasedActions(g))
	if winner, ok := g.Winner(); ok {
		result.Winner = winner.ID
		result.HasWinner = true
		foldFinalState(g, result)
		return result
	}

	for !g.IsGameOver() && len(result.Turns) < maxGameTurns {
		turnLog := e.runTurn(g, agents)
		result.addLosses(turnLog.Losses)
		result.Turns = append(result.Turns, turnLog)
		result.TurnCount = len(result.Turns)
	}

	if winner, ok := g.Winner(); ok {
		result.Winner = winner.ID
		result.HasWinner = true
	}
	foldFinalState(g, result)
	return result
}

// foldFinalState copies the final event stream, end-state, and card identities
// from the game into the result, so consumers never need the live *game.Game.
func foldFinalState(g *game.Game, result *GameResult) {
	result.Events = append([]game.Event(nil), g.Events...)

	result.Cards = make(map[id.ID]CardInfo, len(g.CardInstances))
	for cardID, instance := range g.CardInstances {
		info := CardInfo{Owner: instance.Owner}
		if instance.Def != nil {
			info.Name = instance.Def.Name
			info.ManaValue = instance.Def.ManaValue()
			info.Types = append([]types.Card(nil), instance.Def.Types...)
		}
		result.Cards[cardID] = info
	}

	for i := range g.Players {
		player := g.Players[i]
		result.EndState.Players[i] = PlayerEndState{
			Life:           player.Life,
			Eliminated:     player.Eliminated,
			Hand:           append([]id.ID(nil), player.Hand.All()...),
			LibrarySize:    player.Library.Size(),
			CommanderCasts: player.CommanderCastCount,
		}
	}
}
