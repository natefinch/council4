package rules

import (
	"fmt"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
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
	choices      [][]game.Target
	targetCounts [][]int
	kind         targetChoiceKind
	// err is diagnostic context for invalid card-definition input. Production
	// enumeration currently treats invalid specs as unavailable actions/triggers.
	err error
}

func targetChoicesForSpell(g *game.Game, controller game.PlayerID, card *game.CardDef, chosenModes []int, branch game.CastBranch) targetChoiceResult {
	kickerCount := 0
	if branch.Kicked {
		kickerCount = 1
	}
	return targetChoicesForSpellWithKickerCount(g, controller, card, chosenModes, branch, kickerCount)
}

func targetChoicesForSpellWithKickerCount(g *game.Game, controller game.PlayerID, card *game.CardDef, chosenModes []int, branch game.CastBranch, kickerCount int) targetChoiceResult {
	specs := spellTargetSpecs(card, chosenModes, branch)
	for i := range specs {
		if !specs[i].CountEqualsKickerPlusOne {
			continue
		}
		count := kickerCount + 1
		specs[i].MinTargets = count
		specs[i].MaxTargets = count
	}
	return targetChoicesForSpecs(g, controller, card, 0, game.Event{}, specs)
}

func targetChoicesForBody(g *game.Game, controller game.PlayerID, body game.Ability) targetChoiceResult {
	return targetChoicesForBodyFromSource(g, controller, nil, body)
}

func targetChoicesForBodyFromSource(g *game.Game, controller game.PlayerID, source *game.CardDef, body game.Ability) targetChoiceResult {
	return targetChoicesForBodyFromSourceObject(g, controller, source, 0, body)
}

func targetChoicesForBodyFromSourceObject(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, body game.Ability) targetChoiceResult {
	return targetChoicesForBodyFromSourceObjectWithModes(g, controller, source, sourceObjectID, game.Event{}, body, nil)
}

func targetChoicesForBodyFromSourceObjectWithModes(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, triggerEvent game.Event, body game.Ability, chosenModes []int) targetChoiceResult {
	if body == nil {
		return targetChoiceResult{kind: targetNoTargetsRequired, choices: [][]game.Target{nil}, targetCounts: [][]int{nil}}
	}
	if !modesValidForBody(body, chosenModes) {
		return targetChoiceResult{
			kind: targetInvalidSpec,
			err:  fmt.Errorf("ability has invalid mode selection %v", chosenModes),
		}
	}
	return targetChoicesForSpecs(g, controller, source, sourceObjectID, triggerEvent, bodyTargetSpecs(body, chosenModes))
}

