package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ThousandMoonsCrackshot is the card definition for Thousand Moons Crackshot.
//
// Type: Creature — Human Soldier
// Cost: {1}{W}
//
// Oracle text:
//
//	Whenever this creature attacks, you may pay {2}{W}. When you do, tap target creature.
var ThousandMoonsCrackshot = newThousandMoonsCrackshot

func newThousandMoonsCrackshot() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Thousand Moons Crackshot",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Soldier},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
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
								Primitive: game.Pay{
									Payment: game.ResolutionPayment{
										Prompt: "Pay {2}{W}?",
										ManaCost: opt.Val(cost.Mana{
											cost.O(2),
											cost.W,
										}),
									},
								},
								PublishResult: game.ResultKey("controller-paid"),
							},
							{
								Primitive: game.CreateReflexiveTrigger{
									Trigger: game.ReflexiveTriggerDef{
										Content: game.Mode{
											Targets: []game.TargetSpec{
												game.TargetSpec{
													MinTargets: 1,
													MaxTargets: 1,
													Constraint: "target creature",
													Allow:      game.TargetAllowPermanent,
													Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
												},
											},
											Sequence: []game.Instruction{
												{
													Primitive: game.Tap{
														Object: game.TargetPermanentReference(0),
													},
												},
											},
										}.Ability(),
									},
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "controller-paid",
									Succeeded: game.TriTrue,
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever this creature attacks, you may pay {2}{W}. When you do, tap target creature.
		`,
		},
	}
}
