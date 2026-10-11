package cardgen

import (
	"strings"
	"testing"
)

func TestConditionalDividedDamageRequiresAnnouncementCapture(t *testing.T) {
	for _, text := range []string{
		"Kicker {5}{R}\nDamage Probe deals 5 damage to target creature. If this spell was kicked, it deals 10 damage divided as you choose among any number of targets instead.",
		"Damage Probe deals 5 damage to any target.\nMorbid — Damage Probe deals 10 damage divided as you choose among any number of targets instead if a creature died this turn.",
		"If you have 5 or less life, Damage Probe deals 5 damage divided as you choose among any number of targets. Draw a card.",
		"If you have 5 or less life, draw a card. Otherwise, Damage Probe deals 5 damage divided as you choose among any number of targets.",
	} {
		t.Run(text, func(t *testing.T) {
			card := &ScryfallCard{Name: "Damage Probe", Layout: "normal", TypeLine: "Instant", OracleText: text}
			_, diagnostics, err := GenerateExecutableCardSource(card, "t")
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, diagnostic := range diagnostics {
				found = found || strings.Contains(diagnostic.Detail, "announced allocation receipt")
			}
			if !found {
				t.Fatalf("missing precise allocation capture diagnostic: %+v", diagnostics)
			}
			assertCardUnsupported(t, card)
		})
	}
}

func TestDamageTargetProjectionKeepsModeledNeighbors(t *testing.T) {
	for _, text := range []string{
		"Damage Probe deals 5 damage divided as you choose among any number of targets.",
		"Kicker {5}{R}\nDamage Probe deals 5 damage to target creature.\nIf this spell was kicked, Damage Probe deals 10 damage instead.",
	} {
		t.Run(text, func(t *testing.T) {
			assertCardPaths(t, &ScryfallCard{Name: "Damage Probe", Layout: "normal", TypeLine: "Instant", OracleText: text},
				"SpellAbility.Val.Modes[0].Targets[0].Selection.Exists = true")
		})
	}
}
