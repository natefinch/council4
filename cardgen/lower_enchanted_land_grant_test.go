package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceChamberOfManipulation confirms the Aura shape
// "Enchanted land has '<quoted activated ability>'." lowers into an ability-layer
// continuous effect that grants the quoted activated ability to the attached
// object. The granted ability carries the {T}+discard activation cost and the
// gain-control-of-target-creature-until-end-of-turn effect. The attached object
// is the permanent the Aura enchants regardless of its card type, so an Aura
// enchanting a land names the same closed attached-object group an Aura
// enchanting a creature does.
func TestGenerateExecutableCardSourceChamberOfManipulation(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Chamber of Manipulation",
		Layout:   "normal",
		ManaCost: "{2}{U}{U}",
		TypeLine: "Enchantment — Aura",
		Colors:   []string{"U"},
		OracleText: "Enchant land\n" +
			"Enchanted land has \"{T}, Discard a card: Gain control of target creature until end of turn.\"",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.EnchantStaticAbility(",
		"game.AttachedObjectGroup(game.SourcePermanentReference())",
		"AddAbilities: []game.Ability{",
		"Kind: cost.AdditionalTap,",
		"Kind:   cost.AdditionalDiscard,",
		"Layer:         game.LayerControl,",
		"Duration: game.DurationUntilEndOfTurn,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if strings.Contains(source, "TODO") {
		t.Fatalf("executable source contains TODO:\n%s", source)
	}
}

// TestGenerateExecutableCardSourceEnchantedLandGrant confirms a second
// "Enchanted land has '<quoted activated ability>'." card generates: an Aura
// enchanting a land grants a tap-to-tap-target-creature activated ability to the
// attached land. This exercises the attached-object land subject independent of
// the gain-control granted body.
func TestGenerateExecutableCardSourceEnchantedLandGrant(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Debtor's Pulpit",
		Layout:   "normal",
		ManaCost: "{3}{W}{W}",
		TypeLine: "Enchantment — Aura",
		Colors:   []string{"W"},
		OracleText: "Enchant land\n" +
			"Enchanted land has \"{T}: Tap target creature.\"",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "d")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.EnchantStaticAbility(",
		"game.AttachedObjectGroup(game.SourcePermanentReference())",
		"AddAbilities: []game.Ability{",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if strings.Contains(source, "TODO") {
		t.Fatalf("executable source contains TODO:\n%s", source)
	}
}