// targetChoicesForSpecs enumerates every legal target combination for specs.
// Returns an explicit result kind so callers never infer state from nil-slice shape:
//   - targetNoTargetsRequired when specs is empty
//   - targetLegalChoicesFound when at least one combination is legal (including optional no-target choices)
//   - targetNoLegalChoices when specs are valid but no board-legal combination exists
//   - targetInvalidSpec (with err) when a spec has an invalid min/max range
//
// targetChoicesForSpecs enumerates every legal combination of targets for a
// spell or ability's target specs (CR 115.1: targets are declared as the spell or
// ability is put on the stack; CR 601.2c: a target is announced for each instance
// of the word "target"). Each spec corresponds to one "target" instance; only
// legal targets are offered (CR 115.2, 115.4). The same object can be chosen once
// per instance of "target" (CR 115.3); "another target" specs additionally exclude
// the earlier targets of the same spell or ability, enforcing the card's "another"
// criterion.
func targetChoicesForSpecs(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, triggerEvent game.Event, specs []game.TargetSpec) targetChoiceResult {
	if len(specs) == 0 {
		return targetChoiceResult{kind: targetNoTargetsRequired, choices: [][]game.Target{nil}, targetCounts: [][]int{nil}}
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
	var targetCounts [][]int
	appendTargetChoicesForSpec(g, controller, source, sourceObjectID, triggerEvent, specs, 0, nil, nil, &result, &targetCounts)
	if len(result) == 0 {
		return targetChoiceResult{kind: targetNoLegalChoices}
	}
	for i, targets := range result {
		if _, unique := uniqueTargetCountsForSpecs(g, controller, source, sourceObjectID, triggerEvent, specs, targets); unique {
			targetCounts[i] = nil
		}
	}
	return targetChoiceResult{kind: targetLegalChoicesFound, choices: result, targetCounts: targetCounts}
}

func appendTargetChoicesForSpec(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, triggerEvent game.Event, specs []game.TargetSpec, specIndex int, prefix []game.Target, countPrefix []int, result *[][]game.Target, targetCounts *[][]int) {
	if specIndex >= len(specs) {
		*result = append(*result, append([]game.Target(nil), prefix...))
		*targetCounts = append(*targetCounts, append([]int(nil), countPrefix...))
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
		nextCounts := append(append([]int(nil), countPrefix...), 1)
		appendTargetChoicesForSpec(g, controller, source, sourceObjectID, triggerEvent, specs, specIndex+1, next, nextCounts, result, targetCounts)
		return
	}
	candidates := targetCandidatesForSpec(g, controller, source, sourceObjectID, triggerEvent, &spec)
	if spec.DistinctFromPriorTargets {
		candidates = filterTargetsDistinctFrom(candidates, prefix)
	}
	maxTargets := min(spec.MaxTargets, len(candidates))
	for _, count := range targetCountsForChoices(spec.MinTargets, maxTargets) {
		for _, combination := range targetCombinations(candidates, count) {
			if !targetsSatisfySameGraveyard(g, &spec, combination) {
				continue
			}
			next := append(append([]game.Target(nil), prefix...), combination...)
			nextCounts := append(append([]int(nil), countPrefix...), count)
			appendTargetChoicesForSpec(g, controller, source, sourceObjectID, triggerEvent, specs, specIndex+1, next, nextCounts, result, targetCounts)
		}
	}
}

// targetsSatisfySameGraveyard reports whether targets meet a SameGraveyard spec's
// one-graveyard restriction. A card in a graveyard is in its owner's graveyard
// (CR 404.2), so the restriction holds exactly when every card target shares one
// owner. Non-card targets and specs without the restriction always pass.
func targetsSatisfySameGraveyard(g *game.Game, spec *game.TargetSpec, targets []game.Target) bool {
	if !spec.SameGraveyard {
		return true
	}
	owner := game.PlayerID(-1)
	for _, target := range targets {
		if target.Kind != game.TargetCard {
			continue
		}
		card, ok := g.GetCardInstance(target.CardID)
		if !ok {
			return false
		}
		if owner == game.PlayerID(-1) {
			owner = card.Owner
			continue
		}
		if card.Owner != owner {
			return false
		}
	}
	return true
}

// filterTargetsDistinctFrom drops any candidate that points at the same game
// object as a target already chosen in prior, supporting "another target ..."
// specs that must differ from the earlier targets of the same spell or ability.
// CR 115.3: the same target can't be chosen multiple times for any one instance
// of the word "target".
func filterTargetsDistinctFrom(candidates, prior []game.Target) []game.Target {
	var kept []game.Target
	for _, candidate := range candidates {
		distinct := true
		for _, used := range prior {
			if sameTargetObject(candidate, used) {
				distinct = false
				break
			}
		}
		if distinct {
			kept = append(kept, candidate)
		}
	}
	return kept
}

// sameTargetObject reports whether two targets point at the same game object.
func sameTargetObject(a, b game.Target) bool {
	if a.Kind != b.Kind {
		return false
	}
	switch a.Kind {
	case game.TargetPermanent:
		return a.PermanentID == b.PermanentID
	case game.TargetPlayer:
		return a.PlayerID == b.PlayerID
	case game.TargetStackObject:
		return a.StackObjectID == b.StackObjectID
	case game.TargetCard:
		return a.CardID == b.CardID
	default:
		// TargetDeferred has no concrete object identity to compare.
	}
	return false
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

func targetCandidatesForSpec(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, triggerEvent game.Event, spec *game.TargetSpec) []game.Target {
	return targetCandidatesForSpecChosenBy(g, controller, controller, source, sourceObjectID, triggerEvent, spec)
}

// targetCandidatesForSpecChosenBy returns every object or player that is a legal
// target for one target spec, evaluating the spec's predicate as predicatePlayer.
// Only permanents are candidates unless the spec allows players, stack objects, or
// cards in other zones (CR 115.2, 115.4), and a candidate the source can't legally
// target (e.g. protection, CR 702.16b, or hexproof, CR 702.11b-c) is excluded.
func targetCandidatesForSpecChosenBy(g *game.Game, sourceController, predicatePlayer game.PlayerID, source *game.CardDef, sourceObjectID id.ID, triggerEvent game.Event, spec *game.TargetSpec) []game.Target {
	var candidates []game.Target
	if targetSpecAllowsPlayers(spec) {
		for playerID := range game.PlayerID(game.NumPlayers) {
			target := game.PlayerTarget(playerID)
			if targetMatchesSpec(g, predicatePlayer, sourceObjectID, triggerEvent, spec, target) && !targetProtectedFromSource(g, sourceController, source, sourceObjectID, target) {
				candidates = append(candidates, target)
			}
		}
	}
	if targetSpecAllowsPermanents(spec) {
		for _, permanent := range g.Battlefield {
			target := game.PermanentTarget(permanent.ObjectID)
			if targetMatchesSpec(g, predicatePlayer, sourceObjectID, triggerEvent, spec, target) && !targetProtectedFromSource(g, sourceController, source, sourceObjectID, target) {
				candidates = append(candidates, target)
			}
		}
	}
	if targetSpecAllowsStackObjects(spec) {
		for _, obj := range g.Stack.Objects() {
			target := game.StackObjectTarget(obj.ID)
			if targetMatchesSpec(g, predicatePlayer, sourceObjectID, triggerEvent, spec, target) {
				candidates = append(candidates, target)
			}
		}
	}
	if targetSpecAllowsCards(spec) {
		for _, card := range g.CardInstances {
			target := game.CardTargetWithZoneVersion(card.ID, card.ZoneVersion)
			if targetMatchesSpec(g, predicatePlayer, sourceObjectID, triggerEvent, spec, target) {
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

func (e *Engine) completeSpellAnnouncementTargets(g *game.Game, controller game.PlayerID, card *game.CardDef, chosenModes []int, targets []game.Target, agents [game.NumPlayers]PlayerAgent, log *TurnLog, branch game.CastBranch) ([]game.Target, bool) {
	return e.completeAnnouncementTargets(g, controller, card, 0, spellTargetSpecs(card, chosenModes, branch), targets, agents, log)
}

func (e *Engine) completeAbilityAnnouncementTargets(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, body game.Ability, targets []game.Target, agents [game.NumPlayers]PlayerAgent, log *TurnLog) ([]game.Target, bool) {
	return e.completeAbilityAnnouncementTargetsWithModes(g, controller, source, sourceObjectID, body, nil, targets, agents, log)
}

func (e *Engine) completeAbilityAnnouncementTargetsWithModes(g *game.Game, controller game.PlayerID, source *game.CardDef, sourceObjectID id.ID, body game.Ability, chosenModes []int, targets []game.Target, agents [game.NumPlayers]PlayerAgent, log *TurnLog) ([]game.Target, bool) {
	if body == nil {
		return targets, len(targets) == 0
	}
	if !modesValidForBody(body, chosenModes) {
		return nil, false
	}
	return e.completeAnnouncementTargets(g, controller, source, sourceObjectID, bodyTargetSpecs(body, chosenModes), targets, agents, log)
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
		targets := targetCandidatesForSpecChosenBy(g, controller, opponent, source, sourceObjectID, game.Event{}, spec)
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
		options = append(options, game.ChoiceOption{
			Index:   i,
			Label:   fmt.Sprintf("Player %d", opponent+1),
			Targets: []game.Target{game.PlayerTarget(opponent)},
		})
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
