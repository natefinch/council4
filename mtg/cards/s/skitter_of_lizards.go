package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SkitterOfLizards is the card definition for Skitter of Lizards.
//
// Type: Creature — Lizard
// Cost: {R}
//
// Oracle text:
//
//	Multikicker {1}{R} (You may pay an additional {1}{R} any number of times as you cast this spell.)
//	Haste
//	This creature enters with a +1/+1 counter on it for each time it was kicked.
var SkitterOfLizards = newSkitterOfLizards

func newSkitterOfLizards() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Skitter of Lizards",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Lizard},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Multikicker {1}{R}",
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{
							Cost: cost.Mana{
								cost.O(1),
								cost.R,
							},
							Multi: true,
						},
					},
				},
				game.HasteStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This creature enters with a +1/+1 counter on it for each time it was kicked.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Dynamic: opt.Val(&game.DynamicAmount{
					Kind:       game.DynamicAmountTimesKicked,
					Multiplier: 1,
				})}),
			},
			OracleText: `
			Multikicker {1}{R} (You may pay an additional {1}{R} any number of times as you cast this spell.)
			Haste
			This creature enters with a +1/+1 counter on it for each time it was kicked.
		`,
		},
	}
}
