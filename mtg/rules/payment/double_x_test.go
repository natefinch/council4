package payment

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
)

func TestDoubleXCostRequirements(t *testing.T) {
	manaCost := cost.Mana{cost.X, cost.X, cost.R}
	for _, test := range []struct {
		x           int
		wantGeneric int
	}{
		{x: 0, wantGeneric: 0},
		{x: 3, wantGeneric: 6},
	} {
		colored, generic, ok := costRequirements(&manaCost, test.x)
		if !ok || generic != test.wantGeneric || colored[mana.R] != 1 {
			t.Fatalf("X=%d requirements = colored %#v generic %d ok %v", test.x, colored, generic, ok)
		}
	}
}
