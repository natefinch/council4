package agent

import (
	"math/rand/v2"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/rules"
)

func newSeededRandomAgent(seed uint64) RandomAgent {
	return NewRandomAgent(rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15)))
}

func TestRandomAgentChoosesFromLegalActions(t *testing.T) {
	agent := newSeededRandomAgent(7)
	legal := []action.Action{action.Pass(), action.CastSpell(1, nil, 0, nil), action.CastSpell(2, nil, 0, nil)}

	for range 50 {
		got := agent.ChooseAction(rules.PlayerObservation{}, legal)
		if !containsActionKindCardID(legal, got) {
			t.Fatalf("ChooseAction returned %v, which is not in the legal list", got)
		}
	}
}

func TestRandomAgentEmptyLegalPasses(t *testing.T) {
	agent := newSeededRandomAgent(1)
	if got := agent.ChooseAction(rules.PlayerObservation{}, nil); got.Kind != action.ActionPass {
		t.Errorf("ChooseAction on empty legal = %v, want Pass", got.Kind)
	}
}

func TestRandomAgentDeterministicForSameSeed(t *testing.T) {
	legal := []action.Action{action.Pass(), action.CastSpell(1, nil, 0, nil), action.CastSpell(2, nil, 0, nil)}

	a := newSeededRandomAgent(42)
	b := newSeededRandomAgent(42)
	for i := range 30 {
		gotA := a.ChooseAction(rules.PlayerObservation{}, legal)
		gotB := b.ChooseAction(rules.PlayerObservation{}, legal)
		if gotA.Kind != gotB.Kind {
			t.Fatalf("step %d: agents with the same seed diverged: %v vs %v", i, gotA.Kind, gotB.Kind)
		}
	}
}

func TestRandomAgentChooseChoiceSelectionValid(t *testing.T) {
	agent := newSeededRandomAgent(3)
	request := game.ChoiceRequest{
		Kind:       game.ChoiceMay,
		Options:    []game.ChoiceOption{{Index: 0}, {Index: 1}, {Index: 2}, {Index: 3}},
		MinChoices: 1,
		MaxChoices: 2,
	}
	for range 50 {
		got := agent.ChooseChoice(rules.PlayerObservation{}, request)
		if len(got) < request.MinChoices || len(got) > request.MaxChoices {
			t.Fatalf("selection %v out of bounds [%d,%d]", got, request.MinChoices, request.MaxChoices)
		}
		assertDistinctValidIndices(t, request, got)
	}
}

func TestRandomAgentChooseChoiceOrderIsPermutation(t *testing.T) {
	agent := newSeededRandomAgent(9)
	request := game.ChoiceRequest{
		Kind:       game.ChoiceOrder,
		Options:    []game.ChoiceOption{{Index: 0}, {Index: 1}, {Index: 2}},
		MinChoices: 3,
		MaxChoices: 3,
	}
	for range 50 {
		got := agent.ChooseChoice(rules.PlayerObservation{}, request)
		if len(got) != 3 {
			t.Fatalf("order selection %v, want a permutation of 3 options", got)
		}
		assertDistinctValidIndices(t, request, got)
	}
}

func TestRandomAgentChooseChoiceDamageAllocationCoversAllOptions(t *testing.T) {
	agent := newSeededRandomAgent(11)
	request := game.ChoiceRequest{
		Kind:       game.ChoiceDamageAllocation,
		Options:    []game.ChoiceOption{{Index: 0}, {Index: 1}},
		MinChoices: 5,
		MaxChoices: 5,
	}
	for range 50 {
		got := agent.ChooseChoice(rules.PlayerObservation{}, request)
		if len(got) != 5 {
			t.Fatalf("allocation %v, want 5 total units", got)
		}
		counts := map[int]int{}
		for _, index := range got {
			counts[index]++
		}
		if counts[0] < 1 || counts[1] < 1 {
			t.Errorf("allocation %v does not give every option at least one", got)
		}
	}
}

func TestRandomAgentPlaysFullGameDeterministically(t *testing.T) {
	run := func() *rules.GameResult {
		engine := rules.NewEngine(rand.New(rand.NewPCG(5, 6)))
		var configs [game.NumPlayers]game.PlayerConfig
		for i := range configs {
			for range 8 {
				configs[i].Deck = append(configs[i].Deck, forest())
			}
			for range 4 {
				configs[i].Deck = append(configs[i].Deck, scrySpell())
			}
		}
		g := engine.NewGame(configs)
		agents := [game.NumPlayers]rules.PlayerAgent{}
		for seat := range agents {
			agents[seat] = newSeededRandomAgent(uint64(seat) + 1)
		}
		return engine.RunGame(g, agents)
	}

	first := run()
	second := run()

	if !first.HasWinner && len(first.Losses) == 0 {
		t.Fatal("random-vs-random game did not resolve")
	}
	if first.TurnCount != second.TurnCount || first.HasWinner != second.HasWinner || first.Winner != second.Winner {
		t.Errorf("random game not deterministic: (%d turns, winner %v/%v) vs (%d turns, winner %v/%v)",
			first.TurnCount, first.HasWinner, first.Winner,
			second.TurnCount, second.HasWinner, second.Winner)
	}
}

func assertDistinctValidIndices(t *testing.T, request game.ChoiceRequest, selected []int) {
	t.Helper()
	valid := map[int]bool{}
	for i := range request.Options {
		valid[request.Options[i].Index] = true
	}
	seen := map[int]bool{}
	for _, index := range selected {
		if !valid[index] {
			t.Errorf("selection %v contains invalid option index %d", selected, index)
		}
		if seen[index] {
			t.Errorf("selection %v repeats option index %d", selected, index)
		}
		seen[index] = true
	}
}

func containsActionKindCardID(legal []action.Action, target action.Action) bool {
	for _, act := range legal {
		if act.Kind != target.Kind {
			continue
		}
		if act.Kind != action.ActionCastSpell {
			return true
		}
		a, _ := act.CastSpellPayload()
		b, _ := target.CastSpellPayload()
		if a.CardID == b.CardID {
			return true
		}
	}
	return false
}
