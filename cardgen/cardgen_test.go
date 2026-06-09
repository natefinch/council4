package cardgen

import (
	"testing"
)

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
		{
			"host creature",
			"Host Creature — Beaver",
			[]string{"Host"},
			[]string{"Creature"},
			[]string{"Beaver"},
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
