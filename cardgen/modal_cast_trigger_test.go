package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerModalCastTriggerUpToOne proves a "Whenever you cast a spell, choose
// up to one —" triggered ability lowers to a modal AbilityContent with one mode
// per alternative and an optional (zero-to-one) choice range.
func TestLowerModalCastTriggerUpToOne(t *testing.T) {
	t.Parallel()

	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Hullbreaker Horror",
		Layout:   "normal",
		TypeLine: "Legendary Creature — Elemental",
		OracleText: "Flash\nThis spell can't be countered.\n" +
			"Whenever you cast a spell, choose up to one —\n" +
			"• Return target spell you don't control to its owner's hand.\n" +
			"• Return target nonland permanent to its owner's hand.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %#v, want one", face.TriggeredAbilities)
	}
	ability := face.TriggeredAbilities[0]
	if ability.Trigger.Pattern.Event != game.EventSpellCast {
		t.Fatalf("trigger event = %#v, want spell cast", ability.Trigger.Pattern)
	}
	if !ability.Content.IsModal() {
		t.Fatalf("content = %#v, want modal", ability.Content)
	}
	if ability.Content.MinModes != 0 || ability.Content.MaxModes != 1 ||
		len(ability.Content.Modes) != 2 {
		t.Fatalf("modal range = %d..%d over %d modes, want optional 0..1 over two modes",
			ability.Content.MinModes, ability.Content.MaxModes, len(ability.Content.Modes))
	}
	for i := range ability.Content.Modes {
		if len(ability.Content.Modes[i].Sequence) != 1 {
			t.Fatalf("mode %d sequence = %#v, want a single bounce", i, ability.Content.Modes[i].Sequence)
		}
		if _, ok := ability.Content.Modes[i].Sequence[0].Primitive.(game.Bounce); !ok {
			t.Fatalf("mode %d primitive = %#v, want bounce", i, ability.Content.Modes[i].Sequence[0].Primitive)
		}
	}
}

// TestLowerModalCastTriggerChooseOne proves the mandatory "choose one —" form of
// a cast-triggered modal ability lowers with a 1..1 choice range.
func TestLowerModalCastTriggerChooseOne(t *testing.T) {
	t.Parallel()

	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Test Zephyr",
		Layout:   "normal",
		TypeLine: "Legendary Creature — Bird",
		OracleText: "Whenever you cast a noncreature spell, choose one —\n" +
			"• Return target nonland permanent to its owner's hand.\n" +
			"• Create a 1/1 white Spirit creature token with flying.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %#v, want one", face.TriggeredAbilities)
	}
	ability := face.TriggeredAbilities[0]
	if ability.Content.MinModes != 1 || ability.Content.MaxModes != 1 ||
		len(ability.Content.Modes) != 2 {
		t.Fatalf("modal range = %d..%d over %d modes, want mandatory 1..1 over two modes",
			ability.Content.MinModes, ability.Content.MaxModes, len(ability.Content.Modes))
	}
}
