package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceAttackInTheBox covers the controller "may have"
// causative gate with a self-referential caused action and a delayed-trigger
// consequence: the controller decides ("you may have it get +4/+0 until end of
// turn") whether the source pumps itself, and the accepted branch ("If you do,
// sacrifice it at the beginning of the next end step") schedules a delayed
// sacrifice. The pump publishes whether it happened with no OptionalActor (the
// controller decides), and the delayed-trigger creation is gated on the offer
// having been accepted (TriTrue) so the sacrifice is only scheduled when the
// controller pumped.
func TestGenerateExecutableCardSourceAttackInTheBox(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Attack-in-the-Box",
		Layout:     "normal",
		ManaCost:   "{3}",
		TypeLine:   "Artifact Creature — Construct",
		OracleText: "Whenever this creature attacks, you may have it get +4/+0 until end of turn. If you do, sacrifice it at the beginning of the next end step.",
	}, "a")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Primitive: game.ModifyPT{",
		"PowerDelta:     game.Fixed(4)",
		"Optional:      true",
		"PublishResult: game.ResultKey(\"may-have-action\")",
		"Primitive: game.CreateDelayedTrigger{",
		"Timing: game.DelayedAtBeginningOfNextEndStep",
		"Primitive: game.Sacrifice{",
		"Key:       \"may-have-action\"",
		"Succeeded: game.TriTrue",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
	if strings.Contains(source, "OptionalActor:") {
		t.Fatalf("controller-decided offer set an OptionalActor:\n%s", source)
	}
}
