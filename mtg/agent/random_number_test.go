package agent

import (
	"math"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/rules"
	"github.com/natefinch/council4/opt"
)

func numberRequest(minimum int, maximum opt.V[int], defaults ...int) game.ChoiceRequest {
	return game.ChoiceRequest{
		Kind:        game.ChoiceNumber,
		NumberRange: opt.Val(game.NumberRange{Minimum: minimum, Maximum: maximum}),
		MinChoices:  1, MaxChoices: 1, DefaultSelection: defaults,
	}
}

func TestRandomAgentBoundedNumberIsUniformAndReproducible(t *testing.T) {
	request := numberRequest(0, opt.Val(4), 0)
	first, second := newSeededRandomAgent(11), newSeededRandomAgent(11)
	seen := map[int]int{}
	var sequence []int
	for range 500 {
		got := first.ChooseChoice(rules.PlayerObservation{}, request)
		if len(got) != 1 || got[0] < 0 || got[0] > 4 {
			t.Fatalf("bounded 0..4 answer = %v, want one integer in range", got)
		}
		if again := second.ChooseChoice(rules.PlayerObservation{}, request); !slices.Equal(got, again) {
			t.Fatalf("same seed answered %v then %v", got, again)
		}
		seen[got[0]]++
		sequence = append(sequence, got[0])
	}
	for value := range 5 {
		if seen[value] == 0 {
			t.Fatalf("value %d never chosen in 500 draws: %v", value, seen)
		}
	}
	other := newSeededRandomAgent(12)
	var otherSequence []int
	for range 500 {
		otherSequence = append(otherSequence, other.ChooseChoice(rules.PlayerObservation{}, request)[0])
	}
	if slices.Equal(sequence, otherSequence) {
		t.Fatal("different seeds produced the same 500-draw sequence")
	}
}

func TestRandomAgentNumberExtremeBoundsDoNotOverflow(t *testing.T) {
	for _, tc := range []struct {
		name     string
		min, max int
	}{
		{"single value", 7, 7},
		{"negative", -3, -1},
		{"full int range", math.MinInt, math.MaxInt},
		{"above half range", -1, math.MaxInt},
		{"top edge", math.MaxInt - 1, math.MaxInt},
		{"bottom edge", math.MinInt, math.MinInt + 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			agent := newSeededRandomAgent(5)
			for range 200 {
				got := agent.ChooseChoice(rules.PlayerObservation{}, numberRequest(tc.min, opt.Val(tc.max)))
				if len(got) != 1 || got[0] < tc.min || got[0] > tc.max {
					t.Fatalf("answer %v outside [%d, %d]", got, tc.min, tc.max)
				}
			}
		})
	}
}

