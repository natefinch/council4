package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Treefolk
//
// Type: Token Creature — Treefolk
//
// Oracle text:
//   Reach

// TreefolkToken5d403f8968624ed9ba39890c90ff487c is the card definition for Treefolk.
var TreefolkToken5d403f8968624ed9ba39890c90ff487c = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Treefolk",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Treefolk},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 4}),
		StaticAbilities: []game.StaticAbility{
			game.ReachStaticBody,
		},
		OracleText: `
			Reach
		`,
	},
}
