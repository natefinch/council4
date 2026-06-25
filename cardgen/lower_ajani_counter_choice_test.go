package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

// TestLowerAjaniSingleChoiceCounterPlacement verifies that Ajani Fells the
// Godsire's chapter II sequence lowers its "put a vigilance counter on a
// creature you control" tail clause to an AddCounter that chooses one member of
// the controller's creatures (ChooseOne) rather than every member.
func TestLowerAjaniSingleChoiceCounterPlacement(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Ajani Fells the Godsire",
		Layout:   "saga",
		TypeLine: "Enchantment — Saga",
		OracleText: "(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)\n" +
			"I — Exile target creature an opponent controls with power 3 or greater.\n" +
			"II — Create a 2/1 white Cat Warrior creature token, then put a vigilance counter on a creature you control.\n" +
			"III — Target creature you control gains double strike until end of turn.",
	})
	if len(face.ChapterAbilities) != 3 {
		t.Fatalf("got %d chapter abilities, want 3", len(face.ChapterAbilities))
	}
	secondSeq := face.ChapterAbilities[1].Content.Modes[0].Sequence
	if len(secondSeq) != 2 {
		t.Fatalf("chapter II sequence length = %d, want 2", len(secondSeq))
	}
	add, ok := secondSeq[1].Primitive.(game.AddCounter)
	if !ok {
		t.Fatalf("chapter II primitive[1] = %T, want game.AddCounter", secondSeq[1].Primitive)
	}
	if !add.ChooseOne {
		t.Fatal("AddCounter.ChooseOne = false, want true")
	}
	if !add.Group.Valid() {
		t.Fatal("AddCounter.Group is not set")
	}
	if add.Object.Kind() != game.ObjectReferenceNone {
		t.Fatalf("AddCounter.Object = %#v, want none", add.Object)
	}
	if add.CounterKind != counter.Vigilance {
		t.Fatalf("AddCounter.CounterKind = %v, want vigilance", add.CounterKind)
	}
}
