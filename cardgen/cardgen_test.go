package cardgen

import (
	"go/format"
	"strings"
	"testing"
)

func TestParseManaCostLiteral(t *testing.T) {
	tests := []struct {
		name string
		cost string
		want string
	}{
		{"empty", "", ""},
		{"single colored", "{R}", "cost.R"},
		{"generic plus colors", "{2}{W}{U}", "cost.O(2),cost.W,cost.U"},
		{"variable", "{X}{R}{R}", "cost.X,cost.R,cost.R"},
		{"hybrid", "{W/U}", "cost.HybridMana(mana.W,mana.U)"},
		{"phyrexian", "{W/P}", "cost.PhyrexianMana(mana.W)"},
		{"mono hybrid", "{2/W}", "cost.Twobrid(mana.W)"},
		{"colorless", "{C}", "cost.C"},
		{"snow", "{S}", "cost.S"},
		{"generic only", "{1}", "cost.O(1)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseManaCostLiteral(tt.cost)
			if err != nil {
				t.Fatalf("ParseManaCostLiteral(%q) error: %v", tt.cost, err)
			}
			got = strings.ReplaceAll(got, " ", "")
			got = strings.ReplaceAll(got, "\t", "")
			got = strings.ReplaceAll(got, "\n", "")
			if tt.want == "" {
				if got != "" {
					t.Errorf("ParseManaCostLiteral(%q) = %q, want empty", tt.cost, got)
				}
				return
			}
			if !strings.Contains(got, tt.want) {
				t.Errorf("ParseManaCostLiteral(%q) = %q, want to contain %q", tt.cost, got, tt.want)
			}
		})
	}
}

func TestParseTypeLine(t *testing.T) {
	tests := []struct {
		name       string
		typeLine   string
		supertypes []string
		types      []string
		subtypes   []string
	}{
		{
			"simple creature",
			"Creature — Angel",
			nil,
			[]string{"Creature"},
			[]string{"Angel"},
		},
		{
			"legendary creature",
			"Legendary Creature — Human Wizard",
			[]string{"Legendary"},
			[]string{"Creature"},
			[]string{"Human", "Wizard"},
		},
		{
			"artifact no subtypes",
			"Artifact",
			nil,
			[]string{"Artifact"},
			nil,
		},
		{
			"instant",
			"Instant",
			nil,
			[]string{"Instant"},
			nil,
		},
		{
			"basic land",
			"Basic Land — Forest",
			[]string{"Basic"},
			[]string{"Land"},
			[]string{"Forest"},
		},
		{
			"artifact creature",
			"Artifact Creature — Golem",
			nil,
			[]string{"Artifact", "Creature"},
			[]string{"Golem"},
		},
		{
			"time lord creature subtype",
			"Legendary Creature — Time Lord Doctor",
			[]string{"Legendary"},
			[]string{"Creature"},
			[]string{"Time Lord", "Doctor"},
		},
		{
			"multi-word plane subtype",
			"Plane — Bolas’s Meditation Realm",
			nil,
			[]string{"Plane"},
			[]string{"Bolas’s Meditation Realm"},
		},
		{
			"enchantment",
			"Enchantment",
			nil,
			[]string{"Enchantment"},
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseTypeLine(tt.typeLine)
			if !sliceEqual(got.Supertypes, tt.supertypes) {
				t.Errorf("supertypes = %v, want %v", got.Supertypes, tt.supertypes)
			}
			if !sliceEqual(got.Types, tt.types) {
				t.Errorf("types = %v, want %v", got.Types, tt.types)
			}
			if !sliceEqual(got.Subtypes, tt.subtypes) {
				t.Errorf("subtypes = %v, want %v", got.Subtypes, tt.subtypes)
			}
		})
	}
}

