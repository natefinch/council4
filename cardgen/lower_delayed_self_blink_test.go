package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// delayedSelfBlinkInstructions extracts the exile and delayed-trigger
// put-onto-battlefield instructions of a two-step delayed self-blink mode, where
// the exiled object is the source permanent itself ("Exile this creature. Return
// it … at the beginning of the next end step.") and the return is wrapped in a
// next-end-step delayed trigger rather than resolved immediately.
func delayedSelfBlinkInstructions(t *testing.T, mode game.Mode) (game.Exile, game.PutOnBattlefield) {
	t.Helper()
	if len(mode.Targets) != 0 {
		t.Fatalf("targets = %#v, want none for self-blink", mode.Targets)
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want exile and delayed trigger", mode.Sequence)
	}
	exile, ok := mode.Sequence[0].Primitive.(game.Exile)
	if !ok || exile.Object != game.SourcePermanentReference() || exile.ExileLinkedKey == "" {
		t.Fatalf("exile = %#v, want linked source exile", mode.Sequence[0].Primitive)
	}
	delayed, ok := mode.Sequence[1].Primitive.(game.CreateDelayedTrigger)
	if !ok || delayed.Trigger.Timing != game.DelayedAtBeginningOfNextEndStep {
		t.Fatalf("instruction[1] = %#v, want next-end-step delayed trigger", mode.Sequence[1].Primitive)
	}
	inner := delayed.Trigger.Content.Modes[0].Sequence
	if len(inner) != 1 {
		t.Fatalf("delayed content = %#v, want one put on battlefield", inner)
	}
	put, ok := inner[0].Primitive.(game.PutOnBattlefield)
	if !ok {
		t.Fatalf("delayed content = %#v, want put on battlefield", inner[0].Primitive)
	}
	key, linked := put.Source.LinkedKey()
	if !linked || key != exile.ExileLinkedKey {
		t.Fatalf("put source = %#v, want linked source %q", put.Source, exile.ExileLinkedKey)
	}
	return exile, put
}

func TestLowerDelayedSelfBlinkUnderYourControl(t *testing.T) {
	t.Parallel()
	// Argent Sphinx: "{U}: Exile this creature. Return it to the battlefield
	// under your control at the beginning of the next end step."
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Argent Sphinx",
		Layout:     "normal",
		TypeLine:   "Creature — Sphinx",
		OracleText: "{U}: Exile this creature. Return it to the battlefield under your control at the beginning of the next end step.",
	})
	_, put := delayedSelfBlinkInstructions(t, face.ActivatedAbilities[0].Content.Modes[0])
	if !put.Recipient.Exists || put.Recipient.Val != game.ControllerReference() {
		t.Fatalf("recipient = %#v, want controller", put.Recipient)
	}
	if put.EntryTapped || len(put.EntryCounters) != 0 {
		t.Fatalf("put = %#v, want untapped with no counters", put)
	}
}

func TestLowerDelayedSelfBlinkUnderOwnersControl(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Delayed Self Blink",
		Layout:     "normal",
		TypeLine:   "Creature",
		OracleText: "{2}: Exile this creature. Return it to the battlefield under its owner's control at the beginning of the next end step.",
	})
	_, put := delayedSelfBlinkInstructions(t, face.ActivatedAbilities[0].Content.Modes[0])
	if put.Recipient.Exists {
		t.Fatalf("recipient = %#v, want unset (owner's control)", put.Recipient)
	}
}

func TestLowerDelayedSelfBlinkTapped(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Tapped Self Blink",
		Layout:     "normal",
		TypeLine:   "Creature",
		OracleText: "{2}: Exile this creature. Return it to the battlefield tapped under its owner's control at the beginning of the next end step.",
	})
	_, put := delayedSelfBlinkInstructions(t, face.ActivatedAbilities[0].Content.Modes[0])
	if !put.EntryTapped {
		t.Fatalf("put = %#v, want entry tapped", put)
	}
}

// TestLowerDelayedSelfBlinkReturnThisCreature confirms the "Return this creature"
// direct-object wording (Saltskitter) lowers identically to the "Return it"
// pronoun form, exercising the ReferenceThisObject branch of the return body.
func TestLowerDelayedSelfBlinkReturnThisCreature(t *testing.T) {
	t.Parallel()
	// Saltskitter trigger body.
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Saltskitter",
		Layout:     "normal",
		TypeLine:   "Creature — Wurm",
		OracleText: "Whenever another creature enters, exile this creature. Return this creature to the battlefield under its owner's control at the beginning of the next end step.",
	})
	_, put := delayedSelfBlinkInstructions(t, face.TriggeredAbilities[0].Content.Modes[0])
	if put.Recipient.Exists {
		t.Fatalf("recipient = %#v, want unset (owner's control)", put.Recipient)
	}
}

// TestLowerDelayedSelfBlinkRejectsStandaloneSelfExile confirms a bare delayed
// self-exile with no return clause fails closed rather than being promoted to a
// delayed self-blink.
func TestLowerDelayedSelfBlinkRejectsStandaloneSelfExile(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Standalone Delayed Self Exile",
		Layout:     "normal",
		TypeLine:   "Creature",
		OracleText: "{2}: Exile this creature.",
	})
	if len(diagnostics) == 0 {
		t.Fatal("expected standalone self-exile to fail closed")
	}
}
