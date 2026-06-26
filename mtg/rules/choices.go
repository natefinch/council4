package rules

import (
	"fmt"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// cardChoiceInfo builds the public ChoiceCardInfo for a card instance, for use
// in a ChoiceOption.Card or ChoiceRequest.Subject. It is unset when the card is
// unknown.
func cardChoiceInfo(g *game.Game, cardID id.ID) opt.V[game.ChoiceCardInfo] {
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return opt.V[game.ChoiceCardInfo]{}
	}
	return opt.Val(game.ChoiceCardInfo{
		CardID:    cardID,
		Name:      card.Def.Name,
		Types:     append([]types.Card(nil), card.Def.Types...),
		ManaValue: card.Def.ManaValue(),
		Colors:    append([]color.Color(nil), card.Def.Colors...),
	})
}

// permanentChoiceInfo builds the public ChoiceCardInfo for a permanent.
func permanentChoiceInfo(g *game.Game, permanent *game.Permanent) opt.V[game.ChoiceCardInfo] {
	def, ok := permanentCardDef(g, permanent)
	if !ok {
		return opt.V[game.ChoiceCardInfo]{}
	}
	return opt.Val(game.ChoiceCardInfo{
		CardID:    permanent.CardInstanceID,
		Name:      permanentEffectiveName(g, permanent),
		Types:     append([]types.Card(nil), def.Types...),
		ManaValue: def.ManaValue(),
		Colors:    append([]color.Color(nil), def.Colors...),
	})
}

// ChoiceAgent is implemented by agents that can answer engine-mediated
// decisions beyond normal priority actions. Agents that do not implement it use
// deterministic fallback choices.
type ChoiceAgent interface {
	ChooseChoice(obs PlayerObservation, request game.ChoiceRequest) []int
}

func (e *Engine) chooseChoice(g *game.Game, agents [game.NumPlayers]PlayerAgent, request game.ChoiceRequest, log *TurnLog) []int {
	selected, _ := e.chooseChoiceWithFallback(g, agents, request, log)
	return selected
}

// chooseChoiceWithFallback is chooseChoice but also reports whether the engine had
// to use a deterministic fallback (no agent, no ChoiceAgent, or an invalid agent
// answer), so callers that record their own decision can mirror that flag.
func (e *Engine) chooseChoiceWithFallback(g *game.Game, agents [game.NumPlayers]PlayerAgent, request game.ChoiceRequest, log *TurnLog) ([]int, bool) {
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
	return selected, usedFallback
}

func (*Engine) agentChoice(g *game.Game, agents [game.NumPlayers]PlayerAgent, request game.ChoiceRequest) ([]int, bool) {
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
	case game.ChoicePayment, game.ChoiceScry, game.ChoiceSurveil, game.ChoiceZoneSelection, game.ChoiceSearch, game.ChoiceModal, game.ChoiceResolution, game.ChoiceProliferate, game.ChoicePlayer, game.ChoiceManifest, game.ChoiceDig, game.ChoiceVote, game.ChoicePileSeparate, game.ChoicePileChoose:
		if len(request.Options) == 0 || request.MaxChoices == 0 {
			return nil
		}
		count := min(request.MaxChoices, len(request.Options))
		selected := make([]int, 0, count)
		for i := range count {
			selected = append(selected, request.Options[i].Index)
		}
		return selected
	case game.ChoiceDamageAllocation:
		return defaultDividedAllocation(request.MaxChoices, len(request.Options))
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
	if request.MaxTotalManaValue.Exists && !choiceTotalManaValueValid(request, selected) {
		return false
	}
	switch request.Kind {
	case game.ChoiceOrder:
		return orderSelectionValid(request, selected)
	case game.ChoiceDamageAllocation:
		return damageAllocationSelectionValid(request, selected)
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

// damageAllocationSelectionValid accepts a divided-damage allocation expressed
// as a multiset of option indices: the total count must equal the requested
// total (MinChoices == MaxChoices), every index must be a valid option, and
// every option must receive at least one (CR 601.2d, no zero allocations).
func damageAllocationSelectionValid(request game.ChoiceRequest, selected []int) bool {
	counts := make(map[int]int, len(request.Options))
	for _, index := range selected {
		if !choiceOptionExists(request, index) {
			return false
		}
		counts[index]++
	}
	for _, option := range request.Options {
		if counts[option.Index] < 1 {
			return false
		}
	}
	return true
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

// choiceTotalManaValueValid reports whether the combined mana value of the
// selected options stays within request.MaxTotalManaValue. Options without card
// info contribute zero. Indices that name no option are rejected so an invalid
// index can never bypass the cap by contributing nothing.
func choiceTotalManaValueValid(request game.ChoiceRequest, selected []int) bool {
	total := 0
	for _, index := range selected {
		option, ok := choiceOption(request, index)
		if !ok {
			return false
		}
		if option.Card.Exists {
			total += option.Card.Val.ManaValue
		}
	}
	return total <= request.MaxTotalManaValue.Val
}

func choiceOption(request game.ChoiceRequest, index int) (game.ChoiceOption, bool) {
	for _, option := range request.Options {
		if option.Index == index {
			return option, true
		}
	}
	return game.ChoiceOption{}, false
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
