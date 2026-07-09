package cardgen

import (
	"reflect"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
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
		"card type list increase": {
			oracleText: "Artifact and enchantment spells you cast cost {1} more to cast.",
			modifiers: []game.CostModifier{
				{Kind: game.CostModifierSpell, CardSelection: game.Selection{RequiredTypes: []types.Card{types.Artifact}}, GenericIncrease: 1},
				{Kind: game.CostModifierSpell, CardSelection: game.Selection{RequiredTypes: []types.Card{types.Enchantment}}, GenericIncrease: 1},
			},
		},
		"card type list reduction": {
			oracleText: "Instant and enchantment spells you cast cost {2} less to cast.",
			modifiers: []game.CostModifier{
				{Kind: game.CostModifierSpell, CardSelection: game.Selection{RequiredTypes: []types.Card{types.Instant}}, GenericReduction: 2},
				{Kind: game.CostModifierSpell, CardSelection: game.Selection{RequiredTypes: []types.Card{types.Enchantment}}, GenericReduction: 2},
			},
		},
		"color type pair increase": {
			oracleText: "Red creature spells and green creature spells you cast cost {1} more to cast.",
			modifiers: []game.CostModifier{
				{Kind: game.CostModifierSpell, CardSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}, ColorsAny: []color.Color{color.Red, color.Green}}, GenericIncrease: 1},
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

func TestLowerStaticSpellCostModifierNonHandZone(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Beyond Reducer",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "Spells you cast from anywhere other than your hand cost {2} less to cast.",
	})
	if len(face.StaticAbilities) != 1 || len(face.StaticAbilities[0].Body.RuleEffects) != 1 {
		t.Fatalf("static abilities = %#v, want one non-hand-zone cost effect", face.StaticAbilities)
	}
	modifier := face.StaticAbilities[0].Body.RuleEffects[0].CostModifier
	if modifier.Kind != game.CostModifierSpell ||
		!modifier.CardSelection.Empty() ||
		modifier.GenericReduction != 2 ||
		modifier.SourceZone.Exists ||
		!slices.Equal(modifier.SourceZones, []zone.Type{zone.Graveyard, zone.Exile, zone.Library, zone.Command}) {
		t.Fatalf("modifier = %#v, want non-hand-scoped {2} reduction", modifier)
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

func TestLowerStaticSpellCostModifierManaValueThreshold(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Krosan Drover",
		Layout:     "normal",
		TypeLine:   "Creature — Elf",
		Power:      new("2"),
		Toughness:  new("2"),
		OracleText: "Creature spells you cast with mana value 6 or greater cost {2} less to cast.",
	})
	if len(face.StaticAbilities) != 1 || len(face.StaticAbilities[0].Body.RuleEffects) != 1 {
		t.Fatalf("static abilities = %#v, want one mana-value-threshold cost effect", face.StaticAbilities)
	}
	modifier := face.StaticAbilities[0].Body.RuleEffects[0].CostModifier
	if modifier.Kind != game.CostModifierSpell ||
		len(modifier.CardSelection.RequiredTypes) != 1 ||
		modifier.CardSelection.RequiredTypes[0] != types.Creature ||
		modifier.GenericReduction != 2 ||
		modifier.CardSelection.Power.Exists ||
		!modifier.CardSelection.ManaValue.Exists ||
		modifier.CardSelection.ManaValue.Val != (compare.Int{Op: compare.GreaterOrEqual, Value: 6}) {
		t.Fatalf("modifier = %#v, want mana-value-6 creature {2} reduction", modifier)
	}
}

