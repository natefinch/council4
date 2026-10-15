package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AbbeyGargoyles is the card definition for Abbey Gargoyles.
//
// Type: Creature — Gargoyle
// Cost: {2}{W}{W}{W}
//
// Oracle text:
//
//	Flying, protection from red
var AbbeyGargoyles = newAbbeyGargoyles

func newAbbeyGargoyles() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Abbey Gargoyles",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Gargoyle},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.ProtectionFromColorsStaticAbility(color.Red),
			},
			OracleText: `
			Flying, protection from red
		`,
		},
	}
}
