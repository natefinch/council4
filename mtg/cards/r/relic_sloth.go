package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RelicSloth is the card definition for Relic Sloth.
//
// Type: Creature — Sloth Beast
// Cost: {3}{R}{W}
//
// Oracle text:
//
//	Vigilance
//	Menace (This creature can't be blocked except by two or more creatures.)
var RelicSloth = newRelicSloth

func newRelicSloth() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name: "Relic Sloth",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.W,
			}),
			Colors:    []color.Color{color.Red, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Sloth, types.Beast},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
				game.MenaceStaticBody,
			},
			OracleText: `
			Vigilance
			Menace (This creature can't be blocked except by two or more creatures.)
		`,
		},
	}
}
