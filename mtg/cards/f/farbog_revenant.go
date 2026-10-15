package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FarbogRevenant is the card definition for Farbog Revenant.
//
// Type: Creature — Spirit
// Cost: {2}{B}
//
// Oracle text:
//
//	Skulk (This creature can't be blocked by creatures with greater power.)
//	Lifelink (Damage dealt by this creature also causes you to gain that much life.)
var FarbogRevenant = newFarbogRevenant

func newFarbogRevenant() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Farbog Revenant",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.SkulkStaticBody,
				game.LifelinkStaticBody,
			},
			OracleText: `
			Skulk (This creature can't be blocked by creatures with greater power.)
			Lifelink (Damage dealt by this creature also causes you to gain that much life.)
		`,
		},
	}
}
