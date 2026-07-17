package cardgen

import (
	"strings"
	"testing"
)

func TestGenerateCultivatorColossusComposableMechanics(t *testing.T) {
	t.Parallel()
	star := "*"
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:      "Cultivator Colossus",
		Layout:    "normal",
		ManaCost:  "{4}{G}{G}{G}",
		TypeLine:  "Creature — Plant Beast",
		Power:     &star,
		Toughness: &star,
		OracleText: "Trample\n" +
			"Cultivator Colossus's power and toughness are each equal to the number of lands you control.\n" +
			"When this creature enters, you may put a land card from your hand onto the battlefield tapped. If you do, draw a card and repeat this process.",
	}, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.DynamicValueControllerLandCount",
		"game.Trample",
		"game.RepeatProcess{",
		"ContinueResult: game.ResultKey(\"repeat-process-continue\")",
		"Primitive: game.ChooseFromZone{",
		"RequiredTypes: []types.Card{types.Land}",
		"Optional:      true",
		"PublishResult: game.ResultKey(\"if-you-do\")",
		"Primitive: game.Draw{",
		"Succeeded: game.TriTrue",
		"PublishResult: game.ResultKey(\"repeat-process-continue\")",
	} {
		if !strings.Contains(source, want) {
			t.Errorf("generated source missing %q:\n%s", want, source)
		}
	}
	if strings.Contains(source, "Cultivator Colossus") &&
		!strings.Contains(source, "Name: \"Cultivator Colossus\"") {
		t.Fatal("card-name text leaked into generated mechanics")
	}
}
