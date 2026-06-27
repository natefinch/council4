package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceCopyTokenOverrideReplace covers the
// replacement form of the copy-token characteristic-overriding exception
// ("except it's a 1/1 green Frog" — Croaking Counterpart). The fixed
// power/toughness, color, and subtype replace the copied values, so they lower
// to SetPower/SetToughness/SetColors/SetSubtypes.
func TestGenerateExecutableCardSourceCopyTokenOverrideReplace(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Frog Maker",
		Layout:     "normal",
		ManaCost:   "{1}{G}",
		TypeLine:   "Sorcery",
		OracleText: "Create a token that's a copy of target creature, except it's a 1/1 green Frog.",
		Colors:     []string{"G"},
	}, "f")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Source: game.TokenCopyOf(game.TokenCopySpec{",
		"SetPower:",
		"SetToughness:",
		"game.PT{Value: 1}",
		"SetColors:",
		"[]color.Color{color.Green}",
		"SetSubtypes:",
		"[]types.Sub{types.Frog}",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if strings.Contains(source, "AddColors:") || strings.Contains(source, "AddSubtypes:") {
		t.Fatalf("replacement form must not emit additive fields:\n%s", source)
	}
}

// TestGenerateExecutableCardSourceCopyTokenOverrideAdditiveColorsAndTypes covers
// the additive form whose "in addition to its other colors and types" suffix
// makes both the color and the subtype additive ("except it's not legendary and
// it's a 2/2 black Zombie in addition to its other colors and types" —
// Ratadrabik of Urborg). The color and subtype append to the copied values
// (AddColors/AddSubtypes), the new power/toughness still replaces, and the
// legendary drop lowers to SetNotLegendary.
func TestGenerateExecutableCardSourceCopyTokenOverrideAdditiveColorsAndTypes(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Zombie Maker",
		Layout:   "normal",
		ManaCost: "{2}{B}",
		TypeLine: "Sorcery",
		OracleText: "Create a token that's a copy of target creature, except it's not legendary " +
			"and it's a 2/2 black Zombie in addition to its other colors and types.",
		Colors: []string{"B"},
	}, "z")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"SetNotLegendary: true,",
		"SetPower:",
		"game.PT{Value: 2}",
		"AddColors:",
		"[]color.Color{color.Black}",
		"AddSubtypes:",
		"[]types.Sub{types.Zombie}",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if strings.Contains(source, "SetColors:") || strings.Contains(source, "SetSubtypes:") {
		t.Fatalf("additive form must not emit replacement color/subtype fields:\n%s", source)
	}
}

// TestGenerateExecutableCardSourceCopyTokenOverrideAdditiveTypes covers the
// additive form whose "in addition to its other types" suffix makes the card
// type and subtype additive while colors are absent ("except it's a 1/1 Soldier
// creature in addition to its other types" — Urza, Prince of Kroog). The
// creature card type lowers to AddTypes and the Soldier subtype to AddSubtypes.
func TestGenerateExecutableCardSourceCopyTokenOverrideAdditiveTypes(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Soldier Maker",
		Layout:     "normal",
		ManaCost:   "{6}",
		TypeLine:   "Artifact",
		OracleText: "{6}: Create a token that's a copy of target artifact you control, except it's a 1/1 Soldier creature in addition to its other types.",
	}, "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"AddTypes:",
		"[]types.Card{types.Creature}",
		"AddSubtypes:",
		"[]types.Sub{types.Soldier}",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestGenerateExecutableCardSourceCopyTokenOverrideKeyword covers the additive
// form carrying an inline granted keyword ("except it's a 4/4 Dragon creature
// with flying in addition to its other types" — Will of the Temur). The "with
// flying" keyword joins the copy as an added keyword ability (AddKeywords),
// distinct from the additive card-type and subtype overrides.
func TestGenerateExecutableCardSourceCopyTokenOverrideKeyword(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Dragon Maker",
		Layout:     "normal",
		ManaCost:   "{3}{R}{G}",
		TypeLine:   "Sorcery",
		OracleText: "Create a token that's a copy of target permanent, except it's a 4/4 Dragon creature with flying in addition to its other types.",
		Colors:     []string{"R", "G"},
	}, "d")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"SetPower:",
		"game.PT{Value: 4}",
		"AddTypes:",
		"[]types.Card{types.Creature}",
		"AddSubtypes:",
		"[]types.Sub{types.Dragon}",
		"AddKeywords:",
		"[]game.Keyword{game.Flying}",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestGenerateExecutableCardSourceCopyTokenOverrideNamedFailsClosed covers a
// copy-token exception that renames the token ("except it's a legendary Alien
// named Prisoner Zero" — The Eleventh Hour). A printed token name is not part of
// the supported characteristic-override grammar, so the recognizer must leave
// the create unrecognized and the card must fail closed rather than silently
// dropping the name.
func TestGenerateExecutableCardSourceCopyTokenOverrideNamedFailsClosed(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Named Maker",
		Layout:     "normal",
		ManaCost:   "{2}{U}",
		TypeLine:   "Sorcery",
		OracleText: "Create a token that's a copy of target creature, except it's a legendary Alien named Prisoner Zero.",
		Colors:     []string{"U"},
	}, "n")
	if err != nil {
		t.Fatal(err)
	}
	if source != "" || len(diagnostics) == 0 {
		t.Fatalf("want fail-closed (empty source, diagnostics); got source=%q diagnostics=%#v", source, diagnostics)
	}
}
