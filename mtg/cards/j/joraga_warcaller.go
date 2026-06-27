package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// JoragaWarcaller is the card definition for Joraga Warcaller.
//
// Type: Creature — Elf Warrior
// Cost: {G}
//
// Oracle text:
//
//	Multikicker {1}{G} (You may pay an additional {1}{G} any number of times as you cast this spell.)
//	This creature enters with a +1/+1 counter on it for each time it was kicked.
//	Other Elf creatures you control get +1/+1 for each +1/+1 counter on this creature.
var JoragaWarcaller = newJoragaWarcaller()

func newJoragaWarcaller() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Joraga Warcaller",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Warrior},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{Cost: cost.Mana{cost.O(1), cost.G}, Multi: true},
					},
				},
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerPowerToughnessModify,
							Group: game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Elf")}}, game.SourcePermanentReference()),
							PowerDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:        game.DynamicAmountObjectCounters,
								Multiplier:  1,
								CounterKind: counter.PlusOnePlusOne,
								Object:      game.SourcePermanentReference(),
							}),
							ToughnessDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:        game.DynamicAmountObjectCounters,
								Multiplier:  1,
								CounterKind: counter.PlusOnePlusOne,
								Object:      game.SourcePermanentReference(),
							}),
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
			Other Elf creatures you control get +1/+1 for each +1/+1 counter on this creature.
		`,
		},
	}
}
