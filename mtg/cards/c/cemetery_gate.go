package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CemeteryGate is the card definition for Cemetery Gate.
//
// Type: Creature — Wall
// Cost: {2}{B}
//
// Oracle text:
//
//	Defender (This creature can't attack.)
//	Protection from black
var CemeteryGate = newCemeteryGate

func newCemeteryGate() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Cemetery Gate",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wall},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
				game.ProtectionFromColorsStaticAbility(color.Black),
			},
			OracleText: `
			Defender (This creature can't attack.)
			Protection from black
		`,
		},
	}
}
