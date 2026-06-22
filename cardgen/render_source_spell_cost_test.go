package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestRenderSourceSpellCostModifier(t *testing.T) {
	t.Parallel()
	r := Renderer{}
	tests := []struct {
		name      string
		modifier  game.CostModifier
		wantParts []string
	}{
		{
			name: "per-object reduction with count selection",
			modifier: game.CostModifier{
				Kind:               game.CostModifierSpell,
				PerObjectReduction: 1,
				CountSelection:     &game.Selection{RequiredTypes: []types.Card{types.Creature}},
			},
			wantParts: []string{
				"Kind: game.CostModifierSpell,",
				"PerObjectReduction: 1,",
				"CountSelection:",
				"RequiredTypes:",
				"types.Creature",
			},
		},
		{
			name: "controller-scoped count selection",
			modifier: game.CostModifier{
				Kind:               game.CostModifierSpell,
				PerObjectReduction: 2,
				CountSelection: &game.Selection{
					Controller:    game.ControllerOpponent,
					RequiredTypes: []types.Card{types.Creature},
				},
			},
			wantParts: []string{
				"PerObjectReduction: 2,",
				"CountSelection:",
				"Controller: game.ControllerOpponent",
			},
		},
		{
			name: "creature power-threshold reduction",
			modifier: game.CostModifier{
				Kind:             game.CostModifierSpell,
				MatchCardType:    true,
				CardType:         types.Creature,
				GenericReduction: 2,
				MinPower:         opt.Val(4),
			},
			wantParts: []string{
				"Kind: game.CostModifierSpell,",
				"MatchCardType: true,",
				"CardType: types.Creature,",
				"GenericReduction: 2,",
				"MinPower: opt.Val(4),",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ctx := newRenderCtx()
			got, err := r.renderCostModifier(ctx, test.modifier)
			if err != nil {
				t.Fatalf("renderCostModifier: unexpected error: %v", err)
			}
			for _, part := range test.wantParts {
				if !strings.Contains(got, part) {
					t.Fatalf("rendered %q missing %q", got, part)
				}
			}
		})
	}
}
