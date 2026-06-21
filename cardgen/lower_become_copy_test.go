package cardgen

import (
	goparser "go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestLowerBecomeCopyOfGraveyardCard(t *testing.T) {
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Woodland",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{2}: This land becomes a copy of target permanent card in your graveyard until end of turn.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	if _, err := goparser.ParseFile(token.NewFileSet(), "test_woodland.go", source, goparser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, source)
	}
	for _, want := range []string{
		"game.TargetAllowCard",
		"TargetZone:",
		"zone.Graveyard",
		"game.BecomeCopy{",
		"Card:",
		"game.CardReference{Kind: game.CardReferenceTarget}",
		"UntilEndOfTurn: true,",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
	if strings.Contains(source, "Object:") {
		t.Fatalf("graveyard-card copy must not set Object:\n%s", source)
	}
}
