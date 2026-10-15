package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SynchronousSliver is the card definition for Synchronous Sliver.
//
// Type: Creature — Sliver
// Cost: {4}{U}
//
// Oracle text:
//
//	All Sliver creatures have vigilance.
var SynchronousSliver = newSynchronousSliver

func newSynchronousSliver() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Synchronous Sliver",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Sliver},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "All Sliver creatures have vigilance.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Sliver")}}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Vigilance},
						},
					},
				},
			},
			OracleText: `
			All Sliver creatures have vigilance.
		`,
		},
	}
}
