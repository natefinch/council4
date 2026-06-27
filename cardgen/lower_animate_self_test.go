package cardgen

import (
	goparser "go/parser"
	"go/token"
	"strings"
	"testing"
)

func generateAnimateSelfSource(t *testing.T, card *ScryfallCard, letter string) string {
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

// TestLowerAnimateSelfManlandWithKeyword covers the manland shape: a land that
// becomes a single-color N/N creature with a keyword until end of turn, with the
// trailing "It's still a land." sentence folded into coverage (Faerie Conclave).
func TestLowerAnimateSelfManlandWithKeyword(t *testing.T) {
	source := generateAnimateSelfSource(t, &ScryfallCard{
		Name:       "Faerie Conclave",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "This land enters tapped.\n{T}: Add {U}.\n{1}{U}: This land becomes a 2/1 blue Faerie creature with flying until end of turn. It's still a land.",
	}, "f")
	assertSourceContains(t, source,
		"game.ApplyContinuous{",
		"Object: opt.Val(game.SourcePermanentReference()),",
		"Layer:     game.LayerColor,",
		"SetColors: []color.Color{color.Blue},",
		"Layer:       game.LayerType,",
		"AddTypes:    []types.Card{types.Creature},",
		"AddSubtypes: []types.Sub{types.Faerie},",
		"Layer: game.LayerAbility,",
		"game.Flying,",
		"Layer:        game.LayerPowerToughnessSet,",
		"SetPower:     opt.Val(game.PT{Value: 2}),",
		"SetToughness: opt.Val(game.PT{Value: 1}),",
		"Duration: game.DurationUntilEndOfTurn,",
	)
}

// TestLowerAnimateSelfMulticolorArtifact covers a mana rock that becomes a
// multicolor artifact creature with a keyword (Boros Keyrune): both stated
// colors, the added artifact card type alongside creature, and the keyword.
func TestLowerAnimateSelfMulticolorArtifact(t *testing.T) {
	source := generateAnimateSelfSource(t, &ScryfallCard{
		Name:       "Boros Keyrune",
		Layout:     "normal",
		TypeLine:   "Artifact",
		ManaCost:   "{3}",
		OracleText: "{T}: Add {R} or {W}.\n{R}{W}: This artifact becomes a 1/1 red and white Soldier artifact creature with double strike until end of turn.",
	}, "b")
	assertSourceContains(t, source,
		"SetColors: []color.Color{color.Red, color.White},",
		"AddTypes:    []types.Card{types.Creature, types.Artifact},",
		"AddSubtypes: []types.Sub{types.Soldier},",
		"game.DoubleStrike,",
		"SetPower:     opt.Val(game.PT{Value: 1}),",
		"SetToughness: opt.Val(game.PT{Value: 1}),",
	)
}

// TestLowerAnimateSelfEveryCreatureType covers the colorless "all creature
// types" rider (Mutavault), which adds every creature type rather than named
// subtypes and grants no keywords or colors.
func TestLowerAnimateSelfEveryCreatureType(t *testing.T) {
	source := generateAnimateSelfSource(t, &ScryfallCard{
		Name:       "Mutavault",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{T}: Add {C}.\n{1}: This land becomes a 2/2 creature with all creature types until end of turn. It's still a land.",
	}, "m")
	assertSourceContains(t, source,
		"AddTypes:             []types.Card{types.Creature},",
		"AddEveryCreatureType: true,",
		"SetPower:     opt.Val(game.PT{Value: 2}),",
		"SetToughness: opt.Val(game.PT{Value: 2}),",
	)
	if strings.Contains(source, "AddSubtypes") {
		t.Fatalf("Mutavault must not name subtypes when gaining all creature types:\n%s", source)
	}
	if strings.Contains(source, "game.LayerColor") {
		t.Fatalf("Mutavault is colorless and must not set colors:\n%s", source)
	}
}

// TestLowerAnimateSelfTwoKeywords covers two animated keywords joined by "and"
// (Inkmoth Nexus: flying and infect).
func TestLowerAnimateSelfTwoKeywords(t *testing.T) {
	source := generateAnimateSelfSource(t, &ScryfallCard{
		Name:       "Inkmoth Nexus",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{T}: Add {C}.\n{1}: This land becomes a 1/1 Phyrexian Blinkmoth artifact creature with flying and infect until end of turn. It's still a land. (It deals damage to creatures in the form of -1/-1 counters and to players in the form of poison counters.)",
	}, "i")
	assertSourceContains(t, source,
		"AddTypes:    []types.Card{types.Creature, types.Artifact},",
		"game.Flying,",
		"game.Infect,",
	)
}

// TestAnimateSelfFailsClosedOnVariablePT confirms an X/X self-animation stays
// unsupported (fail closed) rather than fabricating a fixed power/toughness
// (Chimeric Staff).
func TestAnimateSelfFailsClosedOnVariablePT(t *testing.T) {
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Chimeric Staff",
		Layout:     "normal",
		TypeLine:   "Artifact",
		ManaCost:   "{4}",
		OracleText: "{X}: This artifact becomes an X/X Construct artifact creature until end of turn.",
	}, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected diagnostics for X/X self-animation, got none")
	}
}
