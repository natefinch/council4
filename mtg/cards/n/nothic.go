package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Nothic is the card definition for Nothic.
//
// Type: Creature — Horror
// Cost: {4}{B}
//
// Oracle text:
//
//	Weird Insight — When this creature dies, roll a d20.
//	1—9 | You draw a card and you lose 1 life.
//	10—19 | You draw two cards and you lose 2 life.
//	20 | You draw seven cards and you lose 7 life.
var Nothic = newNothic

func newNothic() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Nothic",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Horror},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Weird Insight — When this creature dies, roll a d20.\n1—9 | You draw a card and you lose 1 life.\n10—19 | You draw two cards and you lose 2 life.\n20 | You draw seven cards and you lose 7 life.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							Source:           game.TriggerSourceSelf,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
							StackObjectKind:  game.StackSpell,
							CounterKind:      counter.PlusOnePlusOne,
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
										Primitive: game.Draw{
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
										Primitive: game.LoseLife{
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
										Primitive: game.Draw{
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
										Primitive: game.LoseLife{
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
										Primitive: game.Draw{
											Amount: game.Fixed(7),
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
									game.Instruction{
										Primitive: game.LoseLife{
											Amount: game.Fixed(7),
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
			Weird Insight — When this creature dies, roll a d20.
			1—9 | You draw a card and you lose 1 life.
			10—19 | You draw two cards and you lose 2 life.
			20 | You draw seven cards and you lose 7 life.
		`,
		},
	}
}
