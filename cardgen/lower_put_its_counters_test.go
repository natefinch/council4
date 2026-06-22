package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerPutItsCountersOnTargetSalvage verifies that the singular-pronoun
// counter-salvage wording "put its counters on target creature you control"
// (Star Pupil) lowers to the same kind-agnostic MoveCounters as the established
// "those counters" wording: the counters are read from the triggering event
// permanent (the dying source, via CounterSourceEventPermanent) and placed,
// regardless of kind, on the chosen target permanent.
func TestLowerPutItsCountersOnTargetSalvage(t *testing.T) {
	t.Parallel()
	power, toughness := "0", "0"
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Star Pupil",
		Layout:     "normal",
		TypeLine:   "Creature — Human Wizard",
		ManaCost:   "{W}",
		OracleText: "This creature enters with a +1/+1 counter on it.\nWhen this creature dies, put its counters on target creature you control.",
		Power:      &power,
		Toughness:  &toughness,
	})

	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	ability := face.TriggeredAbilities[0]
	if len(ability.Content.Modes) != 1 || len(ability.Content.Modes[0].Targets) != 1 {
		t.Fatalf("ability content = %+v, want one mode with one target", ability.Content)
	}
	instruction := ability.Content.Modes[0].Sequence[0]
	move, ok := instruction.Primitive.(game.MoveCounters)
	if !ok {
		t.Fatalf("primitive = %T, want game.MoveCounters", instruction.Primitive)
	}
	if move.Source.Kind != game.CounterSourceEventPermanent {
		t.Fatalf("move source = %v, want CounterSourceEventPermanent", move.Source.Kind)
	}
	if !move.AllKinds {
		t.Fatalf("move = %+v, want AllKinds salvage move", move)
	}
	if move.Object != game.TargetPermanentReference(0) {
		t.Fatalf("move destination = %v, want TargetPermanentReference(0)", move.Object)
	}
}

// TestLowerPutItsCountersOnSelfSalvage verifies the self-destination form "put
// its counters on this creature" (Buzzard-Wasp Colony) lowers onto the source
// permanent rather than a target.
func TestLowerPutItsCountersOnSelfSalvage(t *testing.T) {
	t.Parallel()
	power, toughness := "1", "1"
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Buzzard-Wasp Colony",
		Layout:     "normal",
		TypeLine:   "Creature — Bird Insect",
		ManaCost:   "{3}{B}",
		OracleText: "Flying\nWhenever another creature you control dies, if it had counters on it, put its counters on this creature.",
		Power:      &power,
		Toughness:  &toughness,
	})

	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	ability := face.TriggeredAbilities[0]
	if len(ability.Content.Modes) != 1 || len(ability.Content.Modes[0].Targets) != 0 {
		t.Fatalf("ability content = %+v, want one mode with no targets", ability.Content)
	}
	move, ok := ability.Content.Modes[0].Sequence[0].Primitive.(game.MoveCounters)
	if !ok {
		t.Fatalf("primitive = %T, want game.MoveCounters", ability.Content.Modes[0].Sequence[0].Primitive)
	}
	if move.Source.Kind != game.CounterSourceEventPermanent {
		t.Fatalf("move source = %v, want CounterSourceEventPermanent", move.Source.Kind)
	}
	if !move.AllKinds {
		t.Fatalf("move = %+v, want AllKinds salvage move", move)
	}
	if move.Object != game.SourcePermanentReference() {
		t.Fatalf("move destination = %v, want SourcePermanentReference", move.Object)
	}
}
