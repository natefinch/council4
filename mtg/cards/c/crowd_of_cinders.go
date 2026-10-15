package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CrowdOfCinders is the card definition for Crowd of Cinders.
//
// Type: Creature — Elemental
// Cost: {3}{B}
//
// Oracle text:
//
//	Fear (This creature can't be blocked except by artifact creatures and/or black creatures.)
//	Crowd of Cinders's power and toughness are each equal to the number of black permanents you control.
var CrowdOfCinders = newCrowdOfCinders

func newCrowdOfCinders() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Crowd of Cinders",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors:           []color.Color{color.Black},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Elemental},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerColorPermanentCount, Color: color.Black}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerColorPermanentCount, Color: color.Black}),
			StaticAbilities: []game.StaticAbility{
				game.FearStaticBody,
			},
			OracleText: `
			Fear (This creature can't be blocked except by artifact creatures and/or black creatures.)
			Crowd of Cinders's power and toughness are each equal to the number of black permanents you control.
		`,
		},
	}
}
