package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Ornithopter is the card definition for Ornithopter.
//
// Type: Artifact Creature — Thopter
// Cost: {0}
//
// Oracle text:
//
//	Flying
var Ornithopter = newOrnithopter

func newOrnithopter() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Ornithopter",
			ManaCost: opt.Val(cost.Mana{
				cost.O(0),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Thopter},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			OracleText: `
			Flying
		`,
		},
	}
}
