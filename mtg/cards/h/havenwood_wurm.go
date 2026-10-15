package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HavenwoodWurm is the card definition for Havenwood Wurm.
//
// Type: Creature — Wurm
// Cost: {6}{G}
//
// Oracle text:
//
//	Flash (You may cast this spell any time you could cast an instant.)
//	Trample
var HavenwoodWurm = newHavenwoodWurm

func newHavenwoodWurm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Havenwood Wurm",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wurm},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.TrampleStaticBody,
			},
			OracleText: `
			Flash (You may cast this spell any time you could cast an instant.)
			Trample
		`,
		},
	}
}
