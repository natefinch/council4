package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GnarlidPack is the card definition for Gnarlid Pack.
//
// Type: Creature — Beast
// Cost: {1}{G}
//
// Oracle text:
//
//	Multikicker {1}{G} (You may pay an additional {1}{G} any number of times as you cast this spell.)
//	This creature enters with a +1/+1 counter on it for each time it was kicked.
var GnarlidPack = newGnarlidPack

func newGnarlidPack() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Gnarlid Pack",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Multikicker {1}{G}",
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{
							Cost: cost.Mana{
								cost.O(1),
								cost.G,
							},
							Multi: true,
						},
					},
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This creature enters with a +1/+1 counter on it for each time it was kicked.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Dynamic: opt.Val(&game.DynamicAmount{
					Kind:       game.DynamicAmountTimesKicked,
					Multiplier: 1,
				})}),
			},
			OracleText: `
			Multikicker {1}{G} (You may pay an additional {1}{G} any number of times as you cast this spell.)
			This creature enters with a +1/+1 counter on it for each time it was kicked.
		`,
		},
	}
}
