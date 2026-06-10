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
//   Lifelink (Damage dealt by this creature also causes you to gain that much life.)

// VampireToken97af46cb62094ae781cc792894999937 is the card definition for Vampire.
var VampireToken97af46cb62094ae781cc792894999937 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name:      "Vampire",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Vampire},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.LifelinkStaticBody,
		},
		OracleText: `
			Lifelink (Damage dealt by this creature also causes you to gain that much life.)
		`,
	},
}
