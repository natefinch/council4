package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerMoveCounterSingleNamed verifies the generated counter-movement slice:
// a {T} activated ability that moves one +1/+1 counter off the source permanent
// (CounterSourceSelf) onto a single target creature lowers to a MoveCounters
// instruction with the named kind and a fixed amount of one.
func TestLowerMoveCounterSingleNamed(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Weapon Rack",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "{T}: Move a +1/+1 counter from this artifact onto target creature.",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	mode := face.ActivatedAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 1 || mode.Targets[0].Predicate.PermanentTypes[0] != types.Creature {
		t.Fatalf("targets = %+v, want one creature target", mode.Targets)
	}
	move, ok := mode.Sequence[0].Primitive.(game.MoveCounters)
	if !ok ||
		move.AllKinds ||
		move.CounterKind != counter.PlusOnePlusOne ||
		move.Amount.Value() != 1 ||
		move.Object != game.TargetPermanentReference(0) ||
		move.Source.Kind != game.CounterSourceSelf {
		t.Fatalf("primitive = %+v, want one +1/+1 counter from self onto target 0", mode.Sequence[0].Primitive)
	}
}

// TestLowerMoveCountersAll verifies the kind-agnostic "all counters" form lowers
// to a MoveCounters instruction with AllKinds set, moving every counter on the
// source regardless of kind.
func TestLowerMoveCountersAll(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Relic Mover",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "{T}: Move all counters from this artifact onto target creature.",
	})
	move, ok := face.ActivatedAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.MoveCounters)
	if !ok ||
		!move.AllKinds ||
		move.Object != game.TargetPermanentReference(0) ||
		move.Source.Kind != game.CounterSourceSelf {
		t.Fatalf("primitive = %+v, want all-kinds move from self onto target 0",
			face.ActivatedAbilities[0].Content.Modes[0].Sequence[0].Primitive)
	}
}
