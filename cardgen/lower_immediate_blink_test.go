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

// selfBlinkInstructions extracts the exile and put-onto-battlefield instructions
// of a two-step self-blink mode, where the exiled object is the source permanent
// itself ("Exile this creature, then return it …") rather than a target.
func selfBlinkInstructions(t *testing.T, mode game.Mode) (game.Exile, game.PutOnBattlefield) {
	t.Helper()
	if len(mode.Targets) != 0 {
		t.Fatalf("targets = %#v, want none for self-blink", mode.Targets)
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want two instructions", mode.Sequence)
	}
	exile, ok := mode.Sequence[0].Primitive.(game.Exile)
	if !ok || exile.Object != game.SourcePermanentReference() || exile.ExileLinkedKey == "" {
		t.Fatalf("exile = %#v, want linked source exile", mode.Sequence[0].Primitive)
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

func TestLowerSelfBlinkUnderOwnersControl(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Flickering Spirit",
		Layout:     "normal",
		TypeLine:   "Creature — Spirit",
		OracleText: "{3}{W}: Exile this creature, then return it to the battlefield under its owner's control.",
	})
	_, put := selfBlinkInstructions(t, face.ActivatedAbilities[0].Content.Modes[0])
	if put.Recipient.Exists {
		t.Fatalf("recipient = %#v, want unset (owner's control)", put.Recipient)
	}
	if put.EntryTapped || len(put.EntryCounters) != 0 {
		t.Fatalf("put = %#v, want untapped with no counters", put)
	}
}

func TestLowerSelfBlinkTapped(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Self Blink",
		Layout:     "normal",
		TypeLine:   "Creature",
		OracleText: "{2}: Exile this creature, then return it to the battlefield tapped under its owner's control.",
	})
	_, put := selfBlinkInstructions(t, face.ActivatedAbilities[0].Content.Modes[0])
	if !put.EntryTapped {
		t.Fatalf("put = %#v, want entry tapped", put)
	}
}

// TestLowerSelfBlinkStandaloneSelfExileIsPlainExile confirms a bare "Exile this
// creature" (no return) is not promoted to a self-blink: it lowers to a plain
// source-permanent exile, since the self-blink shape requires a return clause.
func TestLowerSelfBlinkStandaloneSelfExileIsPlainExile(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Standalone Self Exile",
		Layout:     "normal",
		TypeLine:   "Creature",
		OracleText: "{2}: Exile this creature.",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	sequence := face.ActivatedAbilities[0].Content.Modes[0].Sequence
	if len(sequence) != 1 {
		t.Fatalf("sequence = %#v, want a single plain exile", sequence)
	}
	exile, ok := sequence[0].Primitive.(game.Exile)
	if !ok || exile.ExileLinkedKey != "" || exile.SourceSpell ||
		exile.Object != game.SourceCardPermanentReference() {
		t.Fatalf("instruction = %#v, want plain source-permanent exile", sequence[0].Primitive)
	}
}

