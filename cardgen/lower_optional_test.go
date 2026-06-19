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

// TestLowerOptionalIfYouDoMultipleGatedEffects verifies that a single "if you
// do" clause may gate several and-joined trailing effects ("you may X. If you
// do, Y and Z"): the optional effect publishes its result and every trailing
// effect is gated on that result having succeeded.
func TestLowerOptionalIfYouDoMultipleGatedEffects(t *testing.T) {
	t.Parallel()
	sequence := lowerSpellSequence(t, "Optional Multi Gate",
		"You may discard a card. If you do, draw a card and you gain 2 life.")
	if len(sequence) != 3 {
		t.Fatalf("sequence = %#v, want three instructions", sequence)
	}
	discard := sequence[0]
	if _, ok := discard.Primitive.(game.Discard); !ok {
		t.Fatalf("instruction[0] = %T, want game.Discard", discard.Primitive)
	}
	if !discard.Optional || discard.PublishResult != optionalIfYouDoResultKey {
		t.Fatalf("instruction[0] discard not wired optional: %#v", discard)
	}
	if discard.ResultGate.Exists {
		t.Fatalf("instruction[0] discard must not be gated: %#v", discard)
	}
	if _, ok := sequence[1].Primitive.(game.Draw); !ok {
		t.Fatalf("instruction[1] = %T, want game.Draw", sequence[1].Primitive)
	}
	if _, ok := sequence[2].Primitive.(game.GainLife); !ok {
		t.Fatalf("instruction[2] = %T, want game.GainLife", sequence[2].Primitive)
	}
	for i := 1; i < len(sequence); i++ {
		gate := sequence[i].ResultGate
		if !gate.Exists || gate.Val.Key != optionalIfYouDoResultKey || gate.Val.Succeeded != game.TriTrue {
			t.Fatalf("instruction[%d] not gated on if-you-do success: %#v", i, sequence[i])
		}
		if sequence[i].Optional || sequence[i].PublishResult != "" {
			t.Fatalf("instruction[%d] gated effect must carry no optional/publish envelope: %#v", i, sequence[i])
		}
	}
}

// TestLowerSingleOptionalEffect verifies that a one-effect "you may X" body
// lowers to a single instruction marked Optional (the runtime asks the
// controller whether to apply it) with no result-publish/gate envelope.
func TestLowerSingleOptionalEffect(t *testing.T) {
	t.Parallel()
	sequence := lowerSpellSequence(t, "Optional Discard", "You may discard a card.")
	if len(sequence) != 1 {
		t.Fatalf("sequence = %#v, want one instruction", sequence)
	}
	instr := sequence[0]
	if _, ok := instr.Primitive.(game.Discard); !ok {
		t.Fatalf("instruction[0] = %T, want game.Discard", instr.Primitive)
	}
	if !instr.Optional {
		t.Fatal("instruction[0].Optional = false, want optional")
	}
	if instr.PublishResult != "" || instr.ResultGate.Exists {
		t.Fatalf("single optional effect must carry no result envelope: %#v", instr)
	}
}

// TestLowerSingleOptionalTargetedEffect verifies that a one-effect "you may X"
// body whose effect targets keeps the mode target (chosen when the spell is put
// on the stack) and marks the resolving instruction Optional.
func TestLowerSingleOptionalTargetedEffect(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Optional Strike",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "You may destroy target creature.",
	})
	if !face.SpellAbility.Exists || len(face.SpellAbility.Val.Modes) != 1 {
		t.Fatalf("spell ability not a single mode: %#v", face.SpellAbility)
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("mode targets = %#v, want one", mode.Targets)
	}
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %#v, want one instruction", mode.Sequence)
	}
	if _, ok := mode.Sequence[0].Primitive.(game.Destroy); !ok {
		t.Fatalf("instruction[0] = %T, want game.Destroy", mode.Sequence[0].Primitive)
	}
	if !mode.Sequence[0].Optional {
		t.Fatal("instruction[0].Optional = false, want optional")
	}
}

