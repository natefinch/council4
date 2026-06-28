package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// optionalWheelSequence asserts the lowered face carries exactly the optional
// whole-hand discard "wheel" template: an optional whole-hand Discard that
// publishes its discard count, followed by a Draw whose dynamic amount reads
// that published count and whose result gate requires the discard to have been
// accepted. It returns the resolved discard result key so callers can confirm
// the two instructions share it.
func optionalWheelSequence(t *testing.T, seq []game.Instruction) {
	t.Helper()
	if len(seq) != 2 {
		t.Fatalf("sequence len = %d, want 2", len(seq))
	}

	discard, ok := seq[0].Primitive.(game.Discard)
	if !ok {
		t.Fatalf("instruction 0 primitive = %#v, want Discard", seq[0].Primitive)
	}
	if !discard.EntireHand {
		t.Fatal("discard EntireHand = false, want true")
	}
	if !seq[0].Optional {
		t.Fatal("discard instruction Optional = false, want true")
	}
	if seq[0].PublishResult == "" {
		t.Fatal("discard instruction PublishResult is empty, want a result key")
	}

	draw, ok := seq[1].Primitive.(game.Draw)
	if !ok {
		t.Fatalf("instruction 1 primitive = %#v, want Draw", seq[1].Primitive)
	}
	dynV := draw.Amount.DynamicAmount()
	if !dynV.Exists {
		t.Fatalf("draw amount is not dynamic: %#v", draw.Amount)
	}
	dyn := dynV.Val
	if dyn.Kind != game.DynamicAmountPreviousEffectResult {
		t.Fatalf("draw dynamic kind = %v, want DynamicAmountPreviousEffectResult", dyn.Kind)
	}
	if dyn.ResultKey != seq[0].PublishResult {
		t.Fatalf("draw result key = %q, want %q", dyn.ResultKey, seq[0].PublishResult)
	}
	if !seq[1].ResultGate.Exists {
		t.Fatal("draw instruction has no ResultGate, want one")
	}
	gate := seq[1].ResultGate.Val
	if gate.Key != seq[0].PublishResult {
		t.Fatalf("draw gate key = %q, want %q", gate.Key, seq[0].PublishResult)
	}
	if gate.Accepted != game.TriTrue {
		t.Fatalf("draw gate Accepted = %v, want TriTrue", gate.Accepted)
	}
}

// TestLowerOptionalWheelDiscardDrawTriggered proves the optional whole-hand
// "wheel" lowers inside a triggered ability for both the verbose "all the cards
// in your hand" wording (Book Devourer, Forgotten Creation) and the terse "your
// hand" wording, producing the optional-discard-then-gated-dynamic-draw
// template.
func TestLowerOptionalWheelDiscardDrawTriggered(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		oracleText string
	}{
		{
			name: "Book Devourer",
			oracleText: "Trample\n" +
				"Whenever this creature deals combat damage to a player, you may discard all the cards in your hand. If you do, draw that many cards.",
		},
		{
			name:       "Forgotten Creation",
			oracleText: "At the beginning of your upkeep, you may discard all the cards in your hand. If you do, draw that many cards.",
		},
		{
			name:       "Terse Wheel",
			oracleText: "At the beginning of your upkeep, you may discard your hand. If you do, draw that many cards.",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       tc.name,
				TypeLine:   "Creature — Zombie Horror",
				OracleText: tc.oracleText,
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
			}
			content := face.TriggeredAbilities[0].Content
			if len(content.Modes) != 1 {
				t.Fatalf("content modes = %d, want 1", len(content.Modes))
			}
			optionalWheelSequence(t, content.Modes[0].Sequence)
		})
	}
}
