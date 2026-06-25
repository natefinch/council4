package cardgen

import (
	"reflect"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
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
				{Kind: game.CostModifierSpell, CardSelection: game.Selection{RequiredTypes: []types.Card{types.Artifact}}, GenericReduction: 1},
			},
		},
		"creature spells increase": {
			oracleText: "Creature spells you cast cost {1} more to cast.",
			modifiers: []game.CostModifier{
				{Kind: game.CostModifierSpell, CardSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}}, GenericIncrease: 1},
			},
		},
		"instant and sorcery reduction": {
			oracleText: "Instant and sorcery spells you cast cost {1} less to cast.",
			modifiers: []game.CostModifier{
				{Kind: game.CostModifierSpell, CardSelection: game.Selection{RequiredTypes: []types.Card{types.Instant}}, GenericReduction: 1},
				{Kind: game.CostModifierSpell, CardSelection: game.Selection{RequiredTypes: []types.Card{types.Sorcery}}, GenericReduction: 1},
			},
		},
		"red spells reduction": {
			oracleText: "Red spells you cast cost {1} less to cast.",
			modifiers: []game.CostModifier{
				{Kind: game.CostModifierSpell, CardSelection: game.Selection{ColorsAny: []color.Color{color.Red}}, GenericReduction: 1},
			},
		},
		"colorless spells reduction": {
			oracleText: "Colorless spells you cast cost {1} less to cast.",
			modifiers: []game.CostModifier{
				{Kind: game.CostModifierSpell, CardSelection: game.Selection{Colorless: true}, GenericReduction: 1},
			},
		},
		"green spells increase": {
			oracleText: "Green spells you cast cost {2} more to cast.",
			modifiers: []game.CostModifier{
				{Kind: game.CostModifierSpell, CardSelection: game.Selection{ColorsAny: []color.Color{color.Green}}, GenericIncrease: 2},
			},
		},
		"black creature spells reduction": {
			oracleText: "Black creature spells you cast cost {1} less to cast.",
			modifiers: []game.CostModifier{
				{Kind: game.CostModifierSpell, CardSelection: game.Selection{ColorsAny: []color.Color{color.Black}, RequiredTypes: []types.Card{types.Creature}}, GenericReduction: 1},
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
					!reflect.DeepEqual(got.CardSelection, want.CardSelection) ||
					got.GenericReduction != want.GenericReduction ||
					got.GenericIncrease != want.GenericIncrease {
					t.Fatalf("rule effect %d cost modifier = %#v, want %#v", i, got, want)
				}
			}
		})
	}
}

func TestLowerStaticSpellColorDisjunctionCostModifier(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		oracleText string
		colors     []color.Color
	}{
		"each spell red or green": {
			oracleText: "Each spell you cast that's red or green costs {1} less to cast.",
			colors:     []color.Color{color.Red, color.Green},
		},
		"color pair and": {
			oracleText: "Blue spells and red spells you cast cost {1} less to cast.",
			colors:     []color.Color{color.Blue, color.Red},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Disjunction Reducer",
				Layout:     "normal",
				TypeLine:   "Artifact",
				OracleText: test.oracleText,
			})
			if len(face.StaticAbilities) != 1 || len(face.StaticAbilities[0].Body.RuleEffects) != 1 {
				t.Fatalf("static abilities = %#v, want one color-disjunction cost effect", face.StaticAbilities)
			}
			modifier := face.StaticAbilities[0].Body.RuleEffects[0].CostModifier
			if modifier.Kind != game.CostModifierSpell ||
				modifier.CardSelection.Colorless ||
				len(modifier.CardSelection.RequiredTypes) != 0 ||
				modifier.GenericReduction != 1 ||
				!slices.Equal(modifier.CardSelection.ColorsAny, test.colors) {
				t.Fatalf("modifier = %#v, want color disjunction %v", modifier, test.colors)
			}
		})
	}
}

