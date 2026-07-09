package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestMatchSelectionPowerToughnessTarget proves the "target N/M creature" filter
// (Pendelhaven, Aegis of the Meek) matches only a creature whose current power
// and toughness both equal the pinned values. A 1/1 creature matches, while a
// 5/5 creature (both differ) and a 0/4 creature (only toughness differs) fail
// closed, so the ability can never reach a non-1/1 creature.
func TestMatchSelectionPowerToughnessTarget(t *testing.T) {
	board := newParityBoard(t)
	g := board.g

	onePowerOneToughness := game.Selection{
		RequiredTypesAny: []types.Card{types.Creature},
		Power:            opt.Val(compare.Int{Op: compare.Equal, Value: 1}),
		Toughness:        opt.Val(compare.Int{Op: compare.Equal, Value: 1}),
	}

	if !matchSelectionForPermanent(g, game.Player1, onePowerOneToughness, board.whiteCreature) {
		t.Error("a 1/1 creature must match a target 1/1 creature filter")
	}
	if matchSelectionForPermanent(g, game.Player1, onePowerOneToughness, board.greenCreatureP2) {
		t.Error("a 5/5 creature must not match a target 1/1 creature filter")
	}
	if matchSelectionForPermanent(g, game.Player1, onePowerOneToughness, board.artifactCreature) {
		t.Error("a 0/4 creature must not match a target 1/1 creature filter (toughness differs)")
	}
}
