package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SpinelessThug is the card definition for Spineless Thug.
//
// Type: Creature — Phyrexian Zombie Mercenary
// Cost: {1}{B}
//
// Oracle text:
//
//	This creature can't block.
var SpinelessThug = newSpinelessThug

func newSpinelessThug() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Spineless Thug",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Zombie, types.Mercenary},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.CantBlockStaticBody,
			},
			OracleText: `
			This creature can't block.
		`,
		},
	}
}
