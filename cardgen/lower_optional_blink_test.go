package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// optionalBlinkMode lowers a triggered or spell body that should produce the
// optional immediate-blink shape and returns its single resolving mode, failing
// the test on any diagnostic or unexpected shell.
func optionalBlinkSpellMode(t *testing.T, name, oracleText string) game.Mode {
	t.Helper()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       name,
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: oracleText,
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not found")
	}
	if len(face.SpellAbility.Val.Modes) != 1 {
		t.Fatalf("modes = %#v, want one", face.SpellAbility.Val.Modes)
	}
	mode := face.SpellAbility.Val.Modes[0]
	if err := game.ValidateInstructionSequence(mode.Sequence, mode.Targets); err != nil {
		t.Fatalf("invalid instruction sequence: %v", err)
	}
	return mode
}

// assertOptionalBlink checks that a lowered blink mode carries one target and a
// two-instruction [Exile, PutOnBattlefield] sequence whose exile is Optional and
// publishes its result and whose put is gated on the exile succeeding.
func assertOptionalBlink(t *testing.T, mode game.Mode) (game.Exile, game.PutOnBattlefield) {
	t.Helper()
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %#v, want one", mode.Targets)
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want two instructions", mode.Sequence)
	}
	exileInstr := mode.Sequence[0]
	exile, ok := exileInstr.Primitive.(game.Exile)
	if !ok || exile.Object != game.TargetPermanentReference(0) || exile.ExileLinkedKey == "" {
		t.Fatalf("instruction[0] = %#v, want linked target exile", exileInstr.Primitive)
	}
	if !exileInstr.Optional {
		t.Fatal("instruction[0].Optional = false, want optional exile")
	}
	if exileInstr.PublishResult != optionalIfYouDoResultKey {
		t.Fatalf("instruction[0].PublishResult = %q, want %q", exileInstr.PublishResult, optionalIfYouDoResultKey)
	}
	if exileInstr.ResultGate.Exists || exileInstr.OptionalActor.Exists {
		t.Fatalf("instruction[0] must carry no gate/actor envelope: %#v", exileInstr)
	}
	putInstr := mode.Sequence[1]
	put, ok := putInstr.Primitive.(game.PutOnBattlefield)
	if !ok {
		t.Fatalf("instruction[1] = %#v, want put on battlefield", putInstr.Primitive)
	}
	key, linked := put.Source.LinkedKey()
	if !linked || key != exile.ExileLinkedKey {
		t.Fatalf("put source = %#v, want linked source %q", put.Source, exile.ExileLinkedKey)
	}
	if putInstr.Optional || putInstr.PublishResult != "" {
		t.Fatalf("instruction[1] gated put must carry no optional/publish envelope: %#v", putInstr)
	}
	if !putInstr.ResultGate.Exists {
		t.Fatal("instruction[1].ResultGate missing, want gate on exile success")
	}
	gate := putInstr.ResultGate.Val
	if gate.Key != optionalIfYouDoResultKey || gate.Succeeded != game.TriTrue {
		t.Fatalf("instruction[1].ResultGate = %#v, want succeeded gate on %q", gate, optionalIfYouDoResultKey)
	}
	return exile, put
}

// TestLowerOptionalBlinkUnderOwnersControl verifies the Soulherder / Felidar
// Guardian shape "you may exile [another] target creature you control, then
// return that card to the battlefield under its owner's control" lowers to an
// optional exile gating an owner's-control return.
func TestLowerOptionalBlinkUnderOwnersControl(t *testing.T) {
	t.Parallel()
	for _, reference := range []string{"that card", "it"} {
		t.Run(reference, func(t *testing.T) {
			t.Parallel()
			mode := optionalBlinkSpellMode(t, "Optional Flicker",
				"You may exile target creature you control, then return "+reference+" to the battlefield under its owner's control.")
			_, put := assertOptionalBlink(t, mode)
			if put.Recipient.Exists {
				t.Fatalf("recipient = %#v, want unset (owner's control)", put.Recipient)
			}
			if put.EntryTapped || len(put.EntryCounters) != 0 {
				t.Fatalf("put = %#v, want untapped with no counters", put)
			}
		})
	}
}

// TestLowerOptionalBlinkUnderYourControl verifies the Conjurer's Closet shape
// "you may exile target creature you control, then return that card to the
// battlefield under your control" gives the returned card to the controller.
func TestLowerOptionalBlinkUnderYourControl(t *testing.T) {
	t.Parallel()
	mode := optionalBlinkSpellMode(t, "Optional Closet",
		"You may exile target creature you control, then return that card to the battlefield under your control.")
	_, put := assertOptionalBlink(t, mode)
	if !put.Recipient.Exists || put.Recipient.Val != game.ControllerReference() {
		t.Fatalf("recipient = %#v, want controller (your control)", put.Recipient)
	}
}

