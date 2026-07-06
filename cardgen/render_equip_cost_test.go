package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestRenderEquipCostReductionAbility proves an unrestricted Equip ability with
// a conditional generic cost modifier (Crown of Gondor) renders as the
// EquipCostReductionActivatedAbility factory call rather than an opaque struct
// literal whose EquipKeyword the renderer cannot emit.
func TestRenderEquipCostReductionAbility(t *testing.T) {
	t.Parallel()

	modifier := game.CostModifier{
		Kind:               game.CostModifierAbility,
		GenericReduction:   3,
		ReductionCondition: opt.Val(game.Condition{ControllerIsMonarch: true}),
	}
	ability := game.EquipCostReductionActivatedAbility(cost.Mana{cost.O(4)}, modifier)

	got, err := (Renderer{}).renderActivatedAbility(newRenderCtx(), &ability)
	if err != nil {
		t.Fatalf("renderActivatedAbility() error = %v", err)
	}
	for _, want := range []string{
		"game.EquipCostReductionActivatedAbility(cost.Mana{cost.O(4)}",
		"GenericReduction: 3",
		"ReductionCondition:",
		"ControllerIsMonarch: true",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered ability missing %q:\n%s", want, got)
		}
	}
}

// TestRenderEquipRestrictedCostReductionAbility proves a subtype-restricted Equip
// ability with a conditional cost modifier renders through its dedicated factory
// call, so the restricted path is not silently dropped.
func TestRenderEquipRestrictedCostReductionAbility(t *testing.T) {
	t.Parallel()

	modifier := game.CostModifier{
		Kind:               game.CostModifierAbility,
		GenericReduction:   1,
		ReductionCondition: opt.Val(game.Condition{ControllerIsMonarch: true}),
	}
	ability := game.EquipRestrictedCostReductionActivatedAbility(
		cost.Mana{cost.O(2)}, nil, []types.Sub{types.Knight}, modifier,
	)

	got, err := (Renderer{}).renderActivatedAbility(newRenderCtx(), &ability)
	if err != nil {
		t.Fatalf("renderActivatedAbility() error = %v", err)
	}
	for _, want := range []string{
		"game.EquipRestrictedCostReductionActivatedAbility(cost.Mana{cost.O(2)}",
		"types.Knight",
		"GenericReduction: 1",
		"ControllerIsMonarch: true",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered ability missing %q:\n%s", want, got)
		}
	}
}
