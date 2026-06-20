package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestRenderActivatedAbilityIncludesSourceCostModifiersWithoutAdditionalCosts(t *testing.T) {
	t.Parallel()

	ability := game.ActivatedAbility{
		ManaCost: opt.Val(cost.Mana{cost.O(1), cost.G}),
		CostModifiers: []game.CostModifier{{
			Kind:               game.CostModifierAbility,
			PerObjectReduction: 1,
			CountSelection: &game.Selection{
				RequiredTypes: []types.Card{types.Creature},
				Supertypes:    []types.Super{types.Legendary},
				Controller:    game.ControllerYou,
			},
		}},
		Content: game.Mode{Sequence: []game.Instruction{{
			Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
		}}}.Ability(),
	}

	got, err := (Renderer{}).renderActivatedAbility(newRenderCtx(), &ability)
	if err != nil {
		t.Fatalf("renderActivatedAbility() error = %v", err)
	}
	for _, want := range []string{
		"CostModifiers: []game.CostModifier",
		"PerObjectReduction: 1",
		"CountSelection:",
		"types.Legendary",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered ability missing %q:\n%s", want, got)
		}
	}
}
