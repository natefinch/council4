package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EarthCultElemental is the card definition for Earth-Cult Elemental.
//
// Type: Creature — Elemental
// Cost: {4}{R}{R}
//
// Oracle text:
//
//	Siege Monster — When this creature enters, roll a d20.
//	1—9 | Each player sacrifices a permanent of their choice.
//	10—19 | Each opponent sacrifices a permanent of their choice.
//	20 | Each opponent sacrifices two permanents of their choice.
var EarthCultElemental = newEarthCultElemental

func newEarthCultElemental() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Earth-Cult Elemental",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Siege Monster — When this creature enters, roll a d20.\n1—9 | Each player sacrifices a permanent of their choice.\n10—19 | Each opponent sacrifices a permanent of their choice.\n20 | Each opponent sacrifices two permanents of their choice.",
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
										Primitive: game.SacrificePermanents{
											PlayerGroup: game.AllPlayersReference(),
											Amount:      game.Fixed(1),
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
										Primitive: game.SacrificePermanents{
											PlayerGroup: game.OpponentsReference(),
											Amount:      game.Fixed(1),
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
										Primitive: game.SacrificePermanents{
											PlayerGroup: game.OpponentsReference(),
											Amount:      game.Fixed(2),
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
			Siege Monster — When this creature enters, roll a d20.
			1—9 | Each player sacrifices a permanent of their choice.
			10—19 | Each opponent sacrifices a permanent of their choice.
			20 | Each opponent sacrifices two permanents of their choice.
		`,
		},
	}
}
