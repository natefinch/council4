package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

// auraCounterTrigger lowers a single-face Aura whose body is a triggered counter
// placement on the enchanted creature and returns the lowered trigger body's
// lone AddCounter instruction.
func auraCounterTrigger(t *testing.T, body string) game.AddCounter {
	t.Helper()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Aura",
		Layout:     "normal",
		TypeLine:   "Enchantment — Aura",
		OracleText: "Enchant creature\n" + body,
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
	return add
}

// expectUnsupportedAuraCounter asserts that an Aura counter placement on the
// enchanted creature fails closed with the "unsupported counter placement"
// diagnostic and lowers no triggered ability.
func expectUnsupportedAuraCounter(t *testing.T, body string) {
	t.Helper()
	faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Aura",
		Layout:     "normal",
		TypeLine:   "Enchantment — Aura",
		OracleText: "Enchant creature\n" + body,
	})
	for i := range faces {
		if len(faces[i].TriggeredAbilities) != 0 {
			t.Fatalf("%q unexpectedly lowered a triggered ability", body)
		}
	}
	for i := range diagnostics {
		if diagnostics[i].Summary == "unsupported counter placement" {
			return
		}
	}
	t.Fatalf("diagnostics = %#v, want unsupported counter placement", diagnostics)
}

func TestLowerAttachedCounterPlacementPlusOne(t *testing.T) {
	t.Parallel()
	add := auraCounterTrigger(t, "At the beginning of your upkeep, put a +1/+1 counter on enchanted creature.")
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

func TestLowerAttachedCounterPlacementPluralMinusOne(t *testing.T) {
	t.Parallel()
	add := auraCounterTrigger(t, "When this Aura enters, put two -1/-1 counters on enchanted creature.")
	if add.Object != game.SourceAttachedPermanentReference() {
		t.Fatalf("object = %#v, want source attached permanent", add.Object)
	}
	if add.CounterKind != counter.MinusOneMinusOne {
		t.Fatalf("counter = %v, want -1/-1", add.CounterKind)
	}
	if add.Amount != game.Fixed(2) {
		t.Fatalf("amount = %#v, want fixed 2", add.Amount)
	}
}

func TestLowerAttachedCounterPlacementThatCreatureUpkeep(t *testing.T) {
	t.Parallel()
	// "that creature" in an upkeep trigger scoped to the enchanted creature's
	// controller names the enchanted creature (Unstable Mutation, Essence Flare),
	// resolved through the source attached-permanent reference.
	add := auraCounterTrigger(t, "Enchanted creature gets +3/+3.\nAt the beginning of the upkeep of enchanted creature's controller, put a -1/-1 counter on that creature.")
	if add.Object != game.SourceAttachedPermanentReference() {
		t.Fatalf("object = %#v, want source attached permanent", add.Object)
	}
	if add.CounterKind != counter.MinusOneMinusOne {
		t.Fatalf("counter = %v, want -1/-1", add.CounterKind)
	}
	if add.Amount != game.Fixed(1) {
		t.Fatalf("amount = %#v, want fixed 1", add.Amount)
	}
}

func TestLowerAttachedCounterPlacementFailsClosed(t *testing.T) {
	t.Parallel()
	// A trailing selector qualifier leaves the recipient unrecognized.
	expectUnsupportedAuraCounter(t, "At the beginning of your upkeep, put a +1/+1 counter on enchanted creature with flying.")
	// "enchanted permanent" is not the recognized recipient.
	expectUnsupportedAuraCounter(t, "At the beginning of your upkeep, put a +1/+1 counter on enchanted permanent.")
	// A player-only counter kind cannot be placed on a permanent.
	expectUnsupportedAuraCounter(t, "At the beginning of your upkeep, put a poison counter on enchanted creature.")
	// A second clause makes the counter clause inexact.
	expectUnsupportedAuraCounter(t, "At the beginning of your upkeep, put a +1/+1 counter on enchanted creature and a +1/+1 counter on this creature.")
}
