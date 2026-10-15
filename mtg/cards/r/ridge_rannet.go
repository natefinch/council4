package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RidgeRannet is the card definition for Ridge Rannet.
//
// Type: Creature — Beast
// Cost: {5}{R}{R}
//
// Oracle text:
//
//	Cycling {2} ({2}, Discard this card: Draw a card.)
var RidgeRannet = newRidgeRannet

func newRidgeRannet() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Ridge Rannet",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 4}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(2)}),
			},
			OracleText: `
			Cycling {2} ({2}, Discard this card: Draw a card.)
		`,
		},
	}
}
