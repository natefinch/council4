package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WayfaringGiant is the card definition for Wayfaring Giant.
//
// Type: Creature — Giant
// Cost: {5}{W}
//
// Oracle text:
//
//	Domain — This creature gets +1/+1 for each basic land type among lands you control.
var WayfaringGiant = newWayfaringGiant

func newWayfaringGiant() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Wayfaring Giant",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Giant},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Domain — This creature gets +1/+1 for each basic land type among lands you control.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessModify,
							PowerDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:        game.DynamicAmountControllerBasicLandTypeCount,
								Multiplier:  1,
								CounterKind: counter.PlusOnePlusOne,
							}),
							ToughnessDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:        game.DynamicAmountControllerBasicLandTypeCount,
								Multiplier:  1,
								CounterKind: counter.PlusOnePlusOne,
							}),
						},
					},
				},
			},
			OracleText: `
			Domain — This creature gets +1/+1 for each basic land type among lands you control.
		`,
		},
	}
}
