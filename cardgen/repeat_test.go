package cardgen

import (
	"strings"
	"testing"
)

func TestGenerateExecutableCardSourceTormentRepeatProcess(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Torment of Hailfire",
		Layout:     "normal",
		ManaCost:   "{X}{B}{B}{B}",
		TypeLine:   "Sorcery",
		OracleText: "Repeat the following process X times. Each opponent loses 3 life unless that player sacrifices a nonland permanent of their choice or discards a card.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.RepeatProcess{",
		"Times: game.Dynamic(game.DynamicAmount{",
		"Kind: game.DynamicAmountX",
		"Body: game.Mode{",
		"game.PunisherEachLoseLife{",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
