package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

// TestLowerDoubleCountersSelf proves "double the number of +1/+1 counters on
// this creature" (Mossborn Hydra) lowers to a counter placement that adds
// counters equal to the source's current count, modeling the doubling with a
// DynamicAmountObjectCounters amount read from the source permanent.
func TestLowerDoubleCountersSelf(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Hydra",
		Layout:     "normal",
		TypeLine:   "Creature — Hydra",
		OracleText: "When this creature enters, double the number of +1/+1 counters on this creature.",
		Power:      new("0"),
		Toughness:  new("0"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %#v, want one", face.TriggeredAbilities)
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %#v, want one instruction", mode.Sequence)
	}
	add, ok := mode.Sequence[0].Primitive.(game.AddCounter)
	if !ok {
		t.Fatalf("primitive = %T, want game.AddCounter", mode.Sequence[0].Primitive)
	}
	if add.Object != game.SourcePermanentReference() {
		t.Fatalf("object = %#v, want source permanent", add.Object)
	}
	if add.CounterKind != counter.PlusOnePlusOne {
		t.Fatalf("counter kind = %v, want PlusOnePlusOne", add.CounterKind)
	}
	dynamicOpt := add.Amount.DynamicAmount()
	if !dynamicOpt.Exists {
		t.Fatalf("amount = %#v, want dynamic object counter count", add.Amount)
	}
	dynamic := dynamicOpt.Val
	if dynamic.Kind != game.DynamicAmountObjectCounters ||
		dynamic.Object != game.SourcePermanentReference() ||
		dynamic.CounterKind != counter.PlusOnePlusOne {
		t.Fatalf("amount = %#v, want object counter count of the source", dynamic)
	}
}