func TestRandomAgentUnboundedNumberAnswersProposedValue(t *testing.T) {
	agent := newSeededRandomAgent(3)
	for _, tc := range []struct {
		name    string
		request game.ChoiceRequest
		want    []int
	}{
		{"in-range default", numberRequest(0, opt.V[int]{}, 6), []int{6}},
		{"default below minimum", numberRequest(1, opt.V[int]{}, 0), []int{1}},
		{"no default", numberRequest(2, opt.V[int]{}), []int{2}},
		{"huge default", numberRequest(0, opt.V[int]{}, math.MaxInt), []int{math.MaxInt}},
	} {
		if got := agent.ChooseChoice(rules.PlayerObservation{}, tc.request); !slices.Equal(got, tc.want) {
			t.Errorf("%s: answer = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestRandomAgentMalformedNumberRequestGetsNoAnswer(t *testing.T) {
	agent := newSeededRandomAgent(3)
	noRange := numberRequest(0, opt.Val(4))
	noRange.NumberRange = opt.V[game.NumberRange]{}
	withOptions := numberRequest(0, opt.Val(4))
	withOptions.Options = []game.ChoiceOption{{Index: 0}}
	for name, request := range map[string]game.ChoiceRequest{
		"no range":        noRange,
		"options present": withOptions,
		"inverted range":  numberRequest(5, opt.Val(4)),
	} {
		if got := agent.ChooseChoice(rules.PlayerObservation{}, request); got != nil {
			t.Errorf("%s: answer = %v, want none", name, got)
		}
	}
}

// randomChoicePolicy passes priority so the scenario stays focused, answers
// every engine choice with RandomAgent, and records numeric requests.
type randomChoicePolicy struct {
	RandomAgent
	numbers *[]game.ChoiceRequest
	answers *[]int
}

func (randomChoicePolicy) ChooseAction(rules.PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (p randomChoicePolicy) ChooseChoice(obs rules.PlayerObservation, request game.ChoiceRequest) []int {
	selected := p.RandomAgent.ChooseChoice(obs, request)
	if request.Kind == game.ChoiceNumber {
		*p.numbers = append(*p.numbers, request)
		*p.answers = append(*p.answers, selected...)
	}
	return selected
}

// TestRandomAgentAnnouncesXForRegisteredWalkingBallista activates registered
// Conduit of Worlds targeting registered Walking Ballista ({X}{X}) in the
// graveyard. Casting it during resolution announces X through the Engine's
// unbounded ChoiceNumber request (CR 601.2b). RandomAgent answers validly for
// every seed, so no invalid choice or fallback is counted, and the answers are
// reproducible for a fixed seed.
func TestRandomAgentAnnouncesXForRegisteredWalkingBallista(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	ballista, conduit := registry.Lookup("Walking Ballista"), registry.Lookup("Conduit of Worlds")
	if ballista == nil || conduit == nil {
		t.Fatal("Walking Ballista or Conduit of Worlds is not registered")
	}
	type outcome struct {
		numbers int
		answers []int
	}
	run := func(seed uint64) outcome {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		g.Turn.ActivePlayer, g.Turn.PriorityPlayer = game.Player1, game.Player1
		g.Turn.Phase, g.Turn.Step = game.PhasePrecombatMain, game.StepNone
		source := addObservedPermanent(g, game.Player1, conduit)
		cardID := g.IDGen.Next()
		g.CardInstances[cardID] = &game.CardInstance{ID: cardID, Def: ballista, Owner: game.Player1}
		g.Players[game.Player1].Graveyard.Add(cardID)
		g.Players[game.Player1].ManaPool.Add(mana.C, 6)
		engine := rules.NewEngineWithPrograms(nil, registry.Programs())
		sim := engine.Simulator()
		var activation action.Action
		found := false
		for _, act := range sim.LegalActions(g, game.Player1) {
			if activate, ok := act.ActivateAbilityPayload(); ok && activate.SourceID == source.ObjectID {
				activation, found = act, true
			}
		}
		if !found {
			t.Fatal("Conduit of Worlds activation is not legal")
		}
		var numbers []game.ChoiceRequest
		var answers []int
		var policies [game.NumPlayers]rules.PlayerAgent
		for seat := range policies {
			policies[seat] = randomChoicePolicy{RandomAgent: newSeededRandomAgent(seed + uint64(seat)), numbers: &numbers, answers: &answers}
		}
		applied, ok := sim.Apply(g, game.Player1, activation, policies)
		if !ok {
			t.Fatalf("seed %d: Conduit activation rejected", seed)
		}
		resolved := sim.ResolvePriority(applied, policies)
		for _, request := range numbers {
			if request.Player != game.Player1 || request.NumberRange.Val.Maximum.Exists || request.NumberRange.Val.Minimum != 0 {
				t.Fatalf("seed %d: numeric request = %+v, want Player1's unbounded X from 0", seed, request)
			}
		}
		// A resolution cast proposes X = 0, the value RandomAgent answers for
		// an unbounded range, so Ballista enters as a 0/0 and dies (CR 704.5f).
		if !slices.Equal(answers, make([]int, len(numbers))) {
			t.Fatalf("seed %d: X answers = %v, want the proposed 0", seed, answers)
		}
		if len(numbers) > 0 && (!resolved.Players[game.Player1].Graveyard.Contains(cardID) || resolved.Stack.Size() != 0) {
			t.Fatalf("seed %d: X=0 Walking Ballista did not resolve and die", seed)
		}
		if resolved.Execution.InvalidChoices != 0 || resolved.Execution.ChoiceFallbacks != 0 {
			t.Fatalf("seed %d: invalid/fallback = %d/%d, want 0/0", seed,
				resolved.Execution.InvalidChoices, resolved.Execution.ChoiceFallbacks)
		}
		return outcome{numbers: len(numbers), answers: answers}
	}
	announced := 0
	for seed := range uint64(16) {
		first, second := run(seed), run(seed)
		if first.numbers != second.numbers || !slices.Equal(first.answers, second.answers) {
			t.Fatalf("seed %d not reproducible: %+v vs %+v", seed, first, second)
		}
		announced += first.numbers
	}
	if announced == 0 {
		t.Fatal("no seed cast Walking Ballista through the X announcement")
	}
}
