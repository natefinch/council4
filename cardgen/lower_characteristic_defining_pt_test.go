package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
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
