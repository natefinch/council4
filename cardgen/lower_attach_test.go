package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerMithrilCoatAttachTrigger verifies the enters-the-battlefield
// auto-attach trigger "When ~ enters, attach it to target legendary creature you
// control." lowers to an Attach primitive attaching the entering permanent (the
// triggering event permanent) to the single chosen target.
func TestLowerMithrilCoatAttachTrigger(t *testing.T) {
	card := &ScryfallCard{
		Name:       "Mithril Coat",
		Layout:     "normal",
		TypeLine:   "Artifact — Equipment",
		OracleText: "Flash\nIndestructible\nWhen Mithril Coat enters, attach it to target legendary creature you control.\nEquipped creature has indestructible.\nEquip {3}",
	}
	face := lowerSingleFace(t, card)
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	content := face.TriggeredAbilities[0].Content
	if len(content.Modes) != 1 || len(content.Modes[0].Targets) != 1 {
		t.Fatalf("content modes/targets = %+v, want one mode with one target", content.Modes)
	}
	sequence := content.Modes[0].Sequence
	if len(sequence) != 1 {
		t.Fatalf("sequence length = %d, want 1", len(sequence))
	}
	attach, ok := sequence[0].Primitive.(game.Attach)
	if !ok {
		t.Fatalf("primitive type = %T, want game.Attach", sequence[0].Primitive)
	}
	if attach.Attachment.Kind() != game.ObjectReferenceEventPermanent {
		t.Fatalf("attachment reference = %v, want event permanent", attach.Attachment.Kind())
	}
	if attach.Target.Kind() != game.ObjectReferenceTargetPermanent || attach.Target.TargetIndex() != 0 {
		t.Fatalf("target reference = %v idx %d, want target permanent 0", attach.Target.Kind(), attach.Target.TargetIndex())
	}
}

// TestLowerOptionalAttachTrigger verifies the optional "you may attach it"
// variant lowers to a single optional Attach instruction.
func TestLowerOptionalAttachTrigger(t *testing.T) {
	card := &ScryfallCard{
		Name:       "Living Cloak",
		Layout:     "normal",
		TypeLine:   "Artifact — Equipment",
		OracleText: "Equip {2}\nWhen Living Cloak enters, you may attach it to target creature you control.",
	}
	face := lowerSingleFace(t, card)
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	trigger := face.TriggeredAbilities[0]
	sequence := trigger.Content.Modes[0].Sequence
	if len(sequence) != 1 {
		t.Fatalf("sequence length = %d, want 1", len(sequence))
	}
	if _, ok := sequence[0].Primitive.(game.Attach); !ok {
		t.Fatalf("primitive type = %T, want game.Attach", sequence[0].Primitive)
	}
	if !trigger.Optional && !sequence[0].Optional {
		t.Fatal("neither trigger nor instruction Optional set, want one true for \"you may attach\"")
	}
}

// TestLowerHammerOfNazahnAttachTrigger verifies the self-or-another-Equipment
// enters trigger "Whenever ~ or another Equipment you control enters, you may
// attach that Equipment to target creature you control." lowers to an optional
// Attach that attaches the triggering event permanent (the entering Equipment,
// which may be a different Equipment than the source) to the chosen target.
func TestLowerHammerOfNazahnAttachTrigger(t *testing.T) {
	card := &ScryfallCard{
		Name:       "Hammer of Nazahn",
		Layout:     "normal",
		TypeLine:   "Artifact — Equipment",
		ManaCost:   "{4}",
		OracleText: "Whenever Hammer of Nazahn or another Equipment you control enters, you may attach that Equipment to target creature you control.\nEquipped creature gets +2/+0 and has indestructible.\nEquip {4}",
	}
	face := lowerSingleFace(t, card)
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	trigger := face.TriggeredAbilities[0]
	if !trigger.Optional {
		t.Fatal("trigger Optional = false, want true for \"you may attach\"")
	}
	if trigger.Trigger.Pattern.Event != game.EventPermanentEnteredBattlefield {
		t.Fatalf("trigger event = %v, want permanent entered battlefield", trigger.Trigger.Pattern.Event)
	}
	if !trigger.Trigger.Pattern.SubjectSelectionOrSelf {
		t.Fatal("trigger SubjectSelectionOrSelf = false, want true for \"~ or another Equipment\"")
	}
	sequence := trigger.Content.Modes[0].Sequence
	if len(sequence) != 1 {
		t.Fatalf("sequence length = %d, want 1", len(sequence))
	}
	attach, ok := sequence[0].Primitive.(game.Attach)
	if !ok {
		t.Fatalf("primitive type = %T, want game.Attach", sequence[0].Primitive)
	}
	if attach.Attachment.Kind() != game.ObjectReferenceEventPermanent {
		t.Fatalf("attachment reference = %v, want event permanent", attach.Attachment.Kind())
	}
	if attach.Target.Kind() != game.ObjectReferenceTargetPermanent || attach.Target.TargetIndex() != 0 {
		t.Fatalf("target reference = %v idx %d, want target permanent 0", attach.Target.Kind(), attach.Target.TargetIndex())
	}
}

// TestLowerAttachToPluralTargetsFailsClosed verifies that attaching to more than
// one target — which the single-attachment Attach primitive does not model —
// fails closed rather than lowering a silently-wrong attachment.
func TestLowerAttachToPluralTargetsFailsClosed(t *testing.T) {
	card := &ScryfallCard{
		Name:       "Bad Attach",
		Layout:     "normal",
		TypeLine:   "Artifact — Equipment",
		OracleText: "When Bad Attach enters, attach it to up to two target creatures you control.\nEquip {2}",
	}
	face := lowerSingleFaceExpectingUnsupported(t, card)
	if len(face.TriggeredAbilities) != 0 {
		t.Fatalf("triggered abilities = %d, want 0 (fail closed)", len(face.TriggeredAbilities))
	}
}
