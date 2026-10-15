package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KnightOfMeadowgrain is the card definition for Knight of Meadowgrain.
//
// Type: Creature — Kithkin Knight
// Cost: {W}{W}
//
// Oracle text:
//
//	First strike
//	Lifelink (Damage dealt by this creature also causes you to gain that much life.)
var KnightOfMeadowgrain = newKnightOfMeadowgrain

func newKnightOfMeadowgrain() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Knight of Meadowgrain",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Kithkin, types.Knight},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
				game.LifelinkStaticBody,
			},
			OracleText: `
			First strike
			Lifelink (Damage dealt by this creature also causes you to gain that much life.)
		`,
		},
	}
}
