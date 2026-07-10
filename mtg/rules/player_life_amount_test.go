package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestPlayerLifeHalvedAmount covers the "loses half their life, rounded up/down"
// amount (Quietus Spike, Virtus the Veiled): the value is the named player's
// current life halved as the effect resolves, rounded per RoundUp (CR 107.4).
func TestPlayerLifeHalvedAmount(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	controllerRef := game.ControllerReference()
	obj := &game.StackObject{Controller: game.Player1}

	cases := []struct {
		life    int
		roundUp bool
		want    int
	}{
		{life: 20, roundUp: true, want: 10},
		{life: 15, roundUp: true, want: 8},
		{life: 15, roundUp: false, want: 7},
		{life: 1, roundUp: true, want: 1},
		{life: 0, roundUp: true, want: 0},
	}
	for _, tc := range cases {
		g.Players[game.Player1].Life = tc.life
		dynamic := game.DynamicAmount{
			Kind:    game.DynamicAmountPlayerLife,
			Player:  &controllerRef,
			Divisor: 2,
			RoundUp: tc.roundUp,
		}
		got := dynamicAmountValue(g, obj, game.Player1, dynamic)
		if got != tc.want {
			t.Errorf("half of %d life (roundUp=%v) = %d, want %d", tc.life, tc.roundUp, got, tc.want)
		}
	}
}
