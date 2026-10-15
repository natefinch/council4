package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Crystacean is the card definition for Crystacean.
//
// Type: Creature — Crab
// Cost: {3}{U}
//
// Oracle text:
//
//	Flash (You may cast this spell any time you could cast an instant.)
var Crystacean = newCrystacean

func newCrystacean() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Crystacean",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Crab},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
			},
			OracleText: `
			Flash (You may cast this spell any time you could cast an instant.)
		`,
		},
	}
}
