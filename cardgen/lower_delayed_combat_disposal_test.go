package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerDelayedEndOfCombatDestroyCapturesEventRelatedPermanent proves the
// basilisk block idiom "Whenever this creature blocks or becomes blocked by a
// creature, destroy that creature at end of combat" (Tangle Asp) lowers to a
// CreateDelayedTrigger whose CapturedObject freezes the opposing combatant
// (the event-related permanent) and whose content destroys that captured
// creature.
func TestLowerDelayedEndOfCombatDestroyCapturesEventRelatedPermanent(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Tangle Asp",
		Layout:     "normal",
		TypeLine:   "Creature — Snake",
		Power:      new("1"),
		Toughness:  new("1"),
		OracleText: "Whenever this creature blocks or becomes blocked by a creature, destroy that creature at end of combat.",
	})
	trigger, destroy := delayedEndOfCombatDestroy(t, face)
	if trigger.Trigger.Timing != game.DelayedAtEndOfCombat {
		t.Fatalf("timing = %v, want DelayedAtEndOfCombat", trigger.Trigger.Timing)
	}
	if !trigger.Trigger.CapturedObject.Exists ||
		trigger.Trigger.CapturedObject.Val.Kind() != game.ObjectReferenceEventRelatedPermanent {
		t.Fatalf("captured object = %#v, want event-related permanent", trigger.Trigger.CapturedObject)
	}
	if destroy.Object.Kind() != game.ObjectReferenceCapturedObject {
		t.Fatalf("destroy object = %v, want captured object", destroy.Object.Kind())
	}
}

// TestLowerDelayedEndOfCombatDestroyCapturesEventPermanent proves the basilisk
// combat-damage idiom "Whenever this creature deals combat damage to a creature,
// destroy that creature at end of combat" (Serpentine Basilisk) lowers with the
// event permanent itself captured.
func TestLowerDelayedEndOfCombatDestroyCapturesEventPermanent(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Serpentine Basilisk",
		Layout:     "normal",
		TypeLine:   "Creature — Lizard Basilisk",
		Power:      new("3"),
		Toughness:  new("3"),
		OracleText: "Whenever this creature deals combat damage to a creature, destroy that creature at end of combat.",
	})
	trigger, destroy := delayedEndOfCombatDestroy(t, face)
	if !trigger.Trigger.CapturedObject.Exists ||
		trigger.Trigger.CapturedObject.Val.Kind() != game.ObjectReferenceEventPermanent {
		t.Fatalf("captured object = %#v, want event permanent", trigger.Trigger.CapturedObject)
	}
	if destroy.Object.Kind() != game.ObjectReferenceCapturedObject {
		t.Fatalf("destroy object = %v, want captured object", destroy.Object.Kind())
	}
}

// TestLowerDelayedEndOfCombatReturnStillUnsupported locks the wave's narrow
// scope: only the destroy verb is captured. The return-to-hand basilisk variant
// ("return that creature to its owner's hand at end of combat") must still fail
// closed rather than silently lowering through the capture path.
func TestLowerDelayedEndOfCombatReturnStillUnsupported(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Wall of Tears",
		Layout:     "normal",
		TypeLine:   "Creature — Wall",
		Power:      new("0"),
		Toughness:  new("3"),
		OracleText: "Defender\nWhenever this creature blocks a creature, return that creature to its owner's hand at end of combat.",
	})
}

// TestLowerDelayedEndOfCombatOptionalDestroyPreservesOptionality proves the
// capture path represents an optional ("you may") basilisk faithfully: the
// compiler hoists the trigger body's "you may" onto the enclosing triggered
// ability, so the ability lowers with Optional set while still capturing and
// destroying the combat creature at end of combat — the "you may" is not dropped.
func TestLowerDelayedEndOfCombatOptionalDestroyPreservesOptionality(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Optional Basilisk",
		Layout:     "normal",
		TypeLine:   "Creature — Basilisk",
		Power:      new("2"),
		Toughness:  new("2"),
		OracleText: "Whenever this creature blocks or becomes blocked by a creature, you may destroy that creature at end of combat.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	if !face.TriggeredAbilities[0].Optional {
		t.Fatal("triggered ability Optional = false, want true (the \"you may\" must be preserved)")
	}
	trigger, destroy := delayedEndOfCombatDestroy(t, face)
	if !trigger.Trigger.CapturedObject.Exists {
		t.Fatalf("captured object = %#v, want captured combat creature", trigger.Trigger.CapturedObject)
	}
	if destroy.Object.Kind() != game.ObjectReferenceCapturedObject {
		t.Fatalf("destroy object = %v, want captured object", destroy.Object.Kind())
	}
}

func delayedEndOfCombatDestroy(t *testing.T, face loweredFaceAbilities) (game.CreateDelayedTrigger, game.Destroy) {
	t.Helper()
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence len = %d, want 1", len(mode.Sequence))
	}
	trigger, ok := mode.Sequence[0].Primitive.(game.CreateDelayedTrigger)
	if !ok {
		t.Fatalf("primitive = %#v, want CreateDelayedTrigger", mode.Sequence[0].Primitive)
	}
	inner := trigger.Trigger.Content.Modes[0]
	if len(inner.Sequence) != 1 {
		t.Fatalf("delayed content sequence len = %d, want 1", len(inner.Sequence))
	}
	destroy, ok := inner.Sequence[0].Primitive.(game.Destroy)
	if !ok {
		t.Fatalf("delayed primitive = %#v, want Destroy", inner.Sequence[0].Primitive)
	}
	return trigger, destroy
}
