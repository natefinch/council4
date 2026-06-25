package cardgen

import (
	"strings"
	"testing"
)

const fieldOfRuinOracle = "{T}: Add {C}.\n{2}, {T}, Sacrifice this land: Destroy target nonbasic land an opponent controls. Each player searches their library for a basic land card, puts it onto the battlefield, then shuffles."

func TestGenerateExecutableCardSourceFieldOfRuin(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Field of Ruin",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: fieldOfRuinOracle,
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "r")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Primitive: game.Destroy{",
		"Object: game.TargetPermanentReference(0),",
		"Primitive: game.Search{",
		"PlayerGroup: game.AllPlayersReference(),",
		"Filter:      game.Selection{RequiredTypes: []types.Card{types.Land}, Supertypes: []types.Super{types.Basic}},",
		"Destination: zone.Battlefield,",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
