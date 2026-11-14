package payment

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
)

func TestAppendLifeCostModifiers(t *testing.T) {
	existing := []cost.Additional{{Kind: cost.AdditionalSacrifice}}

	tests := []struct {
		name      string
		modifiers []game.CostModifier
		wantLife  int
		wantAdded bool
	}{
		{
			name:      "no modifiers leaves costs unchanged",
			modifiers: nil,
			wantAdded: false,
		},
		{
			name:      "generic increase does not add life",
			modifiers: []game.CostModifier{{Kind: game.CostModifierSpell, GenericIncrease: 2}},
			wantAdded: false,
		},
		{
			name:      "single life increase adds one pay-life cost",
			modifiers: []game.CostModifier{{Kind: game.CostModifierSpell, LifeIncrease: 3}},
			wantLife:  3,
			wantAdded: true,
		},
		{
			name: "life increases sum into one pay-life cost",
			modifiers: []game.CostModifier{
				{Kind: game.CostModifierSpell, LifeIncrease: 3},
				{Kind: game.CostModifierSpell, LifeIncrease: 2},
			},
			wantLife:  5,
			wantAdded: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := appendLifeCostModifiers(existing, test.modifiers)
			if !test.wantAdded {
				if len(got) != len(existing) {
					t.Fatalf("appendLifeCostModifiers added a cost: got %d, want %d", len(got), len(existing))
				}
				return
			}
			if len(got) != len(existing)+1 {
				t.Fatalf("appendLifeCostModifiers appended %d costs, want 1", len(got)-len(existing))
			}
			added := got[len(got)-1]
			if added.Kind != cost.AdditionalPayLife {
				t.Fatalf("appended cost kind = %v, want AdditionalPayLife", added.Kind)
			}
			if added.Amount != test.wantLife {
				t.Fatalf("appended pay-life amount = %d, want %d", added.Amount, test.wantLife)
			}
		})
	}
}

// TestApplyManaCostModifiersColoredReduction pins how a colored reduction
// lowers a total cost (CR 601.2f, 118.7b, 118.7c): it removes a symbol of its
// color, including one a colored increase added, and otherwise lowers generic.
func TestApplyManaCostModifiersColoredReduction(t *testing.T) {
	blue := func(generic int, colors ...mana.Color) game.CostModifier {
		return game.CostModifier{Kind: game.CostModifierSpell, GenericReduction: generic, ColoredReduction: colors}
	}
	tests := []struct {
		name      string
		manaCost  cost.Mana
		modifiers []game.CostModifier
		want      string
	}{
		{name: "removes a matching colored symbol and generic", manaCost: cost.Mana{cost.O(2), cost.U, cost.U}, modifiers: []game.CostModifier{blue(1, mana.U)}, want: "{1}{U}"},
		{name: "absent color reduces generic (118.7b)", manaCost: cost.Mana{cost.O(2), cost.R}, modifiers: []game.CostModifier{blue(0, mana.U)}, want: "{1}{R}"},
		{name: "excess colored reduces generic (118.7c)", manaCost: cost.Mana{cost.O(1), cost.U}, modifiers: []game.CostModifier{blue(0, mana.U, mana.U)}, want: ""},
		{name: "never below zero", manaCost: cost.Mana{cost.U}, modifiers: []game.CostModifier{blue(1, mana.U, mana.U)}, want: ""},
		{name: "hybrid is not a blue requirement", manaCost: cost.Mana{cost.O(1), cost.HybridMana(mana.U, mana.R)}, modifiers: []game.CostModifier{blue(0, mana.U)}, want: "{U/R}"},
		{
			name:     "reduction applies after colored and generic increases",
			manaCost: cost.Mana{cost.O(2), cost.U, cost.U},
			modifiers: []game.CostModifier{
				{Kind: game.CostModifierSpell, GenericIncrease: 1, ColoredIncrease: []mana.Color{mana.U}},
				blue(1, mana.U),
			},
			want: "{2}{U}{U}",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manaCost := append(cost.Mana(nil), test.manaCost...)
			got := applyManaCostModifiers(&manaCost, test.modifiers)
			if got.String() != test.want {
				t.Fatalf("total cost = %q, want %q", got.String(), test.want)
			}
			if manaCost.String() != test.manaCost.String() {
				t.Fatalf("printed cost mutated to %q", manaCost.String())
			}
		})
	}
}
