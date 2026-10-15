package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AvianChangeling is the card definition for Avian Changeling.
//
// Type: Creature — Shapeshifter
// Cost: {2}{W}
//
// Oracle text:
//
//	Changeling (This card is every creature type.)
//	Flying
var AvianChangeling = newAvianChangeling

func newAvianChangeling() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Avian Changeling",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Shapeshifter},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.ChangelingStaticBody,
				game.FlyingStaticBody,
			},
			OracleText: `
			Changeling (This card is every creature type.)
			Flying
		`,
		},
	}
}
