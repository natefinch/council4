package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CourtOfGarenbrig is the card definition for Court of Garenbrig.
//
// Type: Enchantment
// Cost: {1}{G}{G}
//
// Oracle text:
//
//	When this enchantment enters, you become the monarch.
//	At the beginning of your upkeep, distribute two +1/+1 counters among up to two target creatures. Then if you're the monarch, double the number of +1/+1 counters on each creature you control.
var CourtOfGarenbrig = newCourtOfGarenbrig()

func newCourtOfGarenbrig() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Court of Garenbrig",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
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
								Primitive: game.BecomeMonarch{
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepUpkeep,
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 0,
								MaxTargets: 2,
								Constraint: "up to two target creatures",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(2),
									Object:      game.AllTargetPermanentsReference(0),
									CounterKind: counter.PlusOnePlusOne,
									Distribute:  true,
								},
							},
							{
								Primitive: game.AddCounter{
									Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									CounterKind: counter.PlusOnePlusOne,
									DoubleKind:  true,
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										ControllerIsMonarch: true,
									}),
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this enchantment enters, you become the monarch.
			At the beginning of your upkeep, distribute two +1/+1 counters among up to two target creatures. Then if you're the monarch, double the number of +1/+1 counters on each creature you control.
		`,
		},
	}
}
