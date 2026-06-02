package types

import "testing"

func TestKnownSubtypeForTypeIncludesComprehensiveRulesFamilies(t *testing.T) {
	tests := []struct {
		name     string
		cardType Card
		subtype  Sub
	}{
		{name: "artifact", cardType: Artifact, subtype: Blood},
		{name: "enchantment", cardType: Enchantment, subtype: Room},
		{name: "land", cardType: Land, subtype: PowerPlant},
		{name: "planeswalker", cardType: Planeswalker, subtype: Chandra},
		{name: "instant spell", cardType: Instant, subtype: Omen},
		{name: "sorcery spell", cardType: Sorcery, subtype: Lesson},
		{name: "creature", cardType: Creature, subtype: TimeLord},
		{name: "kindred creature", cardType: Kindred, subtype: TimeLord},
		{name: "plane", cardType: Plane, subtype: BolassMeditationRealm},
		{name: "dungeon", cardType: Dungeon, subtype: Undercity},
		{name: "battle", cardType: Battle, subtype: Siege},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !KnownSubtypeForType(tt.cardType, tt.subtype) {
				t.Fatalf("KnownSubtypeForType(%v, %v) = false, want true", tt.cardType, tt.subtype)
			}
		})
	}
}

func TestSubtypeFamilyCountsMatchComprehensiveRules20260417(t *testing.T) {
	tests := []struct {
		cardType Card
		want     int
	}{
		{cardType: Artifact, want: 21},
		{cardType: Enchantment, want: 12},
		{cardType: Land, want: 17},
		{cardType: Planeswalker, want: 80},
		{cardType: Instant, want: 5},
		{cardType: Sorcery, want: 5},
		{cardType: Creature, want: 317},
		{cardType: Plane, want: 82},
		{cardType: Dungeon, want: 1},
		{cardType: Battle, want: 1},
	}

	for _, tt := range tests {
		t.Run(string(tt.cardType), func(t *testing.T) {
			if got := len(subtypesByType[tt.cardType]); got != tt.want {
				t.Fatalf("%v subtype count = %d, want %d", tt.cardType, got, tt.want)
			}
		})
	}
}

func TestKnownSubtypeForTypeSeparatesSpacecraftFamilies(t *testing.T) {
	if ArtifactSpacecraft != PlanarSpacecraft {
		t.Fatal("Spacecraft subtype identifiers should preserve the same printed subtype value")
	}
	if !KnownSubtypeForType(Artifact, ArtifactSpacecraft) {
		t.Fatal("artifact Spacecraft subtype was not known for artifacts")
	}
	if !KnownSubtypeForType(Plane, PlanarSpacecraft) {
		t.Fatal("planar Spacecraft subtype was not known for planes")
	}
}
