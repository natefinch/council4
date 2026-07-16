package cardgen

import (
	"strings"
	"testing"
)

func TestGenerateExecutableSakashimaOfAThousandFaces(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Sakashima of a Thousand Faces",
		Layout:   "normal",
		TypeLine: "Legendary Creature — Human Rogue",
		ManaCost: "{3}{U}",
		OracleText: "You may have Sakashima enter as a copy of another creature you control, except it has Sakashima's other abilities.\n" +
			"The \"legend rule\" doesn't apply to permanents you control.\n" +
			"Partner (You can have two commanders if both have partner.)",
		Power:     new("3"),
		Toughness: new("1"),
	}, "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.EntersAsCopyWithRetainedName(",
		"game.EntersAsCopyWithOtherAbilities(",
		"game.LegendRuleDoesNotApplyStaticBody",
		"game.PartnerStaticBody",
		"Controller: game.ControllerYou",
	} {
		if !strings.Contains(source, want) {
			t.Errorf("generated source missing %q:\n%s", want, source)
		}
	}
}