func TestLowerStaticChosenTypeSpellCostModifier(t *testing.T) {
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Chosen Type Reducer",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "As Chosen Type Reducer enters, choose a creature type.\nCreature spells you cast of the chosen type cost {1} less to cast.",
	})
	if len(face.StaticAbilities) != 1 || len(face.StaticAbilities[0].Body.RuleEffects) != 1 {
		t.Fatalf("static abilities = %#v, want one chosen-type cost effect", face.StaticAbilities)
	}
	modifier := face.StaticAbilities[0].Body.RuleEffects[0].CostModifier
	if modifier.Kind != game.CostModifierSpell ||
		len(modifier.CardSelection.RequiredTypes) != 1 ||
		modifier.CardSelection.RequiredTypes[0] != types.Creature ||
		!modifier.ChosenSubtypeFromEntryChoice ||
		modifier.GenericReduction != 1 {
		t.Fatalf("modifier = %#v, want chosen creature type reduction", modifier)
	}
}

func TestLowerStaticSpellSubtypeCostModifier(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		oracleText string
		subtypes   []types.Sub
	}{
		"single subtype": {
			oracleText: "Aura spells you cast cost {1} less to cast.",
			subtypes:   []types.Sub{types.Aura},
		},
		"multi subtype": {
			oracleText: "Aura and Equipment spells you cast cost {1} less to cast.",
			subtypes:   []types.Sub{types.Aura, types.Equipment},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Subtype Reducer",
				Layout:     "normal",
				TypeLine:   "Artifact",
				OracleText: test.oracleText,
			})
			if len(face.StaticAbilities) != 1 || len(face.StaticAbilities[0].Body.RuleEffects) != 1 {
				t.Fatalf("static abilities = %#v, want one subtype cost effect", face.StaticAbilities)
			}
			modifier := face.StaticAbilities[0].Body.RuleEffects[0].CostModifier
			if modifier.Kind != game.CostModifierSpell ||
				len(modifier.CardSelection.ColorsAny) != 0 ||
				len(modifier.CardSelection.RequiredTypes) != 0 ||
				modifier.GenericReduction != 1 ||
				!slices.Equal(modifier.CardSelection.SubtypesAny, test.subtypes) {
				t.Fatalf("modifier = %#v, want subtype filter %v", modifier, test.subtypes)
			}
		})
	}
}

func TestLowerStaticSpellCostModifierGraveyardZone(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Graveyard Reducer",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "Spells you cast from your graveyard cost {1} less to cast.",
	})
	if len(face.StaticAbilities) != 1 || len(face.StaticAbilities[0].Body.RuleEffects) != 1 {
		t.Fatalf("static abilities = %#v, want one graveyard-zone cost effect", face.StaticAbilities)
	}
	modifier := face.StaticAbilities[0].Body.RuleEffects[0].CostModifier
	if modifier.Kind != game.CostModifierSpell ||
		!modifier.CardSelection.Empty() ||
		modifier.GenericReduction != 1 ||
		!modifier.SourceZone.Exists ||
		modifier.SourceZone.Val != zone.Graveyard {
		t.Fatalf("modifier = %#v, want graveyard-scoped {1} reduction", modifier)
	}
}

func TestLowerStaticSpellCostModifierPowerThreshold(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Goreclaw, Terror of Qal Sisma",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Bear",
		Power:      new("4"),
		Toughness:  new("4"),
		OracleText: "Creature spells you cast with power 4 or greater cost {2} less to cast.",
	})
	if len(face.StaticAbilities) != 1 || len(face.StaticAbilities[0].Body.RuleEffects) != 1 {
		t.Fatalf("static abilities = %#v, want one power-threshold cost effect", face.StaticAbilities)
	}
	modifier := face.StaticAbilities[0].Body.RuleEffects[0].CostModifier
	if modifier.Kind != game.CostModifierSpell ||
		len(modifier.CardSelection.RequiredTypes) != 1 ||
		modifier.CardSelection.RequiredTypes[0] != types.Creature ||
		modifier.GenericReduction != 2 ||
		!modifier.CardSelection.Power.Exists ||
		modifier.CardSelection.Power.Val != (compare.Int{Op: compare.GreaterOrEqual, Value: 4}) {
		t.Fatalf("modifier = %#v, want power-4 creature {2} reduction", modifier)
	}
}

func TestLowerStaticSpellCostModifierRejectsUnsupported(t *testing.T) {
	t.Parallel()
	sources := map[string]string{
		"leading condition": "During turns other than yours, spells you cast cost {1} less to cast.",
		"colored mana cost": "Black spells you cast cost {B} more to cast.",
		"unsupported zone":  "Spells you cast from anywhere other than your hand cost {2} less to cast.",
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
