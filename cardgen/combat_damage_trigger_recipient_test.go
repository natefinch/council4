package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourcePluralPlayersCombatDamageTrigger covers the
// "one or more players" combat-damage trigger recipient (Contaminant Grafter):
// it lowers to the same aggregated combat-damage-to-a-player trigger as the
// singular "a player" form.
func TestGenerateExecutableCardSourcePluralPlayersCombatDamageTrigger(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Toxic Grafter",
		Layout:     "normal",
		ManaCost:   "{3}{G}",
		TypeLine:   "Creature — Phyrexian Druid",
		OracleText: "Whenever one or more creatures you control deal combat damage to one or more players, proliferate.",
		Power:      new("2"),
		Toughness:  new("3"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.EventDamageDealt,",
		"OneOrMore:",
		"RequireCombatDamage:",
		"game.DamageRecipientPlayer,",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}

// TestGenerateExecutableCardSourceCombatDamageDelayedTrigger covers a
// combat-damage "this turn" delayed trigger created by an activated ability
// (Flitterwing Nuisance): the "Whenever ... deal combat damage ... this turn,
// <body>" preamble becomes a CreateDelayedTrigger with a this-turn window rather
// than a spurious resolving combat-damage effect.
func TestGenerateExecutableCardSourceCombatDamageDelayedTrigger(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Faerie Nuisance",
		Layout:     "normal",
		ManaCost:   "{1}{U}",
		TypeLine:   "Creature — Faerie Rogue",
		OracleText: "{2}{U}, {T}: Whenever a creature you control deals combat damage to a player this turn, draw a card.",
		Power:      new("1"),
		Toughness:  new("1"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.CreateDelayedTrigger{",
		"RequireCombatDamage:",
		"game.DelayedWindowThisTurn,",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}
