package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

// TestLowerGraftKeyword verifies the "Graft N" keyword (CR 702.57) lowers to its
// two abilities: a static placing N +1/+1 counters as the creature enters, and a
// triggered ability that may move one +1/+1 counter off this creature
// (CounterSourceSelf) onto the entering creature (EventPermanentReference), not
// onto the source.
func TestLowerGraftKeyword(t *testing.T) {
	t.Parallel()
	power, toughness := "0", "0"
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Simic Initiate",
		Layout:     "normal",
		TypeLine:   "Creature — Merfolk Wizard",
		ManaCost:   "{G}",
		OracleText: "Graft 1 (This creature enters with a +1/+1 counter on it. Whenever another creature enters, you may move a +1/+1 counter from this creature onto it.)",
		Power:      &power,
		Toughness:  &toughness,
	})

	if len(face.ReplacementAbilities) != 1 {
		t.Fatalf("replacement abilities = %d, want 1", len(face.ReplacementAbilities))
	}
	placements := face.ReplacementAbilities[0].Replacement.EntersWithCounters
	if len(placements) != 1 ||
		placements[0].Kind != counter.PlusOnePlusOne ||
		placements[0].Amount != 1 {
		t.Fatalf("enters-with-counters = %+v, want one +1/+1 counter", placements)
	}

	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	if !face.TriggeredAbilities[0].Optional {
		t.Fatalf("triggered ability = %+v, want Optional (\"you may\")", face.TriggeredAbilities[0])
	}
	instruction := face.TriggeredAbilities[0].Content.Modes[0].Sequence[0]
	move, ok := instruction.Primitive.(game.MoveCounters)
	if !ok {
		t.Fatalf("primitive = %T, want game.MoveCounters", instruction.Primitive)
	}
	if move.Object != game.EventPermanentReference() {
		t.Fatalf("move destination = %v, want EventPermanentReference (the entering creature)", move.Object)
	}
	if move.Source.Kind != game.CounterSourceSelf {
		t.Fatalf("move source = %v, want CounterSourceSelf", move.Source.Kind)
	}
	if move.AllKinds ||
		move.CounterKind != counter.PlusOnePlusOne ||
		move.Amount.Value() != 1 {
		t.Fatalf("move = %+v, want one named +1/+1 counter", move)
	}
}
