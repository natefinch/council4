package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

// blinkInstructions extracts the exile and put-onto-battlefield instructions of a
// two-step immediate blink spell mode, failing the test if the shape differs.
func blinkInstructions(t *testing.T, mode game.Mode) (game.Exile, game.PutOnBattlefield) {
	t.Helper()
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want two instructions", mode.Sequence)
	}
	exile, ok := mode.Sequence[0].Primitive.(game.Exile)
	if !ok || exile.Object != game.TargetPermanentReference(0) || exile.ExileLinkedKey == "" {
		t.Fatalf("exile = %#v, want linked target exile", mode.Sequence[0].Primitive)
	}
	put, ok := mode.Sequence[1].Primitive.(game.PutOnBattlefield)
	if !ok {
		t.Fatalf("second primitive = %#v, want put on battlefield", mode.Sequence[1].Primitive)
	}
	key, linked := put.Source.LinkedKey()
	if !linked || key != exile.ExileLinkedKey {
		t.Fatalf("put source = %#v, want linked source %q", put.Source, exile.ExileLinkedKey)
	}
	return exile, put
}

func TestLowerImmediateBlinkUnderOwnersControl(t *testing.T) {
	t.Parallel()
	for _, reference := range []string{"that card", "it"} {
		t.Run(reference, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Flicker",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: "Exile target creature you control, then return " + reference + " to the battlefield under its owner's control.",
			})
			mode := face.SpellAbility.Val.Modes[0]
			if len(mode.Targets) != 1 {
				t.Fatalf("targets = %#v, want one", mode.Targets)
			}
			_, put := blinkInstructions(t, mode)
			if put.Recipient.Exists {
				t.Fatalf("recipient = %#v, want unset (owner's control)", put.Recipient)
			}
			if put.EntryTapped || len(put.EntryCounters) != 0 {
				t.Fatalf("put = %#v, want untapped with no counters", put)
			}
		})
	}
}

func TestLowerImmediateBlinkUnderYourControl(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Cloudshift",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Exile target creature you control, then return that card to the battlefield under your control.",
	})
	_, put := blinkInstructions(t, face.SpellAbility.Val.Modes[0])
	if !put.Recipient.Exists || put.Recipient.Val != game.ControllerReference() {
		t.Fatalf("recipient = %#v, want controller", put.Recipient)
	}
}

func TestLowerImmediateBlinkTapped(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Displacer",
		Layout:     "normal",
		TypeLine:   "Creature",
		OracleText: "{2}: Exile another target creature, then return it to the battlefield tapped under its owner's control.",
	})
	_, put := blinkInstructions(t, face.ActivatedAbilities[0].Content.Modes[0])
	if !put.EntryTapped {
		t.Fatalf("put = %#v, want entry tapped", put)
	}
}

func TestLowerImmediateBlinkWithCounter(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Resolve",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Exile target creature you control, then return it to the battlefield under its owner's control with a +1/+1 counter on it.",
	})
	_, put := blinkInstructions(t, face.SpellAbility.Val.Modes[0])
	want := []game.CounterPlacement{{Kind: counter.PlusOnePlusOne, Amount: 1}}
	if len(put.EntryCounters) != 1 || put.EntryCounters[0] != want[0] {
		t.Fatalf("entry counters = %#v, want %#v", put.EntryCounters, want)
	}
}

// TestLowerImmediateBlinkRejectsUnsupportedVariants confirms the immediate blink
// lowerer fails closed for shapes it does not fully model, most importantly the
// leading-position delayed wording whose timing the parser does not capture and
// which must therefore never resolve at once.
func TestLowerImmediateBlinkRejectsUnsupportedVariants(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		// Leading-position delayed return: must not be lowered as an immediate blink.
		"Exile target creature. At the beginning of the next end step, return that card to the battlefield under its owner's control.",
		"Exile target creature. At the beginning of the next end step, return it to the battlefield under its owner's control with a +1/+1 counter on it.",
		// Exile of a non-supported selector still blocks the sequence.
		"Exile target nontoken permanent, then return it to the battlefield under its owner's control.",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			_, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Test Unsupported Blink",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: text,
			})
			if len(diagnostics) == 0 {
				t.Fatal("expected unsupported blink variant to fail closed")
			}
		})
	}
}

