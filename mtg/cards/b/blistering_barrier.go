package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BlisteringBarrier is the card definition for Blistering Barrier.
//
// Type: Creature — Wall
// Cost: {2}{R}
//
// Oracle text:
//
//	Defender (This creature can't attack.)
var BlisteringBarrier = newBlisteringBarrier

func newBlisteringBarrier() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Blistering Barrier",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wall},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
			},
			OracleText: `
			Defender (This creature can't attack.)
		`,
		},
	}
}
