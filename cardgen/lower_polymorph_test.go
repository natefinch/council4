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

// TestLowerNamedBecomePolymorphSpell covers the permanent named-become polymorph
// ("Target nontoken creature becomes a 6/6 legendary Horror creature named
// Fenric and loses all abilities.", The Curse of Fenric II): the creature is
// renamed, made legendary, set to a fixed type/subtype and power/toughness, and
// loses all abilities permanently.
func TestLowerNamedBecomePolymorphSpell(t *testing.T) {
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Named Become Test",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{1}{U}",
		OracleText: "Target nontoken creature becomes a 6/6 legendary Horror creature named Fenric and loses all abilities.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}

	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	if _, err := goparser.ParseFile(token.NewFileSet(), "named_become.go", source, goparser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, source)
	}
	for _, want := range []string{
		"game.ApplyContinuous{",
		"RemoveAllAbilities: true,",
		"game.LayerType,",
		"AddSupertypes: []types.Super{types.Legendary},",
		"SetTypes:      []types.Card{types.Creature},",
		"SetSubtypes:   []types.Sub{types.Horror},",
		"game.LayerText,",
		"SetName: \"Fenric\",",
		"SetPower:     opt.Val(game.PT{Value: 6}),",
		"SetToughness: opt.Val(game.PT{Value: 6}),",
		"Duration: game.DurationPermanent,",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

func TestLowerCyberConversion(t *testing.T) {
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Cyber Conversion",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{U}",
		OracleText: "Turn target creature face down. It's a 2/2 Cyberman artifact creature.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	if _, err := goparser.ParseFile(token.NewFileSet(), "cyber_conversion.go", source, goparser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, source)
	}
	for _, want := range []string{
		"Primitive: game.TurnFaceDown{",
		"Object: game.TargetPermanentReference(0),",
		"Characteristics: opt.Val(game.FaceDownCharacteristics{",
		"Types:     []types.Card{types.Artifact, types.Creature},",
		"Subtypes:  []types.Sub{types.Cyberman},",
		"Power:     game.PT{Value: 2},",
		"Toughness: game.PT{Value: 2},",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
	if strings.Contains(source, "RemoveAllAbilities: true") ||
		strings.Contains(source, "SetColorless: true") ||
		strings.Contains(source, "game.ApplyContinuous") {
		t.Fatalf("generated source contains an unprinted characteristic change:\n%s", source)
	}
}
