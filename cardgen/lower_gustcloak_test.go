package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerGustcloakOptionalUntapAndRemoveFromCombat verifies the optional
// two-effect sequence applies one choice to both instructions.
func TestLowerGustcloakOptionalUntapAndRemoveFromCombat(t *testing.T) {
	t.Parallel()
	const oracleText = "Whenever this creature becomes blocked, you may untap it and remove it from combat."
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Gustcloak Runner",
		Layout:     "normal",
		TypeLine:   "Creature — Human Soldier",
		OracleText: oracleText,
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	ability := face.TriggeredAbilities[0]
	if !ability.Optional {
		t.Fatal("triggered ability is not optional")
	}
	sequence := ability.Content.Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("sequence length = %d, want 2", len(sequence))
	}
	untap, ok := sequence[0].Primitive.(game.Untap)
	if !ok || untap.Object != game.EventPermanentReference() {
		t.Fatalf("instruction 0 = %#v, want event-permanent untap", sequence[0].Primitive)
	}
	remove, ok := sequence[1].Primitive.(game.RemoveFromCombat)
	if !ok || remove.Object != game.EventPermanentReference() {
		t.Fatalf("instruction 1 = %#v, want event-permanent remove-from-combat", sequence[1].Primitive)
	}
}
