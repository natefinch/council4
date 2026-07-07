package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerNakiaWakandanOperative locks in the two recognizers Nakia, Wakandan
// Operative needs: "Whenever your commander enters" as an enters trigger whose
// subject is a commander you control (EventPermanentEnteredBattlefield with
// SubjectSelection.MatchCommander) feeding become-monarch, and "put two +1/+1
// counters on target creature or Vehicle" as a two-counter placement on a
// creature-or-Vehicle union target (Selection.AnyOf).
func TestLowerNakiaWakandanOperative(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Nakia, Wakandan Operative",
		Layout:   "normal",
		TypeLine: "Legendary Creature — Human Warrior Hero",
		OracleText: "Reach\n" +
			"Whenever your commander enters, you become the monarch.\n" +
			"{2}, {T}: Put two +1/+1 counters on target creature or Vehicle. Activate only as a sorcery.",
		Power:     new("3"),
		Toughness: new("3"),
	})

	if len(face.StaticAbilities) != 1 {
		t.Fatalf("got %d static abilities, want 1 (Reach)", len(face.StaticAbilities))
	}
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("got %d activated abilities, want 1", len(face.ActivatedAbilities))
	}

	// Ability 2: "Whenever your commander enters, you become the monarch."
	trigger := face.TriggeredAbilities[0].Trigger
	if trigger.Pattern.Event != game.EventPermanentEnteredBattlefield {
		t.Fatalf("trigger event = %v, want EventPermanentEnteredBattlefield", trigger.Pattern.Event)
	}
	if !trigger.Pattern.SubjectSelection.MatchCommander {
		t.Fatal("trigger SubjectSelection.MatchCommander = false, want true (your commander)")
	}
	if trigger.Pattern.Controller != game.TriggerControllerYou {
		t.Fatalf("trigger controller = %v, want TriggerControllerYou", trigger.Pattern.Controller)
	}
	if trigger.Pattern.UnionEvent != game.EventUnknown {
		t.Fatalf("trigger UnionEvent = %v, want none (plain enters, not enters-or-attacks)", trigger.Pattern.UnionEvent)
	}
	triggerModes := face.TriggeredAbilities[0].Content.Modes
	if len(triggerModes) != 1 || len(triggerModes[0].Sequence) != 1 {
		t.Fatalf("trigger content = %#v, want one mode with one instruction", triggerModes)
	}
	if _, ok := triggerModes[0].Sequence[0].Primitive.(game.BecomeMonarch); !ok {
		t.Fatalf("trigger primitive = %#v, want BecomeMonarch", triggerModes[0].Sequence[0].Primitive)
	}

	// Ability 3: "{2}, {T}: Put two +1/+1 counters on target creature or Vehicle."
	activated := face.ActivatedAbilities[0]
	if activated.Timing != game.SorceryOnly {
		t.Fatalf("activated timing = %v, want SorceryOnly", activated.Timing)
	}
	modes := activated.Content.Modes
	if len(modes) != 1 {
		t.Fatalf("activated modes = %d, want 1", len(modes))
	}
	mode := modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("activated targets = %d, want 1", len(mode.Targets))
	}
	assertCreatureOrVehicleTarget(t, mode.Targets[0])
	if len(mode.Sequence) != 1 {
		t.Fatalf("activated sequence = %d instructions, want 1", len(mode.Sequence))
	}
	addCounter, ok := mode.Sequence[0].Primitive.(game.AddCounter)
	if !ok {
		t.Fatalf("activated primitive = %#v, want AddCounter", mode.Sequence[0].Primitive)
	}
	if addCounter.CounterKind != counter.PlusOnePlusOne {
		t.Fatalf("counter kind = %v, want +1/+1", addCounter.CounterKind)
	}
	if addCounter.Amount.Value() != 2 {
		t.Fatalf("counter amount = %d, want 2", addCounter.Amount.Value())
	}
	if addCounter.Object != game.TargetPermanentReference(0) {
		t.Fatalf("counter object = %#v, want target permanent 0", addCounter.Object)
	}
}

// assertCreatureOrVehicleTarget verifies a single-permanent target spec is the
// "creature or Vehicle" union: a Selection.AnyOf with a creature alternative and
// a Vehicle-subtype alternative.
func assertCreatureOrVehicleTarget(t *testing.T, target game.TargetSpec) {
	t.Helper()
	if target.Allow != game.TargetAllowPermanent {
		t.Fatalf("target allow = %v, want TargetAllowPermanent", target.Allow)
	}
	if target.MinTargets != 1 || target.MaxTargets != 1 {
		t.Fatalf("target cardinality = [%d,%d], want [1,1]", target.MinTargets, target.MaxTargets)
	}
	if !target.Selection.Exists {
		t.Fatal("target selection missing, want creature-or-Vehicle AnyOf")
	}
	anyOf := target.Selection.Val.AnyOf
	if len(anyOf) != 2 {
		t.Fatalf("target AnyOf = %d alternatives, want 2 (creature, Vehicle)", len(anyOf))
	}
	sawCreature := false
	sawVehicle := false
	for _, alt := range anyOf {
		for _, cardType := range alt.RequiredTypesAny {
			if cardType == types.Creature {
				sawCreature = true
			}
		}
		for _, sub := range alt.SubtypesAny {
			if sub == types.Vehicle {
				sawVehicle = true
			}
		}
	}
	if !sawCreature {
		t.Fatalf("target AnyOf missing creature alternative: %#v", anyOf)
	}
	if !sawVehicle {
		t.Fatalf("target AnyOf missing Vehicle alternative: %#v", anyOf)
	}
}
