package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/counter"
)

// TestLowerPrizePigSourceCounterThresholdSequence proves Prize Pig's "Whenever
// you gain life, put that many ribbon counters on this creature. Then if there
// are three or more ribbon counters on this creature, remove those counters and
// untap it. {T}: Add one mana of any color." lowers to a life-gain triggered
// ability whose ordered sequence places ribbon counters from the life gained,
// then — gated on the source holding three or more ribbon counters — removes
// every ribbon counter and untaps the source, alongside an intact {T} any-color
// mana ability.
func TestLowerPrizePigSourceCounterThresholdSequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Prize Pig",
		Layout:     "normal",
		TypeLine:   "Creature — Boar",
		OracleText: "Whenever you gain life, put that many ribbon counters on this creature. Then if there are three or more ribbon counters on this creature, remove those counters and untap it.\n{T}: Add one mana of any color.",
		Power:      new("2"),
		Toughness:  new("2"),
	})

	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("mana abilities = %d, want 1 (the {T} any-color ability)", len(face.ManaAbilities))
	}
	if len(face.ActivatedAbilities) != 0 {
		t.Fatalf("activated abilities = %d, want 0 (the mana ability lowers to a ManaAbility)", len(face.ActivatedAbilities))
	}

	trigger := face.TriggeredAbilities[0]
	if trigger.Trigger.Pattern.Event != game.EventLifeGained {
		t.Fatalf("trigger event = %v, want EventLifeGained", trigger.Trigger.Pattern.Event)
	}
	seq := trigger.Content.Modes[0].Sequence
	if len(seq) != 3 {
		t.Fatalf("sequence length = %d, want 3 (put, remove, untap)", len(seq))
	}

	source := game.SourcePermanentReference()

	// Instruction 0: place ribbon counters equal to the life gained, ungated.
	add, ok := seq[0].Primitive.(game.AddCounter)
	if !ok {
		t.Fatalf("sequence[0] = %T, want game.AddCounter", seq[0].Primitive)
	}
	if add.CounterKind != counter.Ribbon {
		t.Fatalf("add counter kind = %v, want Ribbon", add.CounterKind)
	}
	if add.Object != source {
		t.Fatalf("add object = %#v, want source permanent", add.Object)
	}
	if seq[0].Condition.Exists {
		t.Fatalf("sequence[0] is gated, want the placement ungated: %#v", seq[0].Condition)
	}
	addDyn := add.Amount.DynamicAmount()
	if !addDyn.Exists || addDyn.Val.Kind != game.DynamicAmountEventLifeChange {
		t.Fatalf("add amount = %#v, want DynamicAmountEventLifeChange (life gained)", add.Amount)
	}

	// Instruction 1: remove every ribbon counter from the source, gated on the
	// threshold. The amount reads the source's current ribbon count so all of
	// them are removed.
	remove, ok := seq[1].Primitive.(game.RemoveCounter)
	if !ok {
		t.Fatalf("sequence[1] = %T, want game.RemoveCounter", seq[1].Primitive)
	}
	if remove.CounterKind != counter.Ribbon {
		t.Fatalf("remove counter kind = %v, want Ribbon", remove.CounterKind)
	}
	if remove.Object != source {
		t.Fatalf("remove object = %#v, want source permanent", remove.Object)
	}
	if remove.AllKinds || remove.ChooseKind {
		t.Fatalf("remove kind selection = %#v, want a fixed ribbon kind", remove)
	}
	removeDyn := remove.Amount.DynamicAmount()
	if !removeDyn.Exists || removeDyn.Val.Kind != game.DynamicAmountObjectCounters {
		t.Fatalf("remove amount = %#v, want DynamicAmountObjectCounters (all ribbon counters)", remove.Amount)
	}
	if removeDyn.Val.CounterKind != counter.Ribbon || removeDyn.Val.Object != source {
		t.Fatalf("remove amount source = %#v, want source ribbon count", removeDyn.Val)
	}
	assertRibbonThresholdGate(t, seq[1], source)
	if seq[1].PublishResult != game.ResultKey("counter-threshold-cleared") {
		t.Fatalf("remove publishes %q, want counter-threshold-cleared (single-eval capture)", seq[1].PublishResult)
	}

	// Instruction 2: untap the source. The threshold is checked once at
	// resolution, so the untap chains on the removal's published success rather
	// than re-evaluating the (now-cleared) counter threshold.
	untap, ok := seq[2].Primitive.(game.Untap)
	if !ok {
		t.Fatalf("sequence[2] = %T, want game.Untap", seq[2].Primitive)
	}
	if untap.Object != source {
		t.Fatalf("untap object = %#v, want source permanent", untap.Object)
	}
	if seq[2].Condition.Exists {
		t.Fatalf("sequence[2] re-evaluates the threshold, want it gated on the removal result: %#v", seq[2].Condition)
	}
	if !seq[2].ResultGate.Exists {
		t.Fatal("untap is not result-gated, want it chained on the removal result")
	}
	gate := seq[2].ResultGate.Val
	if gate.Key != game.ResultKey("counter-threshold-cleared") || gate.Succeeded != game.TriTrue {
		t.Fatalf("untap result gate = %#v, want key counter-threshold-cleared succeeded=true", gate)
	}
}

// assertRibbonThresholdGate verifies that an instruction is gated on the source
// permanent holding three or more ribbon counters.
func assertRibbonThresholdGate(t *testing.T, ins game.Instruction, source game.ObjectReference) {
	t.Helper()
	if !ins.Condition.Exists || !ins.Condition.Val.Condition.Exists {
		t.Fatalf("instruction is not gated: %#v", ins.Condition)
	}
	matches := ins.Condition.Val.Condition.Val.ObjectMatches
	if !matches.Exists {
		t.Fatalf("gate has no ObjectMatches selection: %#v", ins.Condition.Val.Condition.Val)
	}
	condObj := ins.Condition.Val.Condition.Val.Object
	if !condObj.Exists || condObj.Val != source {
		t.Fatalf("gate object = %#v, want source permanent", condObj)
	}
	if matches.Val.RequiredCounter != counter.Ribbon {
		t.Fatalf("gate required counter = %v, want Ribbon", matches.Val.RequiredCounter)
	}
	want := compare.Int{Op: compare.GreaterOrEqual, Value: 3}
	if !matches.Val.RequiredCounterCount.Exists || matches.Val.RequiredCounterCount.Val != want {
		t.Fatalf("gate required counter count = %#v, want %#v", matches.Val.RequiredCounterCount, want)
	}
}
