package rules

import (
	"cmp"
	"slices"
	"strconv"
	"strings"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

type mustBlockSearchResult struct {
	satisfied    int
	declarations []game.BlockDeclaration
}

// maximumSatisfiedMustBlockRequirements assigns blocker capacity to required
// attackers optimally. A blocker with additional-block capacity may satisfy more
// than one requirement, while every requirement consumes distinct blockers.
func maximumSatisfiedMustBlockRequirements(g *game.Game, required map[id.ID]int, blockers []*game.Permanent) (int, []game.BlockDeclaration) {
	if len(required) == 0 || len(blockers) == 0 {
		return 0, nil
	}
	if unitMustBlockRequirements(required) {
		return maximumUnitMustBlockMatching(g, required, blockers, nil)
	}
	attackerIDs := make([]id.ID, 0, len(required))
	for attackerID := range required {
		attackerIDs = append(attackerIDs, attackerID)
	}
	slices.Sort(attackerIDs)
	legal := make([][]int, len(attackerIDs))
	for i, attackerID := range attackerIDs {
		attacker, ok := permanentByObjectID(g, attackerID)
		if !ok {
			continue
		}
		for blockerIndex, blocker := range blockers {
			if canBlockAttacker(g, blocker, attacker) {
				legal[i] = append(legal[i], blockerIndex)
			}
		}
	}
	capacities := make([]int, len(blockers))
	for i, blocker := range blockers {
		capacities[i] = blockerBlockLimit(g, blocker)
	}
	memo := make(map[string]mustBlockSearchResult)
	var search func(int) mustBlockSearchResult
	search = func(requirementIndex int) mustBlockSearchResult {
		if requirementIndex == len(attackerIDs) {
			return mustBlockSearchResult{}
		}
		key := strconv.Itoa(requirementIndex)
		for _, capacity := range capacities {
			key += "/" + strconv.Itoa(capacity)
		}
		if cached, ok := memo[key]; ok {
			return mustBlockSearchResult{
				satisfied:    cached.satisfied,
				declarations: append([]game.BlockDeclaration(nil), cached.declarations...),
			}
		}
		best := search(requirementIndex + 1)
		minimum := required[attackerIDs[requirementIndex]]
		chosen := make([]int, 0, minimum)
		var choose func(int)
		choose = func(start int) {
			if len(chosen) == minimum {
				for _, blockerIndex := range chosen {
					capacities[blockerIndex]--
				}
				child := search(requirementIndex + 1)
				for _, blockerIndex := range chosen {
					capacities[blockerIndex]++
				}
				if child.satisfied+1 > best.satisfied {
					declarations := make([]game.BlockDeclaration, 0, len(chosen)+len(child.declarations))
					for _, blockerIndex := range chosen {
						declarations = append(declarations, game.BlockDeclaration{
							Blocker:  blockers[blockerIndex].ObjectID,
							Blocking: attackerIDs[requirementIndex],
						})
					}
					declarations = append(declarations, child.declarations...)
					best = mustBlockSearchResult{satisfied: child.satisfied + 1, declarations: declarations}
				}
				return
			}
			for i := start; i < len(legal[requirementIndex]); i++ {
				blockerIndex := legal[requirementIndex][i]
				if capacities[blockerIndex] == 0 {
					continue
				}
				chosen = append(chosen, blockerIndex)
				choose(i + 1)
				chosen = chosen[:len(chosen)-1]
			}
		}
		choose(0)
		memo[key] = mustBlockSearchResult{
			satisfied:    best.satisfied,
			declarations: append([]game.BlockDeclaration(nil), best.declarations...),
		}
		return best
	}
	result := search(0)
	return result.satisfied, result.declarations
}

func unitMustBlockRequirements(required map[id.ID]int) bool {
	for _, minimum := range required {
		if minimum != 1 {
			return false
		}
	}
	return true
}

// maximumUnitMustBlockMatching computes a maximum bipartite b-matching for
// ordinary one-blocker requirements. Expanding each blocker's block capacity
// into slots keeps the common War-Pride case polynomial instead of exploring
// every blocker subset. When forced is non-nil, that declaration is reserved
// and the remaining matching is completed around it.
func maximumUnitMustBlockMatching(g *game.Game, required map[id.ID]int, blockers []*game.Permanent, forced *game.BlockDeclaration) (int, []game.BlockDeclaration) {
	attackerIDs := make([]id.ID, 0, len(required))
	for attackerID := range required {
		if forced == nil || attackerID != forced.Blocking {
			attackerIDs = append(attackerIDs, attackerID)
		}
	}
	slices.Sort(attackerIDs)

	capacities := make([]int, len(blockers))
	for i, blocker := range blockers {
		capacities[i] = blockerBlockLimit(g, blocker)
	}
	var declarations []game.BlockDeclaration
	if forced != nil {
		blockerIndex := -1
		for i, blocker := range blockers {
			if blocker.ObjectID == forced.Blocker {
				blockerIndex = i
				break
			}
		}
		attacker, attackerOK := permanentByObjectID(g, forced.Blocking)
		if blockerIndex < 0 || !attackerOK || required[forced.Blocking] != 1 ||
			capacities[blockerIndex] == 0 || !canBlockAttacker(g, blockers[blockerIndex], attacker) {
			return 0, nil
		}
		capacities[blockerIndex]--
		declarations = append(declarations, *forced)
	}

	var slots []int
	for blockerIndex, capacity := range capacities {
		for range min(capacity, len(attackerIDs)) {
			slots = append(slots, blockerIndex)
		}
	}
	matchedAttacker := make([]int, len(slots))
	for i := range matchedAttacker {
		matchedAttacker[i] = -1
	}
	var augment func(int, []bool) bool
	augment = func(attackerIndex int, seen []bool) bool {
		attacker, ok := permanentByObjectID(g, attackerIDs[attackerIndex])
		if !ok {
			return false
		}
		for slotIndex, blockerIndex := range slots {
			if seen[slotIndex] || !canBlockAttacker(g, blockers[blockerIndex], attacker) {
				continue
			}
			seen[slotIndex] = true
			if matchedAttacker[slotIndex] == -1 || augment(matchedAttacker[slotIndex], seen) {
				matchedAttacker[slotIndex] = attackerIndex
				return true
			}
		}
		return false
	}
	matched := len(declarations)
	for attackerIndex := range attackerIDs {
		if augment(attackerIndex, make([]bool, len(slots))) {
			matched++
		}
	}
	for slotIndex, attackerIndex := range matchedAttacker {
		if attackerIndex < 0 {
			continue
		}
		declarations = append(declarations, game.BlockDeclaration{
			Blocker:  blockers[slots[slotIndex]].ObjectID,
			Blocking: attackerIDs[attackerIndex],
		})
	}
	slices.SortFunc(declarations, func(a, b game.BlockDeclaration) int {
		if byAttacker := cmp.Compare(a.Blocking, b.Blocking); byAttacker != 0 {
			return byAttacker
		}
		return cmp.Compare(a.Blocker, b.Blocker)
	})
	return matched, declarations
}

// maximumUnitMustBlockDeclarations returns representative maximum matchings
// covering every blocker-attacker edge that can participate in an optimum.
// This exposes alternate assignments without materializing the factorial set of
// complete matchings created by a large War-Pride combat.
func maximumUnitMustBlockDeclarations(g *game.Game, required map[id.ID]int, blockers []*game.Permanent, maximum int) [][]game.BlockDeclaration {
	seen := make(map[string]bool)
	var declarations [][]game.BlockDeclaration
	add := func(candidate []game.BlockDeclaration) {
		parts := make([]string, 0, len(candidate))
		for _, declaration := range candidate {
			parts = append(parts,
				strconv.FormatUint(uint64(declaration.Blocker), 10)+":"+
					strconv.FormatUint(uint64(declaration.Blocking), 10),
			)
		}
		key := strings.Join(parts, "/")
		if !seen[key] {
			seen[key] = true
			declarations = append(declarations, candidate)
		}
	}
	_, preferred := maximumUnitMustBlockMatching(g, required, blockers, nil)
	add(preferred)
	for attackerID := range required {
		attacker, ok := permanentByObjectID(g, attackerID)
		if !ok {
			continue
		}
		for _, blocker := range blockers {
			if !canBlockAttacker(g, blocker, attacker) {
				continue
			}
			forced := game.BlockDeclaration{Blocker: blocker.ObjectID, Blocking: attackerID}
			satisfied, candidate := maximumUnitMustBlockMatching(g, required, blockers, &forced)
			if satisfied == maximum {
				add(candidate)
			}
		}
	}
	return declarations
}
