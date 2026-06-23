package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

// TestLowerSigurdBoastPutOrRemoveCounter checks that Sigurd, Jarl of
// Ravensthorpe lowers fully: its boast activates the put-or-remove lore-counter
// modal, the elided "remove one from it" removes a lore counter (the kind the
// put alternative places) from the same target Saga, and the lore-counter
// placement trigger fires.
func TestLowerSigurdBoastPutOrRemoveCounter(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Sigurd, Jarl of Ravensthorpe",
		Layout:   "normal",
		TypeLine: "Legendary Creature — Human Warrior",
		OracleText: "Vigilance, trample, lifelink\n" +
			"Boast — {1}: Put a lore counter on target Saga you control or remove one from it. (Activate only if this creature attacked this turn and only once each turn.)\n" +
			"Whenever you put a lore counter on a Saga you control, put a +1/+1 counter on up to one other target creature.",
		Power:     new("3"),
		Toughness: new("3"),
	})

	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want one Boast ability", len(face.ActivatedAbilities))
	}
	boast := face.ActivatedAbilities[0]
	if boast.Timing != game.OncePerTurn {
		t.Fatalf("timing = %v, want OncePerTurn", boast.Timing)
	}
	if !boast.ActivationCondition.Exists || !boast.ActivationCondition.Val.EventHistory.Exists {
		t.Fatalf("activation condition = %#v, want attacked-this-turn event history", boast.ActivationCondition)
	}

	content := boast.Content
	if content.MinModes != 1 || content.MaxModes != 1 {
		t.Fatalf("modes = [%d,%d], want choose-one (1,1)", content.MinModes, content.MaxModes)
	}
	if len(content.SharedTargets) != 1 {
		t.Fatalf("shared targets = %d, want one shared Saga target", len(content.SharedTargets))
	}
	if len(content.Modes) != 2 {
		t.Fatalf("modes = %d, want two (put, remove)", len(content.Modes))
	}

	add, ok := content.Modes[0].Sequence[0].Primitive.(game.AddCounter)
	if !ok {
		t.Fatalf("mode 0 primitive = %T, want game.AddCounter", content.Modes[0].Sequence[0].Primitive)
	}
	if add.CounterKind != counter.Lore {
		t.Fatalf("put counter kind = %v, want lore", add.CounterKind)
	}

	remove, ok := content.Modes[1].Sequence[0].Primitive.(game.RemoveCounter)
	if !ok {
		t.Fatalf("mode 1 primitive = %T, want game.RemoveCounter", content.Modes[1].Sequence[0].Primitive)
	}
	if remove.CounterKind != counter.Lore {
		t.Fatalf("removed counter kind = %v, want lore inherited from the put alternative", remove.CounterKind)
	}
	if remove.ChooseKind {
		t.Fatal("removed counter ChooseKind = true, want the elided kind resolved to lore")
	}
	if remove.Amount != game.Fixed(1) {
		t.Fatalf("removed amount = %#v, want one", remove.Amount)
	}

	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want the lore-counter trigger", len(face.TriggeredAbilities))
	}
	trigger := face.TriggeredAbilities[0].Trigger.Pattern
	if trigger.Event != game.EventCountersAdded ||
		!trigger.MatchCounterKind ||
		trigger.CounterKind != counter.Lore {
		t.Fatalf("trigger pattern = %#v, want a lore counters-added trigger", trigger)
	}
}

// TestLowerPutOrRemoveCounterModalNamedKind checks the non-elided put-or-remove
// counter modal, where the removal arm names its own counter kind ("...or
// remove a +1/+1 counter from it"), lowers to a choose-one modal over the shared
// target.
func TestLowerPutOrRemoveCounterModalNamedKind(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Counter Tactician",
		Layout:     "normal",
		TypeLine:   "Creature — Human Warrior",
		OracleText: "Boast — {1}: Put a +1/+1 counter on target creature or remove a +1/+1 counter from it. (Activate only if this creature attacked this turn and only once each turn.)",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want one Boast ability", len(face.ActivatedAbilities))
	}
	content := face.ActivatedAbilities[0].Content
	if content.MinModes != 1 || content.MaxModes != 1 || len(content.Modes) != 2 {
		t.Fatalf("content = %#v, want a two-mode choose-one modal", content)
	}
	remove, ok := content.Modes[1].Sequence[0].Primitive.(game.RemoveCounter)
	if !ok {
		t.Fatalf("mode 1 primitive = %T, want game.RemoveCounter", content.Modes[1].Sequence[0].Primitive)
	}
	if remove.CounterKind != counter.PlusOnePlusOne {
		t.Fatalf("removed counter kind = %v, want +1/+1", remove.CounterKind)
	}
}
