package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerTargetCounterTriggeringLifeGained proves a single-target counter
// placement whose count reads the triggering quantity ("put that many +1/+1
// counters on target creature") lowers inside a life-gain trigger, scaling with
// the life gained (Treebeard, Gracious Host). The amount resolves to the
// triggering life change rather than a fixed count.
func TestLowerTargetCounterTriggeringLifeGained(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Treant",
		Layout:     "normal",
		TypeLine:   "Creature — Treefolk",
		ManaCost:   "{2}{G}{W}",
		OracleText: "Whenever you gain life, put that many +1/+1 counters on target creature.",
		Power:      new("3"),
		Toughness:  new("5"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %d, want 1", len(mode.Targets))
	}
	add, ok := mode.Sequence[0].Primitive.(game.AddCounter)
	if !ok {
		t.Fatalf("primitive = %T, want game.AddCounter", mode.Sequence[0].Primitive)
	}
	dynamic := add.Amount.DynamicAmount()
	if !dynamic.Exists || dynamic.Val.Kind != game.DynamicAmountEventLifeChange {
		t.Fatalf("amount = %#v, want dynamic event life change", add.Amount)
	}
}

// TestLowerTargetCounterTriggeringQuantityFailsClosedInSpell proves the same
// "that many" target counter amount stays unsupported outside a measuring
// trigger, where no triggering quantity exists.
func TestLowerTargetCounterTriggeringQuantityFailsClosedInSpell(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Counters",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{1}{G}",
		OracleText: "Put that many +1/+1 counters on target creature.",
	})
}
