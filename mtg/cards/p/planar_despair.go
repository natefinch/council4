package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PlanarDespair is the card definition for Planar Despair.
//
// Type: Sorcery
// Cost: {3}{B}{B}
//
// Oracle text:
//
//	Domain — All creatures get -1/-1 until end of turn for each basic land type among lands you control.
var PlanarDespair = newPlanarDespair

func newPlanarDespair() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Planar Despair",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Duration: game.DurationPermanent,
									Group:    game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
									Layer:    game.LayerPowerToughnessModify,
									PowerDeltaDynamic: opt.Val(game.DynamicAmount{
										Kind:        game.DynamicAmountControllerBasicLandTypeCount,
										Multiplier:  -1,
										CounterKind: counter.PlusOnePlusOne,
									}),
									ToughnessDeltaDynamic: opt.Val(game.DynamicAmount{
										Kind:        game.DynamicAmountControllerBasicLandTypeCount,
										Multiplier:  -1,
										CounterKind: counter.PlusOnePlusOne,
									}),
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Domain — All creatures get -1/-1 until end of turn for each basic land type among lands you control.
		`,
		},
	}
}
