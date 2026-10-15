package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WallOfVines is the card definition for Wall of Vines.
//
// Type: Creature — Plant Wall
// Cost: {G}
//
// Oracle text:
//
//	Defender (This creature can't attack.)
//	Reach (This creature can block creatures with flying.)
var WallOfVines = newWallOfVines

func newWallOfVines() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Wall of Vines",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Plant, types.Wall},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
				game.ReachStaticBody,
			},
			OracleText: `
			Defender (This creature can't attack.)
			Reach (This creature can block creatures with flying.)
		`,
		},
	}
}
