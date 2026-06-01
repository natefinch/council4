package rules

import (
	"fmt"
	"slices"
	"strings"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
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

func targetChoicesForAbility(g *game.Game, controller game.PlayerID, ability *game.AbilityDef) targetChoiceResult {
	return targetChoicesForAbilityFromSource(g, controller, nil, ability)
}

func targetChoicesForAbilityFromSource(g *game.Game, controller game.PlayerID, source *game.CardDef, ability *game.AbilityDef) targetChoiceResult {
	return targetChoicesForAbilityFromSourceObject(g, controller, source, 0, ability)
}

func targetChoicesForAbilityFromSourceObject(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, ability *game.AbilityDef) targetChoiceResult {
	if ability == nil {
		return targetChoiceResult{kind: targetNoTargetsRequired, choices: [][]game.Target{nil}}
	}
	return targetChoicesForSpecs(g, controller, source, sourceObjectID, ability.Targets)
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
	for _, spec := range specs {
		normalized := normalizeTargetSpec(spec)
		if !targetSpecValid(normalized) {
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
	spec := normalizeTargetSpec(specs[specIndex])
	if !targetSpecValid(spec) {
		return
	}
	if targetSpecUsesExternalChooser(spec) {
		if len(choosingOpponentsForTargetSpec(g, controller, source, sourceObjectID, spec)) == 0 {
			return
		}
		next := append(append([]game.Target(nil), prefix...), game.DeferredTarget())
		appendTargetChoicesForSpec(g, controller, source, sourceObjectID, specs, specIndex+1, next, result)
		return
	}
	candidates := targetCandidatesForSpec(g, controller, source, sourceObjectID, spec)
	maxTargets := min(spec.MaxTargets, len(candidates))
	for _, count := range targetCountsForChoices(spec.MinTargets, maxTargets) {
		for _, combination := range targetCombinations(candidates, count) {
			next := append(append([]game.Target(nil), prefix...), combination...)
			appendTargetChoicesForSpec(g, controller, source, sourceObjectID, specs, specIndex+1, next, result)
		}
	}
}

func targetCountsForChoices(minTargets int, maxTargets int) []int {
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

func targetCandidatesForSpec(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, spec game.TargetSpec) []game.Target {
	return targetCandidatesForSpecChosenBy(g, controller, controller, source, sourceObjectID, spec)
}

func targetCandidatesForSpecChosenBy(g *game.Game, sourceController game.PlayerID, predicatePlayer game.PlayerID, source *game.CardDef, sourceObjectID id.ID, spec game.TargetSpec) []game.Target {
	var candidates []game.Target
	if targetSpecAllowsPlayers(spec) {
		for playerID := game.Player1; playerID < game.NumPlayers; playerID++ {
			target := game.PlayerTarget(playerID)
			if targetMatchesSpec(g, predicatePlayer, sourceObjectID, spec, target) && !targetProtectedFromSource(g, sourceController, source, target) {
				candidates = append(candidates, target)
			}
		}
	}
	if targetSpecAllowsPermanents(spec) {
		for _, permanent := range g.Battlefield {
			target := game.PermanentTarget(permanent.ObjectID)
			if targetMatchesSpec(g, predicatePlayer, sourceObjectID, spec, target) && !targetProtectedFromSource(g, sourceController, source, target) {
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

func targetsValidForAbility(g *game.Game, controller game.PlayerID, ability *game.AbilityDef, targets []game.Target) bool {
	return targetsValidForAbilityFromSource(g, controller, nil, ability, targets)
}

func targetsValidForAbilityFromSource(g *game.Game, controller game.PlayerID, source *game.CardDef, ability *game.AbilityDef, targets []game.Target) bool {
	return targetsValidForAbilityFromSourceObject(g, controller, source, 0, ability, targets)
}

func targetsValidForAbilityFromSourceObject(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, ability *game.AbilityDef, targets []game.Target) bool {
	if ability == nil {
		return len(targets) == 0
	}
	return targetsValidForSpecs(g, controller, source, sourceObjectID, ability.Targets, targets)
}

func targetsValidForSpecs(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, specs []game.TargetSpec, targets []game.Target) bool {
	if len(specs) == 0 {
		return len(targets) == 0
	}
	return targetsValidForSpecFrom(g, controller, source, sourceObjectID, specs, targets, 0, 0)
}

func targetsValidForSpecFrom(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, specs []game.TargetSpec, targets []game.Target, specIndex int, targetIndex int) bool {
	if specIndex == len(specs) {
		return targetIndex == len(targets)
	}
	spec := normalizeTargetSpec(specs[specIndex])
	if !targetSpecValid(spec) {
		return false
	}
	remaining := len(targets) - targetIndex
	maxTargets := min(spec.MaxTargets, remaining)
	for count := spec.MinTargets; count <= maxTargets; count++ {
		slice := targets[targetIndex : targetIndex+count]
		if targetsMatchSpecSlice(g, controller, source, sourceObjectID, spec, slice) &&
			targetsValidForSpecFrom(g, controller, source, sourceObjectID, specs, targets, specIndex+1, targetIndex+count) {
			return true
		}
	}
	return false
}

func targetsMatchSpecSlice(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, spec game.TargetSpec, targets []game.Target) bool {
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
		if seen[target] || !targetMatchesSpec(g, controller, sourceObjectID, spec, target) || targetProtectedFromSource(g, controller, source, target) {
			return false
		}
		seen[target] = true
	}
	return true
}

func spellHasAnyLegalTargets(g *game.Game, card *game.CardDef, controller game.PlayerID, chosenModes []int, targets []game.Target) bool {
	specs := spellTargetSpecs(card, chosenModes)
	return hasAnyLegalTargetForSpecs(g, controller, card, 0, specs, targets)
}

func abilityHasAnyLegalTargets(g *game.Game, ability *game.AbilityDef, controller game.PlayerID, targets []game.Target) bool {
	return abilityHasAnyLegalTargetsFromSource(g, nil, ability, controller, targets)
}

func abilityHasAnyLegalTargetsFromSource(g *game.Game, source *game.CardDef, ability *game.AbilityDef, controller game.PlayerID, targets []game.Target) bool {
	return abilityHasAnyLegalTargetsFromSourceObject(g, source, 0, ability, controller, targets)
}

func abilityHasAnyLegalTargetsFromSourceObject(g *game.Game, source *game.CardDef, sourceObjectID id.ID, ability *game.AbilityDef, controller game.PlayerID, targets []game.Target) bool {
	if ability == nil {
		return len(targets) == 0
	}
	return hasAnyLegalTargetForSpecs(g, controller, source, sourceObjectID, ability.Targets, targets)
}

func hasAnyLegalTargetForSpecs(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, specs []game.TargetSpec, targets []game.Target) bool {
	if len(specs) == 0 {
		return true
	}
	return targetsValidForSpecs(g, controller, source, sourceObjectID, specs, targets)
}

func (e *Engine) completeSpellAnnouncementTargets(g *game.Game, controller game.PlayerID, card *game.CardDef, chosenModes []int, targets []game.Target, agents [game.NumPlayers]PlayerAgent, log *TurnLog) ([]game.Target, bool) {
	return e.completeAnnouncementTargets(g, controller, card, 0, spellTargetSpecs(card, chosenModes), targets, agents, log)
}

func (e *Engine) completeAbilityAnnouncementTargets(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, ability *game.AbilityDef, targets []game.Target, agents [game.NumPlayers]PlayerAgent, log *TurnLog) ([]game.Target, bool) {
	if ability == nil {
		return targets, len(targets) == 0
	}
	return e.completeAnnouncementTargets(g, controller, source, sourceObjectID, ability.Targets, targets, agents, log)
}

func (e *Engine) completeAnnouncementTargets(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, specs []game.TargetSpec, targets []game.Target, agents [game.NumPlayers]PlayerAgent, log *TurnLog) ([]game.Target, bool) {
	if !targetSpecsUseExternalChooser(specs) {
		return append([]game.Target(nil), targets...), true
	}
	if !targetSpecsUseFixedSlots(specs) || len(targets) != len(specs) {
		return nil, false
	}
	completed := append([]game.Target(nil), targets...)
	for i, rawSpec := range specs {
		spec := normalizeTargetSpec(rawSpec)
		if !targetSpecUsesExternalChooser(spec) {
			continue
		}
		target, ok := e.chooseExternalTarget(g, controller, source, sourceObjectID, spec, agents, log)
		if !ok {
			return nil, false
		}
		completed[i] = target
	}
	return completed, targetsValidForSpecs(g, controller, source, sourceObjectID, specs, completed)
}

func targetSpecsUseExternalChooser(specs []game.TargetSpec) bool {
	return slices.ContainsFunc(specs, func(spec game.TargetSpec) bool {
		return targetSpecUsesExternalChooser(normalizeTargetSpec(spec))
	})
}

func targetSpecsUseFixedSlots(specs []game.TargetSpec) bool {
	// External chooser completion maps each TargetSpec to one target slot. Keep
	// variable regular target groups out of this path until a second consumer
	// needs full segmentation support.
	for _, spec := range specs {
		normalized := normalizeTargetSpec(spec)
		if normalized.MinTargets != 1 || normalized.MaxTargets != 1 {
			return false
		}
	}
	return true
}

func (e *Engine) chooseExternalTarget(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, spec game.TargetSpec, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (game.Target, bool) {
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

func (e *Engine) chooseTargetingOpponent(g *game.Game, controller game.PlayerID, spec game.TargetSpec, opponents []game.PlayerID, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (game.PlayerID, bool) {
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

func (e *Engine) chooseTargetFromCandidates(g *game.Game, chooser game.PlayerID, spec game.TargetSpec, candidates []game.Target, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (game.Target, bool) {
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
		return []game.TargetSpec{enchantTargetSpecForCard(card)}
	}
	ability, ok := firstSpellAbility(card)
	if !ok {
		return nil
	}
	if len(ability.Modes) > 0 {
		if !modesValidForAbility(ability, chosenModes) {
			return nil
		}
		var specs []game.TargetSpec
		for _, modeIndex := range chosenModes {
			specs = append(specs, ability.Modes[modeIndex].Targets...)
		}
		return specs
	}
	return ability.Targets
}

func modeChoicesForSpell(card *game.CardDef) [][]int {
	ability, _ := firstSpellAbility(card)
	return modeChoicesForAbility(ability)
}

func modeChoicesForAbility(ability *game.AbilityDef) [][]int {
	if ability == nil || len(ability.Modes) == 0 {
		return [][]int{nil}
	}
	// Modal choices are made before targets/costs are finalized and are locked
	// into the stack object (CR 601.2d, CR 700.2).
	minModes, maxModes := modeChoiceRange(ability)
	if minModes < 0 || maxModes < minModes || maxModes > len(ability.Modes) {
		return nil
	}
	if ability.AllowDuplicateModes {
		return duplicateModeChoices(len(ability.Modes), minModes, maxModes)
	}
	var choices [][]int
	for count := minModes; count <= maxModes; count++ {
		choices = append(choices, modeCombinations(len(ability.Modes), count)...)
	}
	return choices
}

func modesValidForSpell(card *game.CardDef, chosenModes []int) bool {
	ability, ok := firstSpellAbility(card)
	if !ok {
		return len(chosenModes) == 0
	}
	return modesValidForAbility(ability, chosenModes)
}

func modesValidForAbility(ability *game.AbilityDef, chosenModes []int) bool {
	if ability == nil {
		return len(chosenModes) == 0
	}
	if len(ability.Modes) == 0 {
		return len(chosenModes) == 0
	}
	minModes, maxModes := modeChoiceRange(ability)
	if len(chosenModes) < minModes || len(chosenModes) > maxModes {
		return false
	}
	seen := make(map[int]bool, len(chosenModes))
	for i, modeIndex := range chosenModes {
		if modeIndex < 0 || modeIndex >= len(ability.Modes) {
			return false
		}
		if i > 0 && chosenModes[i-1] > modeIndex {
			return false
		}
		// Canonical nondecreasing order avoids representing the same modal
		// choice multiple ways while preserving duplicate-mode templates that
		// explicitly permit repeats (CR 700.2d).
		if !ability.AllowDuplicateModes {
			if seen[modeIndex] {
				return false
			}
			seen[modeIndex] = true
		}
	}
	return true
}

func modeChoiceRange(ability *game.AbilityDef) (int, int) {
	if ability == nil || len(ability.Modes) == 0 {
		return 0, 0
	}
	minModes := ability.MinModes
	maxModes := ability.MaxModes
	if minModes == 0 && maxModes == 0 {
		return 1, 1
	}
	return minModes, maxModes
}

func modeCombinations(modeCount int, chooseCount int) [][]int {
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

func duplicateModeChoices(modeCount int, minModes int, maxModes int) [][]int {
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

func normalizeTargetSpec(spec game.TargetSpec) game.TargetSpec {
	if spec.MinTargets == 0 && spec.MaxTargets == 0 && spec.Constraint != "" {
		spec.MinTargets = 1
		spec.MaxTargets = 1
	}
	return spec
}

func targetSpecValid(spec game.TargetSpec) bool {
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

func targetSpecUsesExternalChooser(spec game.TargetSpec) bool {
	return spec.Chooser != game.TargetChooserController
}

func choosingOpponentsForTargetSpec(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, spec game.TargetSpec) []game.PlayerID {
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

func externalChooserCouldChooseTarget(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, spec game.TargetSpec, target game.Target) bool {
	for _, chooser := range choosingOpponentsForTargetSpec(g, controller, source, sourceObjectID, spec) {
		if slices.Contains(targetCandidatesForSpecChosenBy(g, controller, chooser, source, sourceObjectID, spec), target) {
			return true
		}
	}
	return false
}

func targetSpecAllowsPlayers(spec game.TargetSpec) bool {
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

func targetSpecAllowsPermanents(spec game.TargetSpec) bool {
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

func targetMatchesSpec(g *game.Game, controller game.PlayerID, sourceObjectID id.ID, spec game.TargetSpec, target game.Target) bool {
	switch target.Kind {
	case game.TargetPlayer:
		return playerTargetMatchesSpec(g, controller, spec, target.PlayerID)
	case game.TargetPermanent:
		return permanentTargetMatchesSpec(g, controller, sourceObjectID, spec, target.PermanentID)
	default:
		return false
	}
}

func playerTargetMatchesSpec(g *game.Game, controller game.PlayerID, spec game.TargetSpec, playerID game.PlayerID) bool {
	if !isPlayerAlive(g, playerID) || !targetSpecAllowsPlayers(spec) {
		return false
	}
	switch spec.Predicate.Player {
	case game.PlayerYou:
		return playerID == controller
	case game.PlayerOpponent, game.PlayerNotYou:
		return playerID != controller
	}
	normalized := normalizedTargetConstraint(spec)
	if strings.Contains(normalized, "opponent") && playerID == controller {
		return false
	}
	return true
}

func permanentTargetMatchesSpec(g *game.Game, controller game.PlayerID, sourceObjectID id.ID, spec game.TargetSpec, permanentID id.ID) bool {
	if !targetSpecAllowsPermanents(spec) {
		return false
	}
	permanent, ok := permanentByObjectID(g, permanentID)
	if !ok || permanent.PhasedOut {
		return false
	}
	if spec.Predicate.Another && sourceObjectID != 0 && permanent.ObjectID == sourceObjectID {
		return false
	}
	if !permanentControllerMatchesSpec(g, controller, spec, permanent) {
		return false
	}
	if !structuredPermanentPredicateMatches(g, spec.Predicate, permanent) {
		return false
	}
	if normalizedTargetConstraint(spec) == "any target" {
		return permanentHasType(g, permanent, types.Creature) ||
			permanentHasType(g, permanent, types.Planeswalker) ||
			permanentHasType(g, permanent, types.Battle)
	}
	return permanentTypeMatchesSpec(g, spec, permanent)
}

func structuredPermanentPredicateMatches(g *game.Game, predicate game.TargetPredicate, permanent *game.Permanent) bool {
	if len(predicate.PermanentTypes) > 0 && !slices.ContainsFunc(predicate.PermanentTypes, func(cardType types.Card) bool {
		return permanentHasType(g, permanent, cardType)
	}) {
		return false
	}
	if slices.ContainsFunc(predicate.ExcludedTypes, func(cardType types.Card) bool {
		return permanentHasType(g, permanent, cardType)
	}) {
		return false
	}
	colors := permanentEffectiveColors(g, permanent)
	if len(predicate.Colors) > 0 && !slices.ContainsFunc(predicate.Colors, func(color mana.Color) bool {
		return slices.Contains(colors, color)
	}) {
		return false
	}
	if slices.ContainsFunc(predicate.ExcludedColors, func(color mana.Color) bool {
		return slices.Contains(colors, color)
	}) {
		return false
	}
	if predicate.Tapped == game.TriTrue && !permanent.Tapped {
		return false
	}
	if predicate.Tapped == game.TriFalse && permanent.Tapped {
		return false
	}
	if !combatStateMatches(g, permanent, predicate.CombatState) {
		return false
	}
	if predicate.Keyword != game.KeywordNone && !hasKeyword(g, permanent, predicate.Keyword) {
		return false
	}
	if predicate.ExcludedKeyword != game.KeywordNone && hasKeyword(g, permanent, predicate.ExcludedKeyword) {
		return false
	}
	if predicate.ManaValue.Exists {
		def, ok := permanentCardDef(g, permanent)
		if !ok || !predicate.ManaValue.Val.Matches(def.ManaValue) {
			return false
		}
	}
	if predicate.Power.Exists && !predicate.Power.Val.Matches(effectivePower(g, permanent)) {
		return false
	}
	if predicate.Toughness.Exists {
		toughness, ok := effectiveToughness(g, permanent)
		if !ok || !predicate.Toughness.Val.Matches(toughness) {
			return false
		}
	}
	return true
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

func targetProtectedFromSource(g *game.Game, controller game.PlayerID, source *game.CardDef, target game.Target) bool {
	if target.Kind != game.TargetPermanent {
		return false
	}
	permanent, ok := permanentByObjectID(g, target.PermanentID)
	if !ok {
		return false
	}
	if hasKeyword(g, permanent, game.Hexproof) && effectiveController(g, permanent) != controller {
		return true
	}
	return source != nil && permanentProtectedFromSourceDef(g, permanent, source)
}

func permanentControllerMatchesSpec(g *game.Game, controller game.PlayerID, spec game.TargetSpec, permanent *game.Permanent) bool {
	permanentController := effectiveController(g, permanent)
	switch spec.Predicate.Controller {
	case game.ControllerYou:
		return permanentController == controller
	case game.ControllerOpponent, game.ControllerNotYou:
		return permanentController != controller && isPlayerAlive(g, permanentController)
	}
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

func permanentTypeMatchesSpec(g *game.Game, spec game.TargetSpec, permanent *game.Permanent) bool {
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

func normalizedTargetConstraint(spec game.TargetSpec) string {
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
