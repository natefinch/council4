package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SabertoothNishoba is the card definition for Sabertooth Nishoba.
//
// Type: Creature — Cat Beast Warrior
// Cost: {4}{G}{W}
//
// Oracle text:
//
//	Trample, protection from blue and from red
var SabertoothNishoba = newSabertoothNishoba

func newSabertoothNishoba() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Sabertooth Nishoba",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.W,
			}),
			Colors:    []color.Color{color.Green, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Cat, types.Beast, types.Warrior},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
				game.ProtectionFromColorsStaticAbility(color.Blue, color.Red),
			},
			OracleText: `
			Trample, protection from blue and from red
		`,
		},
	}
}
