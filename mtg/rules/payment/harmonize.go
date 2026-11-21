package payment

import (
	"maps"
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

var harmonizeTapCost = cost.Additional{
	Kind: cost.AdditionalTapPermanents, Amount: 1,
	MatchPermanentType: true, PermanentType: types.Creature,
	Text: "Tap the chosen creature for Harmonize",
}

// Announcing the alternative includes choosing its tap subject (CR 702.180b).
// Expand after other cost choices so each combination has a unique payment
// index, while keeping the original declaration index on every combination.
func expandHarmonizeOptions(s State, player game.PlayerID, options []spellCostOption) []spellCostOption {
	if !slices.ContainsFunc(options, func(option spellCostOption) bool {
		return option.castPermission == SpellCastPermissionHarmonize
	}) {
		return options
	}
	creatures := CandidatePermanentsForCost(s, player, harmonizeTapCost, nil)
	nextIndex := 0
	for _, option := range options {
		nextIndex = max(nextIndex, option.index+1)
	}
	var expanded []spellCostOption
	for _, option := range options {
		if option.castPermission != SpellCastPermissionHarmonize {
			expanded = append(expanded, option)
			continue
		}
		option.harmonizeCreature = opt.Val(id.ID(0))
		expanded = append(expanded, option)
		for _, creature := range creatures {
			option.index = nextIndex
			nextIndex++
			option.harmonizeCreature = opt.Val(creature.ObjectID)
			expanded = append(expanded, option)
		}
	}
	return expanded
}

// The selected creature is not another choice in the payment window. Fix the
// last component to its exact object ID so normal resource reservation and
// payment handle the tap, even when other tap costs are present.
func (o spellCostOption) fixedComponents(fixed map[int]ComponentChoice) (map[int]ComponentChoice, bool) {
	creature := o.harmonizeCreature.Val
	if creature == 0 {
		return fixed, true
	}
	index := len(o.additionalCosts) - 1
	choice := fixed[index]
	if len(choice.Counters) != 0 || len(choice.Objects) > 1 ||
		len(choice.Objects) == 1 && choice.Objects[0] != creature ||
		choice.Finished && len(choice.Objects) != 1 {
		return nil, false
	}
	fixed = maps.Clone(fixed)
	if fixed == nil {
		fixed = make(map[int]ComponentChoice, 1)
	}
	fixed[index] = ComponentChoice{Objects: []id.ID{creature}, Finished: true}
	return fixed, true
}
