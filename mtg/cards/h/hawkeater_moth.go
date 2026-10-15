package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HawkeaterMoth is the card definition for Hawkeater Moth.
//
// Type: Creature — Insect
// Cost: {3}{G}
//
// Oracle text:
//
//	Flying
//	Shroud (This creature can't be the target of spells or abilities.)
var HawkeaterMoth = newHawkeaterMoth

func newHawkeaterMoth() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Hawkeater Moth",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Insect},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.ShroudStaticBody,
			},
			OracleText: `
			Flying
			Shroud (This creature can't be the target of spells or abilities.)
		`,
		},
	}
}
