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
	cast.Targets, ok = e.chooseAnnouncedSpellTargets(g, player, def, cast, nil, nil, nil, agents, log)
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
// targets chosen so far toward one of them, when any do. When open is non-nil,
// it holds the choices of targets still open with that proposal after earlier
// attempts were reversed (CR 733.2), and each spec offers only the targets it
// finds some open choice extending (see openTargetsForSpec). A selection that
// begins only withheld choices, found by searching every choice it begins
// rather than a sample, is invalid and replaced by an open default, so legal
// choices beyond any sample stay open. When no open choice remains, it
// records that and the choice fails.
func (e *Engine) chooseAnnouncedSpellTargets(g *game.Game, player game.PlayerID, def *game.CardDef, cast action.CastSpellAction, preferred []game.Target, payable [][]game.Target, open *proposalTargets, agents [game.NumPlayers]PlayerAgent, log *TurnLog) ([]game.Target, bool) {
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
		if open != nil {
			var closed bool
			candidates, fallback, closed = openTargetsForSpec(g, player, def, specs, index, cast.XValue, targets, counts, candidates, fallback, open)
			if closed {
				open.closePrefix(targets, counts)
				return nil, false
			}
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
		chosen := make([]game.Target, len(selected))
		for i, position := range selected {
			chosen[i] = candidates[position]
		}
		if open != nil {
			// Each search stops at the first open choice it finds, so
			// searching every choice is costly only when none is open.
			extends := func(segment []game.Target) bool {
				return openTargetStatus(g, player, def, specs, index+1, cast.XValue,
					append(slices.Clone(targets), segment...), append(slices.Clone(counts), len(segment)), open, 0) == targetsOpen
			}
			if !extends(chosen) {
				// Every choice of targets this selection begins is withheld,
				// so it is invalid and an open default replaces it: the
				// provisional one if still open, otherwise the one
				// openTargetsForSpec found.
				defaults := make([]game.Target, len(request.DefaultSelection))
				for i, position := range request.DefaultSelection {
					defaults[i] = candidates[position]
				}
				g.Execution.InvalidChoices++
				chosen = fallback
				if len(defaults) >= spec.MinTargets && extends(defaults) {
					chosen = defaults
				}
			}
		}
		targets = append(targets, chosen...)
		counts = append(counts, len(chosen))
	}
	return targets, targetsValidForSpecs(g, player, def, 0, specs, targets) &&
		targetsSatisfyManaValueXForSpecs(g, specs, counts, targets, cast.XValue)
}

// openTargetsStatus is what a search for an open choice of targets found.
type openTargetsStatus int

const (
	// targetsOpen means an open choice of targets was found.
	targetsOpen openTargetsStatus = iota
	// targetsUnknown means none was found among the choices searched, but
	// more remain.
	targetsUnknown
	// targetsClosed means every choice of targets searched for is withheld.
	targetsClosed
)

// openTargetCompletion searches the choices of targets that begin with chosen,
// the targets for the specs before specs[index] split by counts, for one that
// open doesn't withhold (CR 733.2), and returns the first open one found. It
// searches at most limit of them, reporting targetsUnknown when it finds none
// but more remain, or every one of them, lazily, when limit is 0. When
// segmentOK is non-nil, only the choices whose targets for specs[index] it
// accepts are searched.
func openTargetCompletion(g *game.Game, player game.PlayerID, def *game.CardDef, specs []game.TargetSpec, index, xValue int, chosen []game.Target, counts []int, open *proposalTargets, limit int, segmentOK func([]game.Target) bool) ([]game.Target, []int, openTargetsStatus) {
	if open.prefixClosed(chosen, counts) {
		return nil, nil, targetsClosed
	}
	status := targetsClosed
	var completion []game.Target
	var completionCounts []int
	searched := 0
	visit := func(targets []game.Target, targetCounts []int) bool {
		if limit > 0 && searched == limit {
			status = targetsUnknown
			return false
		}
		searched++
		if segmentOK != nil && !segmentOK(targets[len(chosen):len(chosen)+targetCounts[index]]) ||
			open.closed(targets, targetCounts) {
			return true
		}
		completion, completionCounts, status = slices.Clone(targets), slices.Clone(targetCounts), targetsOpen
		return false
	}
	visitTargetChoicesForSpec(g, player, def, 0, game.Event{}, specs, index, chosen, counts,
		&[][]game.Target{}, &[][]int{}, targetChoiceOptions{xValue: opt.Val(xValue), visit: visit})
	return completion, completionCounts, status
}

// openTargetStatus reports whether some choice of targets that begins with
// chosen, split by counts, is open, searching at most limit of them (see
// openTargetCompletion).
func openTargetStatus(g *game.Game, player game.PlayerID, def *game.CardDef, specs []game.TargetSpec, index, xValue int, chosen []game.Target, counts []int, open *proposalTargets, limit int) openTargetsStatus {
	_, _, status := openTargetCompletion(g, player, def, specs, index, xValue, chosen, counts, open, limit, nil)
	return status
}

// openTargetsForSpec returns the candidates for specs[index] to offer after
// chosen, split by counts, and the targets to choose for it by default, when
// open holds the choices of targets still open (CR 733.2). A spec that takes
// at most one target leaves out the candidates a bounded search finds begin
// only withheld choices, so that one withheld choice leaves every other open;
// a spec that takes more offers every candidate, and a selection that begins
// only withheld choices is replaced by the default (see
// chooseAnnouncedSpellTargets). Unless a bounded search finds the default
// begins an open choice, it is moved to the first open choice that begins with
// chosen, searching every one of them, so that the default always begins an
// open choice however many legal choices there are. It reports closed,
// offering nothing, when no choice that begins with chosen is open.
func openTargetsForSpec(g *game.Game, player game.PlayerID, def *game.CardDef, specs []game.TargetSpec, index, xValue int, chosen []game.Target, counts []int, candidates, fallback []game.Target, open *proposalTargets) ([]game.Target, []game.Target, bool) {
	extend := func(targets []game.Target) openTargetsStatus {
		return openTargetStatus(g, player, def, specs, index+1, xValue,
			append(slices.Clone(chosen), targets...), append(slices.Clone(counts), len(targets)), open,
			maxSampledNumericTargetChoices)
	}
	spec := normalizeTargetSpec(&specs[index])
	if spec.MaxTargets == 1 {
		candidates = slices.DeleteFunc(slices.Clone(candidates), func(candidate game.Target) bool {
			return extend([]game.Target{candidate}) == targetsClosed
		})
	}
	if extend(fallback) == targetsOpen {
		return candidates, fallback, false
	}
	// An open choice whose targets for this spec are all offered is
	// preferred; otherwise the targets of the first open choice are offered
	// too, since the payable choices candidates may be narrowed to can all
	// be withheld.
	offered := func(target game.Target) bool {
		return slices.ContainsFunc(candidates, func(candidate game.Target) bool { return sameTargetObject(candidate, target) })
	}
	completion, completionCounts, status := openTargetCompletion(g, player, def, specs, index, xValue, chosen, counts, open, 0,
		func(segment []game.Target) bool {
			return !slices.ContainsFunc(segment, func(target game.Target) bool { return !offered(target) })
		})
	if status != targetsOpen {
		completion, completionCounts, status = openTargetCompletion(g, player, def, specs, index, xValue, chosen, counts, open, 0, nil)
	}
	if status != targetsOpen {
		return nil, nil, true
	}
	segment := completion[len(chosen) : len(chosen)+completionCounts[index]]
	for _, target := range segment {
		if !offered(target) {
			candidates = append(slices.Clone(candidates), target)
		}
	}
	return candidates, segment, false
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
