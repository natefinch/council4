package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GoblinPyromancer is the card definition for Goblin Pyromancer.
//
// Type: Creature — Goblin Wizard
// Cost: {3}{R}
//
// Oracle text:
//
//	When this creature enters, Goblin creatures get +3/+0 until end of turn.
//	At the beginning of the end step, destroy all Goblins.
var GoblinPyromancer = newGoblinPyromancer

func newGoblinPyromancer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Goblin Pyromancer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin, types.Wizard},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:      game.LayerPowerToughnessModify,
											Group:      game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Goblin")}}),
											PowerDelta: 3,
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event: game.EventBeginningOfStep,
							Step:  game.StepEnd,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Destroy{
									Group: game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Goblin")}}),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this creature enters, Goblin creatures get +3/+0 until end of turn.
			At the beginning of the end step, destroy all Goblins.
		`,
		},
	}
}
