package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerMendingShuffleGraveyardIntoLibrary verifies that The Mending of
// Dominaria's chapter III sequence lowers its "shuffle your graveyard into your
// library" tail clause to a ShuffleGraveyardIntoLibrary primitive targeting the
// controller, following the mass land return.
func TestLowerMendingShuffleGraveyardIntoLibrary(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "The Mending of Dominaria",
		Layout:   "saga",
		TypeLine: "Enchantment — Saga",
		OracleText: "(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)\n" +
			"I, II — Mill two cards, then you may return a creature card from your graveyard to your hand.\n" +
			"III — Return all land cards from your graveyard to the battlefield, then shuffle your graveyard into your library.",
	})
	if len(face.ChapterAbilities) != 2 {
		t.Fatalf("got %d chapter abilities, want 2", len(face.ChapterAbilities))
	}
	thirdSeq := face.ChapterAbilities[1].Content.Modes[0].Sequence
	if len(thirdSeq) != 2 {
		t.Fatalf("chapter III sequence length = %d, want 2", len(thirdSeq))
	}
	shuffle, ok := thirdSeq[1].Primitive.(game.ShuffleGraveyardIntoLibrary)
	if !ok {
		t.Fatalf("chapter III primitive[1] = %T, want game.ShuffleGraveyardIntoLibrary", thirdSeq[1].Primitive)
	}
	if shuffle.Player != game.ControllerReference() {
		t.Fatalf("shuffle player = %#v, want controller", shuffle.Player)
	}
}
