package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CoiledTinviper is the card definition for Coiled Tinviper.
//
// Type: Artifact Creature — Snake
// Cost: {3}
//
// Oracle text:
//
//	First strike
var CoiledTinviper = newCoiledTinviper

func newCoiledTinviper() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Coiled Tinviper",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Snake},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
			},
			OracleText: `
			First strike
		`,
		},
	}
}
