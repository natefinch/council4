package cardgen

import (
	goparser "go/parser"
	"go/token"
	"testing"
)

func generateBecomeColorSource(t *testing.T, card *ScryfallCard, letter string) string {
	t.Helper()
	source, diagnostics, err := GenerateExecutableCardSource(card, letter)
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	if _, err := goparser.ParseFile(token.NewFileSet(), "card.go", source, goparser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, source)
	}
	return source
}

// TestLowerBecomeColorSelfColorless covers the source form: an activated ability
// that makes this creature colorless until end of turn (Blazing Blade Askari).
// It targets the source permanent and clears its colors at the color layer.
func TestLowerBecomeColorSelfColorless(t *testing.T) {
	source := generateBecomeColorSource(t, &ScryfallCard{
		Name:       "Blazing Blade Askari",
		Layout:     "normal",
		TypeLine:   "Creature — Human Knight",
		ManaCost:   "{2}{R}",
		OracleText: "{2}: This creature becomes colorless until end of turn.",
	}, "b")
	assertSourceContains(t, source,
		"game.ApplyContinuous{",
		"Object: opt.Val(game.SourcePermanentReference()),",
		"Layer:        game.LayerColor,",
		"SetColorless: true,",
		"Duration: game.DurationUntilEndOfTurn,",
	)
}

// TestLowerBecomeColorTargetNamedColor covers the target form: a tap ability
// that sets a target permanent to a single named color until end of turn
// (Fylamarid). It targets permanent reference 0 and sets the color at the color
// layer.
func TestLowerBecomeColorTargetNamedColor(t *testing.T) {
	source := generateBecomeColorSource(t, &ScryfallCard{
		Name:       "Fylamarid",
		Layout:     "normal",
		TypeLine:   "Creature — Squid",
		ManaCost:   "{3}{U}",
		OracleText: "{T}: Target permanent becomes blue until end of turn.",
	}, "f")
	assertSourceContains(t, source,
		"game.ApplyContinuous{",
		"Object: opt.Val(game.TargetPermanentReference(0)),",
		"Layer:     game.LayerColor,",
		"SetColors: []color.Color{color.Blue},",
		"Duration: game.DurationUntilEndOfTurn,",
	)
}
