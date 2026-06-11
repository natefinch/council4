package rules

import (
	"fmt"
	"slices"
	"strings"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// targetChoiceKind distinguishes the four outcomes of target enumeration so
// callers never need to infer state from nil-slice shape.
type targetChoiceKind int

const (
	// targetNoTargetsRequired: the spell or ability has no target specs.
	targetNoTargetsRequired targetChoiceKind = iota
	// targetLegalChoicesFound: at least one legal target combination exists,
	// including the "choose no targets" combination for optional specs.
	targetLegalChoicesFound
	// targetNoLegalChoices: target specs are present and valid but no legal
	// candidates exist on the current board state.
	targetNoLegalChoices
	// targetInvalidSpec: a target spec has an invalid range (e.g. min > max)
	// and represents a card-definition bug rather than a board-state outcome.
	targetInvalidSpec
)

// targetChoiceResult carries the outcome of target enumeration with an
// explicit kind so all four states are distinguishable without nil inspection.
type targetChoiceResult struct {
	choices [][]game.Target
	kind    targetChoiceKind
	// err is diagnostic context for invalid card-definition input. Production
	// enumeration currently treats invalid specs as unavailable actions/triggers.
	err error
}

func targetChoicesForSpell(g *game.Game, controller game.PlayerID, card *game.CardDef, chosenModes []int) targetChoiceResult {
	specs := spellTargetSpecs(card, chosenModes)
	return targetChoicesForSpecs(g, controller, card, 0, specs)
}

func targetChoicesForBody(g *game.Game, controller game.PlayerID, body game.Ability) targetChoiceResult {
	return targetChoicesForBodyFromSource(g, controller, nil, body)
}

func targetChoicesForBodyFromSource(g *game.Game, controller game.PlayerID, source *game.CardDef, body game.Ability) targetChoiceResult {
	return targetChoicesForBodyFromSourceObject(g, controller, source, 0, body)
}

func targetChoicesForBodyFromSourceObject(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, body game.Ability) targetChoiceResult {
	if body == nil {
		return targetChoiceResult{kind: targetNoTargetsRequired, choices: [][]game.Target{nil}}
	}
	return targetChoicesForSpecs(g, controller, source, sourceObjectID, game.BodyTargets(body))
}

// targetChoicesForSpecs enumerates every legal target combination for specs.
// Returns an explicit result kind so callers never infer state from nil-slice shape:
//   - targetNoTargetsRequired when specs is empty
//   - targetLegalChoicesFound when at least one combination is legal (including optional no-target choices)
//   - targetNoLegalChoices when specs are valid but no board-legal combination exists
//   - targetInvalidSpec (with err) when a spec has an invalid min/max range
func targetChoicesForSpecs(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, specs []game.TargetSpec) targetChoiceResult {
	if len(specs) == 0 {
		return targetChoiceResult{kind: targetNoTargetsRequired, choices: [][]game.Target{nil}}
	}
	for i := range specs {
		spec := &specs[i]
		normalized := normalizeTargetSpec(spec)
		if !targetSpecValid(&normalized) {
			return targetChoiceResult{
				kind: targetInvalidSpec,
				err:  fmt.Errorf("target spec %q has invalid range: min=%d max=%d", spec.Constraint, normalized.MinTargets, normalized.MaxTargets),
			}
		}
	}
	var result [][]game.Target
	appendTargetChoicesForSpec(g, controller, source, sourceObjectID, specs, 0, nil, &result)
	if len(result) == 0 {
		return targetChoiceResult{kind: targetNoLegalChoices}
	}
	return targetChoiceResult{kind: targetLegalChoicesFound, choices: result}
}

func appendTargetChoicesForSpec(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, specs []game.TargetSpec, specIndex int, prefix []game.Target, result *[][]game.Target) {
	if specIndex >= len(specs) {
		*result = append(*result, append([]game.Target(nil), prefix...))
		return
	}
	spec := normalizeTargetSpec(&specs[specIndex])
	if !targetSpecValid(&spec) {
		return
	}
	if targetSpecUsesExternalChooser(&spec) {
		if len(choosingOpponentsForTargetSpec(g, controller, source, sourceObjectID, &spec)) == 0 {
			return
		}
		next := append(append([]game.Target(nil), prefix...), game.DeferredTarget())
		appendTargetChoicesForSpec(g, controller, source, sourceObjectID, specs, specIndex+1, next, result)
		return
	}
	candidates := targetCandidatesForSpec(g, controller, source, sourceObjectID, &spec)
	maxTargets := min(spec.MaxTargets, len(candidates))
	for _, count := range targetCountsForChoices(spec.MinTargets, maxTargets) {
		for _, combination := range targetCombinations(candidates, count) {
			next := append(append([]game.Target(nil), prefix...), combination...)
			appendTargetChoicesForSpec(g, controller, source, sourceObjectID, specs, specIndex+1, next, result)
		}
	}
}

func targetCountsForChoices(minTargets, maxTargets int) []int {
	var counts []int
	start := minTargets
	if minTargets == 0 && maxTargets > 0 {
		start = 1
	}
	for count := start; count <= maxTargets; count++ {
		counts = append(counts, count)
	}
	if minTargets == 0 && start > 0 {
		counts = append(counts, 0)
	}
	return counts
}

func targetCandidatesForSpec(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, spec *game.TargetSpec) []game.Target {
	return targetCandidatesForSpecChosenBy(g, controller, controller, source, sourceObjectID, spec)
}

func targetCandidatesForSpecChosenBy(g *game.Game, sourceController, predicatePlayer game.PlayerID, source *game.CardDef, sourceObjectID id.ID, spec *game.TargetSpec) []game.Target {
	var candidates []game.Target
	if targetSpecAllowsPlayers(spec) {
		for playerID := range game.PlayerID(game.NumPlayers) {
			target := game.PlayerTarget(playerID)
			if targetMatchesSpec(g, predicatePlayer, sourceObjectID, spec, target) && !targetProtectedFromSource(g, sourceController, source, sourceObjectID, target) {
				candidates = append(candidates, target)
			}
		}
	}
	if targetSpecAllowsPermanents(spec) {
		for _, permanent := range g.Battlefield {
			target := game.PermanentTarget(permanent.ObjectID)
			if targetMatchesSpec(g, predicatePlayer, sourceObjectID, spec, target) && !targetProtectedFromSource(g, sourceController, source, sourceObjectID, target) {
				candidates = append(candidates, target)
			}
		}
	}
	if targetSpecAllowsCards(spec) {
		for _, card := range g.CardInstances {
			target := game.CardTargetWithZoneVersion(card.ID, card.ZoneVersion)
			if targetMatchesSpec(g, predicatePlayer, sourceObjectID, spec, target) {
				candidates = append(candidates, target)
			}
		}
	}
	return candidates
}

func targetCombinations(candidates []game.Target, count int) [][]game.Target {
	if count == 0 {
		return [][]game.Target{nil}
	}
	if count > len(candidates) {
		return nil
	}
	var result [][]game.Target
	var walk func(start int, chosen []game.Target)
	walk = func(start int, chosen []game.Target) {
		if len(chosen) == count {
			result = append(result, append([]game.Target(nil), chosen...))
			return
		}
		need := count - len(chosen)
		for i := start; i <= len(candidates)-need; i++ {
			walk(i+1, append(chosen, candidates[i]))
		}
	}
	walk(0, nil)
	return result
}

func targetsValidForSpell(g *game.Game, controller game.PlayerID, card *game.CardDef, chosenModes []int, targets []game.Target) bool {
	specs := spellTargetSpecs(card, chosenModes)
	return targetsValidForSpecs(g, controller, card, 0, specs, targets)
}

func targetsValidForBody(g *game.Game, controller game.PlayerID, body game.Ability, targets []game.Target) bool {
	return targetsValidForBodyFromSource(g, controller, nil, body, targets)
}

func targetsValidForBodyFromSource(g *game.Game, controller game.PlayerID, source *game.CardDef, body game.Ability, targets []game.Target) bool {
	return targetsValidForBodyFromSourceObject(g, controller, source, 0, body, targets)
}

func targetsValidForBodyFromSourceObject(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, body game.Ability, targets []game.Target) bool {
	if body == nil {
		return len(targets) == 0
	}
	return targetsValidForSpecs(g, controller, source, sourceObjectID, game.BodyTargets(body), targets)
}

func targetsValidForSpecs(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, specs []game.TargetSpec, targets []game.Target) bool {
	_, ok := targetCountsForSpecs(g, controller, source, sourceObjectID, specs, targets)
	return ok
}

func targetCountsForSpecs(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, specs []game.TargetSpec, targets []game.Target) ([]int, bool) {
	if len(specs) == 0 {
		return nil, len(targets) == 0
	}
	counts := make([]int, len(specs))
	if !targetCountsForSpecFrom(g, controller, source, sourceObjectID, specs, targets, counts, 0, 0) {
		return nil, false
	}
	return counts, true
}

func targetCountsForSpecFrom(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, specs []game.TargetSpec, targets []game.Target, counts []int, specIndex, targetIndex int) bool {
	if specIndex == len(specs) {
		return targetIndex == len(targets)
	}
	spec := normalizeTargetSpec(&specs[specIndex])
	if !targetSpecValid(&spec) {
		return false
	}
	remaining := len(targets) - targetIndex
	maxTargets := min(spec.MaxTargets, remaining)
	for count := spec.MinTargets; count <= maxTargets; count++ {
		slice := targets[targetIndex : targetIndex+count]
		if targetsMatchSpecSlice(g, controller, source, sourceObjectID, &spec, slice) &&
			targetCountsForSpecFrom(g, controller, source, sourceObjectID, specs, targets, counts, specIndex+1, targetIndex+count) {
			counts[specIndex] = count
			return true
		}
	}
	return false
}

func spellTargetCounts(g *game.Game, controller game.PlayerID, card *game.CardDef, chosenModes []int, targets []game.Target) ([]int, bool) {
	return targetCountsForSpecs(g, controller, card, 0, spellTargetSpecs(card, chosenModes), targets)
}

func bodyTargetCounts(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, body game.Ability, targets []game.Target) ([]int, bool) {
	if body == nil {
		return nil, len(targets) == 0
	}
	return targetCountsForSpecs(g, controller, source, sourceObjectID, game.BodyTargets(body), targets)
}

func targetsMatchSpecSlice(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, spec *game.TargetSpec, targets []game.Target) bool {
	if len(targets) < spec.MinTargets || len(targets) > spec.MaxTargets {
		return false
	}
	if targetSpecUsesExternalChooser(spec) {
		if len(targets) != 1 {
			return false
		}
		if targets[0].Kind == game.TargetDeferred {
			return len(choosingOpponentsForTargetSpec(g, controller, source, sourceObjectID, spec)) > 0
		}
		return externalChooserCouldChooseTarget(g, controller, source, sourceObjectID, spec, targets[0])
	}
	seen := make(map[game.Target]bool, len(targets))
	for _, target := range targets {
		if seen[target] || !targetMatchesSpec(g, controller, sourceObjectID, spec, target) || targetProtectedFromSource(g, controller, source, sourceObjectID, target) {
			return false
		}
		seen[target] = true
	}
	return true
}

func spellHasAnyLegalTargets(g *game.Game, card *game.CardDef, obj *game.StackObject) bool {
	return stackObjectHasAnyLegalTargetsForSpecs(g, card, 0, spellTargetSpecs(card, obj.ChosenModes), obj)
}

func bodyHasAnyLegalTargetsFromSourceObject(g *game.Game, source *game.CardDef, sourceObjectID id.ID, body game.Ability, obj *game.StackObject) bool {
	if body == nil {
		return len(obj.Targets) == 0
	}
	return stackObjectHasAnyLegalTargetsForSpecs(g, source, sourceObjectID, game.BodyTargets(body), obj)
}

func stackObjectHasAnyLegalTargetsForSpecs(g *game.Game, source *game.CardDef, sourceObjectID id.ID, specs []game.TargetSpec, obj *game.StackObject) bool {
	if len(specs) == 0 {
		return true
	}
	counts, ok := resolutionTargetCounts(specs, obj.TargetCounts, len(obj.Targets))
	if !ok {
		return false
	}
	if len(obj.Targets) == 0 {
		return true
	}
	targets := append([]game.Target(nil), obj.Targets...)
	anyLegal := false
	targetIndex := 0
	for specIndex := range specs {
		spec := normalizeTargetSpec(&specs[specIndex])
		for range counts[specIndex] {
			target := targets[targetIndex]
			if targetLegalForSpecAtResolution(g, obj.Controller, source, sourceObjectID, &spec, target) {
				anyLegal = true
			} else {
				targets[targetIndex] = game.DeferredTarget()
			}
			targetIndex++
		}
	}
	obj.Targets = targets
	return anyLegal
}

func resolutionTargetCounts(specs []game.TargetSpec, recorded []int, targetCount int) ([]int, bool) {
	if targetCountsHaveValidCardinality(specs, recorded, targetCount) {
		return recorded, true
	}
	counts := make([]int, len(specs))
	if !targetCountsForCardinality(specs, targetCount, counts, 0, 0) {
		return nil, false
	}
	return counts, true
}

func targetCountsHaveValidCardinality(specs []game.TargetSpec, counts []int, targetCount int) bool {
	if len(counts) != len(specs) {
		return false
	}
	total := 0
	for i := range specs {
		spec := normalizeTargetSpec(&specs[i])
		if !targetSpecValid(&spec) || counts[i] < spec.MinTargets || counts[i] > spec.MaxTargets {
			return false
		}
		total += counts[i]
	}
	return total == targetCount
}

func targetCountsForCardinality(specs []game.TargetSpec, targetCount int, counts []int, specIndex, assigned int) bool {
	if specIndex == len(specs) {
		return assigned == targetCount
	}
	spec := normalizeTargetSpec(&specs[specIndex])
	if !targetSpecValid(&spec) {
		return false
	}
	for count := spec.MinTargets; count <= spec.MaxTargets && assigned+count <= targetCount; count++ {
		counts[specIndex] = count
		if targetCountsForCardinality(specs, targetCount, counts, specIndex+1, assigned+count) {
			return true
		}
	}
	return false
}

func targetLegalForSpecAtResolution(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, spec *game.TargetSpec, target game.Target) bool {
	if targetSpecUsesExternalChooser(spec) {
		return externalChooserCouldChooseTarget(g, controller, source, sourceObjectID, spec, target)
	}
	return targetMatchesSpec(g, controller, sourceObjectID, spec, target) &&
		!targetProtectedFromSource(g, controller, source, sourceObjectID, target)
}

func (e *Engine) completeSpellAnnouncementTargets(g *game.Game, controller game.PlayerID, card *game.CardDef, chosenModes []int, targets []game.Target, agents [game.NumPlayers]PlayerAgent, log *TurnLog) ([]game.Target, bool) {
	return e.completeAnnouncementTargets(g, controller, card, 0, spellTargetSpecs(card, chosenModes), targets, agents, log)
}

func (e *Engine) completeAbilityAnnouncementTargets(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, body game.Ability, targets []game.Target, agents [game.NumPlayers]PlayerAgent, log *TurnLog) ([]game.Target, bool) {
	if body == nil {
		return targets, len(targets) == 0
	}
	return e.completeAnnouncementTargets(g, controller, source, sourceObjectID, game.BodyTargets(body), targets, agents, log)
}

func (e *Engine) completeAnnouncementTargets(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, specs []game.TargetSpec, targets []game.Target, agents [game.NumPlayers]PlayerAgent, log *TurnLog) ([]game.Target, bool) {
	if !targetSpecsUseExternalChooser(specs) {
		return bindCardTargetZoneVersions(g, targets)
	}
	if !targetSpecsUseFixedSlots(specs) || len(targets) != len(specs) {
		return nil, false
	}
	completed := append([]game.Target(nil), targets...)
	for i := range specs {
		spec := normalizeTargetSpec(&specs[i])
		if !targetSpecUsesExternalChooser(&spec) {
			continue
		}
		target, ok := e.chooseExternalTarget(g, controller, source, sourceObjectID, &spec, agents, log)
		if !ok {
			return nil, false
		}
		completed[i] = target
	}
	var ok bool
	completed, ok = bindCardTargetZoneVersions(g, completed)
	if !ok {
		return nil, false
	}
	return completed, targetsValidForSpecs(g, controller, source, sourceObjectID, specs, completed)
}

func bindCardTargetZoneVersions(g *game.Game, targets []game.Target) ([]game.Target, bool) {
	bound := append([]game.Target(nil), targets...)
	for i := range bound {
		if bound[i].Kind != game.TargetCard || bound[i].CardZoneVersionSet {
			continue
		}
		card, ok := g.GetCardInstance(bound[i].CardID)
		if !ok {
			return nil, false
		}
		bound[i].CardZoneVersion = card.ZoneVersion
		bound[i].CardZoneVersionSet = true
	}
	return bound, true
}

func targetSpecsUseExternalChooser(specs []game.TargetSpec) bool {
	for i := range specs {
		normalized := normalizeTargetSpec(&specs[i])
		if targetSpecUsesExternalChooser(&normalized) {
			return true
		}
	}
	return false
}

func targetSpecsUseFixedSlots(specs []game.TargetSpec) bool {
	// External chooser completion maps each TargetSpec to one target slot. Keep
	// variable regular target groups out of this path until a second consumer
	// needs full segmentation support.
	for i := range specs {
		normalized := normalizeTargetSpec(&specs[i])
		if normalized.MinTargets != 1 || normalized.MaxTargets != 1 {
			return false
		}
	}
	return true
}

func (e *Engine) chooseExternalTarget(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, spec *game.TargetSpec, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (game.Target, bool) {
	switch spec.Chooser {
	case game.TargetChooserOpponent:
		opponents := choosingOpponentsForTargetSpec(g, controller, source, sourceObjectID, spec)
		if len(opponents) == 0 {
			return game.Target{}, false
		}
		opponent, ok := e.chooseTargetingOpponent(g, controller, spec, opponents, agents, log)
		if !ok {
			return game.Target{}, false
		}
		targets := targetCandidatesForSpecChosenBy(g, controller, opponent, source, sourceObjectID, spec)
		target, ok := e.chooseTargetFromCandidates(g, opponent, spec, targets, agents, log)
		if !ok {
			return game.Target{}, false
		}
		return target, true
	default:
		return game.Target{}, false
	}
}

func (e *Engine) chooseTargetingOpponent(g *game.Game, controller game.PlayerID, spec *game.TargetSpec, opponents []game.PlayerID, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (game.PlayerID, bool) {
	options := make([]game.ChoiceOption, 0, len(opponents))
	for i, opponent := range opponents {
		options = append(options, game.ChoiceOption{Index: i, Label: fmt.Sprintf("Player %d", opponent+1)})
	}
	selected := e.chooseChoice(g, agents, game.ChoiceRequest{
		Kind:       game.ChoicePlayer,
		Player:     controller,
		Prompt:     fmt.Sprintf("Choose an opponent to choose target: %s", spec.Constraint),
		Options:    options,
		MinChoices: 1,
		MaxChoices: 1,
	}, log)
	if len(selected) != 1 || selected[0] < 0 || selected[0] >= len(opponents) {
		return 0, false
	}
	return opponents[selected[0]], true
}

func (e *Engine) chooseTargetFromCandidates(g *game.Game, chooser game.PlayerID, spec *game.TargetSpec, candidates []game.Target, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (game.Target, bool) {
	if len(candidates) == 0 {
		return game.Target{}, false
	}
	choices := make([][]game.Target, 0, len(candidates))
	for _, candidate := range candidates {
		choices = append(choices, []game.Target{candidate})
	}
	selected := e.chooseChoice(g, agents, targetChoiceRequest(chooser, fmt.Sprintf("Choose target: %s", spec.Constraint), choices), log)
	if len(selected) != 1 || selected[0] < 0 || selected[0] >= len(candidates) {
		return game.Target{}, false
	}
	return candidates[selected[0]], true
}

func spellTargetSpecs(card *game.CardDef, chosenModes []int) []game.TargetSpec {
	if isAuraCard(card) {
		spec, ok := enchantTargetSpecForCard(card)
		if !ok {
			return nil
		}
		return []game.TargetSpec{spec}
	}
	ability, ok := firstSpellAbility(card)
	if !ok {
		return nil
	}
	content := *ability
	if len(content.Modes) > 0 {
		if !modesValidForContent(content, chosenModes) {
			return nil
		}
		if !content.IsModal() {
			specs := append([]game.TargetSpec(nil), content.SharedTargets...)
			return append(specs, content.Modes[0].Targets...)
		}
		specs := append([]game.TargetSpec(nil), content.SharedTargets...)
		for _, modeIndex := range chosenModes {
			specs = append(specs, content.Modes[modeIndex].Targets...)
		}
		return specs
	}
	return game.BodyTargets(*ability)
}

func modeChoicesForSpell(card *game.CardDef) [][]int {
	ability, _ := firstSpellAbility(card)
	if ability == nil {
		return [][]int{nil}
	}
	return modeChoicesForContent(*ability)
}

func modeChoicesForBody(body game.Ability) [][]int {
	if body == nil {
		return [][]int{nil}
	}
	return modeChoicesForContent(game.BodyContent(body))
}

func modeChoicesForContent(content game.AbilityContent) [][]int {
	if len(content.Modes) == 0 || !content.IsModal() {
		return [][]int{nil}
	}
	// Modal choices are made before targets/costs are finalized and are locked
	// into the stack object (CR 601.2d, CR 700.2).
	minModes, maxModes := modeChoiceRangeFromContent(content)
	if minModes < 0 || maxModes < minModes || maxModes > len(content.Modes) {
		return nil
	}
	if content.AllowDuplicateModes {
		return duplicateModeChoices(len(content.Modes), minModes, maxModes)
	}
	var choices [][]int
	for count := minModes; count <= maxModes; count++ {
		choices = append(choices, modeCombinations(len(content.Modes), count)...)
	}
	return choices
}

func modesValidForSpell(card *game.CardDef, chosenModes []int) bool {
	ability, ok := firstSpellAbility(card)
	if !ok {
		return len(chosenModes) == 0
	}
	return modesValidForContent(*ability, chosenModes)
}

func modesValidForBody(body game.Ability, chosenModes []int) bool {
	if body == nil {
		return len(chosenModes) == 0
	}
	return modesValidForContent(game.BodyContent(body), chosenModes)
}

func modesValidForContent(content game.AbilityContent, chosenModes []int) bool {
	if len(content.Modes) == 0 || !content.IsModal() {
		return len(chosenModes) == 0
	}
	minModes, maxModes := modeChoiceRangeFromContent(content)
	if len(chosenModes) < minModes || len(chosenModes) > maxModes {
		return false
	}
	seen := make(map[int]bool, len(chosenModes))
	previousMode := -1
	for i, modeIndex := range chosenModes {
		if modeIndex < 0 || modeIndex >= len(content.Modes) {
			return false
		}
		if i > 0 && previousMode > modeIndex {
			return false
		}
		previousMode = modeIndex
		// Canonical nondecreasing order avoids representing the same modal
		// choice multiple ways while preserving duplicate-mode templates that
		// explicitly permit repeats (CR 700.2d).
		if !content.AllowDuplicateModes {
			if seen[modeIndex] {
				return false
			}
			seen[modeIndex] = true
		}
	}
	return true
}

func modeChoiceRangeFromContent(content game.AbilityContent) (minModes, maxModes int) {
	if len(content.Modes) == 0 {
		return 0, 0
	}
	minModes = content.MinModes
	maxModes = content.MaxModes
	if minModes == 0 && maxModes == 0 {
		return 1, 1
	}
	return minModes, maxModes
}

func modeCombinations(modeCount, chooseCount int) [][]int {
	if chooseCount == 0 {
		return [][]int{nil}
	}
	if chooseCount > modeCount {
		return nil
	}
	var result [][]int
	var walk func(start int, chosen []int)
	walk = func(start int, chosen []int) {
		if len(chosen) == chooseCount {
			result = append(result, append([]int(nil), chosen...))
			return
		}
		need := chooseCount - len(chosen)
		for i := start; i <= modeCount-need; i++ {
			walk(i+1, append(chosen, i))
		}
	}
	walk(0, nil)
	return result
}

func duplicateModeChoices(modeCount, minModes, maxModes int) [][]int {
	var result [][]int
	var walk func(start int, chosen []int)
	walk = func(start int, chosen []int) {
		if len(chosen) >= minModes {
			result = append(result, append([]int(nil), chosen...))
		}
		if len(chosen) == maxModes {
			return
		}
		for i := start; i < modeCount; i++ {
			walk(i, append(chosen, i))
		}
	}
	walk(0, nil)
	return result
}

func normalizeTargetSpec(spec *game.TargetSpec) game.TargetSpec {
	normalized := *spec
	if normalized.MinTargets == 0 && normalized.MaxTargets == 0 && normalized.Constraint != "" {
		normalized.MinTargets = 1
		normalized.MaxTargets = 1
	}
	return normalized
}

func targetSpecValid(spec *game.TargetSpec) bool {
	if spec.MinTargets < 0 || spec.MaxTargets < spec.MinTargets {
		return false
	}
	switch spec.Chooser {
	case game.TargetChooserController:
		return true
	case game.TargetChooserOpponent:
		return spec.MinTargets == 1 && spec.MaxTargets == 1
	default:
		return false
	}
}

func targetSpecUsesExternalChooser(spec *game.TargetSpec) bool {
	return spec.Chooser != game.TargetChooserController
}

func choosingOpponentsForTargetSpec(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, spec *game.TargetSpec) []game.PlayerID {
	if spec.Chooser != game.TargetChooserOpponent {
		return nil
	}
	var players []game.PlayerID
	current := controller
	for range game.NumPlayers - 1 {
		current = g.TurnOrder.NextPriority(current)
		if current == controller {
			break
		}
		if !isPlayerAlive(g, current) {
			continue
		}
		if len(targetCandidatesForSpecChosenBy(g, controller, current, source, sourceObjectID, spec)) > 0 {
			players = append(players, current)
		}
	}
	return players
}

func externalChooserCouldChooseTarget(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, spec *game.TargetSpec, target game.Target) bool {
	for _, chooser := range choosingOpponentsForTargetSpec(g, controller, source, sourceObjectID, spec) {
		if slices.Contains(targetCandidatesForSpecChosenBy(g, controller, chooser, source, sourceObjectID, spec), target) {
			return true
		}
	}
	return false
}

func targetSpecAllowsPlayers(spec *game.TargetSpec) bool {
	if spec.Allow != game.TargetAllowUnspecified {
		return spec.Allow&game.TargetAllowPlayer != 0
	}
	normalized := normalizedTargetConstraint(spec)
	return normalized == "player" ||
		normalized == "target player" ||
		normalized == "opponent" ||
		normalized == "target opponent" ||
		normalized == "any target"
}

func targetSpecAllowsPermanents(spec *game.TargetSpec) bool {
	if spec.Allow != game.TargetAllowUnspecified {
		return spec.Allow&game.TargetAllowPermanent != 0
	}
	normalized := normalizedTargetConstraint(spec)
	if normalized == "any target" {
		return true
	}
	if strings.Contains(normalized, "permanent") ||
		strings.Contains(normalized, "creature") ||
		strings.Contains(normalized, "artifact") ||
		strings.Contains(normalized, "enchantment") ||
		strings.Contains(normalized, "land") ||
		strings.Contains(normalized, "planeswalker") ||
		strings.Contains(normalized, "battle") {
		return true
	}
	return false
}

func targetSpecAllowsCards(spec *game.TargetSpec) bool {
	if spec.Allow != game.TargetAllowUnspecified {
		return spec.Allow&game.TargetAllowCard != 0
	}
	normalized := normalizedTargetConstraint(spec)
	return strings.Contains(normalized, "card") &&
		(strings.Contains(normalized, "graveyard") || strings.Contains(normalized, "library") || strings.Contains(normalized, "hand"))
}

func targetMatchesSpec(g *game.Game, controller game.PlayerID, sourceObjectID id.ID, spec *game.TargetSpec, target game.Target) bool {
	switch target.Kind {
	case game.TargetPlayer:
		return playerTargetMatchesSpec(g, controller, spec, target.PlayerID)
	case game.TargetPermanent:
		return permanentTargetMatchesSpec(g, controller, sourceObjectID, spec, target.PermanentID)
	case game.TargetCard:
		return cardTargetMatchesSpec(g, controller, spec, target)
	default:
		return false
	}
}

func cardTargetMatchesSpec(g *game.Game, controller game.PlayerID, spec *game.TargetSpec, target game.Target) bool {
	if !targetSpecAllowsCards(spec) {
		return false
	}
	cardID := target.CardID
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return false
	}
	if target.CardZoneVersionSet && card.ZoneVersion != target.CardZoneVersion {
		return false
	}
	if spec.TargetZone != zone.None {
		actualZone, ok := cardZone(g, cardID)
		if !ok || actualZone != spec.TargetZone {
			return false
		}
	}
	sel := targetSelection(spec)
	if !sel.Empty() {
		subject := selectionSubject{
			kind:       subjectCard,
			g:          g,
			card:       card,
			controller: card.Owner,
			viewer:     controller,
		}
		if !matchSelection(&subject, &sel) {
			return false
		}
	}
	return true
}

func playerTargetMatchesSpec(g *game.Game, controller game.PlayerID, spec *game.TargetSpec, playerID game.PlayerID) bool {
	if !isPlayerAlive(g, playerID) || !targetSpecAllowsPlayers(spec) {
		return false
	}
	sel := targetSelection(spec)
	if sel.Player != game.PlayerAny {
		return selectionPlayerRelationMatches(sel.Player, playerID, controller)
	}
	normalized := normalizedTargetConstraint(spec)
	if strings.Contains(normalized, "opponent") && playerID == controller {
		return false
	}
	return true
}

func permanentTargetMatchesSpec(g *game.Game, controller game.PlayerID, sourceObjectID id.ID, spec *game.TargetSpec, permanentID id.ID) bool {
	if !targetSpecAllowsPermanents(spec) {
		return false
	}
	permanent, ok := permanentByObjectID(g, permanentID)
	if !ok || permanent.PhasedOut {
		return false
	}
	sel := targetSelection(spec)
	if !sel.Empty() {
		values := effectivePermanentValues(g, permanent)
		subject := selectionSubject{
			kind:           subjectPermanent,
			g:              g,
			permanent:      permanent,
			values:         &values,
			viewer:         controller,
			sourceObjectID: sourceObjectID,
			clampPower:     true,
		}
		if sel.Controller != game.ControllerAny {
			subject.controller = effectiveController(g, permanent)
		}
		if !matchSelection(&subject, &sel) {
			return false
		}
	}
	if sel.Controller == game.ControllerAny && !permanentConstraintControllerMatches(g, controller, spec, permanent) {
		return false
	}
	if normalizedTargetConstraint(spec) == "any target" {
		return permanentHasType(g, permanent, types.Creature) ||
			permanentHasType(g, permanent, types.Planeswalker) ||
			permanentHasType(g, permanent, types.Battle)
	}
	return permanentTypeMatchesSpec(g, spec, permanent)
}

// targetSelection returns the Selection a TargetSpec matches against, preferring
// the explicit Selection and otherwise adapting the legacy TargetPredicate.
func targetSelection(spec *game.TargetSpec) game.Selection {
	if spec.Selection.Exists {
		return spec.Selection.Val
	}
	return spec.Predicate.Selection()
}

func combatStateMatches(g *game.Game, permanent *game.Permanent, filter game.CombatStateFilter) bool {
	if filter == game.CombatStateAny {
		return true
	}
	attacking := false
	blocking := false
	if g.Combat != nil {
		attacking = slices.ContainsFunc(g.Combat.Attackers, func(declaration game.AttackDeclaration) bool {
			return declaration.Attacker == permanent.ObjectID
		})
		blocking = slices.ContainsFunc(g.Combat.Blockers, func(declaration game.BlockDeclaration) bool {
			return declaration.Blocker == permanent.ObjectID
		})
	}
	switch filter {
	case game.CombatStateAttacking:
		return attacking
	case game.CombatStateBlocking:
		return blocking
	case game.CombatStateAttackingOrBlocking:
		return attacking || blocking
	default:
		return true
	}
}

func targetProtectedFromSource(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, target game.Target) bool {
	if target.Kind != game.TargetPermanent {
		return false
	}
	permanent, ok := permanentByObjectID(g, target.PermanentID)
	if !ok {
		return false
	}
	if hasKeyword(g, permanent, game.Shroud) {
		return true
	}
	if hasKeyword(g, permanent, game.Hexproof) && effectiveController(g, permanent) != controller {
		return true
	}
	// Use effective source characteristics when the source is a permanent on
	// the battlefield (CR 702.16c).
	if sourceObjectID != 0 {
		if sourcePermanent, ok2 := permanentByObjectID(g, sourceObjectID); ok2 {
			return permanentProtectedFromPermanentEffective(g, permanent, sourcePermanent)
		}
		// Stack spell: use the selected face's characteristics.
		if stackObj, ok2 := stackObjectByID(g, sourceObjectID); ok2 {
			if chars, ok3 := stackObjectSourceChars(g, stackObj); ok3 {
				return permanentProtectedFromChars(g, permanent, chars)
			}
		}
	}
	// Fall back to the supplied face def (LKI, spell during announcement, etc.).
	return source != nil && permanentProtectedFromSourceDef(g, permanent, source)
}

func permanentConstraintControllerMatches(g *game.Game, controller game.PlayerID, spec *game.TargetSpec, permanent *game.Permanent) bool {
	permanentController := effectiveController(g, permanent)
	normalized := normalizedTargetConstraint(spec)
	switch {
	case strings.Contains(normalized, "you control") || strings.Contains(normalized, "controlled by you"):
		return permanentController == controller
	case strings.Contains(normalized, "opponent controls") ||
		strings.Contains(normalized, "opponents control") ||
		strings.Contains(normalized, "controlled by an opponent") ||
		strings.Contains(normalized, "controlled by opponent"):
		return permanentController != controller && isPlayerAlive(g, permanentController)
	default:
		return true
	}
}

func permanentTypeMatchesSpec(g *game.Game, spec *game.TargetSpec, permanent *game.Permanent) bool {
	if spec.Selection.Exists && normalizedTargetConstraint(spec) == "" {
		return true
	}
	if len(spec.Predicate.PermanentTypes) > 0 || len(spec.Predicate.ExcludedTypes) > 0 {
		return true
	}
	normalized := normalizedTargetConstraint(spec)
	if spec.Allow != game.TargetAllowUnspecified && normalized == "" {
		if spec.Allow&game.TargetAllowPlayer != 0 {
			return permanentHasType(g, permanent, types.Creature) ||
				permanentHasType(g, permanent, types.Planeswalker) ||
				permanentHasType(g, permanent, types.Battle)
		}
		return len(effectivePermanentValues(g, permanent).types) > 0
	}
	if strings.Contains(normalized, "nonland permanent") {
		return !permanentHasType(g, permanent, types.Land)
	}
	if strings.Contains(normalized, "permanent") && !containsAnyPermanentTypeConstraint(normalized) {
		return len(effectivePermanentValues(g, permanent).types) > 0
	}
	allowedTypes := permanentTypesForConstraint(normalized)
	if len(allowedTypes) == 0 {
		return false
	}
	return slices.ContainsFunc(allowedTypes, func(cardType types.Card) bool {
		return permanentHasType(g, permanent, cardType)
	})
}

func containsAnyPermanentTypeConstraint(normalized string) bool {
	return strings.Contains(normalized, "creature") ||
		strings.Contains(normalized, "artifact") ||
		strings.Contains(normalized, "enchantment") ||
		strings.Contains(normalized, "land") ||
		strings.Contains(normalized, "planeswalker") ||
		strings.Contains(normalized, "battle")
}

func permanentTypesForConstraint(normalized string) []types.Card {
	var cardTypes []types.Card
	if strings.Contains(normalized, "creature") {
		cardTypes = append(cardTypes, types.Creature)
	}
	if strings.Contains(normalized, "artifact") {
		cardTypes = append(cardTypes, types.Artifact)
	}
	if strings.Contains(normalized, "enchantment") {
		cardTypes = append(cardTypes, types.Enchantment)
	}
	if strings.Contains(normalized, "land") {
		cardTypes = append(cardTypes, types.Land)
	}
	if strings.Contains(normalized, "planeswalker") {
		cardTypes = append(cardTypes, types.Planeswalker)
	}
	if strings.Contains(normalized, "battle") {
		cardTypes = append(cardTypes, types.Battle)
	}
	return cardTypes
}

func normalizedTargetConstraint(spec *game.TargetSpec) string {
	normalized := strings.ToLower(strings.TrimSpace(spec.Constraint))
	normalized = strings.TrimPrefix(normalized, "target ")
	return strings.Join(strings.Fields(normalized), " ")
}

func isPlayerAlive(g *game.Game, playerID game.PlayerID) bool {
	if playerID < 0 || int(playerID) >= len(g.Players) {
		return false
	}
	player := g.Players[playerID]
	return !player.Eliminated && !g.TurnOrder.IsEliminated(playerID)
}
