package sim

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/agent"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/rules"
)

// panicAgent panics on its first priority decision, standing in for an engine
// bug, unsupported card, or illegal action that aborts a game.
type panicAgent struct{}

func (panicAgent) ChooseAction(_ rules.PlayerObservation, _ []action.Action) action.Action {
	panic("boom: simulated game failure")
}

func TestFailureIsCapturedAndBatchContinues(t *testing.T) {
	const failIndex = 2
	cfg := smokeConfig(5, 1)
	failSeed := GameSeed(cfg.Seed, failIndex)
	cfg.NewAgents = func(gameSeed uint64) [game.NumPlayers]rules.PlayerAgent {
		var agents [game.NumPlayers]rules.PlayerAgent
		for i := range agents {
			agents[i] = agentForSeat()
		}
		if gameSeed == failSeed {
			agents[0] = panicAgent{}
		}
		return agents
	}

	result := Run(cfg)

	if result.FailureCount() != 1 {
		t.Fatalf("FailureCount = %d, want 1", result.FailureCount())
	}
	failure := result.Failures[0]
	if failure.Index != failIndex {
		t.Errorf("failure Index = %d, want %d", failure.Index, failIndex)
	}
	if failure.Seed != failSeed {
		t.Errorf("failure Seed = %d, want %d", failure.Seed, failSeed)
	}
	if !strings.Contains(failure.Reason, "boom") {
		t.Errorf("failure Reason = %q, want it to contain the panic message", failure.Reason)
	}
	if failure.Stack == "" {
		t.Error("failure Stack is empty, want a captured stack trace")
	}
	// The failed game holds the zero result; every other game completed.
	for i := range result.Games {
		completed := result.Games[i].TurnCount > 0 || result.Games[i].HasWinner
		if i == failIndex && completed {
			t.Errorf("failed game %d should hold the zero result", i)
		}
		if i != failIndex && !completed {
			t.Errorf("game %d did not complete despite another game failing", i)
		}
	}
}

func TestFailuresAreOrderedAndDeterministic(t *testing.T) {
	// Two failing games, declared out of index order, must surface in index order.
	cfg := smokeConfig(6, 9)
	failA, failB := 1, 4
	seedA, seedB := GameSeed(cfg.Seed, failA), GameSeed(cfg.Seed, failB)
	cfg.NewAgents = func(gameSeed uint64) [game.NumPlayers]rules.PlayerAgent {
		var agents [game.NumPlayers]rules.PlayerAgent
		for i := range agents {
			agents[i] = agentForSeat()
		}
		if gameSeed == seedA || gameSeed == seedB {
			agents[0] = panicAgent{}
		}
		return agents
	}

	for _, workers := range []int{1, 4} {
		cfg.Workers = workers
		result := Run(cfg)
		if result.FailureCount() != 2 {
			t.Fatalf("workers=%d: FailureCount = %d, want 2", workers, result.FailureCount())
		}
		if result.Failures[0].Index != failA || result.Failures[1].Index != failB {
			t.Errorf("workers=%d: failure indices = %d,%d, want %d,%d",
				workers, result.Failures[0].Index, result.Failures[1].Index, failA, failB)
		}
	}
}

// agentForSeat returns the deterministic agent the smoke configs use.
func agentForSeat() rules.PlayerAgent {
	return agent.FirstLegal{}
}

// unsupportedAgent panics with a rules.UnsupportedError on its first priority
// decision, standing in for a game that resolves a mechanic the engine does not
// yet support (the runtime dispatch path raises this typed value).
type unsupportedAgent struct{}

func (unsupportedAgent) ChooseAction(_ rules.PlayerObservation, _ []action.Action) action.Action {
	panic(rules.UnsupportedError{
		Kind:   game.PrimitiveUnknown,
		Reason: "primitive kind 0 has no registered handler",
	})
}

func TestUnsupportedFailureIsFlagged(t *testing.T) {
	const failIndex = 1
	cfg := smokeConfig(3, 7)
	failSeed := GameSeed(cfg.Seed, failIndex)
	cfg.NewAgents = func(gameSeed uint64) [game.NumPlayers]rules.PlayerAgent {
		var agents [game.NumPlayers]rules.PlayerAgent
		for i := range agents {
			agents[i] = agentForSeat()
		}
		if gameSeed == failSeed {
			agents[0] = unsupportedAgent{}
		}
		return agents
	}

	result := Run(cfg)

	if result.FailureCount() != 1 {
		t.Fatalf("FailureCount = %d, want 1", result.FailureCount())
	}
	failure := result.Failures[0]
	if failure.Index != failIndex {
		t.Errorf("failure Index = %d, want %d", failure.Index, failIndex)
	}
	if !failure.Unsupported {
		t.Error("failure Unsupported = false, want true for an UnsupportedError panic")
	}
	if !strings.Contains(failure.Reason, "unsupported mechanic") {
		t.Errorf("failure Reason = %q, want it to contain the error message", failure.Reason)
	}
}

func TestNonUnsupportedFailureIsNotFlagged(t *testing.T) {
	cfg := smokeConfig(2, 11)
	failSeed := GameSeed(cfg.Seed, 0)
	cfg.NewAgents = func(gameSeed uint64) [game.NumPlayers]rules.PlayerAgent {
		var agents [game.NumPlayers]rules.PlayerAgent
		for i := range agents {
			agents[i] = agentForSeat()
		}
		if gameSeed == failSeed {
			agents[0] = panicAgent{}
		}
		return agents
	}

	result := Run(cfg)

	if result.FailureCount() != 1 {
		t.Fatalf("FailureCount = %d, want 1", result.FailureCount())
	}
	if result.Failures[0].Unsupported {
		t.Error("failure Unsupported = true, want false for a plain panic")
	}
}
