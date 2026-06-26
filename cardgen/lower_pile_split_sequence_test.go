package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestLowerPileSplitOpponentSeparates verifies the Fact or Fiction shape
// ("Reveal the top N cards of your library. An opponent separates those cards
// into two piles. Put one pile into your hand and the other into your
// graveyard.") lowers to a single PileSplit primitive where the opponent
// separates the piles and the controller chooses which to keep.
func TestLowerPileSplitOpponentSeparates(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Fact or Fiction",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Reveal the top five cards of your library. An opponent separates those cards into two piles. Put one pile into your hand and the other into your graveyard.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 0 || len(mode.Sequence) != 1 {
		t.Fatalf("mode = %+v, want no targets and one instruction", mode)
	}
	split, ok := mode.Sequence[0].Primitive.(game.PileSplit)
	if !ok {
		t.Fatalf("primitive = %T, want game.PileSplit", mode.Sequence[0].Primitive)
	}
	if split.Amount != game.Fixed(5) {
		t.Fatalf("split amount = %+v, want 5", split.Amount)
	}
	if !split.SeparatorOpponent || split.ChooserOpponent {
		t.Fatalf("split roles = %+v, want opponent separates, controller chooses", split)
	}
	if split.Kept != zone.Hand || split.Other != zone.Graveyard {
		t.Fatalf("split zones = kept %v other %v, want hand/graveyard", split.Kept, split.Other)
	}
	if split.Player != game.ControllerReference() {
		t.Fatalf("split player = %+v, want controller", split.Player)
	}
}

// TestLowerPileSplitOpponentChooses verifies the Steam Augury shape ("Reveal the
// top N cards of your library and separate them into two piles. An opponent
// chooses one of those piles. Put that pile into your hand and the other into
// your graveyard.") lowers to a PileSplit where the controller separates and the
// opponent chooses.
func TestLowerPileSplitOpponentChooses(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Steam Augury",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Reveal the top five cards of your library and separate them into two piles. An opponent chooses one of those piles. Put that pile into your hand and the other into your graveyard.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	split, ok := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.PileSplit)
	if !ok {
		t.Fatalf("primitive = %T, want game.PileSplit", face.SpellAbility.Val.Modes[0].Sequence[0].Primitive)
	}
	if split.Amount != game.Fixed(5) {
		t.Fatalf("split amount = %+v, want 5", split.Amount)
	}
	if split.SeparatorOpponent || !split.ChooserOpponent {
		t.Fatalf("split roles = %+v, want controller separates, opponent chooses", split)
	}
}

// TestLowerPileSplitTriggered verifies the pile-split sequence lowers inside a
// triggered ability (Sphinx of Uthuun's enter trigger).
func TestLowerPileSplitTriggered(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Sphinx",
		Layout:     "normal",
		TypeLine:   "Creature — Sphinx",
		OracleText: "When Test Sphinx enters, reveal the top five cards of your library. An opponent separates those cards into two piles. Put one pile into your hand and the other into your graveyard.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	split, ok := face.TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.PileSplit)
	if !ok {
		t.Fatalf("primitive = %T, want game.PileSplit", face.TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive)
	}
	if split.Amount != game.Fixed(5) || !split.SeparatorOpponent {
		t.Fatalf("split = %+v, want amount 5 opponent separates", split)
	}
}

// TestLowerPileSplitFailsClosed verifies pile-split shapes the PileSplit
// primitive does not model stay unsupported: a variable "X plus one" reveal
// count (Epiphany at the Drownyard) and an inconsistent role pairing.
func TestLowerPileSplitFailsClosed(t *testing.T) {
	t.Parallel()
	rejected := []string{
		"Reveal the top X plus one cards of your library and separate them into two piles. An opponent chooses one of those piles. Put that pile into your hand and the other into your graveyard.",
		"Reveal the top five cards of your library. An opponent chooses one of those piles. Put one pile into your hand and the other into your graveyard.",
	}
	for _, text := range rejected {
		faces, _ := lowerExecutableFaces(&ScryfallCard{
			Name:       "Test Reject",
			Layout:     "normal",
			TypeLine:   "Instant",
			OracleText: text,
		})
		for _, face := range faces {
			if face.SpellAbility.Exists {
				t.Errorf("OracleText %q lowered a spell ability, want fail closed", text)
			}
		}
	}
}
