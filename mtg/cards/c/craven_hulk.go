package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CravenHulk is the card definition for Craven Hulk.
//
// Type: Creature — Giant Coward
// Cost: {3}{R}
//
// Oracle text:
//
//	This creature can't block alone.
var CravenHulk = newCravenHulk

func newCravenHulk() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Craven Hulk",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Giant, types.Coward},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.CantBlockAloneStaticBody,
			},
			OracleText: `
			This creature can't block alone.
		`,
		},
	}
}
