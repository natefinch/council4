package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Nightmare is the card definition for Nightmare.
//
// Type: Creature — Nightmare Horse
// Cost: {5}{B}
//
// Oracle text:
//
//	Flying (This creature can't be blocked except by creatures with flying or reach.)
//	Nightmare's power and toughness are each equal to the number of Swamps you control.
var Nightmare = newNightmare

func newNightmare() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Nightmare",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.B,
			}),
			Colors:           []color.Color{color.Black},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Nightmare, types.Horse},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerSubtypeCount, Subtype: types.Swamp}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerSubtypeCount, Subtype: types.Swamp}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			OracleText: `
			Flying (This creature can't be blocked except by creatures with flying or reach.)
			Nightmare's power and toughness are each equal to the number of Swamps you control.
		`,
		},
	}
}
