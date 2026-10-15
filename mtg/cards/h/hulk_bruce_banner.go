package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HulkBruceBanner is the card definition for Hulk, Bruce Banner.
//
// Type: Legendary Creature — Gamma Berserker Hero
// Cost: {3}{R}{R}
//
// Oracle text:
//
//	Trample (This creature can deal excess combat damage to the player he's attacking.)
var HulkBruceBanner = newHulkBruceBanner

func newHulkBruceBanner() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Hulk, Bruce Banner",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.R,
			}),
			Colors:     []color.Color{color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Sub("Gamma"), types.Berserker, types.Hero},
			Power:      opt.Val(game.PT{Value: 7}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			OracleText: `
			Trample (This creature can deal excess combat damage to the player he's attacking.)
		`,
		},
	}
}
