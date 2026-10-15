package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Mutilate is the card definition for Mutilate.
//
// Type: Sorcery
// Cost: {2}{B}{B}
//
// Oracle text:
//
//	All creatures get -1/-1 until end of turn for each Swamp you control.
var Mutilate = newMutilate

func newMutilate() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Mutilate",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
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
										Kind:        game.DynamicAmountCountSelector,
										Multiplier:  -1,
										CounterKind: counter.PlusOnePlusOne,
										Group:       game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Swamp")}, Controller: game.ControllerYou}),
									}),
									ToughnessDeltaDynamic: opt.Val(game.DynamicAmount{
										Kind:        game.DynamicAmountCountSelector,
										Multiplier:  -1,
										CounterKind: counter.PlusOnePlusOne,
										Group:       game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Swamp")}, Controller: game.ControllerYou}),
									}),
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			All creatures get -1/-1 until end of turn for each Swamp you control.
		`,
		},
	}
}
