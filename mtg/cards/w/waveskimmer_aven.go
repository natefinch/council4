package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WaveskimmerAven is the card definition for Waveskimmer Aven.
//
// Type: Creature — Bird Soldier
// Cost: {2}{G}{W}{U}
//
// Oracle text:
//
//	Flying
//	Exalted (Whenever a creature you control attacks alone, that creature gets +1/+1 until end of turn.)
var WaveskimmerAven = newWaveskimmerAven

func newWaveskimmerAven() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Waveskimmer Aven",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.W,
				cost.U,
			}),
			Colors:    []color.Color{color.Green, color.Blue, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bird, types.Soldier},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.ExaltedStaticBody,
			},
			OracleText: `
			Flying
			Exalted (Whenever a creature you control attacks alone, that creature gets +1/+1 until end of turn.)
		`,
		},
	}
}
