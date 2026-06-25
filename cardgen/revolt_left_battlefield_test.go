package cardgen

import (
	"strings"
	"testing"
)

func narnamRenegadeCard() *ScryfallCard {
	power, toughness := "1", "2"
	return &ScryfallCard{
		Name:      "Narnam Renegade",
		Layout:    "normal",
		ManaCost:  "{G}",
		TypeLine:  "Creature — Elf Warrior",
		Power:     &power,
		Toughness: &toughness,
		OracleText: "Deathtouch\n" +
			"Revolt — This creature enters with a +1/+1 counter on it if a permanent left the battlefield under your control this turn.",
	}
}

// TestGenerateExecutableCardSourceNarnamRenegade asserts the Revolt
// event-history clause "a permanent left the battlefield under your control this
// turn" lowers to a current-turn zone-change intervening condition matching any
// permanent the controller owns leaving the battlefield, gating the
// enters-with-counters replacement.
func TestGenerateExecutableCardSourceNarnamRenegade(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(narnamRenegadeCard(), "n")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.EntersWithCountersIfReplacement(",
		"EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{",
		"Event:         game.EventZoneChanged,",
		"Controller:    game.TriggerControllerYou,",
		"MatchFromZone: true,",
		"FromZone:      zone.Battlefield,",
		"Window: game.EventHistoryCurrentTurn",
		"game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1}",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
