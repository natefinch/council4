package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

// TestMatchSelectionModified exercises the MatchModified predicate (Envoy of the
// Ancestors). A permanent is modified when it carries a counter or has an Aura
// or Equipment attached; an unmodified permanent must not match.
func TestMatchSelectionModified(t *testing.T) {
	board := newParityBoard(t)
	g := board.g
	modified := game.Selection{MatchModified: true}

	if matchSelectionForPermanent(g, game.Player1, modified, board.whiteCreature) {
		t.Error("an unmodified permanent must not match a MatchModified selection")
	}

	withCounter := board.greenCreatureP2
	withCounter.Counters.Add(counter.PlusOnePlusOne, 1)
	if !matchSelectionForPermanent(g, game.Player1, modified, withCounter) {
		t.Error("a permanent with a counter should match a MatchModified selection")
	}

	equipped := board.whiteCreature
	equipped.Attachments = append(equipped.Attachments, board.equipment.ObjectID)
	if !matchSelectionForPermanent(g, game.Player1, modified, equipped) {
		t.Error("a permanent with an attachment should match a MatchModified selection")
	}
}
