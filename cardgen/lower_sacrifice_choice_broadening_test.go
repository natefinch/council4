package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerSacrificeChoiceHighAmountSequence verifies that a spelled sacrifice
// amount above the historical cap of two ("Each player sacrifices four lands of
// their choice.", "sacrifice three creatures.") reconstructs exactly and lowers
// to a SacrificePermanents with the fixed count, so the surrounding ordered
// sequence (Wildfire's mass damage, Greed's Gambit's leaves-the-battlefield
// edict) lowers end to end. The runtime primitive already carried an unbounded
// fixed amount; only the parser round-trip capped the value.
func TestLowerSacrificeChoiceHighAmountSequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Wildfire",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Each player sacrifices four lands of their choice. Test Wildfire deals 4 damage to each creature.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability missing")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %d instructions, want 2 (sacrifice, damage)", len(mode.Sequence))
	}
	sacrifice, ok := mode.Sequence[0].Primitive.(game.SacrificePermanents)
	if !ok {
		t.Fatalf("first primitive = %T, want game.SacrificePermanents", mode.Sequence[0].Primitive)
	}
	if got := sacrifice.Amount; got != game.Fixed(4) {
		t.Fatalf("sacrifice amount = %+v, want Fixed(4)", got)
	}
	if sacrifice.PlayerGroup != game.AllPlayersReference() {
		t.Fatalf("sacrifice player group = %+v, want all players", sacrifice.PlayerGroup)
	}
}

// TestLowerSacrificeChoiceExcludedSubtype verifies the "non-<subtype>" sacrifice
// noun ("non-Zombie creature", "non-Vampire creature") reconstructs exactly and
// lowers to a SacrificePermanents whose Selection drops that creature subtype
// (Archdemon of Unx, Anowon the Ruin Sage, Dreadfeast Demon, Ruthless Winnower).
func TestLowerSacrificeChoiceExcludedSubtype(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Archdemon",
		Layout:     "normal",
		TypeLine:   "Creature",
		OracleText: "At the beginning of your upkeep, sacrifice a non-Zombie creature, then create a 2/2 black Zombie creature token.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %d instructions, want 2 (sacrifice, create token)", len(mode.Sequence))
	}
	sacrifice, ok := mode.Sequence[0].Primitive.(game.SacrificePermanents)
	if !ok {
		t.Fatalf("first primitive = %T, want game.SacrificePermanents", mode.Sequence[0].Primitive)
	}
	if sacrifice.Amount != game.Fixed(1) {
		t.Fatalf("sacrifice amount = %+v, want Fixed(1)", sacrifice.Amount)
	}
	if sacrifice.Selection.ExcludedSubtype != types.Sub("Zombie") {
		t.Fatalf("sacrifice excluded subtype = %q, want Zombie", sacrifice.Selection.ExcludedSubtype)
	}
	if len(sacrifice.Selection.RequiredTypes) != 1 || sacrifice.Selection.RequiredTypes[0] != types.Creature {
		t.Fatalf("sacrifice required types = %v, want [Creature]", sacrifice.Selection.RequiredTypes)
	}
}

// TestSacrificeChoiceExcludedSubtypeStaysUnsupported guards the fail-closed
// boundary: a sacrifice naming more than one excluded subtype has no canonical
// Oracle wording the round-trip reproduces, so it must stay unsupported rather
// than silently mislower.
func TestSacrificeChoiceExcludedSubtypeStaysUnsupported(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Twin Exclusion",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Each player sacrifices a non-Zombie non-Vampire creature of their choice.",
	})
}
