package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GanglyStompling is the card definition for Gangly Stompling.
//
// Type: Creature — Shapeshifter
// Cost: {2}{R/G}
//
// Oracle text:
//
//	Changeling (This card is every creature type.)
//	Trample
var GanglyStompling = newGanglyStompling

func newGanglyStompling() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Gangly Stompling",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.HybridMana(mana.R, mana.G),
			}),
			Colors:    []color.Color{color.Green, color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Shapeshifter},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.ChangelingStaticBody,
				game.TrampleStaticBody,
			},
			OracleText: `
			Changeling (This card is every creature type.)
			Trample
		`,
		},
	}
}