func TestLowerStaticSpellPerObjectCostModifier(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		oracleText string
		modifier   game.CostModifier
		duringTurn bool
	}{
		"temur battlecrier": {
			oracleText: "During your turn, spells you cast cost {1} less to cast for each creature you control with power 4 or greater.",
			modifier: game.CostModifier{
				Kind:               game.CostModifierSpell,
				PerObjectReduction: 1,
				CountSelection: &game.Selection{
					RequiredTypes: []types.Card{types.Creature},
					Controller:    game.ControllerYou,
					Power:         opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4}),
				},
			},
			duringTurn: true,
		},
		"hamza guardian of arashin": {
			oracleText: "Creature spells you cast cost {1} less to cast for each creature you control with a +1/+1 counter on it.",
			modifier: game.CostModifier{
				Kind:               game.CostModifierSpell,
				PerObjectReduction: 1,
				CardSelection:      game.Selection{RequiredTypes: []types.Card{types.Creature}},
				CountSelection: &game.Selection{
					RequiredTypes:   []types.Card{types.Creature},
					Controller:      game.ControllerYou,
					MatchCounter:    true,
					RequiredCounter: counter.PlusOnePlusOne,
				},
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Crier",
				Layout:     "normal",
				TypeLine:   "Creature — Human",
				Power:      new("2"),
				Toughness:  new("2"),
				OracleText: tc.oracleText,
			})
			if len(face.StaticAbilities) != 1 || len(face.StaticAbilities[0].Body.RuleEffects) != 1 {
				t.Fatalf("static abilities = %#v, want one per-object cost effect", face.StaticAbilities)
			}
			effect := face.StaticAbilities[0].Body.RuleEffects[0]
			if effect.RestrictedDuringControllerTurn != tc.duringTurn {
				t.Fatalf("RestrictedDuringControllerTurn = %v, want %v", effect.RestrictedDuringControllerTurn, tc.duringTurn)
			}
			if effect.AffectedPlayer != game.PlayerYou {
				t.Fatalf("affected player = %v, want PlayerYou", effect.AffectedPlayer)
			}
			modifier := effect.CostModifier
			if modifier.Kind != tc.modifier.Kind || modifier.PerObjectReduction != tc.modifier.PerObjectReduction {
				t.Fatalf("modifier = %#v, want %#v", modifier, tc.modifier)
			}
			if !reflect.DeepEqual(modifier.CardSelection, tc.modifier.CardSelection) {
				t.Fatalf("card selection = %#v, want %#v", modifier.CardSelection, tc.modifier.CardSelection)
			}
			if modifier.CountSelection == nil || !reflect.DeepEqual(*modifier.CountSelection, *tc.modifier.CountSelection) {
				t.Fatalf("count selection = %#v, want %#v", modifier.CountSelection, tc.modifier.CountSelection)
			}
		})
	}
}

