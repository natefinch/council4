package cardgen

import (
	goparser "go/parser"
	"go/token"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerDamagePreventionReplacement(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		oracleText     string
		amount         int
		sourceColors   []color.Color
		sourceTypes    []types.Card
		sourceOpponent bool
	}{
		{
			name:         "Sphere of Law red",
			oracleText:   "If a red source would deal damage to you, prevent 2 of that damage.",
			amount:       2,
			sourceColors: []color.Color{color.Red},
		},
		{
			name:        "Sphere of Purity artifact",
			oracleText:  "If an artifact would deal damage to you, prevent 1 of that damage.",
			amount:      1,
			sourceTypes: []types.Card{types.Artifact},
		},
		{
			name:       "Urza's Armor any source",
			oracleText: "If a source would deal damage to you, prevent 1 of that damage.",
			amount:     1,
		},
		{
			name:           "Protection of the Hekma opponent source",
			oracleText:     "If a source an opponent controls would deal damage to you, prevent 1 of that damage.",
			amount:         1,
			sourceOpponent: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Damage Preventer",
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: test.oracleText,
			})
			if len(face.ReplacementAbilities) != 1 {
				t.Fatalf("got %d replacement abilities, want 1", len(face.ReplacementAbilities))
			}
			replacement := face.ReplacementAbilities[0].Replacement
			if replacement.MatchEvent != game.EventDamageDealt ||
				replacement.ControllerFilter != game.TriggerControllerAny ||
				replacement.DamagePreventAmount != test.amount ||
				!replacement.DamageRecipientController ||
				!slices.Equal(replacement.DamageSourceColors, test.sourceColors) ||
				!slices.Equal(replacement.DamageSourceTypes, test.sourceTypes) ||
				replacement.DamageSourceControllerOpponent != test.sourceOpponent ||
				replacement.Duration != game.DurationPermanent {
				t.Fatalf("replacement = %+v, want damage prevention replacement", replacement)
			}
		})
	}
}

func TestGenerateDamagePreventionReplacementSource(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		wanted     []string
	}{
		{
			name:       "Sphere of Law",
			oracleText: "If a red source would deal damage to you, prevent 2 of that damage.",
			wanted:     []string{"game.DamagePreventionReplacement", "Amount: 2", "color.Red"},
		},
		{
			name:       "Sphere of Purity",
			oracleText: "If an artifact would deal damage to you, prevent 1 of that damage.",
			wanted:     []string{"game.DamagePreventionReplacement", "Amount: 1", "types.Artifact"},
		},
		{
			name:       "Protection of the Hekma",
			oracleText: "If a source an opponent controls would deal damage to you, prevent 1 of that damage.",
			wanted:     []string{"game.DamagePreventionReplacement", "Amount: 1", "SourceControllerOpponent: true"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       test.name,
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: test.oracleText,
			}, "d")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("unexpected diagnostics: %#v", diagnostics)
			}
			for _, wanted := range test.wanted {
				if !strings.Contains(source, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
			}
			if _, err := goparser.ParseFile(token.NewFileSet(), "generated.go", source, goparser.AllErrors); err != nil {
				t.Fatalf("generated source does not parse: %v\n%s", err, source)
			}
		})
	}
}
