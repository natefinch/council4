package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DjinniWindseer is the card definition for Djinni Windseer.
//
// Type: Creature — Djinn
// Cost: {3}{U}
//
// Oracle text:
//
//	Flying
//	When this creature enters, roll a d20.
//	1—9 | Scry 1.
//	10—19 | Scry 2.
//	20 | Scry 3.
var DjinniWindseer = newDjinniWindseer

func newDjinniWindseer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Djinni Windseer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Djinn},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature enters, roll a d20.\n1—9 | Scry 1.\n10—19 | Scry 2.\n20 | Scry 3.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
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
										Primitive: game.Scry{
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
										Primitive: game.Scry{
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
										Primitive: game.Scry{
											Amount: game.Fixed(3),
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
			Flying
			When this creature enters, roll a d20.
			1—9 | Scry 1.
			10—19 | Scry 2.
			20 | Scry 3.
		`,
		},
	}
}
