package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SylvanShepherd is the card definition for Sylvan Shepherd.
//
// Type: Creature — Human Druid
// Cost: {2}{G}
//
// Oracle text:
//
//	Vigilance
//	Whenever this creature attacks, roll a d20.
//	1—9 | You gain 1 life.
//	10—19 | You gain 2 life.
//	20 | You gain 5 life.
var SylvanShepherd = newSylvanShepherd

func newSylvanShepherd() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Sylvan Shepherd",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Druid},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever this creature attacks, roll a d20.\n1—9 | You gain 1 life.\n10—19 | You gain 2 life.\n20 | You gain 5 life.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventAttackerDeclared,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.RollDie{
											Sides: 20,
										},
										PublishResult: game.ResultKey("die-roll-result"),
									},
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
										ResultGate: opt.Val(game.InstructionResultGate{
											Key: game.ResultKey("die-roll-result"),
											AmountRange: opt.Val(game.IntRange{
												Min: 1,
												Max: 9,
											}),
										}),
									},
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Fixed(2),
											Player: game.ControllerReference(),
										},
										ResultGate: opt.Val(game.InstructionResultGate{
											Key: game.ResultKey("die-roll-result"),
											AmountRange: opt.Val(game.IntRange{
												Min: 10,
												Max: 19,
											}),
										}),
									},
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Fixed(5),
											Player: game.ControllerReference(),
										},
										ResultGate: opt.Val(game.InstructionResultGate{
											Key: game.ResultKey("die-roll-result"),
											AmountRange: opt.Val(game.IntRange{
												Min: 20,
												Max: 20,
											}),
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
			Vigilance
			Whenever this creature attacks, roll a d20.
			1—9 | You gain 1 life.
			10—19 | You gain 2 life.
			20 | You gain 5 life.
		`,
		},
	}
}
