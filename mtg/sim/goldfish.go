package sim

import (
	"github.com/natefinch/council4/mtg/agent"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/rules"
)

// GoldfishAgentFactory builds the agent for a reproducible goldfish run.
type GoldfishAgentFactory func(seed uint64) rules.PlayerAgent

// GoldfishConfig describes one single-player, turn-limited simulation.
type GoldfishConfig struct {
	Player   game.PlayerConfig
	Seed     uint64
	MaxTurns int
	NewAgent GoldfishAgentFactory
}

// RunGoldfish plays one deck alone and returns its structured turn log and end
// state. The same config always reproduces the same run.
func RunGoldfish(cfg GoldfishConfig) rules.GameResult {
	if cfg.MaxTurns < 1 {
		panic("goldfish max turns must be positive")
	}
	engine := rules.NewEngine(NewRand(cfg.Seed))
	g := engine.NewGoldfishGame(cfg.Player)
	playerAgent := rules.PlayerAgent(agent.Agent{Strategy: agent.GenericStrategy{}})
	if cfg.NewAgent != nil {
		playerAgent = cfg.NewAgent(cfg.Seed)
	}
	return *engine.RunGoldfish(g, playerAgent, cfg.MaxTurns)
}
