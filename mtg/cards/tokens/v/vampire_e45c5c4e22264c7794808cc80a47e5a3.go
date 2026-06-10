package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Vampire
//
// Type: Token Creature — Vampire
//
// Oracle text:
//   Trample, lifelink, haste

// VampireTokene45c5c4e22264c7794808cc80a47e5a3 is the card definition for Vampire.
var VampireTokene45c5c4e22264c7794808cc80a47e5a3 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black, color.Red),
	CardFace: game.CardFace{
		Name:      "Vampire",
		Colors:    []color.Color{color.Black, color.Red},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Vampire},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.TrampleStaticBody,
			game.LifelinkStaticBody,
			game.HasteStaticBody,
		},
		OracleText: `
			Trample, lifelink, haste
		`,
	},
}
