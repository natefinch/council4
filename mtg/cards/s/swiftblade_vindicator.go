package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SwiftbladeVindicator is the card definition for Swiftblade Vindicator.
//
// Type: Creature — Human Soldier
// Cost: {R}{W}
//
// Oracle text:
//
//	Double strike (This creature deals both first-strike and regular combat damage.)
//	Vigilance (Attacking doesn't cause this creature to tap.)
//	Trample (This creature can deal excess combat damage to the player or planeswalker it's attacking.)
var SwiftbladeVindicator = newSwiftbladeVindicator

func newSwiftbladeVindicator() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name: "Swiftblade Vindicator",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
				cost.W,
			}),
			Colors:    []color.Color{color.Red, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Soldier},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.DoubleStrikeStaticBody,
				game.VigilanceStaticBody,
				game.TrampleStaticBody,
			},
			OracleText: `
			Double strike (This creature deals both first-strike and regular combat damage.)
			Vigilance (Attacking doesn't cause this creature to tap.)
			Trample (This creature can deal excess combat damage to the player or planeswalker it's attacking.)
		`,
		},
	}
}
