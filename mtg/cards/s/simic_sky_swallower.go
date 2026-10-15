package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SimicSkySwallower is the card definition for Simic Sky Swallower.
//
// Type: Creature — Leviathan
// Cost: {5}{G}{U}
//
// Oracle text:
//
//	Flying, trample
//	Shroud (This creature can't be the target of spells or abilities.)
var SimicSkySwallower = newSimicSkySwallower

func newSimicSkySwallower() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Simic Sky Swallower",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.G,
				cost.U,
			}),
			Colors:    []color.Color{color.Green, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Leviathan},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.TrampleStaticBody,
				game.ShroudStaticBody,
			},
			OracleText: `
			Flying, trample
			Shroud (This creature can't be the target of spells or abilities.)
		`,
		},
	}
}
