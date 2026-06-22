package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestMatchSelectionCommander exercises the MatchCommander predicate (Bastion
// Protector). A permanent matches only when its underlying card instance is
// recorded as a commander in Game.CommanderIDs; a non-commander permanent must
// not match.
func TestMatchSelectionCommander(t *testing.T) {
	board := newParityBoard(t)
	g := board.g
	commanderSelection := game.Selection{MatchCommander: true}

	if matchSelectionForPermanent(g, game.Player1, commanderSelection, board.whiteCreature) {
		t.Error("a non-commander permanent must not match a MatchCommander selection")
	}

	g.CommanderIDs[board.whiteCreature.CardInstanceID] = true
	if !matchSelectionForPermanent(g, game.Player1, commanderSelection, board.whiteCreature) {
		t.Error("a commander permanent should match a MatchCommander selection")
	}

	if matchSelectionForPermanent(g, game.Player1, commanderSelection, board.greenCreatureP2) {
		t.Error("a different permanent must not match while only the commander is registered")
	}
}
