package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceObNixilisTheFallen covers the controller "may
// have" causative gate: the controller decides ("you may have target player lose
// 3 life") whether the caused action happens, and a consequence resolves when the
// offer is accepted ("If you do, ..."). Because the controller is the decider the
// offer carries no OptionalActor (the runtime default), the caused life loss still
// targets the recipient ("target player"), the offer publishes whether it
// happened, and the +1/+1 counters resolve for the source only when the offer was
// accepted (TriTrue).
func TestGenerateExecutableCardSourceObNixilisTheFallen(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Ob Nixilis, the Fallen",
		Layout:     "normal",
		ManaCost:   "{3}{B}{B}",
		TypeLine:   "Legendary Creature — Demon",
		OracleText: "Landfall — Whenever a land you control enters, you may have target player lose 3 life. If you do, put three +1/+1 counters on Ob Nixilis.",
	}, "o")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Primitive: game.LoseLife{",
		"Amount: game.Fixed(3)",
		"Player: game.TargetPlayerReference(0)",
		"Optional:      true",
		"PublishResult: game.ResultKey(\"may-have-action\")",
		"Primitive: game.AddCounter{",
		"Object:      game.SourcePermanentReference()",
		"CounterKind: counter.PlusOnePlusOne",
		"Key:       \"may-have-action\"",
		"Succeeded: game.TriTrue",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
	// The controller is the decider, so the offer must not delegate the decision.
	if strings.Contains(source, "OptionalActor:") {
		t.Fatalf("controller-decided offer set an OptionalActor:\n%s", source)
	}
}
