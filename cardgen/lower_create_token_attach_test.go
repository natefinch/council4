package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerCreateTokenThenAttachSequence verifies the ordered pair "create a
// <token>, then attach this Equipment to it." lowers to a token creation that
// publishes its result under a link key, followed by an attach fastening the
// source Equipment onto that linked token. Barbed Spike's first clause is the
// blocked construct broadened here.
func TestLowerCreateTokenThenAttachSequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Barbed Spike",
		Layout:     "normal",
		TypeLine:   "Artifact — Equipment",
		OracleText: "When this Equipment enters, create a 1/1 colorless Thopter artifact creature token with flying, then attach this Equipment to it.\nEquipped creature gets +1/+0.\nEquip {2}",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 0 {
		t.Fatalf("targets = %d, want 0", len(mode.Targets))
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want create then attach", mode.Sequence)
	}
	create, ok := mode.Sequence[0].Primitive.(game.CreateToken)
	if !ok || create.PublishLinked == "" {
		t.Fatalf("create = %#v, want a token creation publishing a link", mode.Sequence[0].Primitive)
	}
	attach, ok := mode.Sequence[1].Primitive.(game.Attach)
	if !ok {
		t.Fatalf("attach = %#v, want game.Attach", mode.Sequence[1].Primitive)
	}
	if attach.Attachment.Kind() != game.ObjectReferenceSourcePermanent {
		t.Fatalf("attachment = %v, want source permanent", attach.Attachment.Kind())
	}
	if attach.Target.Kind() != game.ObjectReferenceLinkedObject ||
		attach.Target.LinkID() != string(create.PublishLinked) {
		t.Fatalf("target = %#v, want linked created token", attach.Target)
	}
}

// TestLowerCreateTokenThenAttachThisSequence verifies the bare "this" wording
// (Bonehoard's Living weapon reminder) lowers identically, attaching the source
// Equipment onto the created token.
func TestLowerCreateTokenThenAttachThisSequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Ancestral Blade",
		Layout:     "normal",
		TypeLine:   "Artifact — Equipment",
		OracleText: "When this Equipment enters, create a 1/1 white Soldier creature token, then attach this Equipment to it.\nEquipped creature gets +1/+1.\nEquip {1}",
	})
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want create then attach", mode.Sequence)
	}
	if _, ok := mode.Sequence[1].Primitive.(game.Attach); !ok {
		t.Fatalf("attach = %#v, want game.Attach", mode.Sequence[1].Primitive)
	}
}

// TestLowerCreateMultipleTokensThenAttachUnsupported confirms the lowering fails
// closed when more than one token is created, since the singular "it" target is
// then ambiguous.
func TestLowerCreateMultipleTokensThenAttachUnsupported(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Twin Spike",
		Layout:     "normal",
		TypeLine:   "Artifact — Equipment",
		OracleText: "When this Equipment enters, create two 1/1 colorless Thopter artifact creature tokens, then attach this Equipment to it.\nEquip {2}",
	})
}
