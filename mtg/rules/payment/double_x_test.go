package payment

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
)

func TestDoubleXNonManaPaymentCost(t *testing.T) {
	manaCost := cost.Mana{cost.X, cost.X, cost.R}
	for _, test := range []struct {
		x           int
		wantGeneric int
	}{
		{x: 0, wantGeneric: 0},
		{x: 3, wantGeneric: 6},
	} {
		c, ok := newNonManaPaymentCost(fakePaymentState{}, game.Player1, &manaCost, test.x, nil)
		if !ok || c.genericDue() != test.wantGeneric {
			t.Fatalf("X=%d generic due = %d ok %v, want %d", test.x, c.genericDue(), ok, test.wantGeneric)
		}
		remaining, _ := c.remaining()
		want := cost.Mana{cost.R}
		if test.wantGeneric > 0 {
			want = cost.Mana{cost.O(test.wantGeneric), cost.R}
		}
		if !slices.Equal(*remaining, want) {
			t.Fatalf("X=%d remaining = %v, want %v", test.x, *remaining, want)
		}
	}
}
