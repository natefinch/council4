package rules

import (
	"fmt"

	"github.com/natefinch/council4/mtg/game"
)

// ChoiceAgent is implemented by agents that can answer engine-mediated
// decisions beyond normal priority actions. Agents that do not implement it use
// deterministic fallback choices.
type ChoiceAgent interface {
	ChooseChoice(obs PlayerObservation, request game.ChoiceRequest) []int
}

func (e *Engine) chooseChoice(g *game.Game, agents [game.NumPlayers]PlayerAgent, request game.ChoiceRequest, log *TurnLog) []int {
	if request.ID == 0 {
		request.ID = nextChoiceID(log)
	}
	selected, usedFallback := e.agentChoice(g, agents, request)
	if !choiceSelectionValid(request, selected) {
		selected = fallbackChoice(request)
		usedFallback = true
	}
	if !choiceSelectionValid(request, selected) {
		panic("invalid fallback choice")
	}
	selected = append([]int(nil), selected...)
	log.addChoice(game.ChoiceDecision{
		Request:      cloneChoiceRequest(request),
		Selected:     selected,
		UsedFallback: usedFallback,
	})
	return selected
}

func (e *Engine) agentChoice(g *game.Game, agents [game.NumPlayers]PlayerAgent, request game.ChoiceRequest) ([]int, bool) {
	agent := agentFor(agents, request.Player)
	if agent == nil {
		return fallbackChoice(request), true
	}
	choiceAgent, ok := agent.(ChoiceAgent)
	if !ok {
		return fallbackChoice(request), true
	}
	return choiceAgent.ChooseChoice(observe(g, request.Player), cloneChoiceRequest(request)), false
}

func nextChoiceID(log *TurnLog) int {
	if log == nil {
		return 0
	}
	return len(log.Choices) + 1
}

func cloneChoiceRequest(request game.ChoiceRequest) game.ChoiceRequest {
	request.Options = append([]game.ChoiceOption(nil), request.Options...)
	request.DefaultSelection = append([]int(nil), request.DefaultSelection...)
	return request
}

func fallbackChoice(request game.ChoiceRequest) []int {
	if choiceSelectionValid(request, request.DefaultSelection) {
		return append([]int(nil), request.DefaultSelection...)
	}
	switch request.Kind {
	case game.ChoiceOrder:
		selected := make([]int, 0, len(request.Options))
		for _, option := range request.Options {
			selected = append(selected, option.Index)
		}
		return selected
	case game.ChoiceMay, game.ChoiceTarget:
		if len(request.Options) == 0 || request.MaxChoices == 0 {
			return nil
		}
		return []int{request.Options[0].Index}
	case game.ChoicePayment, game.ChoiceScry, game.ChoiceSurveil, game.ChoiceZoneSelection, game.ChoiceSearch, game.ChoiceModal:
		if len(request.Options) == 0 || request.MaxChoices == 0 {
			return nil
		}
		count := min(request.MaxChoices, len(request.Options))
		selected := make([]int, 0, count)
		for i := 0; i < count; i++ {
			selected = append(selected, request.Options[i].Index)
		}
		return selected
	default:
		return nil
	}
}

func choiceSelectionValid(request game.ChoiceRequest, selected []int) bool {
	if request.MinChoices < 0 || request.MaxChoices < request.MinChoices {
		return false
	}
	if len(selected) < request.MinChoices || len(selected) > request.MaxChoices {
		return false
	}
	switch request.Kind {
	case game.ChoiceOrder:
		return orderSelectionValid(request, selected)
	default:
		seen := make(map[int]bool, len(selected))
		for _, index := range selected {
			if seen[index] || !choiceOptionExists(request, index) {
				return false
			}
			seen[index] = true
		}
		return true
	}
}

func orderSelectionValid(request game.ChoiceRequest, selected []int) bool {
	if len(selected) != len(request.Options) {
		return false
	}
	seen := make(map[int]bool, len(selected))
	for _, index := range selected {
		if seen[index] || !choiceOptionExists(request, index) {
			return false
		}
		seen[index] = true
	}
	return true
}

func choiceOptionExists(request game.ChoiceRequest, index int) bool {
	for _, option := range request.Options {
		if option.Index == index {
			return true
		}
	}
	return false
}

func mayChoiceRequest(player game.PlayerID, prompt string) game.ChoiceRequest {
	return game.ChoiceRequest{
		Kind:       game.ChoiceMay,
		Player:     player,
		Prompt:     prompt,
		Options:    []game.ChoiceOption{{Index: 0, Label: "No"}, {Index: 1, Label: "Yes"}},
		MinChoices: 1,
		MaxChoices: 1,
		// Preserve pre-choice engine behavior for optional triggered abilities:
		// nil agents apply the trigger's effects.
		DefaultSelection: []int{1},
	}
}

func targetChoiceRequest(player game.PlayerID, prompt string, choices [][]game.Target) game.ChoiceRequest {
	options := make([]game.ChoiceOption, 0, len(choices))
	for i, targets := range choices {
		options = append(options, game.ChoiceOption{
			Index: i,
			Label: fmt.Sprintf("%v", targets),
		})
	}
	return game.ChoiceRequest{
		Kind:       game.ChoiceTarget,
		Player:     player,
		Prompt:     prompt,
		Options:    options,
		MinChoices: 1,
		MaxChoices: 1,
	}
}

func orderChoiceRequest(player game.PlayerID, prompt string, options []game.ChoiceOption) game.ChoiceRequest {
	return game.ChoiceRequest{
		Kind:       game.ChoiceOrder,
		Player:     player,
		Prompt:     prompt,
		Options:    append([]game.ChoiceOption(nil), options...),
		MinChoices: len(options),
		MaxChoices: len(options),
	}
}
