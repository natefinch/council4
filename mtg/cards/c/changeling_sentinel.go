package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ChangelingSentinel is the card definition for Changeling Sentinel.
//
// Type: Creature — Shapeshifter
// Cost: {3}{W}
//
// Oracle text:
//
//	Changeling (This card is every creature type.)
//	Vigilance
var ChangelingSentinel = newChangelingSentinel

func newChangelingSentinel() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Changeling Sentinel",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Shapeshifter},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.ChangelingStaticBody,
				game.VigilanceStaticBody,
			},
			OracleText: `
			Changeling (This card is every creature type.)
			Vigilance
		`,
		},
	}
}
