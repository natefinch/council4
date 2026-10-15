package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WallOfSpears is the card definition for Wall of Spears.
//
// Type: Artifact Creature — Wall
// Cost: {3}
//
// Oracle text:
//
//	Defender (This creature can't attack.)
//	First strike
var WallOfSpears = newWallOfSpears

func newWallOfSpears() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Wall of Spears",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Wall},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
				game.FirstStrikeStaticBody,
			},
			OracleText: `
			Defender (This creature can't attack.)
			First strike
		`,
		},
	}
}
