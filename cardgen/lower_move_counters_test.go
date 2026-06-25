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
	if len(mode.Targets) != 1 || mode.Targets[0].Selection.Val.RequiredTypesAny[0] != types.Creature {
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

// TestLowerMoveCountersFromTargetAnyKind verifies the two-target "Move a counter
// from target permanent you control onto a second target permanent." form
// (Nesting Grounds) lowers to a MoveCounters that reads counters from the first
// target (CounterSourceTarget) and moves one counter of a controller-chosen kind
// onto the second target.
func TestLowerMoveCountersFromTargetAnyKind(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Nesting Grounds",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{T}: Add {C}.\n{1}, {T}: Move a counter from target permanent you control onto a second target permanent. Activate only as a sorcery.",
	})
	var move game.MoveCounters
	for _, ability := range face.ActivatedAbilities {
		if got, ok := ability.Content.Modes[0].Sequence[0].Primitive.(game.MoveCounters); ok {
			move = got
		}
	}
	if !move.ChooseKind ||
		move.AllKinds ||
		move.Amount.Value() != 1 ||
		move.Object != game.TargetPermanentReference(1) ||
		move.Source.Kind != game.CounterSourceTarget ||
		move.Source.Object != game.TargetPermanentReference(0) {
		t.Fatalf("primitive = %+v, want choose-kind move from target 0 onto target 1", move)
	}
}

// TestLowerMoveCountersFromTargetNamed verifies the two-target named-kind form
// ("Move a +1/+1 counter from target creature onto a second target creature." —
// Daghatar the Adamant) lowers to a MoveCounters that reads the named kind from
// the first target and moves one onto the second target.
func TestLowerMoveCountersFromTargetNamed(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Daghatar the Adamant",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Human Warrior",
		OracleText: "{1}{B/G}{B/G}: Move a +1/+1 counter from target creature onto a second target creature.",
	})
	mode := face.ActivatedAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 2 {
		t.Fatalf("targets = %+v, want two creature targets", mode.Targets)
	}
	move, ok := mode.Sequence[0].Primitive.(game.MoveCounters)
	if !ok ||
		move.ChooseKind ||
		move.AllKinds ||
		move.CounterKind != counter.PlusOnePlusOne ||
		move.Amount.Value() != 1 ||
		move.Object != game.TargetPermanentReference(1) ||
		move.Source.Kind != game.CounterSourceTarget ||
		move.Source.Object != game.TargetPermanentReference(0) {
		t.Fatalf("primitive = %+v, want one +1/+1 counter from target 0 onto target 1", mode.Sequence[0].Primitive)
	}
}
