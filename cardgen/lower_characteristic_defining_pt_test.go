package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerCharacteristicDefiningPowerToughness(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		card string
		text string
		want game.DynamicValueKind
	}{
		{
			name: "cards in hand",
			card: "Maro",
			text: "Maro's power and toughness are each equal to the number of cards in your hand.",
			want: game.DynamicValueControllerHandSize,
		},
		{
			name: "lands you control",
			card: "Multani, Maro-Sorcerer",
			text: "This creature's power and toughness are each equal to the number of lands you control.",
			want: game.DynamicValueControllerLandCount,
		},
		{
			name: "cards in graveyard",
			card: "Splinterfright",
			text: "This creature's power and toughness are each equal to the number of cards in your graveyard.",
			want: game.DynamicValueControllerGraveyardSize,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			star := "*"
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       tc.card,
				Layout:     "normal",
				TypeLine:   "Creature — Elemental",
				ManaCost:   "{4}{G}",
				Power:      &star,
				Toughness:  &star,
				OracleText: tc.text,
			})
			if !face.DynamicPower.Exists || !face.DynamicToughness.Exists {
				t.Fatalf("dynamic power/toughness not set: %+v", face)
			}
			if face.DynamicPower.Val.Kind != tc.want || face.DynamicToughness.Val.Kind != tc.want {
				t.Fatalf("dynamic kind = %v/%v, want %v", face.DynamicPower.Val.Kind, face.DynamicToughness.Val.Kind, tc.want)
			}
			if len(face.StaticAbilities) != 0 {
				t.Fatalf("static abilities = %d, want 0 (characteristic-defining ability sets the printed P/T)", len(face.StaticAbilities))
			}
		})
	}
}

func TestLowerCharacteristicDefiningPowerOnly(t *testing.T) {
	t.Parallel()
	star := "*"
	four := "4"
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Adeline, Resplendent Cathar",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Human Soldier",
		ManaCost:   "{1}{W}{W}",
		Power:      &star,
		Toughness:  &four,
		OracleText: "This creature's power is equal to the number of creatures you control.",
	})
	if !face.DynamicPower.Exists {
		t.Fatalf("dynamic power not set: %+v", face)
	}
	if face.DynamicPower.Val.Kind != game.DynamicValueControllerCreatureCount {
		t.Fatalf("dynamic power kind = %v, want creature count", face.DynamicPower.Val.Kind)
	}
	if face.DynamicToughness.Exists {
		t.Fatalf("dynamic toughness must not be set for a power-only CDA: %+v", face.DynamicToughness)
	}
}

func TestLowerCharacteristicDefiningToughnessOffset(t *testing.T) {
	t.Parallel()
	star := "*"
	offsetStar := "1+*"
	face := lowerSingleFace(t, &ScryfallCard{
		Name:      "Tarmogoyf",
		Layout:    "normal",
		TypeLine:  "Creature — Lhurgoyf",
		ManaCost:  "{1}{G}",
		Power:     &star,
		Toughness: &offsetStar,
		OracleText: "Tarmogoyf's power is equal to the number of card types among cards in all graveyards " +
			"and its toughness is equal to that number plus 1.",
	})
	if !face.DynamicPower.Exists || !face.DynamicToughness.Exists {
		t.Fatalf("dynamic power/toughness not set: %+v", face)
	}
	if face.DynamicPower.Val.Kind != game.DynamicValueCardTypesAmongAllGraveyards ||
		face.DynamicPower.Val.Offset != 0 {
		t.Fatalf("dynamic power = %+v, want card-types-among-all-graveyards with no offset", face.DynamicPower.Val)
	}
	if face.DynamicToughness.Val.Kind != game.DynamicValueCardTypesAmongAllGraveyards ||
		face.DynamicToughness.Val.Offset != 1 {
		t.Fatalf("dynamic toughness = %+v, want card-types-among-all-graveyards plus 1", face.DynamicToughness.Val)
	}
}

