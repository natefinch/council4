package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCemeteryProwler exercises the two reusable building
// blocks Cemetery Prowler combines (issue #1569): an enter-or-attack trigger
// that exiles a card from any graveyard under a source-keyed link, and a
// controller cast-cost static whose discount scales with the card types a spell
// shares with the cards exiled with the source. Both abilities must read the
// same exiled-with-source link key so the cost reduction sees the trigger's
// captives.
func TestGenerateExecutableCemeteryProwler(t *testing.T) {
	t.Parallel()
	power := "3"
	toughness := "4"
	card := &ScryfallCard{
		Name:     "Cemetery Prowler",
		Layout:   "normal",
		ManaCost: "{1}{G}{G}",
		TypeLine: "Creature — Wolf",
		OracleText: "Vigilance\n" +
			"Whenever this creature enters or attacks, exile a card from a graveyard.\n" +
			"Spells you cast cost {1} less to cast for each card type they share with cards exiled with this creature.",
		Colors:    []string{"G"},
		Power:     &power,
		Toughness: &toughness,
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.ExileFromGraveyard{",
		"AllOwners:     true,",
		`PublishLinked: game.LinkedKey("exiled-with-source"),`,
		"Kind:           game.RuleEffectCostModifier,",
		"AffectedPlayer: game.PlayerYou,",
		"SharedExiledCardTypeReduction: 1,",
		`ExiledLinkKey:                 game.LinkedKey("exiled-with-source"),`,
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}
