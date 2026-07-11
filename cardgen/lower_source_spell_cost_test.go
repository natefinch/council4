package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
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

// TestLowerSourceSpellCostReductionDynamic proves the dynamic form ("costs {X}
// less to cast, where X is the greatest power among creatures you control")
// lowers to a DynamicReduction cost modifier rather than a per-object reduction.
func TestLowerSourceSpellCostReductionDynamic(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Reducer",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "This spell costs {X} less to cast, where X is the greatest power among creatures you control.",
	})
	modifier := sourceSpellCostReductionModifier(t, face)
	if modifier.Kind != game.CostModifierSpell {
		t.Fatalf("modifier kind = %v, want spell", modifier.Kind)
	}
	if modifier.PerObjectReduction != 0 {
		t.Fatalf("per-object reduction = %d, want 0 for the dynamic form", modifier.PerObjectReduction)
	}
	if modifier.DynamicReduction == nil {
		t.Fatal("dynamic reduction missing; the {X} form must carry a DynamicReduction amount")
	}
	if got := modifier.DynamicReduction.Kind; got != game.DynamicAmountGreatestPowerInGroup {
		t.Fatalf("dynamic reduction kind = %v, want greatest power in group", got)
	}
}

// TestLowerSourceSpellCostReductionTotalManaValue proves the total-mana-value
// dynamic form ("costs {X} less to cast, where X is the total mana value of
// artifacts you control" — Metalwork Colossus) lowers to a DynamicReduction
// cost modifier carrying the total-mana-value-in-group amount.
func TestLowerSourceSpellCostReductionTotalManaValue(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Reducer",
		Layout:     "normal",
		TypeLine:   "Artifact Creature",
		OracleText: "This spell costs {X} less to cast, where X is the total mana value of artifacts you control.",
		ManaCost:   "{11}",
	})
	modifier := sourceSpellCostReductionModifier(t, face)
	if modifier.Kind != game.CostModifierSpell {
		t.Fatalf("modifier kind = %v, want spell", modifier.Kind)
	}
	if modifier.PerObjectReduction != 0 {
		t.Fatalf("per-object reduction = %d, want 0 for the dynamic form", modifier.PerObjectReduction)
	}
	if modifier.DynamicReduction == nil {
		t.Fatal("dynamic reduction missing; the {X} form must carry a DynamicReduction amount")
	}
	if got := modifier.DynamicReduction.Kind; got != game.DynamicAmountTotalManaValueInGroup {
		t.Fatalf("dynamic reduction kind = %v, want total mana value in group", got)
	}
}

// TestLowerSourceSpellCostReductionCardZone proves the card-zone count forms
// ("costs {N} less to cast for each <card> in your graveyard/hand") lower to a
// per-object reduction whose CountZone scopes the count to the caster's own
// graveyard or hand rather than to the battlefield.
func TestLowerSourceSpellCostReductionCardZone(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		oracleText string
		perObject  int
		zone       zone.Type
		required   types.Card
	}{
		"land card in your graveyard": {
			oracleText: "This spell costs {1} less to cast for each land card in your graveyard.",
			perObject:  1,
			zone:       zone.Graveyard,
			required:   types.Land,
		},
		"creature card in your graveyard": {
			oracleText: "This spell costs {2} less to cast for each creature card in your graveyard.",
			perObject:  2,
			zone:       zone.Graveyard,
			required:   types.Creature,
		},
		"artifact card in your hand": {
			oracleText: "This spell costs {1} less to cast for each artifact card in your hand.",
			perObject:  1,
			zone:       zone.Hand,
			required:   types.Artifact,
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
			if !modifier.CountZone.Exists || modifier.CountZone.Val != test.zone {
				t.Fatalf("count zone = %#v, want %v", modifier.CountZone, test.zone)
			}
			if modifier.DynamicReduction != nil {
				t.Fatal("card-zone reduction must use a per-object reduction, not a dynamic reduction")
			}
			if modifier.CountSelection == nil || len(modifier.CountSelection.RequiredTypes) != 1 ||
				modifier.CountSelection.RequiredTypes[0] != test.required {
				t.Fatalf("count selection required types = %#v, want [%v]", modifier.CountSelection, test.required)
			}
		})
	}
}

func TestLowerSourceSpellCostReductionRejectsUnsupported(t *testing.T) {
	t.Parallel()
	sources := map[string]string{
		"variable amount":        "This spell costs {X} less to cast for each creature on the battlefield.",
		"opponent count":         "This spell costs {1} less to cast for each opponent you have.",
		"increase wording":       "This spell costs {1} more to cast for each creature on the battlefield.",
		"sacrifice subject":      "This spell costs {2} less to cast for each creature that died this turn.",
		"permanent in graveyard": "This spell costs {1} less to cast for each permanent card in your graveyard.",
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

// TestLowerSourceSpellCostReductionConditional proves the conditional form
// ("This spell costs {N} less to cast if you control a Wizard") lowers to a flat
// GenericReduction cost modifier gated by a ReductionCondition, rather than a
// per-object or dynamic reduction.
func TestLowerSourceSpellCostReductionConditional(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Reducer",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "This spell costs {2} less to cast if you control a Wizard.",
		ManaCost:   "{2}{R}",
	})
	modifier := sourceSpellCostReductionModifier(t, face)
	if modifier.Kind != game.CostModifierSpell {
		t.Fatalf("modifier kind = %v, want spell", modifier.Kind)
	}

	if modifier.GenericReduction != 2 {
		t.Fatalf("generic reduction = %d, want 2", modifier.GenericReduction)
	}
	if modifier.PerObjectReduction != 0 {
		t.Fatalf("per-object reduction = %d, want 0 for the conditional form", modifier.PerObjectReduction)
	}
	if modifier.DynamicReduction != nil {
		t.Fatal("dynamic reduction must be nil for the conditional form")
	}
	if !modifier.ReductionCondition.Exists {
		t.Fatal("reduction condition missing; the conditional form must carry a ReductionCondition")
	}
	cond := modifier.ReductionCondition.Val
	if !cond.ControlsMatching.Exists {
		t.Fatal("reduction condition must gate on a ControlsMatching board-state predicate")
	}
}

func TestLowerSourceSpellCostReductionTargetsTappedCreature(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Seized from Slumber",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "This spell costs {3} less to cast if it targets a tapped creature.\nDestroy target creature.",
		ManaCost:   "{4}{W}",
	})
	modifier := sourceSpellCostReductionModifier(t, face)
	if modifier.GenericReduction != 3 || !modifier.TargetsTappedCreature ||
		modifier.ReductionCondition.Exists {
		t.Fatalf("modifier = %#v, want tapped-creature target reduction", modifier)
	}
}
