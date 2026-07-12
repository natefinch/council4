package cardgen

import (
	"strings"
	"testing"
)

func TestGenerateCityOfDeathCopyNonSagaToken(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "City of Death",
		Layout:     "saga",
		ManaCost:   "{2}{G}",
		TypeLine:   "Enchantment — Saga",
		Colors:     []string{"G"},
		OracleText: "(As this Saga enters and after your draw step, add a lore counter. Sacrifice after VI.)\nI — Create a Treasure token.\nII, III, IV, V, VI — Create a token that's a copy of target non-Saga token you control.",
	}, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"TokenOnly: true",
		"ExcludedSubtype: types.Sub(\"Saga\")",
		"Source: game.TokenCopySourceObject,",
		"Object: game.TargetPermanentReference(0),",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}
