package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Clamavus is the card definition for Clamavus.
//
// Type: Creature — Human Tyranid Artificer
// Cost: {4}{G}
//
// Oracle text:
//
//	Proclamator Hailer — Each creature you control gets +1/+1 for each +1/+1 counter on it.
var Clamavus = newClamavus

func newClamavus() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Clamavus",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Tyranid, types.Artificer},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Proclamator Hailer — Each creature you control gets +1/+1 for each +1/+1 counter on it.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							Layer:    game.LayerPowerToughnessModify,
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
			OracleText: `
			Proclamator Hailer — Each creature you control gets +1/+1 for each +1/+1 counter on it.
		`,
		},
	}
}
