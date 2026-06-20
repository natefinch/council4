package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLowerPactOfNegationEndToEnd(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Pact of Negation",
		Layout:   "normal",
		TypeLine: "Instant",
		OracleText: "Counter target spell.\n" +
			"At the beginning of your next upkeep, pay {3}{U}{U}. If you don't, you lose the game.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("Pact of Negation produced no spell ability")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want counter plus delayed pay-or-lose", mode.Sequence)
	}
	if _, ok := mode.Sequence[0].Primitive.(game.CounterObject); !ok {
		t.Fatalf("first instruction = %#v, want CounterObject", mode.Sequence[0].Primitive)
	}
	delayed, ok := mode.Sequence[1].Primitive.(game.CreateDelayedTrigger)
	if !ok || delayed.Trigger.Timing != game.DelayedAtBeginningOfNextUpkeep {
		t.Fatalf("delayed instruction = %#v", mode.Sequence[1].Primitive)
	}
	tail := delayed.Trigger.Content.Modes[0].Sequence
	if len(tail) != 2 {
		t.Fatalf("tail = %#v, want pay then lose-game", tail)
	}
	pay, ok := tail[0].Primitive.(game.Pay)
	if !ok || !pay.Payment.ManaCost.Exists {
		t.Fatalf("tail[0] = %#v, want Pay with mana cost", tail[0].Primitive)
	}
	if tail[0].PublishResult == "" {
		t.Fatal("Pay instruction does not publish its result")
	}
	lose, ok := tail[1].Primitive.(game.PlayerLosesGame)
	if !ok {
		t.Fatalf("tail[1] = %#v, want PlayerLosesGame", tail[1].Primitive)
	}
	if lose.Player.Kind() != game.PlayerReferenceController {
		t.Fatalf("lose player = %#v, want controller", lose.Player)
	}
	if !tail[1].ResultGate.Exists ||
		tail[1].ResultGate.Val.Key != tail[0].PublishResult ||
		tail[1].ResultGate.Val.Succeeded != game.TriFalse {
		t.Fatalf("lose-game gate = %#v, want gated on unpaid", tail[1].ResultGate)
	}
}

func TestLowerPactRejectsUnsupportedShapes(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		// Recurring (non-"next") upkeep trigger is not a one-shot delayed pact.
		"Counter target spell.\nAt the beginning of your upkeep, pay {3}{U}{U}. If you don't, you lose the game.",
		// Missing the pay obligation: a bare "next upkeep" trigger stays unsupported.
		"Counter target spell.\nAt the beginning of your next upkeep, you lose the game.",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			_, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Unsupported Pact",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: text,
			})
			if len(diagnostics) == 0 {
				t.Fatal("unsupported pact shape lowered")
			}
		})
	}
}
