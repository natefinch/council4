package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableRapaciousGuest proves the full Rapacious Guest card
// lowers with no diagnostics and that its leaves-the-battlefield trigger emits a
// life-loss whose amount reads the departed permanent's power through the event
// permanent's last-known information, applied to the chosen player target. The
// amount binds to the event permanent rather than the player target, since a
// player has no power.
func TestGenerateExecutableRapaciousGuest(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:      "Rapacious Guest",
		Layout:    "normal",
		TypeLine:  "Creature — Halfling Citizen",
		ManaCost:  "{2}{B}",
		Power:     new("2"),
		Toughness: new("2"),
		OracleText: "Menace\n" +
			"Whenever one or more creatures you control deal combat damage to a player, create a Food token.\n" +
			"Whenever you sacrifice a Food, put a +1/+1 counter on this creature.\n" +
			"When this creature leaves the battlefield, target opponent loses life equal to its power.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.EventZoneChanged",
		"game.TriggerSourceSelf",
		"Primitive: game.LoseLife",
		"game.DynamicAmountObjectPower",
		"Object:     game.EventPermanentReference()",
		"Player: game.TargetPlayerReference(0)",
		"game.TargetAllowPlayer",
		"game.PlayerOpponent",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestGenerateExecutableLeaveLoseLifeSourcePowerIsReusable proves the leaving
// source-power life-loss support is name-independent: an arbitrarily named
// creature with the same departure clause lowers identically.
func TestGenerateExecutableLeaveLoseLifeSourcePowerIsReusable(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		Power:      new("2"),
		Toughness:  new("2"),
		OracleText: "When this creature leaves the battlefield, target opponent loses life equal to its power.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.LoseLife",
		"game.DynamicAmountObjectPower",
		"Object:     game.EventPermanentReference()",
		"Player: game.TargetPlayerReference(0)",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}
