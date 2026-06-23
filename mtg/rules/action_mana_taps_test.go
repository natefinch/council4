package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestRecordActionManaTapsCapturesTappedLands(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	forest := addBasicLandWithManaPermanent(g, game.Player1, "Forest", types.Forest, mana.G)

	log := &TurnLog{}
	log.addAction(&ActionLog{Player: game.Player1, Action: action.Pass()})
	entryIndex := lastEntryIndex(log)
	eventsBefore := len(g.Events)

	// A mana payment taps the land for mana via this helper.
	setPermanentTappedForMana(g, forest)

	recordActionManaTaps(g, log, entryIndex, eventsBefore)

	taps := log.Entries[entryIndex].Action.ManaTaps
	if len(taps) != 1 || taps[0].Source != "Forest" {
		t.Fatalf("ManaTaps = %#v, want one Forest tap", taps)
	}
}

func TestRecordActionManaTapsIgnoresNonManaTaps(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCreaturePermanentWithSupertype(g, game.Player1)

	log := &TurnLog{}
	log.addAction(&ActionLog{Player: game.Player1, Action: action.Pass()})
	entryIndex := lastEntryIndex(log)
	eventsBefore := len(g.Events)

	// Tapping that is not for mana (e.g. attacking) must not be recorded.
	setPermanentTapped(g, creature, true)

	recordActionManaTaps(g, log, entryIndex, eventsBefore)

	if taps := log.Entries[entryIndex].Action.ManaTaps; len(taps) != 0 {
		t.Fatalf("ManaTaps = %#v, want none", taps)
	}
}
