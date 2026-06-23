package cardgen

import (
	goparser "go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestLowerPolymorphSpell(t *testing.T) {
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Turn to Frog",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{1}{U}",
		OracleText: "Until end of turn, target creature loses all abilities and becomes a blue Frog with base power and toughness 1/1.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	if _, err := goparser.ParseFile(token.NewFileSet(), "turn_to_frog.go", source, goparser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, source)
	}
	for _, want := range []string{
		"game.ApplyContinuous{",
		"game.TargetPermanentReference(0)",
		"Layer:              game.LayerAbility,",
		"RemoveAllAbilities: true,",
		"game.LayerColor,",
		"SetColors: []color.Color{color.Blue},",
		"game.LayerType,",
		"SetTypes:    []types.Card{types.Creature},",
		"SetSubtypes: []types.Sub{types.Frog},",
		"game.LayerPowerToughnessSet,",
		"SetPower:     opt.Val(game.PT{Value: 1}),",
		"SetToughness: opt.Val(game.PT{Value: 1}),",
		"Duration: game.DurationUntilEndOfTurn,",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

// TestLowerPolymorphSpellInSequence covers the polymorph clause appearing as the
// first effect of an ordered sequence ("... becomes a green Snake ...\nDraw a
// card."), which lowers through the same single-effect entry point.
func TestLowerPolymorphSpellInSequence(t *testing.T) {
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Snakeform",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{1}{G}",
		OracleText: "Until end of turn, target creature loses all abilities and becomes a green Snake with base power and toughness 1/1.\nDraw a card.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	if _, err := goparser.ParseFile(token.NewFileSet(), "snakeform.go", source, goparser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, source)
	}
	for _, want := range []string{
		"RemoveAllAbilities: true,",
		"SetSubtypes: []types.Sub{types.Snake},",
		"game.Draw",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
