package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ElvishLookout is the card definition for Elvish Lookout.
//
// Type: Creature — Elf
// Cost: {G}
//
// Oracle text:
//
//	Shroud (This creature can't be the target of spells or abilities.)
var ElvishLookout = newElvishLookout

func newElvishLookout() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Elvish Lookout",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.ShroudStaticBody,
			},
			OracleText: `
			Shroud (This creature can't be the target of spells or abilities.)
		`,
		},
	}
}
