package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// voteSpellInstructions builds the canonical voting sequence: a Vote primitive
// publishing the signed margin, an option-0 arm gated on a positive margin, and
// a tie-inclusive option-1 arm gated on a non-positive margin. The two arms
// adjust the controller's life so a test can read which arm executed.
func voteSpellInstructions() []game.Instruction {
	const resultKey = game.ResultKey("vote-result")
	return []game.Instruction{
		{
			Primitive:     game.Vote{Options: []string{"a", "b"}},
			PublishResult: resultKey,
		},
		{
			Primitive: game.GainLife{Amount: game.Fixed(5), Player: game.ControllerReference()},
			ResultGate: opt.Val(game.InstructionResultGate{
				Key:         resultKey,
				AmountRange: opt.Val(game.IntRange{Min: 1, Max: game.NumPlayers}),
			}),
		},
		{
			Primitive: game.LoseLife{Amount: game.Fixed(5), Player: game.ControllerReference()},
			ResultGate: opt.Val(game.InstructionResultGate{
				Key:         resultKey,
				AmountRange: opt.Val(game.IntRange{Min: -game.NumPlayers, Max: 0}),
			}),
		},
	}
}

// voteAgents scripts every player's single vote: votes[i] is the option index
// player i casts.
func voteAgents(votes [game.NumPlayers]int) [game.NumPlayers]PlayerAgent {
	var agents [game.NumPlayers]PlayerAgent
	for i := range votes {
		agents[i] = &choiceOnlyAgent{choices: [][]int{{votes[i]}}}
	}
	return agents
}

// TestVoteOptionZeroMajorityRunsFirstArm verifies the voting interaction tallies
// each player's vote and publishes the signed margin (Options[0] minus
// Options[1]). With every player voting the first option the margin is +4, so
// the option-0 arm (gated on a positive margin) runs and the controller gains
// life while the option-1 arm stays gated off.
func TestVoteOptionZeroMajorityRunsFirstArm(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	before := g.Players[game.Player1].Life
	addInstructionSpellToStack(g, voteSpellInstructions())
	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("Stack.Peek() = false, want the pushed spell")
	}

	engine.resolveTopOfStackWithChoices(g, voteAgents([game.NumPlayers]int{0, 0, 0, 0}), &TurnLog{})

	if got := obj.ResolvedAmounts["vote-result"]; got != game.NumPlayers {
		t.Fatalf("vote margin = %d, want %d (all four votes for option 0)", got, game.NumPlayers)
	}
	if got := g.Players[game.Player1].Life; got != before+5 {
		t.Fatalf("controller life = %d, want %d (option-0 arm gains life)", got, before+5)
	}
}

// TestVoteOptionOneMajorityRunsSecondArm verifies a second-option majority: with
// every player voting the second option the margin is -4, so the tie-inclusive
// option-1 arm (gated on a non-positive margin) runs and the controller loses
// life while the option-0 arm stays gated off.
func TestVoteOptionOneMajorityRunsSecondArm(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	before := g.Players[game.Player1].Life
	addInstructionSpellToStack(g, voteSpellInstructions())
	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("Stack.Peek() = false, want the pushed spell")
	}

	engine.resolveTopOfStackWithChoices(g, voteAgents([game.NumPlayers]int{1, 1, 1, 1}), &TurnLog{})

	if got := obj.ResolvedAmounts["vote-result"]; got != -game.NumPlayers {
		t.Fatalf("vote margin = %d, want %d (all four votes for option 1)", got, -game.NumPlayers)
	}
	if got := g.Players[game.Player1].Life; got != before-5 {
		t.Fatalf("controller life = %d, want %d (option-1 arm loses life)", got, before-5)
	}
}

// TestVoteTieRunsTieInclusiveArm verifies a tied vote resolves to the
// tie-inclusive arm: two votes for each option leave a margin of 0, which the
// strict option-0 arm ([1,4]) excludes and the tie-inclusive option-1 arm
// ([-4,0]) includes, so the controller loses life ("or the vote is tied").
func TestVoteTieRunsTieInclusiveArm(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	before := g.Players[game.Player1].Life
	addInstructionSpellToStack(g, voteSpellInstructions())
	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("Stack.Peek() = false, want the pushed spell")
	}

	engine.resolveTopOfStackWithChoices(g, voteAgents([game.NumPlayers]int{0, 1, 0, 1}), &TurnLog{})

	if got := obj.ResolvedAmounts["vote-result"]; got != 0 {
		t.Fatalf("vote margin = %d, want 0 (two votes each option)", got)
	}
	if got := g.Players[game.Player1].Life; got != before-5 {
		t.Fatalf("controller life = %d, want %d (tie resolves to the tie-inclusive arm)", got, before-5)
	}
}