func TestLowerCharacteristicDefiningLandSubtypeCount(t *testing.T) {
	t.Parallel()
	star := "*"
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Korlash, Heir to Blackblade",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Zombie",
		ManaCost:   "{2}{B}{B}",
		Power:      &star,
		Toughness:  &star,
		OracleText: "Korlash's power and toughness are each equal to the number of Swamps you control.",
	})
	if !face.DynamicPower.Exists || !face.DynamicToughness.Exists {
		t.Fatalf("dynamic power/toughness not set: %+v", face)
	}
	if face.DynamicPower.Val.Kind != game.DynamicValueControllerSubtypeCount ||
		face.DynamicPower.Val.Subtype != types.Swamp {
		t.Fatalf("dynamic power = %+v, want land-subtype count of Swamp", face.DynamicPower.Val)
	}
	if face.DynamicToughness.Val.Subtype != types.Swamp {
		t.Fatalf("dynamic toughness subtype = %q, want Swamp", face.DynamicToughness.Val.Subtype)
	}
}

func TestLowerCharacteristicDefiningCreatureSubtypeCount(t *testing.T) {
	t.Parallel()
	star := "*"
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Goblin Warchief",
		Layout:     "normal",
		TypeLine:   "Creature — Goblin",
		ManaCost:   "{1}{R}{R}",
		Power:      &star,
		Toughness:  &star,
		OracleText: "Goblin Warchief's power and toughness are each equal to the number of Goblins you control.",
	})
	if !face.DynamicPower.Exists || !face.DynamicToughness.Exists {
		t.Fatalf("dynamic power/toughness not set: %+v", face)
	}
	if face.DynamicPower.Val.Kind != game.DynamicValueControllerSubtypeCount ||
		face.DynamicPower.Val.Subtype != types.Goblin {
		t.Fatalf("dynamic power = %+v, want subtype count of Goblin", face.DynamicPower.Val)
	}
}

func TestLowerCharacteristicDefiningColorPermanentCount(t *testing.T) {
	t.Parallel()
	star := "*"
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Crimson Avatar",
		Layout:     "normal",
		TypeLine:   "Creature — Avatar",
		ManaCost:   "{3}{R}",
		Power:      &star,
		Toughness:  &star,
		OracleText: "Crimson Avatar's power and toughness are each equal to the number of red permanents you control.",
	}, "t")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	if strings.Count(source, "game.DynamicValue{Kind: game.DynamicValueControllerColorPermanentCount, Color: color.Red}") != 2 {
		t.Fatalf("generated source missing color permanent count, got:\n%s", source)
	}
}

func TestGenerateExecutableCardSourcePsychosisCrawler(t *testing.T) {
	t.Parallel()
	star := "*"
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:      "Psychosis Crawler",
		Layout:    "normal",
		TypeLine:  "Artifact Creature — Construct",
		ManaCost:  "{5}",
		Power:     &star,
		Toughness: &star,
		OracleText: "Psychosis Crawler's power and toughness are each equal to the number of cards in your hand.\n" +
			"Whenever you draw a card, each opponent loses 1 life.",
	}, "t")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	if !strings.Contains(source, "DynamicPower:") ||
		!strings.Contains(source, "DynamicToughness:") ||
		strings.Count(source, "game.DynamicValue{Kind: game.DynamicValueControllerHandSize}") != 2 {
		t.Fatalf("generated source missing dynamic power/toughness, got:\n%s", source)
	}
}

func TestLowerCharacteristicDefiningPowerToughnessNonSourceFailsClosed(t *testing.T) {
	t.Parallel()
	face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Aura",
		Layout:     "normal",
		TypeLine:   "Enchantment — Aura",
		OracleText: "Enchanted creature's power and toughness are each equal to the number of cards in your hand.",
	})
	if face.DynamicPower.Exists || face.DynamicToughness.Exists {
		t.Fatal("non-source characteristic-defining P/T must not lower to a face dynamic value")
	}
}
