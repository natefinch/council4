package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableDruidOfPurificationEachPlayerChooseDestroy exercises the
// "Starting with you, each player may choose <permanent>. Destroy each permanent
// chosen this way." construct (Druid of Purification): the two-sentence structure
// folds to a single EachPlayerChooseDestroy primitive over the shared candidate
// pool, evaluated relative to the controller so "you don't control" offers every
// chooser the same permanents.
func TestGenerateExecutableDruidOfPurificationEachPlayerChooseDestroy(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Druid of Purification",
		Layout:   "normal",
		ManaCost: "{3}{G}",
		TypeLine: "Creature — Human Druid",
		OracleText: "When this creature enters, starting with you, each player may choose an artifact or enchantment you don't control. " +
			"Destroy each permanent chosen this way.",
		Colors: []string{"G"},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "d")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.EachPlayerChooseDestroy{",
		"RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment}",
		"Controller: game.ControllerNotYou",
		"Optional:  true,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if strings.Contains(source, "PreventRegeneration") {
		t.Fatalf("source should not carry a regeneration rider:\n%s", source)
	}
}
