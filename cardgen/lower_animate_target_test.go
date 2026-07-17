package cardgen

import (
	goparser "go/parser"
	"go/token"
	"strings"
	"testing"
)

func generateAnimateTargetSource(t *testing.T, card *ScryfallCard, letter string) string {
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

// TestLowerAnimateTargetInlineStillALand covers the leading-duration inline form
// where the "that's still a land" relative clause co-occurs with a leading
// "Until end of turn," (Animate Land). The continuous effect must bind to the
// single target land rather than the source permanent.
func TestLowerAnimateTargetInlineStillALand(t *testing.T) {
	source := generateAnimateTargetSource(t, &ScryfallCard{
		Name:       "Animate Land",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{G}",
		OracleText: "Until end of turn, target land becomes a 3/3 creature that's still a land.",
	}, "a")
	assertSourceContains(t, source,
		"Constraint: \"target land\",",
		"Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Land}}),",
		"game.ApplyContinuous{",
		"Object: opt.Val(game.TargetPermanentReference(0)),",
		"Layer:    game.LayerType,",
		"AddTypes: []types.Card{types.Creature},",
		"Layer:        game.LayerPowerToughnessSet,",
		"SetPower:     opt.Val(game.PT{Value: 3}),",
		"SetToughness: opt.Val(game.PT{Value: 3}),",
		"Duration: game.DurationUntilEndOfTurn,",
	)
	if strings.Contains(source, "game.SourcePermanentReference()") {
		t.Fatalf("target animation must bind to the target, not the source:\n%s", source)
	}
}

// TestLowerAnimateTargetSubtypeAndKeyword covers the trailing-sentence form
// ("It's still a land.") with a named subtype and a granted keyword (Hydroform:
// Elemental, flying).
func TestLowerAnimateTargetSubtypeAndKeyword(t *testing.T) {
	source := generateAnimateTargetSource(t, &ScryfallCard{
		Name:       "Hydroform",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{1}{U}",
		OracleText: "Target land becomes a 3/3 Elemental creature with flying until end of turn. It's still a land.",
	}, "h")
	assertSourceContains(t, source,
		"Object: opt.Val(game.TargetPermanentReference(0)),",
		"AddSubtypes: []types.Sub{types.Elemental},",
		"Layer: game.LayerAbility,",
		"game.Flying,",
		"SetPower:     opt.Val(game.PT{Value: 3}),",
		"SetToughness: opt.Val(game.PT{Value: 3}),",
	)
}

// TestLowerAnimateTargetYouControlWithColor covers the "target land you control"
// subject with an explicit color set (Ignition Team: red Elemental), confirming
// the controller restriction and the LayerColor SetColors.
func TestLowerAnimateTargetYouControlWithColor(t *testing.T) {
	source := generateAnimateTargetSource(t, &ScryfallCard{
		Name:       "Loamspeaker Probe",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{2}{R}",
		OracleText: "Target land you control becomes a 4/4 red Elemental creature until end of turn. It's still a land.",
	}, "l")
	assertSourceContains(t, source,
		"Constraint: \"target land you control\",",
		"Controller: game.ControllerYou",
		"Layer:     game.LayerColor,",
		"SetColors: []color.Color{color.Red},",
		"AddSubtypes: []types.Sub{types.Elemental},",
		"SetPower:     opt.Val(game.PT{Value: 4}),",
	)
}

// TestAnimateTargetFailsClosedOnIndefiniteDuration confirms that a land
// animation lacking the until-end-of-turn duration (permanent "lasts
// indefinitely" form) stays unsupported rather than fabricating a duration.
func TestAnimateTargetFailsClosedOnIndefiniteDuration(t *testing.T) {
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Permanent Animator",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{1}{G}",
		OracleText: "Target land becomes a 3/3 creature. It's still a land. This effect lasts indefinitely.",
	}, "p")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected diagnostics for indefinite-duration land animation, got none")
	}
}

// TestLowerAnimateTargetDynamicXX covers Destiny Spinner's variable "X/X"
// animation bound by a trailing "where X is the number of enchantments you
// control" clause. The base power/toughness lower to SetPowerDynamic and
// SetToughnessDynamic reading a controlled-enchantment count (locked at
// resolution) rather than the fixed SetPower/SetToughness of an N/N animation.
func TestLowerAnimateTargetDynamicXX(t *testing.T) {
	ptr := func(s string) *string { return &s }
	source := generateAnimateTargetSource(t, &ScryfallCard{
		Name:       "Destiny Spinner",
		Layout:     "normal",
		TypeLine:   "Enchantment Creature — Human",
		ManaCost:   "{1}{G}",
		Power:      ptr("2"),
		Toughness:  ptr("3"),
		OracleText: "Creature and enchantment spells you control can't be countered.\n{3}{G}: Target land you control becomes an X/X Elemental creature with trample and haste until end of turn, where X is the number of enchantments you control. It's still a land.",
	}, "d")
	assertSourceContains(t, source,
		"Constraint: \"target land you control\",",
		"Controller: game.ControllerYou",
		"AddSubtypes: []types.Sub{types.Elemental},",
		"game.Trample,",
		"game.Haste,",
		"Layer: game.LayerPowerToughnessSet,",
		"SetPowerDynamic: opt.Val(game.DynamicAmount{",
		"Kind:       game.DynamicAmountCountSelector,",
		"Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Enchantment}, Controller: game.ControllerYou}),",
		"SetToughnessDynamic: opt.Val(game.DynamicAmount{",
		"Duration: game.DurationUntilEndOfTurn,",
	)
	if strings.Contains(source, "SetPower:     opt.Val") || strings.Contains(source, "SetToughness: opt.Val") {
		t.Fatalf("dynamic X/X animation must not emit a fixed base power/toughness:\n%s", source)
	}
}

// TestAnimateTargetFailsClosedOnUnsupportedWhereX confirms a variable "X/X"
// animation whose "where X is ..." clause is not the supported "number of
// <permanent type> you control" shape stays unsupported rather than fabricating
// a base power/toughness. A creature-subtype count ("Allies") is not a permanent
// card type, so the X binding fails closed.
func TestAnimateTargetFailsClosedOnUnsupportedWhereX(t *testing.T) {
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Ally Animator",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{2}{G}",
		OracleText: "Target land you control becomes an X/X Elemental creature until end of turn, where X is the number of Allies you control. It's still a land.",
	}, "a")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected diagnostics for an unsupported where-X subtype count, got none")
	}
}
