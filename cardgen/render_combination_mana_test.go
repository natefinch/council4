package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
)

// TestRenderCombinationMana verifies that an AddMana carrying a CombinationColors
// set renders the color slice literal and requests the mana import, for both a
// fixed color subset and an all-five color set with a dynamic amount.
func TestRenderCombinationMana(t *testing.T) {
	t.Parallel()

	ctx := newRenderCtx()
	fixed, err := (Renderer{}).renderPrimitive(ctx, game.AddMana{
		Amount:            game.Fixed(3),
		CombinationColors: []mana.Color{mana.R, mana.G},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, wanted := range []string{
		"game.AddMana",
		"Amount: game.Fixed(3),",
		"CombinationColors: []mana.Color{mana.R, mana.G},",
	} {
		if !strings.Contains(fixed, wanted) {
			t.Fatalf("fixed combination render missing %q:\n%s", wanted, fixed)
		}
	}
	if _, ok := ctx.imports[importMana]; !ok {
		t.Fatal("combination render did not request the mana import")
	}

	dynamic, err := (Renderer{}).renderPrimitive(newRenderCtx(), game.AddMana{
		Amount: game.Dynamic(game.DynamicAmount{
			Kind: game.DynamicAmountControllerBasicLandTypeCount,
		}),
		CombinationColors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(dynamic, "CombinationColors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},") {
		t.Fatalf("dynamic combination render missing all-five colors:\n%s", dynamic)
	}
	if !strings.Contains(dynamic, "game.DynamicAmountControllerBasicLandTypeCount") {
		t.Fatalf("dynamic combination render missing domain amount:\n%s", dynamic)
	}
}
