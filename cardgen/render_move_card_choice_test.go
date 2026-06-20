package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestRenderChosenHandToLibraryMove(t *testing.T) {
	t.Parallel()
	rendered, err := (Renderer{}).renderPrimitive(newRenderCtx(), game.MoveCard{
		Player:      game.ControllerReference(),
		Amount:      game.Fixed(2),
		FromZone:    zone.Hand,
		Destination: zone.Library,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"game.MoveCard",
		"Player: game.ControllerReference()",
		"Amount: game.Fixed(2)",
		"zone.Hand",
		"zone.Library",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered move missing %q:\n%s", want, rendered)
		}
	}
}
