package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLowerRemoveFromCombatSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Recall",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{W}",
		OracleText: "Remove target creature from combat.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("expected a spell ability")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %#v", mode.Sequence)
	}
	remove, ok := mode.Sequence[0].Primitive.(game.RemoveFromCombat)
	if !ok || remove.Object != game.TargetPermanentReference(0) {
		t.Fatalf("primitive = %#v", mode.Sequence[0].Primitive)
	}
}

func TestLowerReconnaissanceRemoveFromCombatAndUntap(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Reconnaissance",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "{0}: Remove target attacking creature you control from combat and untap it.",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	mode := face.ActivatedAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v", mode.Sequence)
	}
	remove, ok := mode.Sequence[0].Primitive.(game.RemoveFromCombat)
	if !ok || remove.Object != game.TargetPermanentReference(0) {
		t.Fatalf("instruction 0 = %#v", mode.Sequence[0].Primitive)
	}
	untap, ok := mode.Sequence[1].Primitive.(game.Untap)
	if !ok || untap.Object != game.TargetPermanentReference(0) {
		t.Fatalf("instruction 1 = %#v", mode.Sequence[1].Primitive)
	}
}

// TestLowerSelfRemoveFromCombat covers the source/back-reference remove-from-
// combat form ("remove it from combat" / "Remove this creature from combat"),
// where the creature taken out of combat is the ability's own source and the
// clause names no target. The parser marks the self form exact and the shared
// referenced-permanent path lowers it to the source permanent reference, the
// same routing tap/untap use for their self forms.
func TestLowerSelfRemoveFromCombat(t *testing.T) {
	t.Parallel()
	for _, oracle := range []string{
		"Whenever this creature attacks, remove this creature from combat.",
		"Whenever this creature attacks, tap this creature and remove it from combat.",
	} {
		face := lowerSingleFace(t, &ScryfallCard{
			Name:       "Test Skirmisher",
			Layout:     "normal",
			TypeLine:   "Creature — Ogre",
			ManaCost:   "{1}{R}",
			OracleText: oracle,
		})
		if len(face.TriggeredAbilities) != 1 {
			t.Fatalf("%q: triggered abilities = %d, want 1", oracle, len(face.TriggeredAbilities))
		}
		mode := face.TriggeredAbilities[0].Content.Modes[0]
		remove, ok := mode.Sequence[len(mode.Sequence)-1].Primitive.(game.RemoveFromCombat)
		if !ok || remove.Object != game.SourcePermanentReference() {
			t.Fatalf("%q: remove primitive = %#v", oracle, mode.Sequence[len(mode.Sequence)-1].Primitive)
		}
	}
}
