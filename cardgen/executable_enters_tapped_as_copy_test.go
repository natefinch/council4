package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceEntersTappedAsCopy covers the self
// enters-as-copy replacement that also enters the permanent tapped ("You may
// have this land enter tapped as a copy of any land on the battlefield." —
// Vesuva). The tapped form must wrap the base enters-as-copy constructor with
// game.EntersTappedAsCopy so existing untapped Clone-family cards stay byte
// identical.
func TestGenerateExecutableCardSourceEntersTappedAsCopy(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Vesuva",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "You may have this land enter tapped as a copy of any land on the battlefield.",
	}, "v")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.EntersTappedAsCopy(game.EntersAsCopyReplacement(",
		"RequiredTypes: []types.Card{types.Land}",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}
