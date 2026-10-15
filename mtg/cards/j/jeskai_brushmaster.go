package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// JeskaiBrushmaster is the card definition for Jeskai Brushmaster.
//
// Type: Creature — Orc Monk
// Cost: {1}{U}{R}{W}
//
// Oracle text:
//
//	Double strike
//	Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
var JeskaiBrushmaster = newJeskaiBrushmaster

func newJeskaiBrushmaster() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Jeskai Brushmaster",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
				cost.R,
				cost.W,
			}),
			Colors:    []color.Color{color.Red, color.Blue, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Orc, types.Monk},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.DoubleStrikeStaticBody,
				game.ProwessStaticBody,
			},
			OracleText: `
			Double strike
			Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
		`,
		},
	}
}