func TestSubtypeToLiteralUsesGameConstants(t *testing.T) {
	tests := []struct {
		name    string
		subtype string
		types   []string
		want    string
	}{
		{name: "creature", subtype: "Bird", types: []string{"Creature"}, want: "types.Bird"},
		{name: "kindred", subtype: "Human", types: []string{"Kindred"}, want: "types.Human"},
		{name: "land", subtype: "Mountain", types: []string{"Land"}, want: "types.Mountain"},
		{name: "artifact", subtype: "Equipment", types: []string{"Artifact"}, want: "types.Equipment"},
		{name: "enchantment", subtype: "Aura", types: []string{"Enchantment"}, want: "types.Aura"},
		{name: "planeswalker", subtype: "Chandra", types: []string{"Planeswalker"}, want: "types.Chandra"},
		{name: "instant spell", subtype: "Omen", types: []string{"Instant"}, want: "types.Omen"},
		{name: "sorcery spell", subtype: "Lesson", types: []string{"Sorcery"}, want: "types.Lesson"},
		{name: "two-word creature", subtype: "Time Lord", types: []string{"Creature"}, want: "types.TimeLord"},
		{name: "artifact spacecraft collision", subtype: "Spacecraft", types: []string{"Artifact"}, want: "types.ArtifactSpacecraft"},
		{name: "planar spacecraft collision", subtype: "Spacecraft", types: []string{"Plane"}, want: "types.PlanarSpacecraft"},
		{name: "multi-word plane", subtype: "Bolas’s Meditation Realm", types: []string{"Plane"}, want: "types.BolassMeditationRealm"},
		{name: "dungeon", subtype: "Undercity", types: []string{"Dungeon"}, want: "types.Undercity"},
		{name: "battle", subtype: "Siege", types: []string{"Battle"}, want: "types.Siege"},
		{name: "unknown", subtype: "Unlisted", types: []string{"Creature"}, want: `types.Sub("Unlisted")`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SubtypeToLiteral(tt.subtype, tt.types); got != tt.want {
				t.Fatalf("SubtypeToLiteral(%q, %+v) = %q, want %q", tt.subtype, tt.types, got, tt.want)
			}
		})
	}
}

