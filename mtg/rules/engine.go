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

	// searchRNG seeds the isolated forward-model engines a SearchAgent uses to
	// look ahead (see search.go). It is a separate stream from rng, seeded
	// deterministically, so search never perturbs the live game's RNG and a game
	// stays reproducible whether or not a searching agent is at the table.
	searchRNG *rand.Rand

	cardImplementations map[string]CardImplementation
}

// searchSeed is the fixed seed for every engine's search RNG stream. Fixed (not
// derived from the live rng) so search randomness is independent of and never
// advances the live game stream; per-engine *rand.Rand objects keep parallel
// games isolated.
const (
	searchSeedHi = 0x5EA5C45EED1234
	searchSeedLo = 0x9E3779B97F4A7C15
)

// NewEngine creates an Engine using rng for deterministic game execution.
func NewEngine(rng *rand.Rand) *Engine {
	if rng == nil {
		rng = rand.New(rand.NewPCG(1, 2))
	}
	return &Engine{
		rng:                 rng,
		searchRNG:           rand.New(rand.NewPCG(searchSeedHi, searchSeedLo)),
		cardImplementations: map[string]CardImplementation{},
	}
}

// NewGame creates a game using the engine's RNG for deterministic setup.
func (e *Engine) NewGame(configs [game.NumPlayers]game.PlayerConfig) *game.Game {
	return game.NewGameWithRand(configs, e.rng)
}

// NewGoldfishGame creates a single-player game using the engine's RNG.
func (e *Engine) NewGoldfishGame(config game.PlayerConfig) *game.Game {
	return game.NewGoldfishGameWithRand(config, e.rng)
}

// RunGame runs a game to completion and returns its structured result.
func (e *Engine) RunGame(g *game.Game, agents [game.NumPlayers]PlayerAgent) *GameResult {
	return e.RunGameWithTurnLimit(g, agents, maxGameTurns)
}

// RunGameWithTurnLimit plays a full multiplayer game like RunGame but stops
// after at most turnLimit turns, so a caller in a constrained environment (for
// example a browser running the engine in WebAssembly) can bound a game that
// would otherwise durdle toward the 1000-turn safety cap and exhaust memory. A
// turnLimit of zero or less, or above maxGameTurns, uses the maxGameTurns safety
// cap. When the limit is reached with no winner the result has HasWinner false.
func (e *Engine) RunGameWithTurnLimit(g *game.Game, agents [game.NumPlayers]PlayerAgent, turnLimit int) *GameResult {
	if turnLimit <= 0 || turnLimit > maxGameTurns {
		turnLimit = maxGameTurns
	}
	result := &GameResult{}
	e.drawOpeningHands(g)
	result.OpeningHand = append([]id.ID(nil), g.Players[game.Player1].Hand.All()...)
	markCurrentTurnEventStart(g)
	result.addLosses(e.applyStateBasedActions(g))
	if winner, ok := g.Winner(); ok {
		result.Winner = winner.ID
		result.HasWinner = true
		foldFinalState(g, result)
		return result
	}

	for !g.IsGameOver() && len(result.Turns) < turnLimit {
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

// RunGoldfish plays one deck alone for at most turnLimit complete turns.
func (e *Engine) RunGoldfish(g *game.Game, agent PlayerAgent, turnLimit int) *GameResult {
	if g.Mode != game.RunModeGoldfish {
		panic("RunGoldfish requires a goldfish game")
	}
	if turnLimit < 1 {
		panic("goldfish turn limit must be positive")
	}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = agent
	result := &GameResult{}
	e.drawOpeningHands(g)
	result.OpeningHand = append([]id.ID(nil), g.Players[game.Player1].Hand.All()...)
	markCurrentTurnEventStart(g)
	result.addLosses(e.applyStateBasedActions(g))
	for !g.IsGameOver() && len(result.Turns) < turnLimit {
		turnLog := e.runTurn(g, agents)
		result.addLosses(turnLog.Losses)
		result.Turns = append(result.Turns, turnLog)
		result.TurnCount = len(result.Turns)
	}
	result.TurnLimitReached = !g.IsGameOver() && len(result.Turns) == turnLimit
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
			info.Faces = nonFrontFaceNames(instance.Def)
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

// nonFrontFaceNames returns the name of each non-front printed face of a card,
// keyed by face index, so a report can name the card by the face actually
// played or cast. It returns nil for single-faced cards.
func nonFrontFaceNames(def *game.CardDef) map[game.FaceIndex]string {
	var names map[game.FaceIndex]string
	for _, index := range def.FaceIndexes() {
		if index == game.FaceFront {
			continue
		}
		face, ok := def.Face(index)
		if !ok || face.Name == "" {
			continue
		}
		if names == nil {
			names = make(map[game.FaceIndex]string)
		}
		names[index] = face.Name
	}
	return names
}
