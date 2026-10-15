package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KithkinRabble is the card definition for Kithkin Rabble.
//
// Type: Creature — Kithkin
// Cost: {3}{W}
//
// Oracle text:
//
//	Vigilance
//	Kithkin Rabble's power and toughness are each equal to the number of white permanents you control.
var KithkinRabble = newKithkinRabble

func newKithkinRabble() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Kithkin Rabble",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:           []color.Color{color.White},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Kithkin},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerColorPermanentCount, Color: color.White}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerColorPermanentCount, Color: color.White}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
			},
			OracleText: `
			Vigilance
			Kithkin Rabble's power and toughness are each equal to the number of white permanents you control.
		`,
		},
	}
}
