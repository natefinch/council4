package cardgen

import (
	"strings"
	"testing"
)

func bloodSpeakerCard() *ScryfallCard {
	power, toughness := "3", "2"
	return &ScryfallCard{
		Name:      "Blood Speaker",
		Layout:    "normal",
		ManaCost:  "{3}{B}",
		TypeLine:  "Creature — Ogre Shaman",
		Power:     &power,
		Toughness: &toughness,
		OracleText: "At the beginning of your upkeep, you may sacrifice this creature. If you do, search your library for a Demon card, reveal that card, put it into your hand, then shuffle.\n" +
			"Whenever a Demon you control enters, return this card from your graveyard to your hand.",
	}
}

// TestGenerateExecutableCardSourceBloodSpeaker asserts the optional "you may
// sacrifice this creature. If you do, search ..." upkeep trigger lowers the
// sacrifice as an optional instruction that publishes an if-you-do result and
// gates the following library search on that result.
func TestGenerateExecutableCardSourceBloodSpeaker(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(bloodSpeakerCard(), "b")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.Sacrifice{",
		"Optional:      true",
		"PublishResult: game.ResultKey(\"if-you-do\")",
		"game.Search{",
		"Filter:      game.Selection{SubtypesAny: []types.Sub{types.Sub(\"Demon\")}}",
		"ResultGate: opt.Val(game.InstructionResultGate{",
		"Key:       \"if-you-do\"",
		"Succeeded: game.TriTrue",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
