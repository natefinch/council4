package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerLookAtTargetPlayerLibrary lowers "look at the top card of target
// player's library." to a player-targeted LookAtLibraryTop naming the target
// player, with the PublishLinked invariant satisfied.
func TestLowerLookAtTargetPlayerLibrary(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Library Peek",
		Layout:     "normal",
		TypeLine:   "Creature — Merfolk Rogue",
		OracleText: "When this creature enters, look at the top card of target player's library.",
	})
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 1 || mode.Targets[0].Allow != game.TargetAllowPlayer {
		t.Fatalf("targets = %+v, want one player target", mode.Targets)
	}
	look, ok := mode.Sequence[0].Primitive.(game.LookAtLibraryTop)
	if !ok {
		t.Fatalf("primitive = %T, want game.LookAtLibraryTop", mode.Sequence[0].Primitive)
	}
	if look.Player != game.TargetPlayerReference(0) {
		t.Fatalf("player = %#v, want target player 0", look.Player)
	}
	if look.PublishLinked == "" {
		t.Fatal("LookAtLibraryTop missing PublishLinked key")
	}
}
