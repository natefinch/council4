package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ChaosChanneler is the card definition for Chaos Channeler.
//
// Type: Creature — Human Shaman
// Cost: {2}{R}{R}
//
// Oracle text:
//
//	Wild Magic Surge — Whenever this creature attacks, roll a d20.
//	1—9 | Exile the top card of your library. You may play it this turn.
//	10—19 | Exile the top two cards of your library. You may play them this turn.
//	20 | Exile the top three cards of your library. You may play them this turn.
var ChaosChanneler = newChaosChanneler

func newChaosChanneler() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Chaos Channeler",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Shaman},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Wild Magic Surge — Whenever this creature attacks, roll a d20.\n1—9 | Exile the top card of your library. You may play it this turn.\n10—19 | Exile the top two cards of your library. You may play them this turn.\n20 | Exile the top three cards of your library. You may play them this turn.",
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
										Primitive: game.ImpulseExile{
											Player:   game.ControllerReference(),
											Amount:   game.Fixed(1),
											Duration: game.DurationThisTurn,
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
										Primitive: game.ImpulseExile{
											Player:   game.ControllerReference(),
											Amount:   game.Fixed(2),
											Duration: game.DurationThisTurn,
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
										Primitive: game.ImpulseExile{
											Player:   game.ControllerReference(),
											Amount:   game.Fixed(3),
											Duration: game.DurationThisTurn,
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
			Wild Magic Surge — Whenever this creature attacks, roll a d20.
			1—9 | Exile the top card of your library. You may play it this turn.
			10—19 | Exile the top two cards of your library. You may play them this turn.
			20 | Exile the top three cards of your library. You may play them this turn.
		`,
		},
	}
}
