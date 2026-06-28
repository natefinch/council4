package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestDynamicAmountHalfLibraryDivisorRounds verifies that a CountCardsInZone
// amount carrying a Divisor halves the counted library, rounding down by default
// and up when RoundUp is set. This backs the "mills half their library, rounded
// up/down" family (Traumatize, Fleet Swallower); CR 107.4.
func TestDynamicAmountHalfLibraryDivisorRounds(t *testing.T) {
	for _, tc := range []struct {
		name      string
		librarySz int
		roundUp   bool
		want      int
	}{
		{"seven rounded down", 7, false, 3},
		{"seven rounded up", 7, true, 4},
		{"six rounded down", 6, false, 3},
		{"six rounded up", 6, true, 3},
		{"one rounded down", 1, false, 0},
		{"one rounded up", 1, true, 1},
		{"empty", 0, false, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			for i := 0; i < tc.librarySz; i++ {
				addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Card"}})
			}
			obj := &game.StackObject{Controller: game.Player1}
			player := game.ControllerReference()
			got := dynamicAmountValue(g, obj, game.Player1, game.DynamicAmount{
				Kind:      game.DynamicAmountCountCardsInZone,
				Player:    &player,
				CardZone:  zone.Library,
				Selection: &game.Selection{},
				Divisor:   2,
				RoundUp:   tc.roundUp,
			})
			if got != tc.want {
				t.Fatalf("half of %d (roundUp=%v) = %d, want %d", tc.librarySz, tc.roundUp, got, tc.want)
			}
		})
	}
}
