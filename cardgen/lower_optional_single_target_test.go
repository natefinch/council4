package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

// TestLowerOptionalSingleTargetSharedPronounSequence verifies that an ordered
// effect sequence whose later clause depends on the singular object pronoun "it"
// referring to an "up to one target" optional single target lowers as a shared
// target. The runtime no-ops both instructions when no target is chosen, so the
// optional cardinality flows onto a single target spec shared by both clauses
// (Utrom Scientists, Petrifying Meddler, Collector's Case).
func TestLowerOptionalSingleTargetSharedPronounSequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Optional Stun Test",
		Layout:     "normal",
		TypeLine:   "Creature — Test",
		OracleText: "When Optional Stun Test enters, tap up to one target creature and put a stun counter on it.",
	})
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 1 ||
		mode.Targets[0].MinTargets != 0 ||
		mode.Targets[0].MaxTargets != 1 {
		t.Fatalf("targets = %#v, want one up-to-one target", mode.Targets)
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want two instructions", mode.Sequence)
	}
	tap, ok := mode.Sequence[0].Primitive.(game.Tap)
	if !ok || tap.Object != game.TargetPermanentReference(0) {
		t.Fatalf("sequence[0] = %#v, want Tap on target 0", mode.Sequence[0].Primitive)
	}
	add, ok := mode.Sequence[1].Primitive.(game.AddCounter)
	if !ok ||
		add.Object != game.TargetPermanentReference(0) ||
		add.CounterKind != counter.Stun ||
		add.Amount != game.Fixed(1) {
		t.Fatalf("sequence[1] = %#v, want AddCounter stun on target 0", mode.Sequence[1].Primitive)
	}
}
