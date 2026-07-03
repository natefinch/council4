package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerLookAtHandRemapsSharedPlayer verifies a "look at target player's hand"
// clause composes in an ordered sequence (Urza's Bauble): the LookAtHand
// primitive addresses the clause's target player and the sequence lowers rather
// than failing to remap the LookAtHand player reference.
func TestLowerLookAtHandRemapsSharedPlayer(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bauble",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "{T}, Sacrifice this artifact: Look at a card at random in target player's hand. You draw a card at the beginning of the next turn's upkeep.",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("got %d activated abilities, want 1", len(face.ActivatedAbilities))
	}
	mode := face.ActivatedAbilities[0].Content.Modes[0]
	var look *game.LookAtHand
	for i := range mode.Sequence {
		if prim, ok := mode.Sequence[i].Primitive.(game.LookAtHand); ok {
			look = &prim
		}
	}
	if look == nil {
		t.Fatalf("sequence has no LookAtHand instruction: %+v", mode.Sequence)
	}
	if look.Player != game.TargetPlayerReference(0) {
		t.Fatalf("LookAtHand player = %+v, want the target player 0", look.Player)
	}
}
