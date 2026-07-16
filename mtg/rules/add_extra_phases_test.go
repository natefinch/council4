package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
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

func TestAddExtraPhasesCountRunsExactlyTwoCombatsWithoutMainPhases(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatCreaturePermanent(g, game.Player1, game.Vigilance)

	addEffectSpellToStack(g, game.Player1, game.AddExtraPhases{CombatCount: 2}, nil)
	engine.resolveTopOfStack(g, &TurnLog{})

	log := TurnLog{}
	engine.runExtraPhases(g, allFirstLegalAgents(), &log)

	if len(log.Phases) != 2 {
		t.Fatalf("phases = %#v, want exactly two", log.Phases)
	}
	for i, phase := range log.Phases {
		if phase.Phase != game.PhaseCombat {
			t.Fatalf("phase[%d] = %v, want combat", i, phase.Phase)
		}
	}
	wantSteps := []game.Step{
		game.StepBeginningOfCombat,
		game.StepDeclareAttackers,
		game.StepDeclareBlockers,
		game.StepCombatDamage,
		game.StepEndOfCombat,
		game.StepBeginningOfCombat,
		game.StepDeclareAttackers,
		game.StepDeclareBlockers,
		game.StepCombatDamage,
		game.StepEndOfCombat,
	}
	if len(log.Steps) != len(wantSteps) {
		t.Fatalf("steps = %#v, want %#v", log.Steps, wantSteps)
	}
	for i, want := range wantSteps {
		if log.Steps[i].Step != want {
			t.Fatalf("step[%d] = %v, want %v", i, log.Steps[i].Step, want)
		}
	}
}

func TestAddExtraPhasesMultipleResolutionsAndExistingSequence(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ExtraPhases = []game.Phase{game.PhaseCombat, game.PhasePostcombatMain}

	for range 2 {
		addEffectSpellToStack(g, game.Player1, game.AddExtraPhases{CombatCount: 2}, nil)
		engine.resolveTopOfStack(g, &TurnLog{})
	}

	want := []game.Phase{
		game.PhaseCombat,
		game.PhaseCombat,
		game.PhaseCombat,
		game.PhaseCombat,
		game.PhaseCombat,
		game.PhasePostcombatMain,
	}
	if len(g.Turn.ExtraPhases) != len(want) {
		t.Fatalf("queued phases = %#v, want %#v", g.Turn.ExtraPhases, want)
	}
	for i := range want {
		if g.Turn.ExtraPhases[i] != want[i] {
			t.Fatalf("queued phase[%d] = %v, want %v", i, g.Turn.ExtraPhases[i], want[i])
		}
	}
}

// TestFirstCombatPhaseOfTurnConditionGate proves the FirstCombatPhaseOfTurn
// condition (Raiyuu, Storm's Edge) is satisfied during the turn's first combat
// phase and fails during a later (additional) combat phase, so the gate cannot
// re-trigger itself into an infinite chain of extra combat phases.
func TestFirstCombatPhaseOfTurnConditionGate(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	ctx := conditionContext{controller: game.Player1}
	cond := opt.Val(game.Condition{FirstCombatPhaseOfTurn: true})

	g.Turn.CombatPhasesThisTurn = 1
	if !conditionSatisfied(g, ctx, cond) {
		t.Fatal("condition must be satisfied during the first combat phase")
	}
	g.Turn.CombatPhasesThisTurn = 2
	if conditionSatisfied(g, ctx, cond) {
		t.Fatal("condition must fail during an additional combat phase")
	}
}

// TestFirstCombatPhaseGatedExtraCombatRunsOnce proves that an additional combat
// phase gated by FirstCombatPhaseOfTurn (Raiyuu) untaps the attacker, lets it
// attack again in the extra combat phase, and does not chain further: the gate
// is false during the second combat phase so no third phase is queued.
func TestFirstCombatPhaseGatedExtraCombatRunsOnce(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatCreaturePermanentWithPower(g, game.Player1, 3)

	// Simulate that the turn's first combat phase has begun.
	g.Turn.CombatPhasesThisTurn = 1
	g.Turn.Phase = game.PhaseCombat

	gate := opt.Val(game.EffectCondition{
		Condition: opt.Val(game.Condition{FirstCombatPhaseOfTurn: true}),
	})
	if !conditionSatisfied(g, conditionContext{controller: game.Player1}, gate.Val.Condition) {
		t.Fatal("gate must pass during first combat phase before queueing")
	}
	g.Turn.ExtraPhases = append(g.Turn.ExtraPhases, game.PhaseCombat)

	startLife := g.Players[game.Player2].Life
	log := TurnLog{}
	engine.runExtraPhases(g, allFirstLegalAgents(), &log)

	if g.Players[game.Player2].Life >= startLife {
		t.Fatalf("defending player life = %d, want less than %d (extra combat phase did not run)",
			g.Players[game.Player2].Life, startLife)
	}
	if g.Turn.CombatPhasesThisTurn != 2 {
		t.Fatalf("combat phases this turn = %d, want 2", g.Turn.CombatPhasesThisTurn)
	}
	// The gate must now be false, so a Raiyuu-style trigger would not queue more.
	if conditionSatisfied(g, conditionContext{controller: game.Player1}, gate.Val.Condition) {
		t.Fatal("gate must fail after the additional combat phase (would chain forever)")
	}
	if len(g.Turn.ExtraPhases) != 0 {
		t.Fatalf("extra phases not drained: %#v", g.Turn.ExtraPhases)
	}
}
