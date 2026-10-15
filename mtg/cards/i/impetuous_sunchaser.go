package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ImpetuousSunchaser is the card definition for Impetuous Sunchaser.
//
// Type: Creature — Human Soldier
// Cost: {1}{R}
//
// Oracle text:
//
//	Flying, haste
//	This creature attacks each combat if able.
var ImpetuousSunchaser = newImpetuousSunchaser

func newImpetuousSunchaser() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Impetuous Sunchaser",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Soldier},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.HasteStaticBody,
				game.MustAttackStaticBody,
			},
			OracleText: `
			Flying, haste
			This creature attacks each combat if able.
		`,
		},
	}
}