// TestLowerImmediateBlinkRejectsUnsupportedVariants confirms the immediate blink
// lowerer fails closed for shapes it does not fully model, such as an exile of a
// selector the sequence cannot model.
func TestLowerImmediateBlinkRejectsUnsupportedVariants(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		// Exile of a non-supported selector still blocks the sequence. A
		// "historic" permanent target is not representable, so it fails closed.
		"Exile target historic permanent, then return it to the battlefield under its owner's control.",
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

// TestLowerLeadingDelayedBlinkLowersAsDelayed confirms the leading-position
// delayed wording ("At the beginning of the next end step, return that card ...")
// is captured by the parser and lowered as a delayed return, identical to the
// trailing-position wording, rather than resolving the return at once.
func TestLowerLeadingDelayedBlinkLowersAsDelayed(t *testing.T) {
	t.Parallel()
	face, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Leading Delayed Blink",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Exile target creature. At the beginning of the next end step, return that card to the battlefield under its owner's control.",
	})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %v, want none", diagnostics)
	}
	sequence := face[0].SpellAbility.Val.Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("sequence length = %d, want 2", len(sequence))
	}
	delayed, ok := sequence[1].Primitive.(game.CreateDelayedTrigger)
	if !ok {
		t.Fatalf("sequence[1] = %T, want CreateDelayedTrigger", sequence[1].Primitive)
	}
	if delayed.Trigger.Timing != game.DelayedAtBeginningOfNextEndStep {
		t.Fatalf("timing = %v, want next end step", delayed.Trigger.Timing)
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

// TestLowerGroupBlinkAnyNumberDelayed confirms the unbounded "any number of
// target" group blink (Eerie Interlude) lowers as a single linked exile of every
// chosen permanent plus one delayed group return, rather than unrolling a fixed
// slot per target.
func TestLowerGroupBlinkAnyNumberDelayed(t *testing.T) {
	t.Parallel()
	mode := groupBlinkMode(t,
		"Exile any number of target creatures you control. Return those cards to the battlefield under their owner's control at the beginning of the next end step.")
	if len(mode.Targets) != 1 || mode.Targets[0].MinTargets != 0 || mode.Targets[0].MaxTargets != 99 {
		t.Fatalf("targets = %#v, want one any-number spec", mode.Targets)
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want one exile and one delayed trigger", mode.Sequence)
	}
	exile, ok := mode.Sequence[0].Primitive.(game.Exile)
	if !ok || exile.Object != game.AllTargetPermanentsReference(0) || exile.ExileLinkedKey == "" {
		t.Fatalf("exile = %#v, want linked all-target-permanents exile", mode.Sequence[0].Primitive)
	}
	delayed, ok := mode.Sequence[1].Primitive.(game.CreateDelayedTrigger)
	if !ok || delayed.Trigger.Timing != game.DelayedAtBeginningOfNextEndStep {
		t.Fatalf("instruction[1] = %#v, want next-end-step delayed trigger", mode.Sequence[1].Primitive)
	}
	put, ok := delayed.Trigger.Content.Modes[0].Sequence[0].Primitive.(game.PutOnBattlefield)
	if !ok {
		t.Fatalf("delayed content = %#v, want one put on battlefield", delayed.Trigger.Content.Modes[0].Sequence)
	}
	key, linked := put.Source.LinkedKey()
	if !linked || key != exile.ExileLinkedKey {
		t.Fatalf("put source = %#v, want linked source %q", put.Source, exile.ExileLinkedKey)
	}
}

// TestLowerGroupBlinkAnyNumberImmediate confirms the unbounded "any number of
// target" group blink returns immediately when the return is connected with
// "then" rather than delayed to the next end step.
func TestLowerGroupBlinkAnyNumberImmediate(t *testing.T) {
	t.Parallel()
	mode := groupBlinkMode(t,
		"Exile any number of target creatures you control, then return those cards to the battlefield under their owner's control.")
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want one exile and one put", mode.Sequence)
	}
	exile, ok := mode.Sequence[0].Primitive.(game.Exile)
	if !ok || exile.Object != game.AllTargetPermanentsReference(0) || exile.ExileLinkedKey == "" {
		t.Fatalf("exile = %#v, want linked all-target-permanents exile", mode.Sequence[0].Primitive)
	}
	put, ok := mode.Sequence[1].Primitive.(game.PutOnBattlefield)
	if !ok {
		t.Fatalf("instruction[1] = %#v, want put on battlefield", mode.Sequence[1].Primitive)
	}
	key, linked := put.Source.LinkedKey()
	if !linked || key != exile.ExileLinkedKey {
		t.Fatalf("put source = %#v, want linked source %q", put.Source, exile.ExileLinkedKey)
	}
}

// TestLowerMassGroupBlink confirms the untargeted mass blink (Ghostway) lowers
// as one group exile of every controlled permanent under a single linked key
// plus one delayed group return, with no target spec.
func TestLowerMassGroupBlink(t *testing.T) {
	t.Parallel()
	mode := groupBlinkMode(t,
		"Exile each creature you control. Return those cards to the battlefield under their owner's control at the beginning of the next end step.")
	if len(mode.Targets) != 0 {
		t.Fatalf("targets = %#v, want none for mass blink", mode.Targets)
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want one exile and one delayed trigger", mode.Sequence)
	}
	exile, ok := mode.Sequence[0].Primitive.(game.Exile)
	if !ok || !exile.Group.Valid() || exile.ExileLinkedKey == "" {
		t.Fatalf("exile = %#v, want linked group exile", mode.Sequence[0].Primitive)
	}
	delayed, ok := mode.Sequence[1].Primitive.(game.CreateDelayedTrigger)
	if !ok || delayed.Trigger.Timing != game.DelayedAtBeginningOfNextEndStep {
		t.Fatalf("instruction[1] = %#v, want next-end-step delayed trigger", mode.Sequence[1].Primitive)
	}
	put, ok := delayed.Trigger.Content.Modes[0].Sequence[0].Primitive.(game.PutOnBattlefield)
	if !ok {
		t.Fatalf("delayed content = %#v, want one put on battlefield", delayed.Trigger.Content.Modes[0].Sequence)
	}
	key, linked := put.Source.LinkedKey()
	if !linked || key != exile.ExileLinkedKey {
		t.Fatalf("put source = %#v, want linked source %q", put.Source, exile.ExileLinkedKey)
	}
}
