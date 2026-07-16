package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/counter"
)

const agitatorAntOracle = "At the beginning of your end step, each player may put two +1/+1 counters on a creature they control. Goad each creature that had counters put on it this way. (Until your next turn, those creatures attack each combat if able and attack a player other than you if able.)"

func TestParseOptionalCounterForEachPlayerGoadSequence(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(agitatorAntOracle, Context{CardName: "Agitator Ant"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %d, want 1", len(document.Abilities))
	}
	if coverage := DocumentCoverage(document); !coverage.Complete {
		t.Fatalf("coverage = %#v, want complete", coverage)
	}
	clause := document.Abilities[0].OptionalCounterForEachPlayer
	if clause == nil {
		t.Fatal("OptionalCounterForEachPlayer = nil")
	}
	if clause.PlayerContext != EffectContextEachPlayer ||
		clause.Pool.Kind != SelectionCreature ||
		!clause.Amount.Known ||
		clause.Amount.Value != 2 ||
		clause.CounterKind != counter.PlusOnePlusOne {
		t.Fatalf("clause = %#v", clause)
	}
	for i := range 2 {
		if got := len(document.Abilities[0].Sentences[i].Effects); got != 0 {
			t.Fatalf("sentence %d effects = %d, want consumed", i, got)
		}
	}
}
