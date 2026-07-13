package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FlowerfootSwordmaster is the card definition for FlowerfootSwordmaster.
//
// Type: Creature — Mouse Soldier
// Cost: {W}
//
// Oracle text:
//
//	Offspring {2} (You may pay an additional {2} as you cast this spell. If you do, when this creature enters, create a 1/1 token copy of it.)
//	Valiant — Whenever this creature becomes the target of a spell or ability you control for the first time each turn, Mice you control get +1/+0 until end of turn.
var FlowerfootSwordmaster = newFlowerfootSwordmaster

func newFlowerfootSwordmaster() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Flowerfoot Swordmaster",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Mouse, types.Soldier},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.OffspringStaticAbility(cost.Mana{cost.O(2)}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.OffspringEnterTriggeredAbility(),
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventObjectBecameTarget,
							Source:          game.TriggerSourceSelf,
							CauseController: game.TriggerControllerYou,
						},
					},
					MaxTriggersPerTurn: 1,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:      game.LayerPowerToughnessModify,
											Group:      game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Mouse")}, Controller: game.ControllerYou}),
											PowerDelta: 1,
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Offspring {2} (You may pay an additional {2} as you cast this spell. If you do, when this creature enters, create a 1/1 token copy of it.)
			Valiant — Whenever this creature becomes the target of a spell or ability you control for the first time each turn, Mice you control get +1/+0 until end of turn.
		`,
		},
	}
}
