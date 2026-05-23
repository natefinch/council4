package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

func TestCheckStateBasedActionsEliminatesPlayers(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*game.Player, *game.Game)
		wantReason LossReason
	}{
		{
			name:       "zero life",
			wantReason: LossReasonZeroLife,
			setup: func(p *game.Player, g *game.Game) {
				p.Life = 0
			},
		},
		{
			name:       "lethal poison",
			wantReason: LossReasonPoisonCounters,
			setup: func(p *game.Player, g *game.Game) {
				p.PoisonCounters = 10
			},
		},
		{
			name:       "lethal commander damage",
			wantReason: LossReasonCommanderDamage,
			setup: func(p *game.Player, g *game.Game) {
				p.CommanderDamage[id.ID(99)] = 21
			},
		},
		{
			name:       "failed draw",
			wantReason: LossReasonEmptyLibraryDraw,
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

			changed, losses := engine.checkStateBasedActions(g)
			if !changed {
				t.Fatal("checkStateBasedActions() = false, want true")
			}
			if len(losses) != 1 {
				t.Fatalf("losses = %d, want 1", len(losses))
			}
			if losses[0].Player != player.ID {
				t.Fatalf("loss player = %v, want %v", losses[0].Player, player.ID)
			}
			if losses[0].Reason != tt.wantReason {
				t.Fatalf("loss reason = %q, want %q", losses[0].Reason, tt.wantReason)
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
	changed, losses := engine.checkStateBasedActions(g)
	if changed {
		t.Fatal("checkStateBasedActions() = true for stable eliminated player, want false")
	}
	if len(losses) != 0 {
		t.Fatalf("losses = %d, want 0", len(losses))
	}
}

func TestCheckStateBasedActionsClearsFailedDrawForAlreadyEliminatedPlayer(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	if !engine.eliminatePlayer(g, game.Player1) {
		t.Fatal("eliminatePlayer() = false, want true")
	}
	g.FailedDraws[game.Player1] = true

	changed, losses := engine.checkStateBasedActions(g)
	if changed {
		t.Fatal("checkStateBasedActions() = true for already eliminated player, want false")
	}
	if len(losses) != 0 {
		t.Fatalf("losses = %d, want 0", len(losses))
	}
	if g.FailedDraws[game.Player1] {
		t.Fatal("failed draw flag was not cleared")
	}
}

func TestApplyStateBasedActionsReturnsLosses(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Players[game.Player1].Life = 0

	losses := engine.applyStateBasedActions(g)

	if len(losses) != 1 {
		t.Fatalf("losses = %d, want 1", len(losses))
	}
	if losses[0].Reason != LossReasonZeroLife {
		t.Fatalf("loss reason = %q, want %q", losses[0].Reason, LossReasonZeroLife)
	}
}
