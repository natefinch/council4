package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

func targetsValidForSpell(g *game.Game, controller game.PlayerID, card *game.CardDef, chosenModes []int, targets []game.Target, branch game.CastBranch) bool {
	specs := spellTargetSpecs(card, chosenModes, branch)
	return targetsValidForSpecs(g, controller, card, 0, specs, targets)
}

func targetsValidForBody(g *game.Game, controller game.PlayerID, body game.Ability, targets []game.Target) bool {
	return targetsValidForBodyFromSource(g, controller, nil, body, targets)
}

func targetsValidForBodyFromSource(g *game.Game, controller game.PlayerID, source *game.CardDef, body game.Ability, targets []game.Target) bool {
	return targetsValidForBodyFromSourceObject(g, controller, source, 0, body, targets)
}

func targetsValidForBodyFromSourceObject(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, body game.Ability, targets []game.Target) bool {
	return targetsValidForBodyFromSourceObjectWithModes(g, controller, source, sourceObjectID, body, nil, targets)
}

func targetsValidForBodyFromSourceObjectWithModes(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, body game.Ability, chosenModes []int, targets []game.Target) bool {
	if body == nil {
		return len(targets) == 0
	}
	if !modesValidForBody(body, chosenModes) {
		return false
	}
	return targetsValidForSpecs(g, controller, source, sourceObjectID, bodyTargetSpecs(body, chosenModes), targets)
}

// targetsValidForSpecs reports whether a set of chosen targets is a legal
// selection for a spell or ability's target specs at announcement time (CR 115,
// CR 601.2c): the right number of legal targets is chosen for each instance of
// the word "target", with no target repeated within one instance (CR 115.3).
func targetsValidForSpecs(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, specs []game.TargetSpec, targets []game.Target) bool {
	_, ok := targetCountsForSpecs(g, controller, source, sourceObjectID, game.Event{}, specs, targets)
	return ok
}

func targetCountsForSpecs(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, triggerEvent game.Event, specs []game.TargetSpec, targets []game.Target) ([]int, bool) {
	if len(specs) == 0 {
		return nil, len(targets) == 0
	}
	counts := make([]int, len(specs))
	if !targetCountsForSpecFrom(g, controller, source, sourceObjectID, triggerEvent, specs, targets, counts, 0, 0) {
		return nil, false
	}
	return counts, true
}

func targetCountsForSpecFrom(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, triggerEvent game.Event, specs []game.TargetSpec, targets []game.Target, counts []int, specIndex, targetIndex int) bool {
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
		if targetsMatchSpecSlice(g, controller, source, sourceObjectID, triggerEvent, &spec, slice) &&
			targetSliceDistinctFromPrior(&spec, targets[:targetIndex], slice) &&
			targetCountsForSpecFrom(g, controller, source, sourceObjectID, triggerEvent, specs, targets, counts, specIndex+1, targetIndex+count) {
			counts[specIndex] = count
			return true
		}
	}
	return false
}

func spellTargetCounts(g *game.Game, controller game.PlayerID, card *game.CardDef, chosenModes []int, targets []game.Target, branch game.CastBranch) ([]int, bool) {
	return targetCountsForSpecs(g, controller, card, 0, game.Event{}, spellTargetSpecs(card, chosenModes, branch), targets)
}

// spellTargetCountsMatchX reports whether every CountEqualsX target spec of the
// spell has exactly xValue targets chosen for it. A CountEqualsX spec binds its
// resolving target count to the spell's chosen X ("Exile X target creatures"), so
// a cast is legal only when the announced target count for that spec equals X.
// Spells without a CountEqualsX spec are unaffected.
func spellTargetCountsMatchX(g *game.Game, controller game.PlayerID, card *game.CardDef, chosenModes []int, targets []game.Target, xValue int, branch game.CastBranch) bool {
	specs := spellTargetSpecs(card, chosenModes, branch)
	requiresMatch := false
	for i := range specs {
		if specs[i].CountEqualsX {
			requiresMatch = true
			break
		}
	}
	if !requiresMatch {
		return true
	}
	counts, ok := spellTargetCounts(g, controller, card, chosenModes, targets, branch)
	if !ok {
		return false
	}
	for i := range specs {
		if specs[i].CountEqualsX && (i >= len(counts) || counts[i] != xValue) {
			return false
		}
	}
	return true
}

