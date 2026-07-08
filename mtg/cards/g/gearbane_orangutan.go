package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GearbaneOrangutan is the card definition for Gearbane Orangutan.
//
// Type: Creature — Ape
// Cost: {2}{R}
//
// Oracle text:
//
//	Reach
//	When this creature enters, choose one —
//	• Destroy up to one target artifact.
//	• Sacrifice an artifact. If you do, put two +1/+1 counters on this creature.
var GearbaneOrangutan = newGearbaneOrangutan

func newGearbaneOrangutan() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Gearbane Orangutan",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Ape},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Destroy up to one target artifact.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 0,
										MaxTargets: 1,
										Constraint: "up to one target artifact",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact}}),
									},
								},
								Sequence: []game.Instruction{
									{
										Primitive: game.Destroy{
											Object: game.TargetPermanentReference(0),
										},
									},
								},
							},
							game.Mode{
								Text: "Sacrifice an artifact. If you do, put two +1/+1 counters on this creature.",
								Sequence: []game.Instruction{
									{
										Primitive: game.SacrificePermanents{
											Amount:    game.Fixed(1),
											Player:    game.ControllerReference(),
											Selection: game.Selection{RequiredTypes: []types.Card{types.Artifact}},
										},
										PublishResult: game.ResultKey("if-you-do"),
									},
									{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(2),
											Object:      game.SourcePermanentReference(),
											CounterKind: counter.PlusOnePlusOne,
										},
										ResultGate: opt.Val(game.InstructionResultGate{
											Key:       "if-you-do",
											Succeeded: game.TriTrue,
										}),
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			Reach
			When this creature enters, choose one —
			• Destroy up to one target artifact.
			• Sacrifice an artifact. If you do, put two +1/+1 counters on this creature.
		`,
		},
	}
}
