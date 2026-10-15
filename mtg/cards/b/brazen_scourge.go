package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BrazenScourge is the card definition for Brazen Scourge.
//
// Type: Creature — Gremlin
// Cost: {1}{R}{R}
//
// Oracle text:
//
//	Haste (This creature can attack and {T} as soon as it comes under your control.)
var BrazenScourge = newBrazenScourge

func newBrazenScourge() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Brazen Scourge",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Gremlin},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
			},
			OracleText: `
			Haste (This creature can attack and {T} as soon as it comes under your control.)
		`,
		},
	}
}
