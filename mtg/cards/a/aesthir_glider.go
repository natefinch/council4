package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AesthirGlider is the card definition for Aesthir Glider.
//
// Type: Artifact Creature — Bird Construct
// Cost: {3}
//
// Oracle text:
//
//	Flying
//	This creature can't block.
var AesthirGlider = newAesthirGlider

func newAesthirGlider() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Aesthir Glider",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Bird, types.Construct},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.CantBlockStaticBody,
			},
			OracleText: `
			Flying
			This creature can't block.
		`,
		},
	}
}
