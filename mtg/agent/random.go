package agent

import (
	"math/rand/v2"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/rules"
)

// RandomAgent plays uniformly at random: it picks a uniformly random legal
// action and a valid random selection for engine-mediated choices. It is a
// comparison baseline that smarter strategies can be measured against.
//
// Each RandomAgent owns its injected *rand.Rand, so games are reproducible for a
// fixed seed and parallel simulations never share an RNG. Give each seat in each
// game its own RandomAgent with its own seeded RNG.
type RandomAgent struct {
	rng *rand.Rand
}

// Compile-time checks that a RandomAgent drives both engine decision points.
var (
	_ rules.PlayerAgent = RandomAgent{}
	_ rules.ChoiceAgent = RandomAgent{}
)

// NewRandomAgent returns a RandomAgent that draws its decisions from rng.
func NewRandomAgent(rng *rand.Rand) RandomAgent {
	return RandomAgent{rng: rng}
}

// ChooseAction implements rules.PlayerAgent by picking a uniformly random legal
// action. Choosing from the engine's legal list guarantees the action is legal.
func (a RandomAgent) ChooseAction(_ rules.PlayerObservation, legal []action.Action) action.Action {
	if len(legal) == 0 {
		return action.Pass()
	}
	return legal[a.rng.IntN(len(legal))]
}

// ChooseChoice implements rules.ChoiceAgent by returning a valid random
// selection for the request. Ordering choices get a random permutation and
// divided-damage choices get a random allocation giving each option at least
// one; other choices get a random count of distinct options within the
// requested bounds.
func (a RandomAgent) ChooseChoice(_ rules.PlayerObservation, request game.ChoiceRequest) []int {
	switch request.Kind {
	case game.ChoiceOrder:
		return a.randomPermutation(request)
	case game.ChoiceDamageAllocation, game.ChoiceCounterAllocation:
		return a.randomAllocation(request)
	case game.ChoiceManaCombination:
		return a.randomCombination(request)
	default:
		return a.randomSelection(request)
	}
}

// randomPermutation returns the option indices in a uniformly random order.
func (a RandomAgent) randomPermutation(request game.ChoiceRequest) []int {
	indices := optionIndices(request)
	a.rng.Shuffle(len(indices), func(i, j int) {
		indices[i], indices[j] = indices[j], indices[i]
	})
	return indices
}

// randomAllocation distributes MinChoices (== MaxChoices) units across the
// options as a multiset of option indices, giving every option at least one.
func (a RandomAgent) randomAllocation(request game.ChoiceRequest) []int {
	indices := optionIndices(request)
	total := request.MinChoices
	if len(indices) == 0 || total < len(indices) {
		return nil
	}
	selected := append([]int(nil), indices...)
	for range total - len(indices) {
		selected = append(selected, indices[a.rng.IntN(len(indices))])
	}
	a.rng.Shuffle(len(selected), func(i, j int) {
		selected[i], selected[j] = selected[j], selected[i]
	})
	return selected
}

// randomCombination distributes MinChoices (== MaxChoices) units freely across
// the options as a multiset of option indices, allowing any option to receive
// zero. It backs ChoiceManaCombination, whose "add N mana in any combination of
// <colors>" split places no per-color minimum.
func (a RandomAgent) randomCombination(request game.ChoiceRequest) []int {
	indices := optionIndices(request)
	total := request.MinChoices
	if len(indices) == 0 || total <= 0 {
		return nil
	}
	selected := make([]int, total)
	for i := range selected {
		selected[i] = indices[a.rng.IntN(len(indices))]
	}
	return selected
}

// randomSelection returns a random count of distinct option indices within the
// requested bounds.
func (a RandomAgent) randomSelection(request game.ChoiceRequest) []int {
	indices := optionIndices(request)
	hi := min(request.MaxChoices, len(indices))
	if hi < request.MinChoices {
		return nil
	}
	count := request.MinChoices + a.rng.IntN(hi-request.MinChoices+1)
	a.rng.Shuffle(len(indices), func(i, j int) {
		indices[i], indices[j] = indices[j], indices[i]
	})
	return indices[:count]
}

func optionIndices(request game.ChoiceRequest) []int {
	indices := make([]int, len(request.Options))
	for i := range request.Options {
		indices[i] = request.Options[i].Index
	}
	return indices
}
