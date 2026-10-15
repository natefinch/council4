package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TwoHeadedCerberus is the card definition for Two-Headed Cerberus.
//
// Type: Creature — Dog
// Cost: {1}{R}{R}
//
// Oracle text:
//
//	Double strike (This creature deals both first-strike and regular combat damage.)
var TwoHeadedCerberus = newTwoHeadedCerberus

func newTwoHeadedCerberus() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Two-Headed Cerberus",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dog},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.DoubleStrikeStaticBody,
			},
			OracleText: `
			Double strike (This creature deals both first-strike and regular combat damage.)
		`,
		},
	}
}
