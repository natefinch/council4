package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AvenFleetwing is the card definition for Aven Fleetwing.
//
// Type: Creature — Bird Soldier
// Cost: {3}{U}
//
// Oracle text:
//
//	Flying
//	Hexproof (This creature can't be the target of spells or abilities your opponents control.)
var AvenFleetwing = newAvenFleetwing

func newAvenFleetwing() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Aven Fleetwing",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bird, types.Soldier},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.HexproofStaticBody,
			},
			OracleText: `
			Flying
			Hexproof (This creature can't be the target of spells or abilities your opponents control.)
		`,
		},
	}
}
