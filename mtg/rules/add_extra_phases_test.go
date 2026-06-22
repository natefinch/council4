package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestAddExtraPhasesQueuesAndRunsExtraCombat resolves an AddExtraPhases effect
// and then drains the queue, proving that the queued additional combat phase
// actually runs: the active player's creature attacks the opponent during the
// extra combat phase and deals damage that would not occur without it.
func TestAddExtraPhasesQueuesAndRunsExtraCombat(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	g.Turn.Phase = game.PhasePostcombatMain
	addEffectSpellToStack(g, game.Player1, game.AddExtraPhases{Combat: true, Main: true}, nil)
	engine.resolveTopOfStack(g, &TurnLog{})

	want := []game.Phase{game.PhaseCombat, game.PhasePostcombatMain}
	if len(g.Turn.ExtraPhases) != len(want) {
		t.Fatalf("queued extra phases = %#v, want %#v", g.Turn.ExtraPhases, want)
	}
	for i, phase := range want {
		if g.Turn.ExtraPhases[i] != phase {
			t.Fatalf("queued extra phase[%d] = %v, want %v", i, g.Turn.ExtraPhases[i], phase)
		}
	}

	startLife := g.Players[game.Player2].Life
	log := TurnLog{}
	engine.runExtraPhases(g, allFirstLegalAgents(), &log)

	if len(g.Turn.ExtraPhases) != 0 {
		t.Fatalf("extra phases not drained: %#v", g.Turn.ExtraPhases)
	}
	if g.Players[game.Player2].Life >= startLife {
		t.Fatalf("defending player life = %d, want less than %d (extra combat phase did not run)",
			g.Players[game.Player2].Life, startLife)
	}
	if g.Turn.Phase != game.PhasePostcombatMain {
		t.Fatalf("final phase = %v, want PhasePostcombatMain", g.Turn.Phase)
	}
}

// TestAddExtraPhasesCombatOnlyQueuesSingleCombat proves the combat-only form
// queues exactly one extra combat phase and no extra main phase.
func TestAddExtraPhasesCombatOnlyQueuesSingleCombat(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	addEffectSpellToStack(g, game.Player1, game.AddExtraPhases{Combat: true}, nil)
	engine.resolveTopOfStack(g, &TurnLog{})

	if len(g.Turn.ExtraPhases) != 1 || g.Turn.ExtraPhases[0] != game.PhaseCombat {
		t.Fatalf("queued extra phases = %#v, want one combat phase", g.Turn.ExtraPhases)
	}
}
