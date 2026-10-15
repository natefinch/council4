package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TwoHeadedSliver is the card definition for Two-Headed Sliver.
//
// Type: Creature — Sliver
// Cost: {1}{R}
//
// Oracle text:
//
//	All Sliver creatures have menace. (They can't be blocked except by two or more creatures.)
var TwoHeadedSliver = newTwoHeadedSliver

func newTwoHeadedSliver() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Two-Headed Sliver",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Sliver},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "All Sliver creatures have menace. (They can't be blocked except by two or more creatures.)",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Sliver")}}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Menace},
						},
					},
				},
			},
			OracleText: `
			All Sliver creatures have menace. (They can't be blocked except by two or more creatures.)
		`,
		},
	}
}