// spellTargetsSatisfyManaValueX reports whether every target chosen for a
// ManaValueAtMostX spec of the spell has mana value at most the spell's chosen X
// ("target creature with mana value X or less", Dominate). The Selection matcher
// is X-blind, so announcement over-generates every creature and this check
// enforces the X-derived upper bound at cast time (CR 601.2c). Spells without a
// ManaValueAtMostX spec are unaffected.
func spellTargetsSatisfyManaValueX(g *game.Game, controller game.PlayerID, card *game.CardDef, chosenModes []int, targets []game.Target, xValue int, branch game.CastBranch) bool {
	specs := spellTargetSpecs(card, chosenModes, branch)
	requiresMatch := false
	for i := range specs {
		if specs[i].ManaValueAtMostX {
			requiresMatch = true
			break
		}
	}
	if !requiresMatch {
		return true
	}
	counts, ok := spellTargetCounts(g, controller, card, chosenModes, targets, branch)
	if !ok {
		return false
	}
	targetIndex := 0
	for i := range specs {
		count := 0
		if i < len(counts) {
			count = counts[i]
		}
		if targetIndex+count > len(targets) {
			return false
		}
		slice := targets[targetIndex : targetIndex+count]
		targetIndex += count
		if !specs[i].ManaValueAtMostX {
			continue
		}
		for j := range slice {
			if !targetManaValueAtMost(g, slice[j], xValue) {
				return false
			}
		}
	}
	return true
}

// targetManaValueAtMost reports whether a permanent target's mana value is at
// most bound. Only permanent targets carry a mana value, so any other target
// kind, or a permanent whose card definition is unavailable, fails closed: a
// ManaValueAtMostX bound never silently admits an object it cannot measure. It
// reads mana value the same way the Selection matcher does (permanentCardDef →
// ManaValue), so the X-derived bound and a fixed Selection.ManaValue bound treat
// tokens, copies, and face-down permanents identically.
func targetManaValueAtMost(g *game.Game, target game.Target, bound int) bool {
	if target.Kind != game.TargetPermanent {
		return false
	}
	permanent, ok := permanentByObjectID(g, target.PermanentID)
	if !ok {
		return false
	}
	def, ok := permanentCardDef(g, permanent)
	if !ok {
		return false
	}
	return def.ManaValue() <= bound
}

func bodyTargetCounts(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, body game.Ability, targets []game.Target) ([]int, bool) {
	return bodyTargetCountsWithModes(g, controller, source, sourceObjectID, game.Event{}, body, nil, targets)
}

func bodyTargetCountsWithModes(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, triggerEvent game.Event, body game.Ability, chosenModes []int, targets []game.Target) ([]int, bool) {
	if body == nil {
		return nil, len(targets) == 0
	}
	if !modesValidForBody(body, chosenModes) {
		return nil, false
	}
	return targetCountsForSpecs(g, controller, source, sourceObjectID, triggerEvent, bodyTargetSpecs(body, chosenModes), targets)
}

func bodyTargetCountsWithModesAndRecorded(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, body game.Ability, chosenModes, recorded []int, targets []game.Target) ([]int, bool) {
	if body == nil {
		return nil, len(recorded) == 0 && len(targets) == 0
	}
	if !modesValidForBody(body, chosenModes) {
		return nil, false
	}
	specs := bodyTargetSpecs(body, chosenModes)
	if len(recorded) > 0 {
		if !recordedTargetCountsValidForSpecs(g, controller, source, sourceObjectID, specs, recorded, targets) {
			return nil, false
		}
		return append([]int(nil), recorded...), true
	}
	return uniqueTargetCountsForSpecs(g, controller, source, sourceObjectID, game.Event{}, specs, targets)
}

