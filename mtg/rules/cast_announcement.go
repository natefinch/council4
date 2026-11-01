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
	cast, ok := e.chooseCastAnnouncements(g, player, def, cast, free, agents, log)
	if !ok {
		return cast, false
	}
	return e.chooseCastTargets(g, player, def, cast, agents, log)
}

// chooseCastAnnouncements takes the caster's X, multikicker and mode choices
// for cast (CR 601.2b).
func (e *Engine) chooseCastAnnouncements(g *game.Game, player game.PlayerID, def *game.CardDef, cast action.CastSpellAction, free bool, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (action.CastSpellAction, bool) {
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
	return cast, true
}

// chooseCastTargets keeps cast's targets when they are legal and otherwise asks
// the caster to choose them (CR 601.2c).
func (e *Engine) chooseCastTargets(g *game.Game, player game.PlayerID, def *game.CardDef, cast action.CastSpellAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (action.CastSpellAction, bool) {
	branch := castBranchForCast(cast)
	if spellTargetsValidForCast(g, player, def, cast.ChosenModes, cast.Targets, cast.XValue,
		effectiveKickerCount(cast.KickerPaid, cast.KickerCount), branch) {
		return cast, true
	}
	var ok bool
	cast.Targets, ok = e.chooseAnnouncedSpellTargets(g, player, def, cast, nil, nil, agents, log)
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

// chooseAnnouncedSpellTargets asks the caster to choose cast's targets, spec by
// spec (CR 601.2c). preferred holds the caster's provisional targets, such as
// those named by its cast action before its Phyrexian payments were announced.
// When they are still candidates, each choice defaults to the matching targets
// of preferred, marked provisional so that the caster keeps them unless it
// deliberately changes them; otherwise it defaults to the first legal
// completion. When payable is non-nil, it holds the legal target choices with
// which the cost already proposed for the spell can be paid (see
// payableTargetChoices), and each spec offers only targets that extend the
// targets chosen so far toward one of them, when any do.
func (e *Engine) chooseAnnouncedSpellTargets(g *game.Game, player game.PlayerID, def *game.CardDef, cast action.CastSpellAction, preferred []game.Target, payable [][]game.Target, agents [game.NumPlayers]PlayerAgent, log *TurnLog) ([]game.Target, bool) {
	branch := castBranchForCast(cast)
	specs := spellTargetSpecs(def, cast.ChosenModes, branch)
	payableChoices, restricted := segmentTargetChoices(g, player, def, specs, payable)
	var preferredCounts []int
	if len(preferred) > 0 {
		preferredCounts, _ = spellTargetCounts(g, player, def, cast.ChosenModes, preferred, branch)
	}
	var targets []game.Target
	var counts []int
	preferredOffset := 0
	for index := range specs {
		var preferredSpec []game.Target
		if index < len(preferredCounts) {
			preferredSpec = preferred[preferredOffset : preferredOffset+preferredCounts[index]]
			preferredOffset += preferredCounts[index]
		}
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
		fallback := completions[0][len(targets) : len(targets)+completionCounts[0][index]]
		if restricted {
			candidates, fallback = payableTargetsForSpec(payableChoices, targets, counts, candidates, fallback)
		}
		options := make([][]game.Target, len(candidates))
		for position, candidate := range candidates {
			options[position] = []game.Target{candidate}
		}
		request := targetChoiceRequest(player, "Choose spell targets.", options)
		request.MinChoices, request.MaxChoices = spec.MinTargets, min(spec.MaxTargets, len(candidates))
		request.Private = cast.SourceZone == zone.Hand || cast.SourceZone == zone.Library
		request.DefaultSelection = targetPositions(candidates, preferredSpec)
		request.Provisional = index < len(preferredCounts) && len(request.DefaultSelection) == len(preferredSpec) &&
			len(preferredSpec) >= request.MinChoices && len(preferredSpec) <= request.MaxChoices
		if !request.Provisional {
			request.DefaultSelection = targetPositions(candidates, fallback)
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

// targetPositions returns the position in candidates of each of targets that
// is a candidate.
func targetPositions(candidates, targets []game.Target) []int {
	var positions []int
	for _, target := range targets {
		position := slices.IndexFunc(candidates, func(candidate game.Target) bool { return sameTargetObject(candidate, target) })
		if position >= 0 {
			positions = append(positions, position)
		}
	}
	return positions
}

// segmentedTargetChoice is a complete choice of targets for a spell together
// with the number of them chosen for each of its target specs.
type segmentedTargetChoice struct {
	targets []game.Target
	counts  []int
}

// segmentTargetChoices splits each of choices into its targets for each of
// specs. It reports false, leaving the choice of targets unrestricted, when
// choices is empty or one of them can't be split.
func segmentTargetChoices(g *game.Game, player game.PlayerID, def *game.CardDef, specs []game.TargetSpec, choices [][]game.Target) ([]segmentedTargetChoice, bool) {
	if len(choices) == 0 {
		return nil, false
	}
	segmented := make([]segmentedTargetChoice, 0, len(choices))
	for _, choice := range choices {
		counts, ok := targetCountsForSpecs(g, player, def, 0, game.Event{}, specs, choice)
		if !ok {
			return nil, false
		}
		segmented = append(segmented, segmentedTargetChoice{targets: choice, counts: counts})
	}
	return segmented, true
}

// payableTargetsForSpec returns the candidates for the next target spec that
// belong to that spec's targets in some choice of choices beginning with
// chosen, split by counts, and the first such choice's targets for that spec.
// When no choice begins with chosen, every candidate is returned with fallback.
func payableTargetsForSpec(choices []segmentedTargetChoice, chosen []game.Target, counts []int, candidates, fallback []game.Target) ([]game.Target, []game.Target) {
	index := len(counts)
	var segments [][]game.Target
	for _, choice := range choices {
		if len(choice.counts) <= index || !slices.Equal(choice.counts[:index], counts) ||
			!slices.EqualFunc(choice.targets[:len(chosen)], chosen, sameTargetObject) {
			continue
		}
		segments = append(segments, choice.targets[len(chosen):len(chosen)+choice.counts[index]])
	}
	if len(segments) == 0 {
		return candidates, fallback
	}
	payable := slices.DeleteFunc(slices.Clone(candidates), func(candidate game.Target) bool {
		return !slices.ContainsFunc(segments, func(segment []game.Target) bool {
			return slices.ContainsFunc(segment, func(target game.Target) bool { return sameTargetObject(target, candidate) })
		})
	})
	return payable, segments[0]
}
