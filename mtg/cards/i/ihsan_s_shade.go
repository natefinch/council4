package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// IhsanSShade is the card definition for Ihsan's Shade.
//
// Type: Legendary Creature — Shade Knight
// Cost: {3}{B}{B}{B}
//
// Oracle text:
//
//	Protection from white
var IhsanSShade = newIhsanSShade

func newIhsanSShade() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Ihsan's Shade",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
				cost.B,
				cost.B,
			}),
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Shade, types.Knight},
			Power:      opt.Val(game.PT{Value: 5}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.ProtectionFromColorsStaticAbility(color.White),
			},
			OracleText: `
			Protection from white
		`,
		},
	}
}