func recordedTargetCountsValidForSpecs(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, specs []game.TargetSpec, recorded []int, targets []game.Target) bool {
	if len(recorded) != len(specs) {
		return false
	}
	targetIndex := 0
	for i := range specs {
		spec := normalizeTargetSpec(&specs[i])
		count := recorded[i]
		if !targetSpecValid(&spec) ||
			count < spec.MinTargets ||
			count > spec.MaxTargets ||
			targetIndex+count > len(targets) ||
			!targetsMatchSpecSlice(g, controller, source, sourceObjectID, game.Event{}, &spec, targets[targetIndex:targetIndex+count]) ||
			!targetSliceDistinctFromPrior(&spec, targets[:targetIndex], targets[targetIndex:targetIndex+count]) {
			return false
		}
		targetIndex += count
	}
	return targetIndex == len(targets)
}

func uniqueTargetCountsForSpecs(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, triggerEvent game.Event, specs []game.TargetSpec, targets []game.Target) ([]int, bool) {
	var matches [][]int
	appendMatchingTargetCounts(g, controller, source, sourceObjectID, triggerEvent, specs, targets, 0, 0, nil, &matches)
	if len(matches) != 1 {
		return nil, false
	}
	return matches[0], true
}

func appendMatchingTargetCounts(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, triggerEvent game.Event, specs []game.TargetSpec, targets []game.Target, specIndex, targetIndex int, prefix []int, matches *[][]int) {
	if len(*matches) > 1 {
		return
	}
	if specIndex == len(specs) {
		if targetIndex == len(targets) {
			*matches = append(*matches, append([]int(nil), prefix...))
		}
		return
	}
	spec := normalizeTargetSpec(&specs[specIndex])
	if !targetSpecValid(&spec) {
		return
	}
	maxTargets := min(spec.MaxTargets, len(targets)-targetIndex)
	for count := spec.MinTargets; count <= maxTargets; count++ {
		slice := targets[targetIndex : targetIndex+count]
		if !targetsMatchSpecSlice(g, controller, source, sourceObjectID, triggerEvent, &spec, slice) ||
			!targetSliceDistinctFromPrior(&spec, targets[:targetIndex], slice) {
			continue
		}
		next := append(append([]int(nil), prefix...), count)
		appendMatchingTargetCounts(g, controller, source, sourceObjectID, triggerEvent, specs, targets, specIndex+1, targetIndex+count, next, matches)
	}
}

// targetSliceDistinctFromPrior reports whether a spec's chosen targets satisfy
// its DistinctFromPriorTargets requirement: when set, none of the slice's
// objects may equal an object already chosen for an earlier spec ("... another
// target creature"). It is a no-op for the default unset case.
func targetSliceDistinctFromPrior(spec *game.TargetSpec, prior, slice []game.Target) bool {
	if !spec.DistinctFromPriorTargets {
		return true
	}
	for _, chosen := range slice {
		for _, used := range prior {
			if sameTargetObject(chosen, used) {
				return false
			}
		}
	}
	return true
}

func targetsMatchSpecSlice(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, triggerEvent game.Event, spec *game.TargetSpec, targets []game.Target) bool {
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
		if seen[target] || !targetMatchesSpec(g, controller, sourceObjectID, triggerEvent, spec, target) || targetProtectedFromSource(g, controller, source, sourceObjectID, target) {
			return false
		}
		seen[target] = true
	}
	return targetsSatisfySameGraveyard(g, spec, targets)
}

func spellHasAnyLegalTargets(g *game.Game, card *game.CardDef, obj *game.StackObject) bool {
	if obj.Overloaded && card.Overload.Exists {
		card = overloadSpellDef(card)
	}
	return stackObjectHasAnyLegalTargetsForSpecs(g, card, 0, spellTargetSpecs(card, obj.ChosenModes, castBranchForObject(obj)), obj)
}

