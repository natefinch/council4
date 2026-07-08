package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AntManColonyCommander is the card definition for Ant-Man, Colony Commander.
//
// Type: Legendary Creature — Human Rogue Hero
// Cost: {1}{G}{U}
//
// Oracle text:
//
//	Whenever Ant-Man attacks, you may pay {1}. When you do, put a +1/+1 counter on target creature.
//	Whenever you put a +1/+1 counter on a creature, create a 1/1 green Insect creature token. This ability triggers only once each turn.
var AntManColonyCommander = newAntManColonyCommander()

func newAntManColonyCommander() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Ant-Man, Colony Commander",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.U,
			}),
			Colors:     []color.Color{color.Green, color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Rogue, types.Hero},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
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
										Prompt: "Pay {1}?",
										ManaCost: opt.Val(cost.Mana{
											cost.O(1),
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
													Primitive: game.AddCounter{
														Amount:      game.Fixed(1),
														Object:      game.TargetPermanentReference(0),
														CounterKind: counter.PlusOnePlusOne,
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
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventCountersAdded,
							CauseController:  game.TriggerControllerYou,
							MatchCounterKind: true,
							CounterKind:      counter.PlusOnePlusOne,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					MaxTriggersPerTurn: 1,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(antManColonyCommanderToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever Ant-Man attacks, you may pay {1}. When you do, put a +1/+1 counter on target creature.
			Whenever you put a +1/+1 counter on a creature, create a 1/1 green Insect creature token. This ability triggers only once each turn.
		`,
		},
	}
}

var antManColonyCommanderToken = newAntManColonyCommanderToken()

func newAntManColonyCommanderToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Insect",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Insect},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
