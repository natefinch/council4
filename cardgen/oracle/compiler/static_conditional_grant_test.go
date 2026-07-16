package compiler

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
)

// TestCompileConditionalStaticPermanentManaAbilityGrant proves the static
// permanent-ability grant recognizer accepts a leading duration condition and
// attaches it alongside the grant, so The World Tree's "As long as you control
// six or more lands, lands you control have '{T}: Add one mana of any color.'"
// compiles to a controlled-land mana-ability grant gated on controlling six or
// more lands.
func TestCompileConditionalStaticPermanentManaAbilityGrant(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		`As long as you control six or more lands, lands you control have "{T}: Add one mana of any color."`,
		pipelineContext{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Static == nil || len(ability.Static.Declarations) != 1 {
		t.Fatalf("static semantics = %#v, want one declaration", ability.Static)
	}
	declaration := ability.Static.Declarations[0]
	if declaration.Kind != StaticDeclarationContinuous ||
		declaration.Group.Domain != StaticGroupSourceControllerPermanents ||
		!slices.Equal(declaration.Group.Selection.RequiredTypes, []types.Card{types.Land}) ||
		declaration.Continuous == nil ||
		declaration.Continuous.Layer != StaticLayerAbility ||
		declaration.Continuous.Operation != StaticContinuousGrantManaAbility {
		t.Fatalf("declaration = %#v, want controlled-land mana-ability grant", declaration)
	}
	granted := declaration.Continuous.GrantedMana
	if granted == nil || !granted.TapCost || granted.Amount != 1 || !granted.AnyColor {
		t.Fatalf("granted mana ability = %#v, want tap for one mana of any color", granted)
	}
	condition := declaration.Condition
	if condition == nil ||
		condition.Predicate != ConditionPredicateControllerControls ||
		condition.Threshold != 6 ||
		!slices.Equal(condition.Selection.RequiredTypes, []types.Card{types.Land}) {
		t.Fatalf("condition = %#v, want controls six or more lands", condition)
	}
}
