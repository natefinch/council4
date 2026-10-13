package cardgen

import "testing"

func TestPermanentUnionSharedTailWholeCardRefuses(t *testing.T) {
	for _, tail := range []string{
		"with flying", "without flying", "with power 3 or less", "with toughness 3 or less",
		"with lesser power", "with greater power", "with power 3 and toughness 3",
	} {
		for _, text := range []string{
			"Tap target creature or Vehicle " + tail + ".",
			"Return target Vehicle or creature " + tail + " to its owner's hand. Draw a card.",
		} {
			t.Run(text, func(t *testing.T) {
				assertCardUnsupported(t, &ScryfallCard{
					Name: "Union Tail Refusal", Layout: "normal", TypeLine: "Instant", OracleText: text,
				})
			})
		}
	}
}

func TestPermanentUnionExplicitTailWholeCardControls(t *testing.T) {
	for _, phrase := range []string{
		"creature with flying or Vehicle with flying",
		"creature without flying or Vehicle without flying",
		"creature with power 3 or less or Vehicle with power 3 or less",
		"creature with toughness 3 or less or Vehicle with toughness 3 or less",
		"creature with flying or Vehicle",
	} {
		t.Run(phrase, func(t *testing.T) {
			assertCardPaths(t, &ScryfallCard{
				Name: "Union Tail Control", Layout: "normal", TypeLine: "Instant", OracleText: "Tap target " + phrase + ".",
			}, "Modes[0].Targets[0].Allow = game.TargetAllowPermanent")
		})
	}
}
