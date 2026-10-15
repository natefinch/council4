package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HarmoniousGrovestrider is the card definition for Harmonious Grovestrider.
//
// Type: Creature — Beast
// Cost: {3}{G}{G}
//
// Oracle text:
//
//	Ward {2} (Whenever this creature becomes the target of a spell or ability an opponent controls, counter it unless that player pays {2}.)
//	Harmonious Grovestrider's power and toughness are each equal to the number of lands you control.
var HarmoniousGrovestrider = newHarmoniousGrovestrider

func newHarmoniousGrovestrider() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Harmonious Grovestrider",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
				cost.G,
			}),
			Colors:           []color.Color{color.Green},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Beast},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerLandCount}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerLandCount}),
			StaticAbilities: []game.StaticAbility{
				game.WardStaticAbility(cost.Mana{cost.O(2)}),
			},
			OracleText: `
			Ward {2} (Whenever this creature becomes the target of a spell or ability an opponent controls, counter it unless that player pays {2}.)
			Harmonious Grovestrider's power and toughness are each equal to the number of lands you control.
		`,
		},
	}
}
