package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

func TestCheckStateBasedActionsEliminatesPlayers(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*game.Player, *game.Game)
	}{
		{
			name: "zero life",
			setup: func(p *game.Player, g *game.Game) {
				p.Life = 0
			},
		},
		{
			name: "lethal poison",
			setup: func(p *game.Player, g *game.Game) {
				p.PoisonCounters = 10
			},
		},
		{
			name: "lethal commander damage",
			setup: func(p *game.Player, g *game.Game) {
				p.CommanderDamage[id.ID(99)] = 21
			},
		},
		{
			name: "failed draw",
			setup: func(p *game.Player, g *game.Game) {
				g.FailedDraws[p.ID] = true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			player := g.Players[game.Player1]
			tt.setup(player, g)

			if !engine.checkStateBasedActions(g) {
				t.Fatal("checkStateBasedActions() = false, want true")
			}
			if !player.Eliminated {
				t.Fatal("player was not marked eliminated")
			}
			if !g.TurnOrder.IsEliminated(player.ID) {
				t.Fatal("turn order was not marked eliminated")
			}
			if g.FailedDraws[player.ID] {
				t.Fatal("failed draw flag was not cleared")
			}
		})
	}
}

func TestCheckStateBasedActionsAlreadyEliminatedIsStable(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	if !engine.eliminatePlayer(g, game.Player1) {
		t.Fatal("first eliminatePlayer() = false, want true")
	}
	if engine.checkStateBasedActions(g) {
		t.Fatal("checkStateBasedActions() = true for stable eliminated player, want false")
	}
}

func TestCheckStateBasedActionsClearsFailedDrawForAlreadyEliminatedPlayer(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	if !engine.eliminatePlayer(g, game.Player1) {
		t.Fatal("eliminatePlayer() = false, want true")
	}
	g.FailedDraws[game.Player1] = true

	if engine.checkStateBasedActions(g) {
		t.Fatal("checkStateBasedActions() = true for already eliminated player, want false")
	}
	if g.FailedDraws[game.Player1] {
		t.Fatal("failed draw flag was not cleared")
	}
}
