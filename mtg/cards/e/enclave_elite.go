package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EnclaveElite is the card definition for Enclave Elite.
//
// Type: Creature — Merfolk Soldier
// Cost: {2}{U}
//
// Oracle text:
//
//	Multikicker {1}{U} (You may pay an additional {1}{U} any number of times as you cast this spell.)
//	Islandwalk (This creature can't be blocked as long as defending player controls an Island.)
//	This creature enters with a +1/+1 counter on it for each time it was kicked.
var EnclaveElite = newEnclaveElite

func newEnclaveElite() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Enclave Elite",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Merfolk, types.Soldier},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Multikicker {1}{U}",
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{
							Cost: cost.Mana{
								cost.O(1),
								cost.U,
							},
							Multi: true,
						},
					},
				},
				game.IslandwalkStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This creature enters with a +1/+1 counter on it for each time it was kicked.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Dynamic: opt.Val(&game.DynamicAmount{
					Kind:       game.DynamicAmountTimesKicked,
					Multiplier: 1,
				})}),
			},
			OracleText: `
			Multikicker {1}{U} (You may pay an additional {1}{U} any number of times as you cast this spell.)
			Islandwalk (This creature can't be blocked as long as defending player controls an Island.)
			This creature enters with a +1/+1 counter on it for each time it was kicked.
		`,
		},
	}
}
