package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ApexHawks is the card definition for Apex Hawks.
//
// Type: Creature — Bird
// Cost: {2}{W}
//
// Oracle text:
//
//	Multikicker {1}{W} (You may pay an additional {1}{W} any number of times as you cast this spell.)
//	Flying
//	This creature enters with a +1/+1 counter on it for each time it was kicked.
var ApexHawks = newApexHawks

func newApexHawks() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Apex Hawks",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bird},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Multikicker {1}{W}",
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{
							Cost: cost.Mana{
								cost.O(1),
								cost.W,
							},
							Multi: true,
						},
					},
				},
				game.FlyingStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This creature enters with a +1/+1 counter on it for each time it was kicked.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Dynamic: opt.Val(&game.DynamicAmount{
					Kind:       game.DynamicAmountTimesKicked,
					Multiplier: 1,
				})}),
			},
			OracleText: `
			Multikicker {1}{W} (You may pay an additional {1}{W} any number of times as you cast this spell.)
			Flying
			This creature enters with a +1/+1 counter on it for each time it was kicked.
		`,
		},
	}
}
