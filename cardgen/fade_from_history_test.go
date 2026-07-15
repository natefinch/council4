package cardgen

import (
	"strings"
	"testing"
)

func fadeFromHistoryCard() *ScryfallCard {
	return &ScryfallCard{
		Name:     "Fade from History",
		Layout:   "normal",
		ManaCost: "{2}{G}{G}",
		TypeLine: "Sorcery",
		OracleText: "Each player who controls an artifact or enchantment creates a 2/2 green Bear creature token. " +
			"Then destroy all artifacts and enchantments.",
	}
}

// TestGenerateExecutableCardSourceFadeFromHistory asserts the "Each player who
// controls an artifact or enchantment creates a 2/2 green Bear creature token.
// Then destroy all artifacts and enchantments." sequence lowers the conditional
// per-player token creation to a group CreateToken whose recipient group carries
// the ControllingMatching artifact-or-enchantment filter, ordered before the mass
// destroy so the created tokens survive.
func TestGenerateExecutableCardSourceFadeFromHistory(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(fadeFromHistoryCard(), "f")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.CreateToken{",
		"RecipientGroup: game.AllPlayersReference().ControllingMatching(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment}})",
		"game.Destroy{",
		"Group: game.BattlefieldGroup(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment}})",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
	// The token creation must appear before the mass destroy so the created Bear
	// tokens survive the destroy.
	createIndex := strings.Index(source, "game.CreateToken{")
	destroyIndex := strings.Index(source, "game.Destroy{")
	if createIndex < 0 || destroyIndex < 0 || createIndex > destroyIndex {
		t.Fatalf("CreateToken (index %d) must precede Destroy (index %d)", createIndex, destroyIndex)
	}
}