func TestCardNameToVarName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"Lightning Bolt", "LightningBolt"},
		{"Sol Ring", "SolRing"},
		{"Swords to Plowshares", "SwordsToPlowshares"},
		{"Serra Angel", "SerraAngel"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CardNameToVarName(tt.name); got != tt.want {
				t.Errorf("CardNameToVarName(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestCardNameToFileName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"Lightning Bolt", "lightning_bolt"},
		{"Sol Ring", "sol_ring"},
		{"Swords to Plowshares", "swords_to_plowshares"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CardNameToFileName(tt.name); got != tt.want {
				t.Errorf("CardNameToFileName(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestCardNameToPackageLetter(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"Lightning Bolt", "l"},
		{"Sol Ring", "s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CardNameToPackageLetter(tt.name); got != tt.want {
				t.Errorf("CardNameToPackageLetter(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestGenerateCardSource(t *testing.T) {
	card := &ScryfallCard{
		Name:          "Lightning Bolt",
		ManaCost:      "{R}",
		CMC:           1.0,
		TypeLine:      "Instant",
		OracleText:    "Lightning Bolt deals 3 damage to any target.",
		Colors:        []string{"R"},
		ColorIdentity: []string{"R"},
	}

	got, err := GenerateCardSource(card, "l")
	if err != nil {
		t.Fatalf("GenerateCardSource error: %v", err)
	}
	assertGoSourceFormats(t, got)

	checks := []string{
		"package l",
		`Name: "Lightning Bolt"`,
		"cost.R",
		"types.Instant",
		"Colors: []color.Color{color.Red}",
		"ColorIdentity: color.NewIdentity(color.Red)",
		"Abilities: []game.AbilityDef{}",
		"Oracle text:",
		"Lightning Bolt deals 3 damage to any target.",
	}

	assertSourceContainsAll(t, got, checks)
}

func TestGenerateCardSourceCreature(t *testing.T) {
	power := "4"
	toughness := "4"
	card := &ScryfallCard{
		Name:          "Serra Angel",
		ManaCost:      "{3}{W}{W}",
		CMC:           5.0,
		TypeLine:      "Creature — Angel",
		OracleText:    "Flying\nVigilance (Attacking doesn't cause this creature to tap.)",
		Colors:        []string{"W"},
		ColorIdentity: []string{"W"},
		Power:         &power,
		Toughness:     &toughness,
	}

	got, err := GenerateCardSource(card, "s")
	if err != nil {
		t.Fatalf("GenerateCardSource error: %v", err)
	}
	assertGoSourceFormats(t, got)

	checks := []string{
		"package s",
		`Name: "Serra Angel"`,
		"types.Creature",
		"types.Angel",
		"Power: opt.Val(game.PT{Value: 4})",
		"Toughness: opt.Val(game.PT{Value: 4})",
	}

	assertSourceContainsAll(t, got, checks)
}

func TestGenerateCardSourceModalDFC(t *testing.T) {
	card := &ScryfallCard{
		Name:          "Front Spell // Back Land",
		Layout:        "modal_dfc",
		ColorIdentity: []string{"G"},
		CardFaces: []ScryfallCardFace{
			{
				Name:       "Front Spell",
				ManaCost:   "{2}{G}",
				TypeLine:   "Sorcery",
				OracleText: "Create a token.",
				Colors:     []string{"G"},
			},
			{
				Name:       "Back Land",
				TypeLine:   "Land — Forest",
				OracleText: "Back Land enters tapped.",
			},
		},
	}

	got, err := GenerateCardSource(card, "f")
	if err != nil {
		t.Fatalf("GenerateCardSource error: %v", err)
	}
	assertGoSourceFormats(t, got)

	checks := []string{
		"Layout: game.LayoutModalDFC",
		"Back: opt.Val(game.CardFace",
		`Name: "Front Spell"`,
		`Name: "Back Land"`,
		"types.Sorcery",
		"types.Land",
		"types.Forest",
		"color.NewIdentity(color.Green)",
		"game.EntersTappedReplacement",
	}
	assertSourceContainsAll(t, got, checks)
}

func TestGenerateCardSourceReversibleEmitsSeparateDefs(t *testing.T) {
	card := &ScryfallCard{
		Name:          "Side A // Side B",
		Layout:        "reversible_card",
		ColorIdentity: []string{"R", "W"},
		CardFaces: []ScryfallCardFace{
			{Name: "Side A", ManaCost: "{R}", TypeLine: "Creature — Goblin", OracleText: "Haste", Colors: []string{"R"}},
			{Name: "Side B", ManaCost: "{W}", TypeLine: "Creature — Soldier", OracleText: "Vigilance", Colors: []string{"W"}},
		},
	}

	got, err := GenerateCardSource(card, "s")
	if err != nil {
		t.Fatalf("GenerateCardSource error: %v", err)
	}
	assertGoSourceFormats(t, got)

	checks := []string{
		"var SideA = &game.CardDef",
		"var SideB = &game.CardDef",
		"Layout: game.LayoutReversibleCard",
		"color.NewIdentity(color.Red, color.White)",
	}
	assertSourceContainsAll(t, got, checks)
	if strings.Contains(got, "Back: opt.Val(game.CardFace") {
		t.Fatalf("reversible card generated face-selectable definition:\n%s", got)
	}
}

func sliceEqual(a, b []string) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func assertGoSourceFormats(t *testing.T, source string) {
	t.Helper()
	if _, err := format.Source([]byte(source)); err != nil {
		t.Fatalf("generated source is not valid Go: %v\n%s", err, source)
	}
}

func assertSourceContainsAll(t *testing.T, source string, checks []string) {
	t.Helper()
	compactSource := compactWhitespace(source)
	for _, check := range checks {
		if !strings.Contains(compactSource, compactWhitespace(check)) {
			t.Errorf("output missing %q\nfull output:\n%s", check, source)
		}
	}
}

func compactWhitespace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
