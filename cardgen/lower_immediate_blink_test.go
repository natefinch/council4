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
		// Plural group blink is not modeled yet.
		"Exile up to two target creatures you control, then return those cards to the battlefield under their owner's control.",
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
