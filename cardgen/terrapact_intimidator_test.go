package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceTerrapactIntimidator covers the non-controller
// "may have" causative gate: a target opponent decides ("target opponent may
// have you create ...") whether the controller performs a caused action, and a
// consequence resolves when that offer is declined ("If they don't, ..."). The
// caused create must be asked of the target opponent (its OptionalActor names the
// target player, not the controller) even though the controller is the actor, the
// offer publishes whether it happened, and the +1/+1 counters resolve for the
// source only when the offer was declined (TriFalse).
func TestGenerateExecutableCardSourceTerrapactIntimidator(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Terrapact Intimidator",
		Layout:     "normal",
		ManaCost:   "{3}{G}",
		TypeLine:   "Creature — Lizard",
		OracleText: "When this creature enters, target opponent may have you create two Lander tokens. If they don't, put two +1/+1 counters on this creature. (A Lander token is an artifact with \"{2}, {T}, Sacrifice this token: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.\")",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Primitive: game.CreateToken{",
		"Amount: game.Fixed(2)",
		"Optional:      true",
		"OptionalActor: opt.Val(game.TargetPlayerReference(0))",
		"PublishResult: game.ResultKey(\"may-have-action\")",
		"Primitive: game.AddCounter{",
		"Object:      game.SourcePermanentReference()",
		"CounterKind: counter.PlusOnePlusOne",
		"Key:       \"may-have-action\"",
		"Succeeded: game.TriFalse",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
