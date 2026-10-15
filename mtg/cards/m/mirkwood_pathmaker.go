package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MirkwoodPathmaker is the card definition for Mirkwood Pathmaker.
//
// Type: Creature — Elf Ranger
// Cost: {2}{G}
//
// Oracle text:
//
//	Mirkwood Pathmaker's power and toughness are each equal to the number of lands you control.
var MirkwoodPathmaker = newMirkwoodPathmaker

func newMirkwoodPathmaker() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Mirkwood Pathmaker",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:           []color.Color{color.Green},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Elf, types.Ranger},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerLandCount}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerLandCount}),
			OracleText: `
			Mirkwood Pathmaker's power and toughness are each equal to the number of lands you control.
		`,
		},
	}
}
