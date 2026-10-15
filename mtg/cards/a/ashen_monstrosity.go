package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AshenMonstrosity is the card definition for Ashen Monstrosity.
//
// Type: Creature — Spirit
// Cost: {5}{R}{R}
//
// Oracle text:
//
//	Haste
//	This creature attacks each combat if able.
var AshenMonstrosity = newAshenMonstrosity

func newAshenMonstrosity() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Ashen Monstrosity",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 7}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
				game.MustAttackStaticBody,
			},
			OracleText: `
			Haste
			This creature attacks each combat if able.
		`,
		},
	}
}