func TestLowerStaticSpellCostModifierRejectsUnsupported(t *testing.T) {
	t.Parallel()
	sources := map[string]string{
		"leading condition":      "During turns other than yours, spells you cast cost {1} less to cast.",
		"colorless mana cost":    "Black spells you cast cost {C} more to cast.",
		"colored cost reduction": "Black spells you cast cost {B} less to cast.",
		"unsupported zone":       "Spells you cast from your library cost {1} less to cast.",
		"unsupported zone list":  "Spells you cast from your graveyard or from exile cost {2} less to cast.",
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

// TestLowerStaticSpellCostModifierCasterAndExcludedType covers the broadened
// shapes: cast-cost modifiers scoped to all players or opponents (not just the
// controller) and modifiers filtered by an excluded card type ("Noncreature
// spells ...").
func TestLowerStaticSpellCostModifierCasterAndExcludedType(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		oracleText     string
		affectedPlayer game.PlayerRelation
		modifier       game.CostModifier
	}{
		"all players bare": {
			oracleText:     "Spells cost {1} more to cast.",
			affectedPlayer: game.PlayerAny,
			modifier:       game.CostModifier{Kind: game.CostModifierSpell, GenericIncrease: 1},
		},
		"opponents bare": {
			oracleText:     "Spells your opponents cast cost {1} more to cast.",
			affectedPlayer: game.PlayerOpponent,
			modifier:       game.CostModifier{Kind: game.CostModifierSpell, GenericIncrease: 1},
		},
		"noncreature all players": {
			oracleText:     "Noncreature spells cost {1} more to cast.",
			affectedPlayer: game.PlayerAny,
			modifier:       game.CostModifier{Kind: game.CostModifierSpell, CardSelection: game.Selection{ExcludedTypes: []types.Card{types.Creature}}, GenericIncrease: 1},
		},
		"noncreature controller": {
			oracleText:     "Noncreature spells you cast cost {1} less to cast.",
			affectedPlayer: game.PlayerYou,
			modifier:       game.CostModifier{Kind: game.CostModifierSpell, CardSelection: game.Selection{ExcludedTypes: []types.Card{types.Creature}}, GenericReduction: 1},
		},
		"nonartifact opponents": {
			oracleText:     "Nonartifact spells your opponents cast cost {2} more to cast.",
			affectedPlayer: game.PlayerOpponent,
			modifier:       game.CostModifier{Kind: game.CostModifierSpell, CardSelection: game.Selection{ExcludedTypes: []types.Card{types.Artifact}}, GenericIncrease: 2},
		},
		"colored controller": {
			oracleText:     "Black spells you cast cost {B} more to cast.",
			affectedPlayer: game.PlayerYou,
			modifier:       game.CostModifier{Kind: game.CostModifierSpell, CardSelection: game.Selection{ColorsAny: []color.Color{color.Black}}, ColoredIncrease: []mana.Color{mana.B}},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Taxer",
				Layout:     "normal",
				TypeLine:   "Artifact",
				OracleText: test.oracleText,
			})
			if len(face.StaticAbilities) != 1 {
				t.Fatalf("static abilities = %d, want 1", len(face.StaticAbilities))
			}
			effects := face.StaticAbilities[0].Body.RuleEffects
			if len(effects) != 1 {
				t.Fatalf("rule effects = %#v, want 1", effects)
			}
			effect := effects[0]
			if effect.Kind != game.RuleEffectCostModifier {
				t.Fatalf("rule effect kind = %v, want cost modifier", effect.Kind)
			}
			if effect.AffectedPlayer != test.affectedPlayer {
				t.Fatalf("affected player = %v, want %v", effect.AffectedPlayer, test.affectedPlayer)
			}
			got := effect.CostModifier
			if got.Kind != test.modifier.Kind ||
				!reflect.DeepEqual(got.CardSelection, test.modifier.CardSelection) ||
				got.GenericReduction != test.modifier.GenericReduction ||
				got.GenericIncrease != test.modifier.GenericIncrease ||
				!reflect.DeepEqual(got.ColoredIncrease, test.modifier.ColoredIncrease) {
				t.Fatalf("cost modifier = %#v, want %#v", got, test.modifier)
			}
		})
	}
}

func TestLowerStaticSpellCardTypeListOpponentCostModifier(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		oracleText     string
		affectedPlayer game.PlayerRelation
		modifiers      []game.CostModifier
	}{
		"type pair opponents": {
			oracleText:     "Artifact and enchantment spells your opponents cast cost {2} more to cast.",
			affectedPlayer: game.PlayerOpponent,
			modifiers: []game.CostModifier{
				{Kind: game.CostModifierSpell, CardSelection: game.Selection{RequiredTypes: []types.Card{types.Artifact}}, GenericIncrease: 2},
				{Kind: game.CostModifierSpell, CardSelection: game.Selection{RequiredTypes: []types.Card{types.Enchantment}}, GenericIncrease: 2},
			},
		},
		"type triple opponents": {
			oracleText:     "Artifact, instant, and sorcery spells your opponents cast cost {1} more to cast.",
			affectedPlayer: game.PlayerOpponent,
			modifiers: []game.CostModifier{
				{Kind: game.CostModifierSpell, CardSelection: game.Selection{RequiredTypes: []types.Card{types.Artifact}}, GenericIncrease: 1},
				{Kind: game.CostModifierSpell, CardSelection: game.Selection{RequiredTypes: []types.Card{types.Instant}}, GenericIncrease: 1},
				{Kind: game.CostModifierSpell, CardSelection: game.Selection{RequiredTypes: []types.Card{types.Sorcery}}, GenericIncrease: 1},
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Taxer",
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
				if effect.AffectedPlayer != test.affectedPlayer {
					t.Fatalf("rule effect %d affected player = %v, want %v", i, effect.AffectedPlayer, test.affectedPlayer)
				}
				got := effect.CostModifier
				want := test.modifiers[i]
				if got.Kind != want.Kind ||
					!reflect.DeepEqual(got.CardSelection, want.CardSelection) ||
					got.GenericIncrease != want.GenericIncrease {
					t.Fatalf("rule effect %d cost modifier = %#v, want %#v", i, got, want)
				}
			}
		})
	}
}
