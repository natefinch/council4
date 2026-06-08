package cardgen

import (
	"strings"
	"testing"
)

func TestGenerateExecutableCardSourceKeywords(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Keyword Bear",
		Layout:     "normal",
		ManaCost:   "{1}{G}",
		TypeLine:   "Creature — Bear",
		OracleText: "Flying, vigilance",
		Power:      new("2"),
		Toughness:  new("2"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "k")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.FlyingStaticBody",
		"game.VigilanceStaticBody",
		"StaticAbilities: []game.StaticAbility",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if strings.Contains(source, "TODO") {
		t.Fatalf("executable source contains TODO:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceRejectsPartialAbility(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Drawing Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature enters, draw a card.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "d")
	if err != nil {
		t.Fatal(err)
	}
	if source != "" || len(diagnostics) == 0 {
		t.Fatalf("source = %q, diagnostics = %#v", source, diagnostics)
	}
	if got := diagnostics[0].Summary; got != "unsupported triggered ability" {
		t.Fatalf("summary = %q", got)
	}
}

func TestGenerateExecutableCardSourceRejectsPartiallyRecognizedKeywordLine(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Bayou Dragonfly",
		Layout:     "normal",
		TypeLine:   "Creature — Insect",
		OracleText: "Flying; swampwalk (This creature can't be blocked as long as defending player controls a Swamp.)",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "b")
	if err != nil {
		t.Fatal(err)
	}
	if source != "" || len(diagnostics) == 0 {
		t.Fatalf("source = %q, diagnostics = %#v", source, diagnostics)
	}
	if got := diagnostics[0].Summary; got != "unsupported mixed keyword ability" {
		t.Fatalf("summary = %q", got)
	}
}

func TestGenerateExecutableCardSourceExplainsUnsupportedAbility(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		typeLine   string
		oracleText string
		summary    string
		detail     string
	}{
		"spell": {
			typeLine:   "Sorcery",
			oracleText: "Draw two cards.",
			summary:    "unsupported spell ability",
			detail:     "does not yet lower spell abilities",
		},
		"activated": {
			typeLine:   "Creature — Bear",
			oracleText: "{T}: Add {G}.",
			summary:    "unsupported activated ability",
			detail:     "does not yet lower activated abilities",
		},
		"static rules text": {
			typeLine:   "Enchantment",
			oracleText: "Creatures you control get +1/+1.",
			summary:    "unsupported static ability",
			detail:     "non-keyword static rules text",
		},
		"parameterized keyword": {
			typeLine:   "Creature — Snake",
			oracleText: "Toxic 1",
			summary:    "unsupported parameterized keyword",
			detail:     `Toxic with parameter "1"`,
		},
		"keyword without template": {
			typeLine:   "Creature — Dinosaur",
			oracleText: "Ward",
			summary:    "unsupported keyword ability",
			detail:     "no reusable game template for Ward",
		},
		"modal": {
			typeLine:   "Sorcery",
			oracleText: "Choose one —\n• Draw a card.\n• Destroy target creature.",
			summary:    "unsupported modal ability",
			detail:     "does not yet lower modal abilities",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Unsupported Example",
				Layout:     "normal",
				TypeLine:   test.typeLine,
				OracleText: test.oracleText,
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "u")
			if err != nil {
				t.Fatal(err)
			}
			if source != "" || len(diagnostics) == 0 {
				t.Fatalf("source = %q, diagnostics = %#v", source, diagnostics)
			}
			if got := diagnostics[0].Summary; got != test.summary {
				t.Fatalf("summary = %q, want %q", got, test.summary)
			}
			if got := diagnostics[0].Detail; !strings.Contains(got, test.detail) {
				t.Fatalf("detail = %q, want substring %q", got, test.detail)
			}
		})
	}
}

func TestGenerateExecutableCardSourceVanilla(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:      "Vanilla Bear",
		Layout:    "normal",
		TypeLine:  "Creature — Bear",
		Power:     new("2"),
		Toughness: new("2"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "v")
	if err != nil {
		t.Fatal(err)
	}
	if source == "" || len(diagnostics) != 0 || strings.Contains(source, "TODO") {
		t.Fatalf("source = %q, diagnostics = %#v", source, diagnostics)
	}
}

func TestGenerateExecutableCardSourceRejectsUnknownTypeLine(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Decklist",
		Layout:   "token",
		TypeLine: "Card",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "d")
	if err != nil {
		t.Fatal(err)
	}
	if source != "" || len(diagnostics) == 0 {
		t.Fatalf("source = %q, diagnostics = %#v", source, diagnostics)
	}
}

func TestGenerateExecutableCardSourceKeepsSameNamedFacesPositional(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:   "Insect // Insect",
		Layout: "reversible_card",
		CardFaces: []ScryfallCardFace{
			{Name: "Insect", TypeLine: "Creature — Insect", OracleText: "Flying"},
			{Name: "Insect", TypeLine: "Creature — Insect", OracleText: "Haste"},
		},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "i")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if strings.Count(source, "game.FlyingStaticBody") != 1 ||
		strings.Count(source, "game.HasteStaticBody") != 1 {
		t.Fatalf("face abilities were not kept positional:\n%s", source)
	}
}
