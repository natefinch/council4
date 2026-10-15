package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LightningStormkin is the card definition for Lightning Stormkin.
//
// Type: Creature — Elemental Wizard
// Cost: {U}{R}
//
// Oracle text:
//
//	Flying
//	Haste (This creature can attack and {T} as soon as it comes under your control.)
var LightningStormkin = newLightningStormkin

func newLightningStormkin() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Lightning Stormkin",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
				cost.R,
			}),
			Colors:    []color.Color{color.Red, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental, types.Wizard},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.HasteStaticBody,
			},
			OracleText: `
			Flying
			Haste (This creature can attack and {T} as soon as it comes under your control.)
		`,
		},
	}
}
