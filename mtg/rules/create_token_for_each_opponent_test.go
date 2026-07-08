package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// villainTokenDef is the 2/1 black Villain creature token Endless Ranks of
// HYDRA creates for each opponent.
func villainTokenDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Villain",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Villain},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}}
}

// TestCreateTokenForEachOpponentScalesWithOpponents drives the real resolution
// of Endless Ranks of HYDRA's create — a CreateToken whose amount is the dynamic
// opponent count — and confirms it creates one token per opponent rather than a
// flat single token. With all three opponents alive it creates three tokens;
// once two opponents are eliminated it creates one, proving the count scales
// with the live opponent count as the effect resolves (CR 608.2c).
func TestCreateTokenForEachOpponentScalesWithOpponents(t *testing.T) {
	for _, tc := range []struct {
		name      string
		eliminate []game.PlayerID
		want      int
	}{
		{name: "three opponents", eliminate: nil, want: 3},
		{name: "one opponent", eliminate: []game.PlayerID{game.Player3, game.Player4}, want: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			for _, id := range tc.eliminate {
				g.Players[id].Eliminated = true
			}

			obj := &game.StackObject{Controller: game.Player1}
			resolveInstruction(engine, g, obj, game.CreateToken{
				Amount: game.Dynamic(game.DynamicAmount{
					Kind:       game.DynamicAmountOpponentCount,
					Multiplier: 1,
				}),
				Source: game.TokenDef(villainTokenDef()),
			}, &TurnLog{})

			if got := tokensByController(g)[game.Player1]; got != tc.want {
				t.Fatalf("Player1 tokens = %d, want %d (one per opponent)", got, tc.want)
			}
		})
	}
}
