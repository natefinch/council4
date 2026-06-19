package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// sourceSpellCostReductionModifier extracts the single AffectedSource spell cost
// modifier a face carries, failing the test if the face does not hold exactly one.
func sourceSpellCostReductionModifier(t *testing.T, face loweredFaceAbilities) game.CostModifier {
	t.Helper()
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want 1", len(face.StaticAbilities))
	}
	effects := face.StaticAbilities[0].Body.RuleEffects
	if len(effects) != 1 {
		t.Fatalf("rule effects = %d, want 1", len(effects))
	}
	effect := effects[0]
	if effect.Kind != game.RuleEffectCostModifier {
		t.Fatalf("rule effect kind = %v, want cost modifier", effect.Kind)
	}
	if !effect.AffectedSource {
		t.Fatal("rule effect is not AffectedSource; the reduction must apply only to the source spell")
	}
	return effect.CostModifier
}

func TestLowerSourceSpellCostReduction(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		oracleText string
		perObject  int
		selection  game.Selection
	}{
		"each creature on the battlefield": {
			oracleText: "This spell costs {1} less to cast for each creature on the battlefield.",
			perObject:  1,
			selection:  game.Selection{RequiredTypes: []types.Card{types.Creature}},
		},
		"each creature you control": {
			oracleText: "This spell costs {2} less to cast for each creature you control.",
			perObject:  2,
			selection:  game.Selection{Controller: game.ControllerYou, RequiredTypes: []types.Card{types.Creature}},
		},
		"each creature your opponents control": {
			oracleText: "This spell costs {1} less to cast for each creature your opponents control.",
			perObject:  1,
			selection:  game.Selection{Controller: game.ControllerOpponent, RequiredTypes: []types.Card{types.Creature}},
		},
		"each artifact you control": {
			oracleText: "This spell costs {1} less to cast for each artifact you control.",
			perObject:  1,
			selection:  game.Selection{Controller: game.ControllerYou, RequiredTypes: []types.Card{types.Artifact}},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Reducer",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: test.oracleText,
			})
			modifier := sourceSpellCostReductionModifier(t, face)
			if modifier.Kind != game.CostModifierSpell {
				t.Fatalf("modifier kind = %v, want spell", modifier.Kind)
			}
			if modifier.PerObjectReduction != test.perObject {
				t.Fatalf("per-object reduction = %d, want %d", modifier.PerObjectReduction, test.perObject)
			}
			if modifier.CountSelection.Controller != test.selection.Controller {
				t.Fatalf("count selection controller = %v, want %v", modifier.CountSelection.Controller, test.selection.Controller)
			}
			if len(modifier.CountSelection.RequiredTypes) != len(test.selection.RequiredTypes) {
				t.Fatalf("count selection required types = %#v, want %#v", modifier.CountSelection.RequiredTypes, test.selection.RequiredTypes)
			}
			for i, want := range test.selection.RequiredTypes {
				if modifier.CountSelection.RequiredTypes[i] != want {
					t.Fatalf("count selection required type %d = %v, want %v", i, modifier.CountSelection.RequiredTypes[i], want)
				}
			}
		})
	}
}

func TestLowerSourceSpellCostReductionRejectsUnsupported(t *testing.T) {
	t.Parallel()
	sources := map[string]string{
		"graveyard count":   "This spell costs {1} less to cast for each creature card in your graveyard.",
		"variable amount":   "This spell costs {X} less to cast for each creature on the battlefield.",
		"opponent count":    "This spell costs {1} less to cast for each opponent you have.",
		"increase wording":  "This spell costs {1} more to cast for each creature on the battlefield.",
		"sacrifice subject": "This spell costs {2} less to cast for each creature that died this turn.",
	}
	for name, source := range sources {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
				Name:       "Test Reducer",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: source,
			})
			for _, static := range face.StaticAbilities {
				for _, effect := range static.Body.RuleEffects {
					if effect.CostModifier.PerObjectReduction != 0 {
						t.Fatalf("source %q produced a partial per-object reduction", source)
					}
				}
			}
		})
	}
}

// TestLowerBlasphemousActEndToEnd proves the whole card lowers: the cost-reduction
// ability becomes a source-scoped cost modifier while the already-supported
// 13-damage body is retained.
func TestLowerBlasphemousActEndToEnd(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Blasphemous Act",
		Layout:     "normal",
		ManaCost:   "{8}{R}",
		TypeLine:   "Sorcery",
		OracleText: "This spell costs {1} less to cast for each creature on the battlefield.\nBlasphemous Act deals 13 damage to each creature.",
		Colors:     []string{"R"},
	})
	modifier := sourceSpellCostReductionModifier(t, face)
	if modifier.PerObjectReduction != 1 {
		t.Fatalf("per-object reduction = %d, want 1", modifier.PerObjectReduction)
	}
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability missing; the 13-damage body must be retained")
	}
	sequence := face.SpellAbility.Val.Modes[0].Sequence
	if len(sequence) != 1 {
		t.Fatalf("spell ability sequence = %d instructions, want 1", len(sequence))
	}
	damage, ok := sequence[0].Primitive.(game.Damage)
	if !ok {
		t.Fatalf("spell ability primitive = %T, want game.Damage", sequence[0].Primitive)
	}
	if damage.Amount.IsDynamic() || damage.Amount.Value() != 13 {
		t.Fatalf("damage amount = %#v, want fixed 13", damage.Amount)
	}
}
