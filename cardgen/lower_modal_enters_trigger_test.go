package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

// TestLowerModalEntersTriggerSharedSourceReference proves an enters-the-
// battlefield "choose one —" trigger lowers even though its trigger subject
// ("When this creature enters") records a content-level source reference shared
// across the modes. Each mode lowers independently, so the redundant source
// reference must not block the modal, and a mode that itself references the
// source ("Put a +1/+1 counter on this creature.") resolves to the source
// permanent.
func TestLowerModalEntersTriggerSharedSourceReference(t *testing.T) {
	t.Parallel()

	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Test Trusty Retriever",
		Layout:   "normal",
		TypeLine: "Creature — Dog",
		OracleText: "When this creature enters, choose one —\n" +
			"• Put a +1/+1 counter on this creature.\n" +
			"• Return target artifact or enchantment card from your graveyard to your hand.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %#v, want one", face.TriggeredAbilities)
	}
	ability := face.TriggeredAbilities[0]
	if !ability.Content.IsModal() {
		t.Fatalf("content = %#v, want modal", ability.Content)
	}
	if ability.Content.MinModes != 1 || ability.Content.MaxModes != 1 ||
		len(ability.Content.Modes) != 2 {
		t.Fatalf("modal range = %d..%d over %d modes, want 1..1 over two modes",
			ability.Content.MinModes, ability.Content.MaxModes, len(ability.Content.Modes))
	}
	first := ability.Content.Modes[0]
	if len(first.Sequence) != 1 {
		t.Fatalf("mode 1 sequence = %#v, want a single counter placement", first.Sequence)
	}
	add, ok := first.Sequence[0].Primitive.(game.AddCounter)
	if !ok {
		t.Fatalf("mode 1 primitive = %#v, want AddCounter", first.Sequence[0].Primitive)
	}
	if add.Object != game.SourcePermanentReference() || add.CounterKind != counter.PlusOnePlusOne {
		t.Fatalf("mode 1 counter = %#v, want +1/+1 on the source permanent", add)
	}
}

// TestLowerModalEntersTriggerUnusedSourceReference proves the shared source
// reference is tolerated even when no mode references it: every mode of this
// enters trigger acts on its own target or the controller, yet the trigger
// subject still records a content-level source reference.
func TestLowerModalEntersTriggerUnusedSourceReference(t *testing.T) {
	t.Parallel()

	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Test Cleanup Crew",
		Layout:   "normal",
		TypeLine: "Creature — Human Citizen",
		OracleText: "When this creature enters, choose one —\n" +
			"• Destroy target artifact.\n" +
			"• Destroy target enchantment.\n" +
			"• Exile target card from a graveyard.\n" +
			"• You gain 4 life.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %#v, want one", face.TriggeredAbilities)
	}
	content := face.TriggeredAbilities[0].Content
	if !content.IsModal() {
		t.Fatalf("content = %#v, want modal", content)
	}
	if content.MinModes != 1 || content.MaxModes != 1 || len(content.Modes) != 4 {
		t.Fatalf("modal range = %d..%d over %d modes, want 1..1 over four modes",
			content.MinModes, content.MaxModes, len(content.Modes))
	}
}
