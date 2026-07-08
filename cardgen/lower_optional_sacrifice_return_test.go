package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

// TestLowerOptionalSacrificeReturnWithCounters verifies the anchor card
// Heart-Shaped Herb's activated ability: "You may sacrifice a creature. If you
// do, return that card to the battlefield under its owner's control with three
// +1/+1 counters on it and you become the monarch." The optional sacrifice
// publishes its success and the sacrificed permanent as a linked object; the
// return reads that linked card, puts it onto the battlefield with the +1/+1
// counters, and both the return and the become-monarch clause are gated on the
// sacrifice having happened.
func TestLowerOptionalSacrificeReturnWithCounters(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Test Heart-Shaped Herb",
		Layout:   "normal",
		TypeLine: "Artifact",
		OracleText: "{2}, {T}, Sacrifice this artifact: You may sacrifice a creature. " +
			"If you do, return that card to the battlefield under its owner's control " +
			"with three +1/+1 counters on it and you become the monarch.",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	if face.ActivatedAbilities[0].ActivationCondition.Exists {
		t.Fatalf("activation condition = %+v, want none (the \"if you do\" gate belongs in the body)",
			face.ActivatedAbilities[0].ActivationCondition.Val)
	}
	mode := face.ActivatedAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 3 {
		t.Fatalf("sequence = %d instructions, want 3 (sacrifice, return, become monarch)", len(mode.Sequence))
	}

	sacrifice, ok := mode.Sequence[0].Primitive.(game.SacrificePermanents)
	if !ok {
		t.Fatalf("first primitive = %T, want game.SacrificePermanents", mode.Sequence[0].Primitive)
	}
	if !mode.Sequence[0].Optional {
		t.Fatal("sacrifice instruction is not Optional")
	}
	if mode.Sequence[0].PublishResult != optionalIfYouDoResultKey {
		t.Fatalf("sacrifice PublishResult = %q, want %q", mode.Sequence[0].PublishResult, optionalIfYouDoResultKey)
	}
	if sacrifice.PublishLinked != sacrificedCreatureLinkKey {
		t.Fatalf("sacrifice PublishLinked = %q, want %q", sacrifice.PublishLinked, sacrificedCreatureLinkKey)
	}

	put, ok := mode.Sequence[1].Primitive.(game.PutOnBattlefield)
	if !ok {
		t.Fatalf("second primitive = %T, want game.PutOnBattlefield", mode.Sequence[1].Primitive)
	}
	if put.Source != game.LinkedBattlefieldSource(sacrificedCreatureLinkKey) {
		t.Fatalf("return Source = %+v, want linked sacrificed creature", put.Source)
	}
	wantCounters := []game.CounterPlacement{{Kind: counter.PlusOnePlusOne, Amount: 3}}
	if len(put.EntryCounters) != len(wantCounters) || put.EntryCounters[0] != wantCounters[0] {
		t.Fatalf("return EntryCounters = %+v, want %+v", put.EntryCounters, wantCounters)
	}

	if _, ok := mode.Sequence[2].Primitive.(game.BecomeMonarch); !ok {
		t.Fatalf("third primitive = %T, want game.BecomeMonarch", mode.Sequence[2].Primitive)
	}

	for i := 1; i <= 2; i++ {
		gate := mode.Sequence[i].ResultGate
		if !gate.Exists {
			t.Fatalf("gated instruction %d has no ResultGate", i)
		}
		if gate.Val.Key != optionalIfYouDoResultKey || gate.Val.Succeeded != game.TriTrue {
			t.Fatalf("gated instruction %d gate = %+v, want if-you-do succeeded", i, gate.Val)
		}
	}
}
