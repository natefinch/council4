package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Treefolk is the card definition for Treefolk.
//
// Type: Token Creature — Treefolk
//
// Oracle text:
//   Reach
//   This creature's power and toughness are each equal to the number of lands you control.

// TreefolkToken46e30c0315cd46eda122c47cbca87368 is the card definition for Treefolk.
var TreefolkToken46e30c0315cd46eda122c47cbca87368 = newTreefolkToken46e30c0315cd46eda122c47cbca87368()

func newTreefolkToken46e30c0315cd46eda122c47cbca87368() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:             "Treefolk",
			Colors:           []color.Color{color.Green},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Treefolk},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerLandCount}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerLandCount}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
			},
			OracleText: `
			Reach
			This creature's power and toughness are each equal to the number of lands you control.
		`,
		},
	}
}
