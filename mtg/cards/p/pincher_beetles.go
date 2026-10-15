package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PincherBeetles is the card definition for Pincher Beetles.
//
// Type: Creature — Insect
// Cost: {2}{G}
//
// Oracle text:
//
//	Shroud (This creature can't be the target of spells or abilities.)
var PincherBeetles = newPincherBeetles

func newPincherBeetles() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Pincher Beetles",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Insect},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.ShroudStaticBody,
			},
			OracleText: `
			Shroud (This creature can't be the target of spells or abilities.)
		`,
		},
	}
}
