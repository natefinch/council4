package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MatcaRioters is the card definition for Matca Rioters.
//
// Type: Creature — Human Warrior
// Cost: {2}{G}
//
// Oracle text:
//
//	Domain — Matca Rioters's power and toughness are each equal to the number of basic land types among lands you control.
var MatcaRioters = newMatcaRioters

func newMatcaRioters() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Matca Rioters",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:           []color.Color{color.Green},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Human, types.Warrior},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerBasicLandTypeCount}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerBasicLandTypeCount}),
			OracleText: `
			Domain — Matca Rioters's power and toughness are each equal to the number of basic land types among lands you control.
		`,
		},
	}
}
