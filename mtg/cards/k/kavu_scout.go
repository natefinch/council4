package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KavuScout is the card definition for Kavu Scout.
//
// Type: Creature — Kavu Scout
// Cost: {2}{R}
//
// Oracle text:
//
//	Domain — This creature gets +1/+0 for each basic land type among lands you control.
var KavuScout = newKavuScout

func newKavuScout() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Kavu Scout",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Kavu, types.Scout},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Domain — This creature gets +1/+0 for each basic land type among lands you control.",
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
						},
					},
				},
			},
			OracleText: `
			Domain — This creature gets +1/+0 for each basic land type among lands you control.
		`,
		},
	}
}
