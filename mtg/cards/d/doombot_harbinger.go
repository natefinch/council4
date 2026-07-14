package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// DoombotHarbinger is the card definition for Doombot Harbinger.
//
// Type: Artifact Creature — Robot Villain
// Cost: {2}{B}
//
// Oracle text:
//
//	Flying
//	When this creature enters, you may mill four cards. (You may put the top four cards of your library into your graveyard.)
//	When this creature dies, you may exile this card. When you do, return target creature card from your graveyard to your hand.
var DoombotHarbinger = newDoombotHarbinger

func newDoombotHarbinger() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Doombot Harbinger",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Robot, types.Villain},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
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
					Optional: true,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Mill{
									Amount: game.Fixed(4),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							Source:           game.TriggerSourceSelf,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Exile{
									Object: game.SourceCardPermanentReference(),
								},
								Optional:      true,
								PublishResult: game.ResultKey("if-you-do"),
							},
							{
								Primitive: game.CreateReflexiveTrigger{
									Trigger: game.ReflexiveTriggerDef{
										Content: game.Mode{
											Targets: []game.TargetSpec{
												game.TargetSpec{
													MinTargets: 1,
													MaxTargets: 1,
													Constraint: "target creature card from your graveyard",
													Allow:      game.TargetAllowCard,
													TargetZone: zone.Graveyard,
													Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
												},
											},
											Sequence: []game.Instruction{
												{
													Primitive: game.MoveCard{
														Card:        game.CardReference{Kind: game.CardReferenceTarget},
														FromZone:    zone.Graveyard,
														Destination: zone.Hand,
													},
												},
											},
										}.Ability(),
									},
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "if-you-do",
									Succeeded: game.TriTrue,
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying
			When this creature enters, you may mill four cards. (You may put the top four cards of your library into your graveyard.)
			When this creature dies, you may exile this card. When you do, return target creature card from your graveyard to your hand.
		`,
		},
	}
}
