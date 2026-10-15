package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HordeOfBoggarts is the card definition for Horde of Boggarts.
//
// Type: Creature — Goblin
// Cost: {3}{R}
//
// Oracle text:
//
//	Menace (This creature can't be blocked except by two or more creatures.)
//	Horde of Boggarts's power and toughness are each equal to the number of red permanents you control.
var HordeOfBoggarts = newHordeOfBoggarts

func newHordeOfBoggarts() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Horde of Boggarts",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors:           []color.Color{color.Red},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Goblin},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerColorPermanentCount, Color: color.Red}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerColorPermanentCount, Color: color.Red}),
			StaticAbilities: []game.StaticAbility{
				game.MenaceStaticBody,
			},
			OracleText: `
			Menace (This creature can't be blocked except by two or more creatures.)
			Horde of Boggarts's power and toughness are each equal to the number of red permanents you control.
		`,
		},
	}
}
