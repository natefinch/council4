package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

// TestLowerMetallicMimicChosenTypeGroupEntersWithCounters proves Metallic
// Mimic's "Each other creature you control of the chosen type enters with an
// additional +1/+1 counter on it." lowers to a group enters-with-counters
// replacement whose recipient is restricted to the creature subtype the source
// chose as it entered (SubtypeChoiceSourceEntry), the chosen-type sibling of the
// literal-subtype "Each other Elf you control enters with..." group form.
func TestLowerMetallicMimicChosenTypeGroupEntersWithCounters(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:      "Metallic Mimic",
		Layout:    "normal",
		TypeLine:  "Artifact Creature — Shapeshifter",
		ManaCost:  "{2}",
		Power:     new("2"),
		Toughness: new("1"),
		OracleText: "As this creature enters, choose a creature type.\n" +
			"This creature is the chosen type in addition to its other types.\n" +
			"Each other creature you control of the chosen type enters with an additional +1/+1 counter on it.",
	})
	if len(face.ReplacementAbilities) != 2 {
		t.Fatalf("got %d replacement abilities, want 2", len(face.ReplacementAbilities))
	}
	var group *game.ReplacementEffect
	for i := range face.ReplacementAbilities {
		if face.ReplacementAbilities[i].Replacement.EntersWithCountersOthers {
			group = &face.ReplacementAbilities[i].Replacement
		}
	}
	if group == nil {
		t.Fatalf("no group enters-with-counters replacement lowered: %#v", face.ReplacementAbilities)
	}
	if group.EntersWithCountersRecipient == nil {
		t.Fatal("group enters-with-counters has no recipient selection")
	}
	if group.EntersWithCountersRecipient.SubtypeChoice != game.SubtypeChoiceSourceEntry {
		t.Fatalf("recipient subtype choice = %v, want SubtypeChoiceSourceEntry", group.EntersWithCountersRecipient.SubtypeChoice)
	}
	if len(group.EntersWithCounters) != 1 || group.EntersWithCounters[0].Kind != counter.PlusOnePlusOne {
		t.Fatalf("counter placement = %#v", group.EntersWithCounters)
	}
}

// TestGenerateMetallicMimicChosenTypeGroupSource proves the full card generates
// without diagnostics, threading the entry-type choice, the chosen-type subtype
// addition, and the chosen-type group enters-with-counters replacement.
func TestGenerateMetallicMimicChosenTypeGroupSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:      "Metallic Mimic",
		Layout:    "normal",
		TypeLine:  "Artifact Creature — Shapeshifter",
		ManaCost:  "{2}",
		Power:     new("2"),
		Toughness: new("1"),
		OracleText: "As this creature enters, choose a creature type.\n" +
			"This creature is the chosen type in addition to its other types.\n" +
			"Each other creature you control of the chosen type enters with an additional +1/+1 counter on it.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.EntryTypeChoiceReplacement(",
		"AddSubtypeFromEntryChoice: game.EntryTypeChoiceKey",
		"game.EntersWithCountersGroupReplacement(",
		"SubtypeChoice: game.SubtypeChoiceSourceEntry",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}
