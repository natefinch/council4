package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceMidnightClock exercises the full reusable
// pipeline Midnight Clock relies on: a mana ability, an activated ability and an
// each-upkeep trigger that both place an hour counter on the source, and a
// self-sourced Nth-counter threshold trigger whose content shuffles the
// controller's hand and graveyard into their library, draws a fixed number of
// cards, then exiles the source permanent.
func TestGenerateExecutableCardSourceMidnightClock(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Midnight Clock",
		Layout:   "normal",
		ManaCost: "{2}{U}",
		TypeLine: "Artifact",
		OracleText: "{T}: Add {U}.\n" +
			"{2}{U}: Put an hour counter on this artifact.\n" +
			"At the beginning of each upkeep, put an hour counter on this artifact.\n" +
			"When the twelfth hour counter is put on this artifact, shuffle your hand and graveyard into your library, then draw seven cards. Exile this artifact.",
	}, "m")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		// Tap-for-mana ability.
		"game.TapManaAbility(mana.U)",
		// Activated ability places an hour counter on the source.
		"game.AddCounter{",
		"CounterKind: counter.Hour",
		"Object:      game.SourcePermanentReference()",
		// Each-upkeep trigger (no Player field => fires on every player's upkeep).
		"Event: game.EventBeginningOfStep",
		"Step:  game.StepUpkeep",
		// Nth-counter threshold trigger, keyed to hour counters crossing twelve.
		"Type: game.TriggerWhen",
		"Event:            game.EventCountersAdded",
		"Source:           game.TriggerSourceSelf",
		"MatchCounterKind: true",
		"CounterKind:      counter.Hour",
		"CounterThreshold: 12",
		// Content: shuffle hand+graveyard into library, draw seven, exile source.
		"game.ShuffleGraveyardIntoLibrary{",
		"IncludeHand: true",
		"game.Draw{",
		"Amount: game.Fixed(7)",
		"game.Exile{",
		"Object: game.SourceCardPermanentReference()",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
	// The each-upkeep trigger must not be pinned to a single player.
	if strings.Contains(source, "Player: game.TriggerPlayerYou") {
		t.Fatalf("upkeep trigger was pinned to a single player:\n%s", source)
	}
}
