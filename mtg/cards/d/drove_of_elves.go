package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DroveOfElves is the card definition for Drove of Elves.
//
// Type: Creature — Elf
// Cost: {3}{G}
//
// Oracle text:
//
//	Hexproof
//	Drove of Elves's power and toughness are each equal to the number of green permanents you control.
var DroveOfElves = newDroveOfElves

func newDroveOfElves() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Drove of Elves",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors:           []color.Color{color.Green},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Elf},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerColorPermanentCount, Color: color.Green}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerColorPermanentCount, Color: color.Green}),
			StaticAbilities: []game.StaticAbility{
				game.HexproofStaticBody,
			},
			OracleText: `
			Hexproof
			Drove of Elves's power and toughness are each equal to the number of green permanents you control.
		`,
		},
	}
}
