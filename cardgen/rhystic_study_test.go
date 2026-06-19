package cardgen

import (
	"strings"
	"testing"
)

func TestGenerateExecutableCardSourceRhysticStudy(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Rhystic Study",
		Layout:     "normal",
		ManaCost:   "{2}{U}",
		TypeLine:   "Enchantment",
		OracleText: "Whenever an opponent casts a spell, you may draw a card unless that player pays {1}.",
	}, "r")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.TriggerControllerOpponent",
		"game.Pay",
		"game.EventPlayerReference()",
		"cost.O(1)",
		"PublishResult: game.ResultKey(\"unless-paid\")",
		"game.Draw",
		"Optional: true",
		"Succeeded: game.TriFalse",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
