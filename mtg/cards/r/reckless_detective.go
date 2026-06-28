package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RecklessDetective is the card definition for Reckless Detective.
//
// Type: Creature — Devil Detective
// Cost: {1}{R}
//
// Oracle text:
//
//	Whenever this creature attacks, you may sacrifice an artifact or discard a card. If you do, draw a card and this creature gets +2/+0 until end of turn.
var RecklessDetective = newRecklessDetective()

func newRecklessDetective() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Reckless Detective",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Devil, types.Detective},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventAttackerDeclared,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.SacrificePermanents{
									Amount:    game.Fixed(1),
									Player:    game.ControllerReference(),
									Selection: game.Selection{RequiredTypes: []types.Card{types.Artifact}},
								},
								Optional:      true,
								PublishResult: game.ResultKey("disjunctive-cost-a"),
							},
							{
								Primitive: game.Discard{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:      "disjunctive-cost-a",
									Accepted: game.TriFalse,
								}),
								Optional:      true,
								PublishResult: game.ResultKey("disjunctive-cost-b"),
							},
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "disjunctive-cost-a",
									Succeeded: game.TriTrue,
								}),
							},
							{
								Primitive: game.ModifyPT{
									Object:         game.SourcePermanentReference(),
									PowerDelta:     game.Fixed(2),
									ToughnessDelta: game.Fixed(0),
									Duration:       game.DurationUntilEndOfTurn,
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "disjunctive-cost-a",
									Succeeded: game.TriTrue,
								}),
							},
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "disjunctive-cost-b",
									Succeeded: game.TriTrue,
								}),
							},
							{
								Primitive: game.ModifyPT{
									Object:         game.SourcePermanentReference(),
									PowerDelta:     game.Fixed(2),
									ToughnessDelta: game.Fixed(0),
									Duration:       game.DurationUntilEndOfTurn,
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "disjunctive-cost-b",
									Succeeded: game.TriTrue,
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever this creature attacks, you may sacrifice an artifact or discard a card. If you do, draw a card and this creature gets +2/+0 until end of turn.
		`,
		},
	}
}
