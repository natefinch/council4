package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceCoinOfFate covers Coin of Fate's activated
// ability: its cost exiles two creature cards from the graveyard and sacrifices
// the artifact, and its resolution "An opponent chooses one of the exiled cards.
// You put that card on the bottom of your library and return the other to the
// battlefield tapped. You become the monarch." lowers the opponent-choice
// disposal to a PartitionExiledCostCards primitive followed by BecomeMonarch.
// The zero-effect "An opponent chooses one of the exiled cards." antecedent is a
// credited rider, so the whole ability lowers without diagnostics.
func TestGenerateExecutableCardSourceCoinOfFate(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Coin of Fate",
		Layout:   "normal",
		ManaCost: "{1}{W}",
		TypeLine: "Artifact",
		OracleText: "When this artifact enters, surveil 1.\n" +
			"{3}{W}, {T}, Exile two creature cards from your graveyard, Sacrifice this artifact: An opponent chooses one of the exiled cards. You put that card on the bottom of your library and return the other to the battlefield tapped. You become the monarch.",
	}, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Kind:          cost.AdditionalExile,",
		"Amount:        2,",
		"CardType:      types.Creature,",
		"Kind:   cost.AdditionalSacrificeSource,",
		"Primitive: game.PartitionExiledCostCards{",
		"ChooserOpponent:       true,",
		"ChosenToLibraryBottom: true,",
		"OtherEntersTapped:     true,",
		"Primitive: game.BecomeMonarch{",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}

// TestGenerateExecutableCardSourceControllerChoiceExiledSplitFailsClosed confirms
// the opponent-choice disposal is text-blind: a "You choose one of the exiled
// cards." variant is not recognized, so the resolution does not lower to a
// PartitionExiledCostCards primitive and the card is reported unsupported.
func TestGenerateExecutableCardSourceControllerChoiceExiledSplitFailsClosed(t *testing.T) {
	t.Parallel()
	source, diagnostics, _ := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Controller Coin",
		Layout:     "normal",
		ManaCost:   "{1}{W}",
		TypeLine:   "Artifact",
		OracleText: "{3}{W}, {T}, Exile two creature cards from your graveyard, Sacrifice this artifact: You choose one of the exiled cards. You put that card on the bottom of your library and return the other to the battlefield tapped. You become the monarch.",
	}, "t")
	if len(diagnostics) == 0 {
		t.Fatalf("expected diagnostics for unrecognized controller-choice disposal, got source:\n%s", source)
	}
	if strings.Contains(source, "game.PartitionExiledCostCards{") {
		t.Fatalf("controller-choice disposal unexpectedly lowered to PartitionExiledCostCards:\n%s", source)
	}
}