// groupBlinkMode lowers a group blink spell and returns its single mode, failing
// the test if the spell did not lower or carries diagnostics.
func groupBlinkMode(t *testing.T, oracle string) game.Mode {
	t.Helper()
	face, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Group Blink",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: oracle,
	})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v, want none", diagnostics)
	}
	if len(face) != 1 {
		t.Fatalf("faces = %d, want one", len(face))
	}
	return face[0].SpellAbility.Val.Modes[0]
}

func TestLowerGroupBlinkImmediate(t *testing.T) {
	t.Parallel()
	// "Exile up to two target creatures you control, then return those cards to
	// the battlefield under their owner's control." (Illusionist's Stratagem)
	mode := groupBlinkMode(t,
		"Exile up to two target creatures you control, then return those cards to the battlefield under their owner's control.")
	if len(mode.Targets) != 1 || mode.Targets[0].MaxTargets != 2 {
		t.Fatalf("targets = %#v, want one spec with max two", mode.Targets)
	}
	if len(mode.Sequence) != 4 {
		t.Fatalf("sequence = %#v, want two exiles and two puts", mode.Sequence)
	}
	for i := range 2 {
		exile, ok := mode.Sequence[i].Primitive.(game.Exile)
		if !ok || exile.Object != game.TargetPermanentReference(i) || exile.ExileLinkedKey == "" {
			t.Fatalf("instruction[%d] = %#v, want linked target exile", i, mode.Sequence[i].Primitive)
		}
		put, ok := mode.Sequence[2+i].Primitive.(game.PutOnBattlefield)
		if !ok {
			t.Fatalf("instruction[%d] = %#v, want put on battlefield", 2+i, mode.Sequence[2+i].Primitive)
		}
		key, linked := put.Source.LinkedKey()
		if !linked || key != exile.ExileLinkedKey {
			t.Fatalf("put source = %#v, want linked source %q", put.Source, exile.ExileLinkedKey)
		}
		if put.Recipient.Exists {
			t.Fatalf("recipient = %#v, want unset (owner's control)", put.Recipient)
		}
	}
}

func TestLowerGroupBlinkUnderYourControl(t *testing.T) {
	t.Parallel()
	// "Exile two target artifacts, creatures, and/or lands you control, then
	// return those cards to the battlefield under your control." (Ghostly Flicker)
	mode := groupBlinkMode(t,
		"Exile two target artifacts, creatures, and/or lands you control, then return those cards to the battlefield under your control.")
	if len(mode.Targets) != 1 || mode.Targets[0].MaxTargets != 2 {
		t.Fatalf("targets = %#v, want one spec with max two", mode.Targets)
	}
	if len(mode.Sequence) != 4 {
		t.Fatalf("sequence = %#v, want two exiles and two puts", mode.Sequence)
	}
	put, ok := mode.Sequence[2].Primitive.(game.PutOnBattlefield)
	if !ok || !put.Recipient.Exists || put.Recipient.Val != game.ControllerReference() {
		t.Fatalf("instruction[2] = %#v, want put under controller", mode.Sequence[2].Primitive)
	}
}

func TestLowerGroupBlinkDelayed(t *testing.T) {
	t.Parallel()
	// "Exile any number of target creatures you control. Return those cards to
	// the battlefield ... at the beginning of the next end step." returns inside
	// a delayed trigger. "any number" exiles do not lower exactly, so use a fixed
	// count to exercise the delayed group return.
	mode := groupBlinkMode(t,
		"Exile two target creatures you control. Return those cards to the battlefield under their owner's control at the beginning of the next end step.")
	if len(mode.Sequence) != 3 {
		t.Fatalf("sequence = %#v, want two exiles and one delayed trigger", mode.Sequence)
	}
	delayed, ok := mode.Sequence[2].Primitive.(game.CreateDelayedTrigger)
	if !ok {
		t.Fatalf("instruction[2] = %#v, want delayed trigger", mode.Sequence[2].Primitive)
	}
	if delayed.Trigger.Timing != game.DelayedAtBeginningOfNextEndStep {
		t.Fatalf("timing = %v, want next end step", delayed.Trigger.Timing)
	}
	if len(delayed.Trigger.Content.Modes[0].Sequence) != 2 {
		t.Fatalf("delayed content = %#v, want two puts", delayed.Trigger.Content.Modes[0].Sequence)
	}
}
