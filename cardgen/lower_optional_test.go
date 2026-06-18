package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// lowerSpellSequence lowers a sorcery body and returns its resolving
// instruction sequence, failing the test on any diagnostic.
func lowerSpellSequence(t *testing.T, name, oracleText string) []game.Instruction {
	t.Helper()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       name,
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: oracleText,
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not found")
	}
	if len(face.SpellAbility.Val.Modes) != 1 {
		t.Fatalf("modes = %#v, want one", face.SpellAbility.Val.Modes)
	}
	sequence := face.SpellAbility.Val.Modes[0].Sequence
	if err := game.ValidateInstructionSequence(sequence, face.SpellAbility.Val.Modes[0].Targets); err != nil {
		t.Fatalf("invalid instruction sequence: %v", err)
	}
	return sequence
}

func TestLowerOptionalIfYouDoDiscardDraw(t *testing.T) {
	t.Parallel()
	sequence := lowerSpellSequence(t, "Optional Flow Test", "You may discard a card. If you do, draw two cards.")
	if len(sequence) != 2 {
		t.Fatalf("sequence = %#v, want two instructions", sequence)
	}
	discard := sequence[0]
	if _, ok := discard.Primitive.(game.Discard); !ok {
		t.Fatalf("instruction[0] = %T, want game.Discard", discard.Primitive)
	}
	if !discard.Optional {
		t.Fatal("instruction[0].Optional = false, want optional")
	}
	if discard.PublishResult != optionalIfYouDoResultKey {
		t.Fatalf("instruction[0].PublishResult = %q, want %q", discard.PublishResult, optionalIfYouDoResultKey)
	}
	draw := sequence[1]
	if _, ok := draw.Primitive.(game.Draw); !ok {
		t.Fatalf("instruction[1] = %T, want game.Draw", draw.Primitive)
	}
	if !draw.ResultGate.Exists {
		t.Fatal("instruction[1].ResultGate missing")
	}
	gate := draw.ResultGate.Val
	if gate.Key != optionalIfYouDoResultKey || gate.Succeeded != game.TriTrue {
		t.Fatalf("instruction[1].ResultGate = %#v, want succeeded gate on %q", gate, optionalIfYouDoResultKey)
	}
}

func TestLowerOptionalIfYouDoAfterLeadingEffect(t *testing.T) {
	t.Parallel()
	sequence := lowerSpellSequence(t, "Singe",
		"Singe deals 3 damage to target creature. You may discard a card. If you do, draw a card.")
	if len(sequence) != 3 {
		t.Fatalf("sequence = %#v, want three instructions", sequence)
	}
	if _, ok := sequence[0].Primitive.(game.Damage); !ok {
		t.Fatalf("instruction[0] = %T, want game.Damage", sequence[0].Primitive)
	}
	if sequence[0].Optional || sequence[0].PublishResult != "" || sequence[0].ResultGate.Exists {
		t.Fatalf("leading damage must carry no optional-flow envelope: %#v", sequence[0])
	}
	if !sequence[1].Optional || sequence[1].PublishResult != optionalIfYouDoResultKey {
		t.Fatalf("instruction[1] discard not wired optional: %#v", sequence[1])
	}
	if !sequence[2].ResultGate.Exists || sequence[2].ResultGate.Val.Succeeded != game.TriTrue {
		t.Fatalf("instruction[2] draw not gated on success: %#v", sequence[2])
	}
}

// TestLowerOptionalFlowFailsClosed verifies that optional-flow variants outside
// the supported "you may X. If you do, Y" pair remain rejected with the
// optional-effect diagnostic rather than lowering to silently-wrong behavior.
func TestLowerOptionalFlowFailsClosed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
	}{
		{"otherwise branch", "You may discard a card. If you do, draw a card. Otherwise, draw a card."},
		{"if you don't branch", "You may discard a card. If you don't, draw a card."},
		{"single optional effect", "You may discard a card."},
		{"two optional effects", "You may discard a card. If you do, you may draw a card."},
		{"optional without if-you-do", "You may discard a card. Draw a card."},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Optional Flow Reject",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: test.oracleText,
			}, "o")
			if err != nil {
				t.Fatal(err)
			}
			if source != "" {
				t.Fatalf("source = %q, want no partial card", source)
			}
			if len(diagnostics) == 0 {
				t.Fatal("diagnostics = none, want fail-closed rejection")
			}
		})
	}
}
