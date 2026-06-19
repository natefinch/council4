package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerStaticSpellCostModifier(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		oracleText string
		modifiers  []game.CostModifier
	}{
		"all spells reduction": {
			oracleText: "Spells you cast cost {1} less to cast.",
			modifiers: []game.CostModifier{
				{Kind: game.CostModifierSpell, GenericReduction: 1},
			},
		},
		"artifact spells reduction": {
			oracleText: "Artifact spells you cast cost {1} less to cast.",
			modifiers: []game.CostModifier{
				{Kind: game.CostModifierSpell, MatchCardType: true, CardType: types.Artifact, GenericReduction: 1},
			},
		},
		"creature spells increase": {
			oracleText: "Creature spells you cast cost {1} more to cast.",
			modifiers: []game.CostModifier{
				{Kind: game.CostModifierSpell, MatchCardType: true, CardType: types.Creature, GenericIncrease: 1},
			},
		},
		"instant and sorcery reduction": {
			oracleText: "Instant and sorcery spells you cast cost {1} less to cast.",
			modifiers: []game.CostModifier{
				{Kind: game.CostModifierSpell, MatchCardType: true, CardType: types.Instant, GenericReduction: 1},
				{Kind: game.CostModifierSpell, MatchCardType: true, CardType: types.Sorcery, GenericReduction: 1},
			},
		},
		"red spells reduction": {
			oracleText: "Red spells you cast cost {1} less to cast.",
			modifiers: []game.CostModifier{
				{Kind: game.CostModifierSpell, MatchColor: true, Color: color.Red, GenericReduction: 1},
			},
		},
		"colorless spells reduction": {
			oracleText: "Colorless spells you cast cost {1} less to cast.",
			modifiers: []game.CostModifier{
				{Kind: game.CostModifierSpell, MatchColor: true, GenericReduction: 1},
			},
		},
		"green spells increase": {
			oracleText: "Green spells you cast cost {2} more to cast.",
			modifiers: []game.CostModifier{
				{Kind: game.CostModifierSpell, MatchColor: true, Color: color.Green, GenericIncrease: 2},
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Reducer",
				Layout:     "normal",
				TypeLine:   "Artifact",
				OracleText: test.oracleText,
			})
			if len(face.StaticAbilities) != 1 {
				t.Fatalf("static abilities = %d, want 1", len(face.StaticAbilities))
			}
			effects := face.StaticAbilities[0].Body.RuleEffects
			if len(effects) != len(test.modifiers) {
				t.Fatalf("rule effects = %#v, want %d", effects, len(test.modifiers))
			}
			for i, effect := range effects {
				if effect.Kind != game.RuleEffectCostModifier {
					t.Fatalf("rule effect %d kind = %v, want cost modifier", i, effect.Kind)
				}
				if effect.AffectedPlayer != game.PlayerYou {
					t.Fatalf("rule effect %d affected player = %v, want you", i, effect.AffectedPlayer)
				}
				got := effect.CostModifier
				want := test.modifiers[i]
				if got.Kind != want.Kind ||
					got.MatchCardType != want.MatchCardType ||
					got.CardType != want.CardType ||
					got.MatchColor != want.MatchColor ||
					got.Color != want.Color ||
					got.GenericReduction != want.GenericReduction ||
					got.GenericIncrease != want.GenericIncrease {
					t.Fatalf("rule effect %d cost modifier = %#v, want %#v", i, got, want)
				}
			}
		})
	}
}

func TestLowerStaticSpellCostModifierRejectsUnsupported(t *testing.T) {
	t.Parallel()
	sources := map[string]string{
		"subtype filter":    "Dragon spells you cast cost {2} less to cast.",
		"leading condition": "During turns other than yours, spells you cast cost {1} less to cast.",
		"colored mana cost": "Black spells you cast cost {B} more to cast.",
		"color and type":    "Red creature spells you cast cost {1} less to cast.",
	}
	for name, source := range sources {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Test Reducer",
				Layout:     "normal",
				TypeLine:   "Artifact",
				OracleText: source,
			})
			if len(diagnostics) == 0 {
				t.Fatalf("source %q lowered without a capability diagnostic", source)
			}
		})
	}
}
