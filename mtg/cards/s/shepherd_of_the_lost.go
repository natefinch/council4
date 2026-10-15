package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ShepherdOfTheLost is the card definition for Shepherd of the Lost.
//
// Type: Creature — Angel
// Cost: {4}{W}
//
// Oracle text:
//
//	Flying, first strike, vigilance
var ShepherdOfTheLost = newShepherdOfTheLost

func newShepherdOfTheLost() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Shepherd of the Lost",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Angel},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.FirstStrikeStaticBody,
				game.VigilanceStaticBody,
			},
			OracleText: `
			Flying, first strike, vigilance
		`,
		},
	}
}
