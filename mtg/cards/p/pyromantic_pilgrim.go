package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PyromanticPilgrim is the card definition for Pyromantic Pilgrim.
//
// Type: Creature — Human Wizard
// Cost: {2}{R}
//
// Oracle text:
//
//	Haste (This creature can attack and {T} as soon as it comes under your control.)
var PyromanticPilgrim = newPyromanticPilgrim

func newPyromanticPilgrim() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Pyromantic Pilgrim",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Wizard},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
			},
			OracleText: `
			Haste (This creature can attack and {T} as soon as it comes under your control.)
		`,
		},
	}
}
