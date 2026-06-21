package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// applyRuleFromActivated returns the lone ApplyRule primitive of a face's first
// activated ability, failing the test if the shape is unexpected.
func applyRuleFromActivated(t *testing.T, face loweredFaceAbilities) game.ApplyRule {
	t.Helper()
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	mode := face.ActivatedAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %#v, want one primitive", mode.Sequence)
	}
	apply, ok := mode.Sequence[0].Primitive.(game.ApplyRule)
	if !ok {
		t.Fatalf("primitive = %T, want game.ApplyRule", mode.Sequence[0].Primitive)
	}
	return apply
}

// TestLowerNextSpellCantBeCountered proves "The next spell you cast this turn
// can't be countered." (Mistrise Village) lowers to a turn-scoped ApplyRule
// carrying a controller-scoped RuleEffectCantBeCountered limited to the next
// spell.
func TestLowerNextSpellCantBeCountered(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Tutor Land",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{T}: The next spell you cast this turn can't be countered.",
	})
	apply := applyRuleFromActivated(t, face)
	if apply.Duration != game.DurationThisTurn {
		t.Fatalf("duration = %v, want DurationThisTurn", apply.Duration)
	}
	if len(apply.RuleEffects) != 1 {
		t.Fatalf("rule effects = %#v, want one", apply.RuleEffects)
	}
	effect := apply.RuleEffects[0]
	if effect.Kind != game.RuleEffectCantBeCountered {
		t.Fatalf("kind = %v, want RuleEffectCantBeCountered", effect.Kind)
	}
	if effect.AffectedController != game.ControllerYou {
		t.Fatalf("affected controller = %v, want ControllerYou", effect.AffectedController)
	}
	if !effect.AppliesToNextSpellOnly {
		t.Fatal("expected AppliesToNextSpellOnly to be set")
	}
}

// TestLowerSpellsCantBeCounteredThisTurn proves the all-spells form "Spells you
// cast this turn can't be countered." lowers without the next-spell limiter, so
// every spell the controller casts this turn is uncounterable.
func TestLowerSpellsCantBeCounteredThisTurn(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Anthem Land",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{T}: Spells you cast this turn can't be countered.",
	})
	apply := applyRuleFromActivated(t, face)
	effect := apply.RuleEffects[0]
	if effect.Kind != game.RuleEffectCantBeCountered {
		t.Fatalf("kind = %v, want RuleEffectCantBeCountered", effect.Kind)
	}
	if effect.AffectedController != game.ControllerYou {
		t.Fatalf("affected controller = %v, want ControllerYou", effect.AffectedController)
	}
	if effect.AppliesToNextSpellOnly {
		t.Fatal("expected AppliesToNextSpellOnly to be unset for the all-spells form")
	}
}

// TestMistriseVillageCompiles proves the full Mistrise Village card compiles
// with no diagnostics now that the next-spell cant-be-countered activated
// ability is supported.
func TestMistriseVillageCompiles(t *testing.T) {
	t.Parallel()
	_, diags, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Mistrise Village",
		Layout:   "normal",
		TypeLine: "Land",
		OracleText: "This land enters tapped unless you control a Mountain or a Forest.\n" +
			"{T}: Add {U}.\n" +
			"{U}, {T}: The next spell you cast this turn can't be countered.",
	}, "t")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("diagnostics = %#v, want none", diags)
	}
}
