package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerDrawReplacementDig verifies the Underrealm Lich draw-replacement dig
// lowers to a DrawCardDigReplacement carrying the look count, take count, and
// graveyard remainder, and that the card's other ability still lowers.
func TestLowerDrawReplacementDig(t *testing.T) {
	t.Parallel()
	power, toughness := "4", "3"
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Underrealm Lich",
		Layout:   "normal",
		TypeLine: "Creature — Zombie Elf Shaman",
		ManaCost: "{3}{B}{G}",
		OracleText: "If you would draw a card, instead look at the top three cards of your library, then put one into your hand and the rest into your graveyard.\n" +
			"Pay 4 life: This creature gains indestructible until end of turn. Tap it.",
		Power:     &power,
		Toughness: &toughness,
	})
	if len(face.ReplacementAbilities) != 1 {
		t.Fatalf("replacement abilities = %d, want 1", len(face.ReplacementAbilities))
	}
	repl := face.ReplacementAbilities[0].Replacement
	if repl.DrawCardDigLook != 3 {
		t.Errorf("DrawCardDigLook = %d, want 3", repl.DrawCardDigLook)
	}
	if repl.DrawCardDigTake != 1 {
		t.Errorf("DrawCardDigTake = %d, want 1", repl.DrawCardDigTake)
	}
	if repl.DrawCardDigRemainder != game.DigRemainderGraveyard {
		t.Errorf("DrawCardDigRemainder = %v, want graveyard", repl.DrawCardDigRemainder)
	}
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1 (pay-life ability)", len(face.ActivatedAbilities))
	}
}

// TestLowerDrawReplacementDigLibraryBottom verifies the library-bottom remainder
// variant lowers to the matching runtime remainder.
func TestLowerDrawReplacementDigLibraryBottom(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Dig Test",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "If you would draw a card, instead look at the top two cards of your library, then put one into your hand and the other on the bottom of your library.",
	})
	if len(face.ReplacementAbilities) != 1 {
		t.Fatalf("replacement abilities = %d, want 1", len(face.ReplacementAbilities))
	}
	repl := face.ReplacementAbilities[0].Replacement
	if repl.DrawCardDigLook != 2 || repl.DrawCardDigTake != 1 {
		t.Errorf("look/take = %d/%d, want 2/1", repl.DrawCardDigLook, repl.DrawCardDigTake)
	}
	if repl.DrawCardDigRemainder != game.DigRemainderLibraryBottom {
		t.Errorf("DrawCardDigRemainder = %v, want library bottom", repl.DrawCardDigRemainder)
	}
}