func bodyHasAnyLegalTargetsFromSourceObject(g *game.Game, source *game.CardDef, sourceObjectID id.ID, body game.Ability, obj *game.StackObject) bool {
	if body == nil {
		return len(obj.Targets) == 0
	}
	if !modesValidForBody(body, obj.ChosenModes) {
		return false
	}
	return stackObjectHasAnyLegalTargetsForSpecs(g, source, sourceObjectID, bodyTargetSpecs(body, obj.ChosenModes), obj)
}

// stackObjectHasAnyLegalTargetsForSpecs rechecks the legality of a resolving
// spell or ability's chosen targets (CR 608.2b). Each target is re-evaluated
// against its spec; targets that are no longer legal are replaced with a deferred
// marker so the resolution code skips its effect on them, and the spell or ability
// resolves only if at least one target is still legal. If every target is illegal
// the caller does not resolve it (CR 608.2b: it is removed from the stack and, if
// a spell, put into its owner's graveyard).
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
	// A resolving triggered ability carries its triggering event on the stack
	// object; thread it into resolution-time legality so "that player" style
	// predicates (Garland, Royal Kidnapper) resolve against the opponent who
	// caused the trigger rather than failing closed. Spells and events without a
	// trigger event contribute a zero event, matching prior behaviour.
	var triggerEvent game.Event
	if obj.HasTriggerEvent {
		triggerEvent = obj.TriggerEvent
	}
	anyLegal := false
	targetIndex := 0
	for specIndex := range specs {
		spec := normalizeTargetSpec(&specs[specIndex])
		for range counts[specIndex] {
			target := targets[targetIndex]
			if targetLegalForSpecAtResolution(g, obj.Controller, source, sourceObjectID, triggerEvent, &spec, target) &&
				(!spec.ManaValueAtMostX || targetManaValueAtMost(g, target, obj.XValue)) {
				anyLegal = true
			} else {
				targets[targetIndex] = game.DeferredTargetFrom(target)
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

// targetLegalForSpecAtResolution reports whether a single target is still a legal
// target for its spec at resolution (CR 608.2b): it must still match the spec's
// predicate and the source must still be able to target it (e.g. it has not gained
// protection, CR 702.16b, or hexproof, CR 702.11b-c). A target that has left the
// zone it was in when targeted no longer matches its spec and so is illegal.
func targetLegalForSpecAtResolution(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, triggerEvent game.Event, spec *game.TargetSpec, target game.Target) bool {
	if targetSpecUsesExternalChooser(spec) {
		return externalChooserCouldChooseTarget(g, controller, source, sourceObjectID, spec, target)
	}
	return targetMatchesSpec(g, controller, sourceObjectID, triggerEvent, spec, target) &&
		!targetProtectedFromSource(g, controller, source, sourceObjectID, target)
}

// spellTargetSpecs returns the target specs a spell announces targets against on
// the given cast branch. Specs gated to a cast branch that is not active
// (an unpromised gift's promised-only target, an unkicked spell's kicked-only
// target) are neutralised so they require no target, while their ordinal slot is
// preserved (CR 601.2c). Ungated spells return their full spec list unchanged.
func spellTargetSpecs(card *game.CardDef, chosenModes []int, branch game.CastBranch) []game.TargetSpec {
	return applyCastBranchToSpecs(spellTargetSpecsRaw(card, chosenModes), branch)
}

func spellTargetSpecsRaw(card *game.CardDef, chosenModes []int) []game.TargetSpec {
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
		if !lockedModesValidForContent(content, chosenModes) {
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
	return game.BodyTargets(ability)
}

func bodyTargetSpecs(body game.Ability, chosenModes []int) []game.TargetSpec {
	if body == nil {
		return nil
	}
	content := game.BodyContent(body)
	if !content.IsModal() {
		return game.BodyTargets(body)
	}
	specs := append([]game.TargetSpec(nil), content.SharedTargets...)
	for _, modeIndex := range chosenModes {
		specs = append(specs, content.Modes[modeIndex].Targets...)
	}
	return specs
}

func modeChoicesForSpell(card *game.CardDef) [][]int {
	ability, _ := firstSpellAbility(card)
	if ability == nil {
		return [][]int{nil}
	}
	return modeChoicesForContent(*ability)
}

func modeChoicesForSpellAt(g *game.Game, playerID game.PlayerID, card *game.CardDef) [][]int {
	ability, _ := firstSpellAbility(card)
	if ability == nil {
		return [][]int{nil}
	}
	minModes, maxModes := modeChoiceRangeFromContentAt(g, playerID, *ability)
	return modeChoicesForContentRange(*ability, minModes, maxModes)
}

func modeChoicesForBody(body game.Ability) [][]int {
	if body == nil {
		return [][]int{nil}
	}
	return modeChoicesForContent(game.BodyContent(body))
}

func modeChoicesForContent(content game.AbilityContent) [][]int {
	minModes, maxModes := modeChoiceRangeFromContent(content)
	return modeChoicesForContentRange(content, minModes, maxModes)
}

func modeChoicesForContentRange(content game.AbilityContent, minModes, maxModes int) [][]int {
	if len(content.Modes) == 0 || !content.IsModal() {
		return [][]int{nil}
	}
	// Modal choices are made before targets/costs are finalized and are locked
	// into the stack object (CR 601.2d, CR 700.2).
	if !modeChoiceRangeValid(content, minModes, maxModes) {
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

func modesValidForSpellAt(g *game.Game, playerID game.PlayerID, card *game.CardDef, chosenModes []int) bool {
	ability, ok := firstSpellAbility(card)
	if !ok {
		return len(chosenModes) == 0
	}
	minModes, maxModes := modeChoiceRangeFromContentAt(g, playerID, *ability)
	return modesValidForContentRange(*ability, chosenModes, minModes, maxModes)
}

func modesValidForBody(body game.Ability, chosenModes []int) bool {
	if body == nil {
		return len(chosenModes) == 0
	}
	return modesValidForContent(game.BodyContent(body), chosenModes)
}

func modesValidForContent(content game.AbilityContent, chosenModes []int) bool {
	minModes, maxModes := modeChoiceRangeFromContent(content)
	return modesValidForContentRange(content, chosenModes, minModes, maxModes)
}

func lockedModesValidForContent(content game.AbilityContent, chosenModes []int) bool {
	minModes, maxModes := modeChoiceRangeFromContent(content)
	maxModes += content.ModeChoiceBonus.AdditionalMaxModes
	return modesValidForContentRange(content, chosenModes, minModes, maxModes)
}

func modesValidForContentRange(content game.AbilityContent, chosenModes []int, minModes, maxModes int) bool {
	if len(content.Modes) == 0 || !content.IsModal() {
		return len(chosenModes) == 0
	}
	if !modeChoiceRangeValid(content, minModes, maxModes) {
		return false
	}
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

func modeChoiceRangeValid(content game.AbilityContent, minModes, maxModes int) bool {
	return minModes >= 0 &&
		maxModes >= minModes &&
		(content.AllowDuplicateModes || maxModes <= len(content.Modes))
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

func modeChoiceRangeFromContentAt(g *game.Game, playerID game.PlayerID, content game.AbilityContent) (minModes, maxModes int) {
	minModes, maxModes = modeChoiceRangeFromContent(content)
	bonus := content.ModeChoiceBonus
	if bonus.Condition == game.ModeChoiceConditionControlsCommander && playerControlsCommander(g, playerID) {
		maxModes += bonus.AdditionalMaxModes
	}
	return minModes, maxModes
}

func playerControlsCommander(g *game.Game, playerID game.PlayerID) bool {
	for _, permanent := range g.Battlefield {
		if !permanent.PhasedOut &&
			effectiveController(g, permanent) == playerID &&
			permanentContainsCommander(g, permanent) {
			return true
		}
	}
	return false
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
