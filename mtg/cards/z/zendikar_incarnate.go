package z

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ZendikarIncarnate is the card definition for Zendikar Incarnate.
//
// Type: Creature — Elemental
// Cost: {2}{R}{G}
//
// Oracle text:
//
//	Zendikar Incarnate's power is equal to the number of lands you control.
var ZendikarIncarnate = newZendikarIncarnate

func newZendikarIncarnate() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Zendikar Incarnate",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
				cost.G,
			}),
			Colors:       []color.Color{color.Green, color.Red},
			Types:        []types.Card{types.Creature},
			Subtypes:     []types.Sub{types.Elemental},
			Power:        opt.Val(game.PT{IsStar: true}),
			Toughness:    opt.Val(game.PT{Value: 4}),
			DynamicPower: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerLandCount}),
			OracleText: `
			Zendikar Incarnate's power is equal to the number of lands you control.
		`,
		},
	}
}
