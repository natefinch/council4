package payment

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/cost"
)

func TestSelectionForAdditionalCostRequiresNontoken(t *testing.T) {
	t.Parallel()
	selection, ok := SelectionForAdditionalCost(cost.Additional{RequireNonToken: true})
	if !ok || !selection.NonToken || selection.TokenOnly {
		t.Fatalf("selection = %#v (ok=%v), want nontoken", selection, ok)
	}
}
