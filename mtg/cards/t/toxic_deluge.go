package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ToxicDeluge is the card definition for Toxic Deluge.
//
// Type: Sorcery
// Cost: {2}{B}
//
// Oracle text:
//
//	As an additional cost to cast this spell, pay X life.
//	All creatures get -X/-X until end of turn.
var ToxicDeluge = newToxicDeluge

func newToxicDeluge() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Toxic Deluge",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			AdditionalCosts: []cost.Additional{
				{
					Kind:        cost.AdditionalPayLife,
					Text:        "pay X life",
					AmountFromX: true,
				},
			},
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
										Kind:        game.DynamicAmountX,
										Multiplier:  -1,
										CounterKind: counter.PlusOnePlusOne,
									}),
									ToughnessDeltaDynamic: opt.Val(game.DynamicAmount{
										Kind:        game.DynamicAmountX,
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
			As an additional cost to cast this spell, pay X life.
			All creatures get -X/-X until end of turn.
		`,
		},
	}
}
