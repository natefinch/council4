package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerGoadTargetCreature verifies "Goad target creature." lowers to a
// single Goad primitive on the chosen target permanent, reusing the existing
// runtime goad keyword-action mechanic (CR 701.38). The reminder text is
// stripped, leaving one target and one instruction.
func TestLowerGoadTargetCreature(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Goader",
		Layout:     "normal",
		ManaCost:   "{1}{R}",
		TypeLine:   "Sorcery",
		OracleText: "Goad target creature. (Until your next turn, that creature attacks each combat if able and attacks a player other than you if able.)",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %d, want 1", len(mode.Targets))
	}
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %#v, want a single goad", mode.Sequence)
	}
	goad, ok := mode.Sequence[0].Primitive.(game.Goad)
	if !ok {
		t.Fatalf("primitive = %T, want game.Goad", mode.Sequence[0].Primitive)
	}
	if goad.Object != game.TargetPermanentReference(0) {
		t.Fatalf("goad object = %#v, want the chosen target permanent", goad.Object)
	}
}

// TestLowerGoadReferenceForm verifies goad lowers on a back-reference subject —
// "goad that creature" bound to the triggering event's related permanent, with
// no chosen target — through the referenced-permanent lowering path rather than
// the target path.
func TestLowerGoadReferenceForm(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Goad Reference",
		Layout:     "normal",
		ManaCost:   "{2}{R}",
		TypeLine:   "Creature — Goblin",
		OracleText: "Whenever this creature blocks or becomes blocked by a creature, goad that creature.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 0 {
		t.Fatalf("targets = %d, want 0 (the subject is a back-reference, not a target)", len(mode.Targets))
	}
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %#v, want a single goad", mode.Sequence)
	}
	goad, ok := mode.Sequence[0].Primitive.(game.Goad)
	if !ok {
		t.Fatalf("primitive = %T, want game.Goad", mode.Sequence[0].Primitive)
	}
	if goad.Object == game.TargetPermanentReference(0) {
		t.Fatalf("goad object = %#v, want a back-reference, not a chosen target", goad.Object)
	}
}
