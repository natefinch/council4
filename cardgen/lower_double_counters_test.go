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

// TestLowerDoubleCountersTargetSingleKind proves "double the number of +1/+1
// counters on target creature" (Gilder Bairn-adjacent) lowers to a dynamic
// counter placement bound to the activated ability's permanent target.
func TestLowerDoubleCountersTargetSingleKind(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Doubler",
		Layout:     "normal",
		TypeLine:   "Creature — Frog",
		OracleText: "{T}: Double the number of +1/+1 counters on target creature.",
		Power:      new("0"),
		Toughness:  new("0"),
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %#v, want one", face.ActivatedAbilities)
	}
	content := face.ActivatedAbilities[0].Content
	if len(content.Modes[0].Targets) != 1 ||
		content.Modes[0].Targets[0].Allow != game.TargetAllowPermanent {
		t.Fatalf("targets = %#v, want one permanent target", content.Modes[0].Targets)
	}
	add, ok := content.Modes[0].Sequence[0].Primitive.(game.AddCounter)
	if !ok {
		t.Fatalf("primitive = %T, want game.AddCounter", content.Modes[0].Sequence[0].Primitive)
	}
	if add.AllKinds {
		t.Fatal("AllKinds = true, want false for single-kind doubling")
	}
	if add.Object != game.TargetPermanentReference(0) {
		t.Fatalf("object = %#v, want target permanent", add.Object)
	}
	dynamic := add.Amount.DynamicAmount()
	if !dynamic.Exists || dynamic.Val.Kind != game.DynamicAmountObjectCounters ||
		dynamic.Val.Object != game.TargetPermanentReference(0) ||
		dynamic.Val.CounterKind != counter.PlusOnePlusOne {
		t.Fatalf("amount = %#v, want object counter count of the target", add.Amount)
	}
}

// TestLowerDoubleCountersTargetAllKinds proves "double the number of each kind of
// counter on target ..." (Vorel of the Hull Clade) lowers to a single
// AddCounter{AllKinds} bound to the permanent target.
func TestLowerDoubleCountersTargetAllKinds(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Vorel of the Hull Clade",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Merfolk Wizard",
		OracleText: "{G}{U}, {T}: Double the number of each kind of counter on target artifact, creature, or land.",
		Power:      new("1"),
		Toughness:  new("4"),
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %#v, want one", face.ActivatedAbilities)
	}
	content := face.ActivatedAbilities[0].Content
	if len(content.Modes[0].Targets) != 1 ||
		content.Modes[0].Targets[0].Allow != game.TargetAllowPermanent {
		t.Fatalf("targets = %#v, want one permanent target", content.Modes[0].Targets)
	}
	add, ok := content.Modes[0].Sequence[0].Primitive.(game.AddCounter)
	if !ok {
		t.Fatalf("primitive = %T, want game.AddCounter", content.Modes[0].Sequence[0].Primitive)
	}
	if !add.AllKinds {
		t.Fatal("AllKinds = false, want true for each-kind doubling")
	}
	if add.Object != game.TargetPermanentReference(0) {
		t.Fatalf("object = %#v, want target permanent", add.Object)
	}
}

// TestLowerDoubleCountersGroup proves "double the number of +1/+1 counters on
// each creature you control" (Bristly Bill, Spine Sower) lowers to a group
// AddCounter with DoubleKind, doubling the single kind on each controlled
// creature.
func TestLowerDoubleCountersGroup(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Group Doubler",
		Layout:     "normal",
		TypeLine:   "Creature — Plant",
		OracleText: "{3}{G}{G}: Double the number of +1/+1 counters on each creature you control.",
		Power:      new("0"),
		Toughness:  new("0"),
	})
	add, ok := face.ActivatedAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.AddCounter)
	if !ok {
		t.Fatalf("primitive = %T, want game.AddCounter", face.ActivatedAbilities[0].Content.Modes[0].Sequence[0].Primitive)
	}
	if !add.DoubleKind || !add.Group.Valid() || add.CounterKind != counter.PlusOnePlusOne {
		t.Fatalf("add = %#v, want group DoubleKind on +1/+1", add)
	}
}

// TestLowerDoubleCountersEventPermanent proves a triggered "double the number of
// +1/+1 counters on it" doubling whose "it" is the triggering event permanent
// (Byrke, Long Ear of the Law; Seismic Tutelage's attacking enchanted creature)
// lowers to a dynamic placement bound to that event permanent.
func TestLowerDoubleCountersEventPermanent(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Pump",
		Layout:     "normal",
		TypeLine:   "Creature — Rabbit",
		OracleText: "Whenever a creature you control with a +1/+1 counter on it attacks, double the number of +1/+1 counters on it.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	add, ok := face.TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.AddCounter)
	if !ok {
		t.Fatalf("primitive = %T, want game.AddCounter", face.TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive)
	}
	if add.Object != game.EventPermanentReference() {
		t.Fatalf("object = %#v, want event permanent", add.Object)
	}
	if d := add.Amount.DynamicAmount(); !d.Exists || d.Val.Object != game.EventPermanentReference() {
		t.Fatalf("amount = %#v, want event-permanent counter count", add.Amount)
	}
}
