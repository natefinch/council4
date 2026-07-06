package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerCloakTopCardEffect verifies the exact cloak spell "Cloak the top
// card of your library." lowers to a single Manifest primitive carrying the
// Cloak flag, reusing the manifest machinery to put the top card onto the
// battlefield face down as a 2/2 with ward {2}.
func TestLowerCloakTopCardEffect(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Cloaking Ritual",
		Layout:     "normal",
		ManaCost:   "{2}{U}",
		TypeLine:   "Sorcery",
		OracleText: "Cloak the top card of your library. (To cloak a card, put it onto the battlefield face down as a 2/2 creature with ward {2}. Turn it face up any time for its mana cost if it's a creature card.)",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %#v, want single cloak", mode.Sequence)
	}
	manifest, ok := mode.Sequence[0].Primitive.(game.Manifest)
	if !ok || !manifest.Cloak {
		t.Fatalf("primitive = %#v, want game.Manifest{Cloak: true}", mode.Sequence[0].Primitive)
	}
}

// TestLowerCloakThenAttachSequence verifies the ordered pair "cloak the top
// card of your library, then attach this Equipment to it." (Cryptic Coat)
// lowers to a cloak that publishes its result under a link key followed by an
// attach fastening the source Equipment onto that linked cloaked permanent.
func TestLowerCloakThenAttachSequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Cryptic Coat",
		Layout:     "normal",
		ManaCost:   "{1}{U}",
		TypeLine:   "Artifact — Equipment",
		OracleText: "When this Equipment enters, cloak the top card of your library, then attach this Equipment to it. (To cloak a card, put it onto the battlefield face down as a 2/2 creature with ward {2}. Turn it face up any time for its mana cost if it's a creature card.)\nEquipped creature gets +1/+0 and can't be blocked.\n{1}{U}: Return this Equipment to its owner's hand.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 0 {
		t.Fatalf("targets = %d, want 0", len(mode.Targets))
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want cloak then attach", mode.Sequence)
	}
	manifest, ok := mode.Sequence[0].Primitive.(game.Manifest)
	if !ok || !manifest.Cloak || manifest.PublishLinked == "" {
		t.Fatalf("cloak = %#v, want a cloak publishing a link", mode.Sequence[0].Primitive)
	}
	attach, ok := mode.Sequence[1].Primitive.(game.Attach)
	if !ok {
		t.Fatalf("attach = %#v, want game.Attach", mode.Sequence[1].Primitive)
	}
	if attach.Attachment.Kind() != game.ObjectReferenceSourcePermanent {
		t.Fatalf("attachment = %v, want source permanent", attach.Attachment.Kind())
	}
	if attach.Target.Kind() != game.ObjectReferenceLinkedObject ||
		attach.Target.LinkID() != string(manifest.PublishLinked) {
		t.Fatalf("target = %#v, want linked cloaked permanent", attach.Target)
	}
}
