package cardgen

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
)

// Selection's dynamic MV evaluator has no resolving stack object. Remapping
// such an anchor cannot make a chosen-target bound executable.
func searchSelectionHasTargetBound(selection game.Selection) bool {
	if slices.ContainsFunc(selection.AnyOf, searchSelectionHasTargetBound) {
		return true
	}
	if !selection.ManaValueDynamic.Exists || selection.ManaValueDynamic.Val.Group == nil {
		return false
	}
	group := *selection.ManaValueDynamic.Val.Group
	if anchor, ok := group.Anchor(); ok && objectReferenceCarriesTargetIndex(anchor) {
		return true
	}
	if exclude, ok := group.Exclusion(); ok && objectReferenceCarriesTargetIndex(exclude) {
		return true
	}
	if player, ok := group.PlayerAnchor(); ok {
		if player.Kind() == game.PlayerReferenceTargetPlayer {
			return true
		}
		if object, ok := player.Object(); ok && objectReferenceCarriesTargetIndex(object) {
			return true
		}
	}
	if players, ok := group.PlayerGroup(); ok {
		if players.Kind == game.PlayerGroupReferenceTargetedPlayers ||
			players.ControlsMatching != nil && searchSelectionHasTargetBound(*players.ControlsMatching) {
			return true
		}
	}
	return searchSelectionHasTargetBound(group.Selection())
}

func searchSpecHasTargetBound(spec game.SearchSpec) bool {
	if searchSelectionHasTargetBound(spec.Filter) {
		return true
	}
	return slices.ContainsFunc(spec.SlotFilters, searchSelectionHasTargetBound)
}
