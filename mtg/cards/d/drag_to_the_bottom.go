package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DragToTheBottom is the card definition for Drag to the Bottom.
//
// Type: Sorcery
// Cost: {2}{B}{B}
//
// Oracle text:
//
//	Domain — Each creature gets -X/-X until end of turn, where X is 1 plus the number of basic land types among lands you control.
var DragToTheBottom = newDragToTheBottom

func newDragToTheBottom() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Drag to the Bottom",
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
										Kind:        game.DynamicAmountControllerBasicLandTypeCount,
										Multiplier:  -1,
										Addend:      1,
										CounterKind: counter.PlusOnePlusOne,
									}),
									ToughnessDeltaDynamic: opt.Val(game.DynamicAmount{
										Kind:        game.DynamicAmountControllerBasicLandTypeCount,
										Multiplier:  -1,
										Addend:      1,
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
			Domain — Each creature gets -X/-X until end of turn, where X is 1 plus the number of basic land types among lands you control.
		`,
		},
	}
}
