package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerRevealTopPartitionGraveyardRemainder verifies the Mulch shape
// ("Reveal the top N cards of your library. Put all land cards revealed this way
// into your hand and the rest into your graveyard.") lowers to a single
// RevealTopPartition primitive whose filter selects lands and whose remainder is
// the graveyard.
func TestLowerRevealTopPartitionGraveyardRemainder(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Mulch",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Reveal the top four cards of your library. Put all land cards revealed this way into your hand and the rest into your graveyard.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 0 || len(mode.Sequence) != 1 {
		t.Fatalf("mode = %+v, want no targets and one instruction", mode)
	}
	partition, ok := mode.Sequence[0].Primitive.(game.RevealTopPartition)
	if !ok {
		t.Fatalf("primitive = %T, want game.RevealTopPartition", mode.Sequence[0].Primitive)
	}
	if partition.Amount != game.Fixed(4) {
		t.Fatalf("partition amount = %+v, want 4", partition.Amount)
	}
	if partition.Remainder != game.DigRemainderGraveyard {
		t.Fatalf("partition remainder = %v, want graveyard", partition.Remainder)
	}
	if partition.Player != game.ControllerReference() {
		t.Fatalf("partition player = %+v, want controller", partition.Player)
	}
	if len(partition.Selection.RequiredTypes) != 1 || partition.Selection.RequiredTypes[0] != types.Land {
		t.Fatalf("partition selection = %+v, want required type Land", partition.Selection)
	}
}

// TestLowerRevealTopPartitionLibraryBottomRemainder verifies the Goblin
// Ringleader shape, where the typed filter selects a subtype and the remainder
// goes to the bottom of the library.
func TestLowerRevealTopPartitionLibraryBottomRemainder(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Goblin Ringleader",
		Layout:     "normal",
		TypeLine:   "Creature — Goblin",
		OracleText: "When this creature enters, reveal the top four cards of your library. Put all Goblin cards revealed this way into your hand and the rest on the bottom of your library in any order.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	sequence := face.TriggeredAbilities[0].Content.Modes[0].Sequence
	if len(sequence) != 1 {
		t.Fatalf("sequence length = %d, want 1", len(sequence))
	}
	partition, ok := sequence[0].Primitive.(game.RevealTopPartition)
	if !ok {
		t.Fatalf("primitive = %T, want game.RevealTopPartition", sequence[0].Primitive)
	}
	if partition.Amount != game.Fixed(4) {
		t.Fatalf("partition amount = %+v, want 4", partition.Amount)
	}
	if partition.Remainder != game.DigRemainderLibraryBottom {
		t.Fatalf("partition remainder = %v, want library bottom", partition.Remainder)
	}
	if len(partition.Selection.SubtypesAny) != 1 || partition.Selection.SubtypesAny[0] != types.Sub("Goblin") {
		t.Fatalf("partition selection = %+v, want subtype Goblin", partition.Selection)
	}
}