// TestLowerOptionalBlinkTriggeredAbility verifies the most-played form — an
// enters/end-step trigger whose whole body is the optional blink — wires the
// optional exile and gated return on the triggered ability's resolving sequence
// (the trigger itself stays mandatory; the controller chooses on resolution).
func TestLowerOptionalBlinkTriggeredAbility(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Optional Flicker Beast",
		Layout:     "normal",
		TypeLine:   "Creature — Beast",
		OracleText: "When this creature enters, you may exile another target creature you control, then return that card to the battlefield under its owner's control.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	ta := face.TriggeredAbilities[0]
	if ta.Optional {
		t.Fatal("triggered ability Optional = true, want false (optionality on the exile instruction)")
	}
	if len(ta.Content.Modes) != 1 {
		t.Fatalf("modes = %#v, want one", ta.Content.Modes)
	}
	assertOptionalBlink(t, ta.Content.Modes[0])
}

// assertOptionalDelayedBlink checks that a lowered delayed optional blink mode
// carries one target and a two-instruction [Exile, CreateDelayedTrigger] sequence
// whose exile is Optional and publishes its result and whose delayed return is
// gated on the exile succeeding. The wrapped delayed trigger fires at the next
// end step and puts the linked card back onto the battlefield.
func assertOptionalDelayedBlink(t *testing.T, mode game.Mode) (game.Exile, game.PutOnBattlefield) {
	t.Helper()
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %#v, want one", mode.Targets)
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want two instructions", mode.Sequence)
	}
	exileInstr := mode.Sequence[0]
	exile, ok := exileInstr.Primitive.(game.Exile)
	if !ok || exile.Object != game.TargetPermanentReference(0) || exile.ExileLinkedKey == "" {
		t.Fatalf("instruction[0] = %#v, want linked target exile", exileInstr.Primitive)
	}
	if !exileInstr.Optional || exileInstr.PublishResult != optionalIfYouDoResultKey {
		t.Fatalf("instruction[0] = %#v, want optional publishing exile", exileInstr)
	}
	delayedInstr := mode.Sequence[1]
	delayed, ok := delayedInstr.Primitive.(game.CreateDelayedTrigger)
	if !ok {
		t.Fatalf("instruction[1] = %#v, want create delayed trigger", delayedInstr.Primitive)
	}
	if delayed.Trigger.Timing != game.DelayedAtBeginningOfNextEndStep {
		t.Fatalf("delayed timing = %v, want next end step", delayed.Trigger.Timing)
	}
	if !delayedInstr.ResultGate.Exists ||
		delayedInstr.ResultGate.Val.Key != optionalIfYouDoResultKey ||
		delayedInstr.ResultGate.Val.Succeeded != game.TriTrue {
		t.Fatalf("instruction[1].ResultGate = %#v, want succeeded gate on %q", delayedInstr.ResultGate, optionalIfYouDoResultKey)
	}
	put, ok := delayed.Trigger.Content.Modes[0].Sequence[0].Primitive.(game.PutOnBattlefield)
	if !ok {
		t.Fatalf("delayed content = %#v, want put on battlefield", delayed.Trigger.Content.Modes[0].Sequence[0].Primitive)
	}
	key, linked := put.Source.LinkedKey()
	if !linked || key != exile.ExileLinkedKey {
		t.Fatalf("put source = %#v, want linked source %q", put.Source, exile.ExileLinkedKey)
	}
	return exile, put
}

// TestLowerOptionalDelayedBlink verifies the Astral Slide / Astral Drift shape
// "you may exile target creature. If you do, return that card to the battlefield
// under its owner's control at the beginning of the next end step" lowers to an
// optional exile gating a delayed-trigger return — the delayed sibling of the
// immediate optional blink.
func TestLowerOptionalDelayedBlink(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Optional Delayed Flicker",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Whenever a player cycles a card, you may exile target creature. If you do, return that card to the battlefield under its owner's control at the beginning of the next end step.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	ta := face.TriggeredAbilities[0]
	if ta.Optional {
		t.Fatal("triggered ability Optional = true, want false (optionality on the exile instruction)")
	}
	mode := ta.Content.Modes[0]
	if err := game.ValidateInstructionSequence(mode.Sequence, mode.Targets); err != nil {
		t.Fatalf("invalid instruction sequence: %v", err)
	}
	_, put := assertOptionalDelayedBlink(t, mode)
	if put.Recipient.Exists {
		t.Fatalf("recipient = %#v, want unset (owner's control)", put.Recipient)
	}
}

// TestLowerOptionalBlinkFailsClosed verifies optional exile-then-return variants
// outside the supported immediate single-target blink remain rejected with a
// fail-closed diagnostic rather than lowering to silently-wrong behavior.
func TestLowerOptionalBlinkFailsClosed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		typeLine   string
		oracleText string
	}{
		// Plural group blink is not modeled by the underlying immediate-blink
		// lowerer, so the optional wrapper must fail closed too.
		{"plural group blink", "Instant",
			"You may exile up to two target creatures you control, then return those cards to the battlefield under their owners' control."},
		// An unsupported exile selector still blocks the sequence underneath the
		// optional wrapper.
		{"unsupported selector", "Instant",
			"You may exile target multicolored permanent, then return it to the battlefield under its owner's control."},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Optional Blink Reject",
				Layout:     "normal",
				TypeLine:   test.typeLine,
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
