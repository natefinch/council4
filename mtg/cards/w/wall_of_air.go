package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WallOfAir is the card definition for Wall of Air.
//
// Type: Creature — Wall
// Cost: {1}{U}{U}
//
// Oracle text:
//
//	Defender, flying (This creature can't attack, and it can block creatures with flying.)
var WallOfAir = newWallOfAir

func newWallOfAir() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Wall of Air",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wall},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
				game.FlyingStaticBody,
			},
			OracleText: `
			Defender, flying (This creature can't attack, and it can block creatures with flying.)
		`,
		},
	}
}
