package cardgen

import (
	goparser "go/parser"
	"go/token"
	"strings"
	"testing"
)

func generateSetBasePTSource(t *testing.T, card *ScryfallCard, letter string) string {
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

func assertSourceContains(t *testing.T, source string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

// TestLowerSetBasePTGroupVariableX covers the activated group form with the
// X/X base power/toughness set plus the every-creature-type rider (Mirror
// Entity), which is the headline shape of issue #1565.
func TestLowerSetBasePTGroupVariableX(t *testing.T) {
	source := generateSetBasePTSource(t, &ScryfallCard{
		Name:       "Mirror Entity",
		Layout:     "normal",
		TypeLine:   "Creature — Shapeshifter",
		ManaCost:   "{2}{W}",
		OracleText: "Changeling (This card is every creature type.)\n{X}: Until end of turn, creatures you control have base power and toughness X/X and gain all creature types.",
	}, "m")
	assertSourceContains(t, source,
		"game.ApplyContinuous{",
		"game.LayerPowerToughnessSet,",
		"Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),",
		"SetPowerDynamic: opt.Val(game.DynamicAmount{",
		"SetToughnessDynamic: opt.Val(game.DynamicAmount{",
		"Kind: game.DynamicAmountX,",
		"game.LayerType,",
		"AddEveryCreatureType: true,",
		"Duration: game.DurationUntilEndOfTurn,",
	)
}

// TestLowerSetBasePTGroupSpellVariableX covers the resolving spell group form
// with X/X and no type rider (Biomass Mutation).
func TestLowerSetBasePTGroupSpellVariableX(t *testing.T) {
	source := generateSetBasePTSource(t, &ScryfallCard{
		Name:       "Biomass Mutation",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{X}{G}{U}",
		OracleText: "Creatures you control have base power and toughness X/X until end of turn.",
	}, "b")
	assertSourceContains(t, source,
		"game.LayerPowerToughnessSet,",
		"Controller: game.ControllerYou",
		"SetPowerDynamic: opt.Val(game.DynamicAmount{",
		"SetToughnessDynamic: opt.Val(game.DynamicAmount{",
	)
	if strings.Contains(source, "AddEveryCreatureType") {
		t.Fatalf("Biomass Mutation must not gain creature types:\n%s", source)
	}
}

// TestLowerSetBasePTGroupOpponentFixed covers the opponent-controlled group with
// a fixed 0/1 set (Flatline).
func TestLowerSetBasePTGroupOpponentFixed(t *testing.T) {
	source := generateSetBasePTSource(t, &ScryfallCard{
		Name:       "Flatline",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{2}{B}",
		OracleText: "Creatures your opponents control have base power and toughness 0/1 until end of turn.",
	}, "f")
	assertSourceContains(t, source,
		"game.LayerPowerToughnessSet,",
		"Controller: game.ControllerOpponent",
		"SetPower:     opt.Val(game.PT{Value: 0}),",
		"SetToughness: opt.Val(game.PT{Value: 1}),",
	)
}

// TestLowerSetBasePTTargetFixed covers the single-target fixed form (Square Up).
func TestLowerSetBasePTTargetFixed(t *testing.T) {
	source := generateSetBasePTSource(t, &ScryfallCard{
		Name:       "Square Up",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{G}",
		OracleText: "Target creature has base power and toughness 4/4 until end of turn.",
	}, "s")
	assertSourceContains(t, source,
		"game.LayerPowerToughnessSet,",
		"game.TargetPermanentReference(0)",
		"SetPower:     opt.Val(game.PT{Value: 4}),",
		"SetToughness: opt.Val(game.PT{Value: 4}),",
	)
}

// TestLowerSetBasePTActivatedTargetFixed covers an activated targeted set form
// (Gigantomancer).
func TestLowerSetBasePTActivatedTargetFixed(t *testing.T) {
	source := generateSetBasePTSource(t, &ScryfallCard{
		Name:       "Gigantomancer",
		Layout:     "normal",
		TypeLine:   "Creature — Human Wizard",
		ManaCost:   "{3}{G}",
		OracleText: "{1}: Target creature you control has base power and toughness 7/7 until end of turn.",
	}, "g")
	assertSourceContains(t, source,
		"game.LayerPowerToughnessSet,",
		"game.TargetPermanentReference(0)",
		"SetPower:     opt.Val(game.PT{Value: 7}),",
		"SetToughness: opt.Val(game.PT{Value: 7}),",
	)
}

// TestLowerSetBasePTSourceFixed covers the source form "This creature has base
// power and toughness N/N until end of turn." (Marsh Flitter). The "This
// creature" subject carries the inherent source self-reference, which must not
// block the source-form lowering.
func TestLowerSetBasePTSourceFixed(t *testing.T) {
	source := generateSetBasePTSource(t, &ScryfallCard{
		Name:       "Marsh Flitter",
		Layout:     "normal",
		TypeLine:   "Creature — Faerie Goblin",
		ManaCost:   "{3}{B}",
		OracleText: "Sacrifice a Goblin: This creature has base power and toughness 3/3 until end of turn.",
	}, "m")
	assertSourceContains(t, source,
		"game.LayerPowerToughnessSet,",
		"Object: opt.Val(game.SourcePermanentReference()),",
		"SetPower:     opt.Val(game.PT{Value: 3}),",
		"SetToughness: opt.Val(game.PT{Value: 3}),",
	)
}

// TestSetBasePTFailsClosedOnUnsupportedRider confirms a base power/toughness set
// carrying an extra keyword rider stays unsupported (fail closed) rather than
// silently dropping the keyword.
func TestSetBasePTFailsClosedOnUnsupportedRider(t *testing.T) {
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Imaginary Trample Set",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{G}",
		OracleText: "Target creature has base power and toughness 4/4 and gains trample until end of turn.",
	}, "i")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected diagnostics for unsupported keyword rider, got none")
	}
}
