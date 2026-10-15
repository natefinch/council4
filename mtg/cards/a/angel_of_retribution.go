package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AngelOfRetribution is the card definition for Angel of Retribution.
//
// Type: Creature — Angel
// Cost: {6}{W}
//
// Oracle text:
//
//	Flying, first strike
var AngelOfRetribution = newAngelOfRetribution

func newAngelOfRetribution() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Angel of Retribution",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Angel},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.FirstStrikeStaticBody,
			},
			OracleText: `
			Flying, first strike
		`,
		},
	}
}