// TestLowerSingleOptionalLifeGain verifies a one-effect "You may gain N life."
// body lowers to a single GainLife instruction marked Optional. The optional
// life effect reconstructs its canonical clause byte-exactly, so the exact life
// recognizer now accepts it and the single-optional-effect path marks it.
func TestLowerSingleOptionalLifeGain(t *testing.T) {
	t.Parallel()
	sequence := lowerSpellSequence(t, "Optional Gain", "You may gain 3 life.")
	if len(sequence) != 1 {
		t.Fatalf("sequence = %#v, want one instruction", sequence)
	}
	gain, ok := sequence[0].Primitive.(game.GainLife)
	if !ok {
		t.Fatalf("instruction[0] = %T, want game.GainLife", sequence[0].Primitive)
	}
	if gain.Amount != game.Fixed(3) {
		t.Errorf("amount = %#v, want fixed 3", gain.Amount)
	}
	if !sequence[0].Optional {
		t.Fatal("instruction[0].Optional = false, want optional")
	}
	if sequence[0].PublishResult != "" || sequence[0].ResultGate.Exists {
		t.Fatalf("single optional effect must carry no result envelope: %#v", sequence[0])
	}
}

// TestLowerSingleOptionalTokenCreation verifies a one-effect "You may create ...
// token." body lowers to a single CreateToken instruction marked Optional.
func TestLowerSingleOptionalTokenCreation(t *testing.T) {
	t.Parallel()
	sequence := lowerSpellSequence(t, "Optional Token", "You may create a 1/1 white Soldier creature token.")
	if len(sequence) != 1 {
		t.Fatalf("sequence = %#v, want one instruction", sequence)
	}
	if _, ok := sequence[0].Primitive.(game.CreateToken); !ok {
		t.Fatalf("instruction[0] = %T, want game.CreateToken", sequence[0].Primitive)
	}
	if !sequence[0].Optional {
		t.Fatal("instruction[0].Optional = false, want optional")
	}
}

// TestLowerTriggerOptionalLifeGain verifies an enters-trigger whose whole body is
// a resolving "you may gain N life" marks the triggered ability Optional (the
// trigger fires, then the controller is asked whether to gain), with the lone
// instruction left mandatory.
func TestLowerTriggerOptionalLifeGain(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Optional Gain Beast",
		Layout:     "normal",
		TypeLine:   "Creature — Beast",
		OracleText: "When this creature enters, you may gain 2 life.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	ta := face.TriggeredAbilities[0]
	if !ta.Optional {
		t.Error("triggered ability Optional = false, want true")
	}
	sequence := ta.Content.Modes[0].Sequence
	if len(sequence) != 1 {
		t.Fatalf("sequence = %#v, want one instruction", sequence)
	}
	if _, ok := sequence[0].Primitive.(game.GainLife); !ok {
		t.Fatalf("instruction[0] = %T, want game.GainLife", sequence[0].Primitive)
	}
	if sequence[0].Optional {
		t.Error("instruction[0].Optional = true, want false (optionality on the ability)")
	}
}

// TestLowerOptionalFlowFailsClosed verifies that optional-flow variants outside
// the supported "you may X. If you do, Y" pair and single-optional-effect shapes
// remain rejected with the optional-effect diagnostic rather than lowering to
// silently-wrong behavior.
func TestLowerOptionalFlowFailsClosed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
	}{
		{"otherwise branch", "You may discard a card. If you do, draw a card. Otherwise, draw a card."},
		{"if you don't branch", "You may discard a card. If you don't, draw a card."},
		{"two optional effects", "You may discard a card. If you do, you may draw a card."},
		{"optional without if-you-do", "You may discard a card. Draw a card."},
		// An independent effect after the gated "if you do" tail ("Scry 2.")
		// does not structurally contain the gate condition, so it would resolve
		// unconditionally. The flow must reject the whole body rather than gate
		// only part of it.
		{"if-you-do independent tail", "You may discard a card. If you do, draw a card. Scry 2."},
		// Single optional effect whose inner effect (putting a permanent from
		// hand onto the battlefield) is itself unsupported must still fail
		// closed rather than emit a partial card.
		{"single optional unsupported inner", "You may put a creature card from your hand onto the battlefield."},
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
