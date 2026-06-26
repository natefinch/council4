package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

// TestLowerEquippedCounterPlacement proves an Equipment counter placement on the
// equipped creature lowers onto the source's attached-permanent reference, the
// same reference the Aura "enchanted creature" recipient uses.
func TestLowerEquippedCounterPlacement(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Equipment",
		Layout:     "normal",
		TypeLine:   "Artifact — Equipment",
		OracleText: "Whenever equipped creature deals combat damage, put a +1/+1 counter on equipped creature.\nEquip {2}",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	modes := face.TriggeredAbilities[0].Content.Modes
	if len(modes) != 1 || len(modes[0].Sequence) != 1 {
		t.Fatalf("trigger body = %#v, want one instruction", modes)
	}
	add, ok := modes[0].Sequence[0].Primitive.(game.AddCounter)
	if !ok {
		t.Fatalf("instruction = %#v, want AddCounter", modes[0].Sequence[0].Primitive)
	}
	if add.Object != game.SourceAttachedPermanentReference() {
		t.Fatalf("object = %#v, want source attached permanent", add.Object)
	}
	if add.CounterKind != counter.PlusOnePlusOne {
		t.Fatalf("counter = %v, want +1/+1", add.CounterKind)
	}
	if add.Amount != game.Fixed(1) {
		t.Fatalf("amount = %#v, want fixed 1", add.Amount)
	}
}

// TestLowerForEachSelfCounterPlacement proves the "for each <group>" dynamic
// count form lowers a self counter placement with a non-fixed amount; the count
// is supplied by the shared dynamic-amount lowerer rather than a literal.
func TestLowerForEachSelfCounterPlacement(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Beast",
		Layout:     "normal",
		TypeLine:   "Creature — Beast",
		OracleText: "When this creature enters, put a +1/+1 counter on this creature for each creature you control.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	modes := face.TriggeredAbilities[0].Content.Modes
	if len(modes) != 1 || len(modes[0].Sequence) != 1 {
		t.Fatalf("trigger body = %#v, want one instruction", modes)
	}
	add, ok := modes[0].Sequence[0].Primitive.(game.AddCounter)
	if !ok {
		t.Fatalf("instruction = %#v, want AddCounter", modes[0].Sequence[0].Primitive)
	}
	if add.CounterKind != counter.PlusOnePlusOne {
		t.Fatalf("counter = %v, want +1/+1", add.CounterKind)
	}
	// The for-each count resolves through the shared dynamic-amount lowerer, so
	// the placement amount must be dynamic rather than a literal one.
	if !add.Amount.IsDynamic() {
		t.Fatalf("amount = %#v, want dynamic for-each count", add.Amount)
	}
}
