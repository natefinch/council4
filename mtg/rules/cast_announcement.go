package rules

import (
	"fmt"
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func (e *Engine) chooseCastNumbers(g *game.Game, player game.PlayerID, def *game.CardDef, cast action.CastSpellAction, free bool, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (action.CastSpellAction, bool) {
	private := cast.SourceZone == zone.Hand || cast.SourceZone == zone.Library
	manaX := costHasVariableMana(manaCostPtr(def.ManaCost))
	if free && manaX {
		cast.XValue = 0
	} else if manaX || additionalCostsUseX(def.AdditionalCosts) {
		minimum := 0
		if additionalCostsRequirePositiveX(def.AdditionalCosts) {
			minimum = 1
		}
		cast.XValue = e.chooseAnnouncementNumber(g, player, "Choose X.", minimum, max(minimum, cast.XValue), private, agents, log)
	}
	if spellHasMultikicker(def) {
		cast.KickerCount = e.chooseAnnouncementNumber(g, player, "Choose how many times to pay multikicker.", 0,
			effectiveKickerCount(cast.KickerPaid, cast.KickerCount), private, agents, log)
		cast.KickerPaid = cast.KickerCount > 0
	}
	branch := castBranchForCast(cast)
	if !modesValidForSpellAtBranch(g, player, def, cast.ChosenModes, branch) {
		modes := modeChoicesForSpellAtBranch(g, player, def, branch)
		if len(modes) == 0 {
			return cast, false
		}
		options := make([]game.ChoiceOption, len(modes))
		for index, choice := range modes {
			options[index] = game.ChoiceOption{Index: index, Label: fmt.Sprintf("Modes %v", choice)}
		}
		selected := e.chooseChoice(g, agents, game.ChoiceRequest{
			Kind: game.ChoiceModal, Player: player, Prompt: "Choose spell modes.", Options: options,
			MinChoices: 1, MaxChoices: 1, DefaultSelection: []int{0}, Private: private,
		}, log)
		cast.ChosenModes = slices.Clone(modes[selected[0]])
	}
	if spellTargetsValidForCast(g, player, def, cast.ChosenModes, cast.Targets, cast.XValue,
		effectiveKickerCount(cast.KickerPaid, cast.KickerCount), branch) {
		return cast, true
	}
	var ok bool
	cast.Targets, ok = e.chooseAnnouncedSpellTargets(g, player, def, cast, agents, log)
	return cast, ok
}

func (e *Engine) chooseAnnouncementNumber(g *game.Game, player game.PlayerID, prompt string, minimum, fallback int, private bool, agents [game.NumPlayers]PlayerAgent, log *TurnLog) int {
	selected := e.chooseChoice(g, agents, game.ChoiceRequest{
		Kind: game.ChoiceNumber, Player: player, Prompt: prompt, Private: private,
		NumberRange: opt.Val(game.NumberRange{Minimum: minimum}),
		MinChoices:  1, MaxChoices: 1, DefaultSelection: []int{fallback},
	}, log)
	return selected[0]
}

func (e *Engine) chooseAnnouncedSpellTargets(g *game.Game, player game.PlayerID, def *game.CardDef, cast action.CastSpellAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) ([]game.Target, bool) {
	specs := spellTargetSpecs(def, cast.ChosenModes, castBranchForCast(cast))
	var targets []game.Target
	var counts []int
	for index := range specs {
		spec := normalizeTargetSpec(&specs[index])
		if !targetSpecValid(&spec) {
			return nil, false
		}
		if targetSpecUsesExternalChooser(&spec) {
			targets = append(targets, game.DeferredTarget())
			counts = append(counts, 1)
			continue
		}
		var completions [][]game.Target
		var completionCounts [][]int
		appendTargetChoicesForSpec(g, player, def, 0, game.Event{}, specs, index, targets, counts,
			&completions, &completionCounts, targetChoiceOptions{limit: 1, xValue: opt.Val(cast.XValue)})
		if len(completions) == 0 {
			return nil, false
		}
		candidates := targetCandidatesForSpec(g, player, def, 0, game.Event{}, &spec)
		candidates = filterAnnouncedManaValueTargets(g, spec, candidates, opt.Val(cast.XValue))
		if spec.DistinctFromPriorTargets {
			candidates = filterTargetsDistinctFrom(candidates, targets)
		}
		options := make([][]game.Target, len(candidates))
		for position, candidate := range candidates {
			options[position] = []game.Target{candidate}
		}
		request := targetChoiceRequest(player, "Choose spell targets.", options)
		request.MinChoices, request.MaxChoices = spec.MinTargets, min(spec.MaxTargets, len(candidates))
		request.Private = cast.SourceZone == zone.Hand || cast.SourceZone == zone.Library
		request.DefaultSelection = nil
		for _, target := range completions[0][len(targets) : len(targets)+completionCounts[0][index]] {
			position := slices.IndexFunc(candidates, func(candidate game.Target) bool { return sameTargetObject(candidate, target) })
			request.DefaultSelection = append(request.DefaultSelection, position)
		}
		selected := e.chooseChoice(g, agents, request, log)
		for _, position := range selected {
			targets = append(targets, candidates[position])
		}
		counts = append(counts, len(selected))
	}
	return targets, targetsValidForSpecs(g, player, def, 0, specs, targets) &&
		targetsSatisfyManaValueXForSpecs(g, specs, counts, targets, cast.XValue)
}
