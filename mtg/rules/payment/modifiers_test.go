package payment

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
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
