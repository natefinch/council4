package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerChosenColorSpellCastTrigger verifies the "Whenever you cast a spell
// of the chosen color" trigger (Prism Ring) lowers its spell-cast pattern to a
// CardSelection bound to the source's entry-time color choice, alongside the
// "As this artifact enters, choose a color." entry replacement.
func TestLowerChosenColorSpellCastTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Prism Ring",
		Layout:   "normal",
		ManaCost: "{2}",
		TypeLine: "Artifact",
		OracleText: "As this artifact enters, choose a color.\n" +
			"Whenever you cast a spell of the chosen color, you gain 1 life.",
	})

	if len(face.ReplacementAbilities) != 1 || !face.ReplacementAbilities[0].Replacement.EntryColorChoice {
		t.Fatalf("replacement abilities = %#v, want one entry-time color choice", face.ReplacementAbilities)
	}

	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want one chosen-color spell-cast trigger", len(face.TriggeredAbilities))
	}
	pattern := face.TriggeredAbilities[0].Trigger.Pattern
	if pattern.Event != game.EventSpellCast ||
		pattern.Controller != game.TriggerControllerYou ||
		pattern.CardSelection.ColorChoice != game.ColorChoiceSourceEntry {
		t.Fatalf("trigger pattern = %#v, want chosen-color spell-cast", pattern)
	}
}
