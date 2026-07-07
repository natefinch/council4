package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceFightForTheThrone covers the ordered spell
// sequence that puts a +1/+1 counter on target creature you control, then has it
// fight target creature an opponent controls, and finally sets up a delayed
// "dies this turn" trigger bound to the fought opponent's creature. The fight's
// second target (RelatedObject) is captured (CapturedDyingObject) so the delayed
// trigger fires only when that specific creature dies (EventPermanentDied,
// DyingObjectCaptured: true). The trigger carries the "if you control your
// commander" intervening condition (ControllerControlsCommander) and its effect
// is to become the monarch.
func TestGenerateExecutableCardSourceFightForTheThrone(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Fight for the Throne",
		Layout:   "normal",
		ManaCost: "{1}{G}",
		TypeLine: "Instant",
		OracleText: "Put a +1/+1 counter on target creature you control. " +
			"Then it fights target creature an opponent controls. " +
			"When the creature an opponent controls dies this turn, if you control your commander, you become the monarch.",
	}, "f")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		// effect[0]: +1/+1 counter on the first target (you control).
		"Primitive: game.AddCounter{",
		"Object:      game.TargetPermanentReference(0)",
		"CounterKind: counter.PlusOnePlusOne",
		// effect[1]: the pumped creature fights the opponent's creature.
		"Primitive: game.Fight{",
		"Object:        game.TargetPermanentReference(0)",
		"RelatedObject: game.TargetPermanentReference(1)",
		// effect[2]: delayed dies trigger bound to the fought opponent's creature.
		"Primitive: game.CreateDelayedTrigger{",
		"Event:               game.EventPermanentDied",
		"DyingObjectCaptured: true",
		"Window:              game.DelayedWindowThisTurn",
		"CapturedDyingObject: opt.Val(game.TargetPermanentReference(1))",
		// The intervening "if you control your commander" condition.
		"InterveningCondition: opt.Val(game.Condition{",
		"ControllerControlsCommander: true",
		// The delayed trigger's effect is to become the monarch.
		"Primitive: game.BecomeMonarch{",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
	// The delayed trigger must bind the captured dying object rather than gate on a
	// static source filter.
	if strings.Contains(source, "Source:              game.TriggerSourceSelf") {
		t.Fatalf("delayed trigger kept a self source filter instead of binding the captured dying object:\n%s", source)
	}
}
