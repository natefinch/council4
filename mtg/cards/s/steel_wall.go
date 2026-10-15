package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SteelWall is the card definition for Steel Wall.
//
// Type: Artifact Creature — Wall
// Cost: {1}
//
// Oracle text:
//
//	Defender (This creature can't attack.)
var SteelWall = newSteelWall

func newSteelWall() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Steel Wall",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Wall},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
			},
			OracleText: `
			Defender (This creature can't attack.)
		`,
		},
	}
}
