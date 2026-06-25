package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerOptionalNestedYouMay verifies the nested double-optional reflexive
// body "you may X. If you do, you may Y": X is optional and publishes its
// result, and the gated consequence Y is BOTH gated on that result having
// succeeded AND itself Optional, so the engine asks the controller whether to
// perform Y only when X happened. The runtime evaluates the result gate before
// the optional prompt, so a declined or failed X skips Y entirely.
func TestLowerOptionalNestedYouMay(t *testing.T) {
	t.Parallel()
	sequence := lowerSpellSequence(t, "Nested You May Test",
		"You may discard a card. If you do, you may draw a card.")
	if len(sequence) != 2 {
		t.Fatalf("sequence = %#v, want two instructions", sequence)
	}
	discard := sequence[0]
	if _, ok := discard.Primitive.(game.Discard); !ok {
		t.Fatalf("instruction[0] = %T, want game.Discard", discard.Primitive)
	}
	if !discard.Optional || discard.PublishResult != optionalIfYouDoResultKey {
		t.Fatalf("instruction[0] = %#v, want optional publishing %q", discard, optionalIfYouDoResultKey)
	}
	draw := sequence[1]
	if _, ok := draw.Primitive.(game.Draw); !ok {
		t.Fatalf("instruction[1] = %T, want game.Draw", draw.Primitive)
	}
	if !draw.Optional {
		t.Fatal("instruction[1].Optional = false, want the nested \"you may\" to be optional")
	}
	if gate := draw.ResultGate; !gate.Exists ||
		gate.Val.Key != optionalIfYouDoResultKey || gate.Val.Succeeded != game.TriTrue {
		t.Fatalf("instruction[1].ResultGate = %#v, want succeeded gate on %q", draw.ResultGate, optionalIfYouDoResultKey)
	}
}

// TestLowerOptionalMandatoryPublishNestedYouMay verifies the mandatory-publish
// reflexive body "X. If you do, you may Y" where X is a mandatory effect that
// can fail (a sacrifice): X publishes its result without being marked Optional,
// and the gated consequence Y is both gated on X having succeeded and itself
// Optional.
func TestLowerOptionalMandatoryPublishNestedYouMay(t *testing.T) {
	t.Parallel()
	sequence := lowerSpellSequence(t, "Mandatory Publish You May Test",
		"Sacrifice a creature. If you do, you may draw a card.")
	if len(sequence) != 2 {
		t.Fatalf("sequence = %#v, want two instructions", sequence)
	}
	sacrifice := sequence[0]
	if _, ok := sacrifice.Primitive.(game.SacrificePermanents); !ok {
		t.Fatalf("instruction[0] = %T, want game.SacrificePermanents", sacrifice.Primitive)
	}
	if sacrifice.Optional {
		t.Fatal("instruction[0].Optional = true, want the mandatory publisher to stay non-optional")
	}
	if sacrifice.PublishResult != optionalIfYouDoResultKey {
		t.Fatalf("instruction[0].PublishResult = %q, want %q", sacrifice.PublishResult, optionalIfYouDoResultKey)
	}
	draw := sequence[1]
	if _, ok := draw.Primitive.(game.Draw); !ok {
		t.Fatalf("instruction[1] = %T, want game.Draw", draw.Primitive)
	}
	if !draw.Optional {
		t.Fatal("instruction[1].Optional = false, want the nested \"you may\" to be optional")
	}
	if gate := draw.ResultGate; !gate.Exists ||
		gate.Val.Key != optionalIfYouDoResultKey || gate.Val.Succeeded != game.TriTrue {
		t.Fatalf("instruction[1].ResultGate = %#v, want succeeded gate on %q", draw.ResultGate, optionalIfYouDoResultKey)
	}
}

// TestLowerOptionalGatedSubjectContinuation verifies a gated tail that ends with
// a prior-subject continuation sentence ("... put a +1/+1 counter on target
// creature. It gains trample until end of turn."): the continuation does not
// itself repeat the "If you do" gate but back-references the subject introduced
// inside the gated tail, so it belongs to the same gated consequence and is
// gated on the optional effect having succeeded.
func TestLowerOptionalGatedSubjectContinuation(t *testing.T) {
	t.Parallel()
	sequence := lowerSpellSequence(t, "Gated Continuation Test",
		"You may discard a card. If you do, put a +1/+1 counter on target creature. It gains trample until end of turn.")
	if len(sequence) != 3 {
		t.Fatalf("sequence = %#v, want three instructions", sequence)
	}
	discard := sequence[0]
	if _, ok := discard.Primitive.(game.Discard); !ok {
		t.Fatalf("instruction[0] = %T, want game.Discard", discard.Primitive)
	}
	if !discard.Optional || discard.PublishResult != optionalIfYouDoResultKey {
		t.Fatalf("instruction[0] = %#v, want optional publishing %q", discard, optionalIfYouDoResultKey)
	}
	if _, ok := sequence[1].Primitive.(game.AddCounter); !ok {
		t.Fatalf("instruction[1] = %T, want game.AddCounter", sequence[1].Primitive)
	}
	if gate := sequence[1].ResultGate; !gate.Exists ||
		gate.Val.Key != optionalIfYouDoResultKey || gate.Val.Succeeded != game.TriTrue {
		t.Fatalf("instruction[1].ResultGate = %#v, want succeeded gate on %q", sequence[1].ResultGate, optionalIfYouDoResultKey)
	}
	if _, ok := sequence[2].Primitive.(game.ApplyContinuous); !ok {
		t.Fatalf("instruction[2] = %T, want game.ApplyContinuous", sequence[2].Primitive)
	}
	if gate := sequence[2].ResultGate; !gate.Exists ||
		gate.Val.Key != optionalIfYouDoResultKey || gate.Val.Succeeded != game.TriTrue {
		t.Fatalf("instruction[2].ResultGate = %#v, want the continuation gated on %q", sequence[2].ResultGate, optionalIfYouDoResultKey)
	}
}
