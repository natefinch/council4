package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RograkhSonOfRohgahh is the card definition for Rograkh, Son of Rohgahh.
//
// Type: Legendary Creature — Kobold Warrior
// Cost: {0}
//
// Oracle text:
//
//	First strike, menace, trample
//	Partner (You can have two commanders if both have partner.)
var RograkhSonOfRohgahh = newRograkhSonOfRohgahh

func newRograkhSonOfRohgahh() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Rograkh, Son of Rohgahh",
			ManaCost: opt.Val(cost.Mana{
				cost.O(0),
			}),
			Colors:     []color.Color{color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Kobold, types.Warrior},
			Power:      opt.Val(game.PT{Value: 0}),
			Toughness:  opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
				game.MenaceStaticBody,
				game.TrampleStaticBody,
				game.PartnerStaticBody,
			},
			OracleText: `
			First strike, menace, trample
			Partner (You can have two commanders if both have partner.)
		`,
		},
	}
}
