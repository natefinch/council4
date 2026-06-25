package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerAfflictKeyword verifies the Afflict keyword expands to a
// becomes-blocked trigger whose body makes the defending player lose life. The
// life-loss subject is the defending player reference (CR 702.131), distinct
// from the spell's controller.
func TestLowerAfflictKeyword(t *testing.T) {
	t.Parallel()
	power := "2"
	toughness := "2"
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Afflicter",
		Layout:     "normal",
		TypeLine:   "Creature — Zombie",
		ManaCost:   "{3}{R}",
		OracleText: "Afflict 3 (Whenever this creature becomes blocked, defending player loses 3 life.)",
		Power:      &power,
		Toughness:  &toughness,
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	loseLife, ok := mode.Sequence[0].Primitive.(game.LoseLife)
	if !ok {
		t.Fatalf("primitive = %T, want game.LoseLife", mode.Sequence[0].Primitive)
	}
	if loseLife.Player != game.DefendingPlayerReference() {
		t.Fatalf("life loss player = %v, want DefendingPlayerReference", loseLife.Player)
	}
	if loseLife.Amount != game.Fixed(3) {
		t.Fatalf("life loss amount = %v, want Fixed(3)", loseLife.Amount)
	}
}
