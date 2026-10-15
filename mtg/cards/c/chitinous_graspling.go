package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ChitinousGraspling is the card definition for Chitinous Graspling.
//
// Type: Creature — Shapeshifter
// Cost: {3}{G/U}
//
// Oracle text:
//
//	Changeling (This card is every creature type.)
//	Reach
var ChitinousGraspling = newChitinousGraspling

func newChitinousGraspling() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Chitinous Graspling",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.HybridMana(mana.G, mana.U),
			}),
			Colors:    []color.Color{color.Green, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Shapeshifter},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.ChangelingStaticBody,
				game.ReachStaticBody,
			},
			OracleText: `
			Changeling (This card is every creature type.)
			Reach
		`,
		},
	}
}
