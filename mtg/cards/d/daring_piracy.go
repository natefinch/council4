package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// DaringPiracy is the card definition for Daring Piracy.
//
// Type: Enchantment
// Cost: {2}{R}
//
// Oracle text:
//
//	At the beginning of combat on your turn, create a 1/1 red Pirate creature token with menace and haste. Exile it at the beginning of the next end step.
var DaringPiracy = newDaringPiracy

func newDaringPiracy() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Daring Piracy",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Enchantment},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "At the beginning of combat on your turn, create a 1/1 red Pirate creature token with menace and haste. Exile it at the beginning of the next end step.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:           game.EventBeginningOfStep,
							Controller:      game.TriggerControllerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							Step:            game.StepBeginningOfCombat,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount:        game.Fixed(1),
											Source:        game.TokenDef(daringPiracyToken),
											PublishLinked: game.LinkedKey("sequence-effect-0-product"),
										},
										LocalProducts: game.LocalProducts{
											Links: []game.LinkedKey{game.LinkedKey("sequence-effect-0-product")},
										},
									},
									game.Instruction{
										Primitive: game.CreateDelayedTrigger{
											Trigger: game.DelayedTriggerDef{
												Timing: game.DelayedAtBeginningOfNextEndStep,
												Content: game.AbilityContent{
													Modes: []game.Mode{
														game.Mode{
															Sequence: []game.Instruction{
																game.Instruction{
																	Primitive: game.MovePermanent{
																		Group:       game.CapturedObjectsGroup(),
																		Destination: zone.Exile,
																	},
																},
															},
														},
													},
													MinModes: 1,
													MaxModes: 1,
												},
												CapturedObjectGroup: opt.Val(game.LinkedObjectReference("sequence-effect-0-product")),
											},
										},
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
			At the beginning of combat on your turn, create a 1/1 red Pirate creature token with menace and haste. Exile it at the beginning of the next end step.
		`,
		},
	}
}

var daringPiracyToken = newDaringPiracyToken()

func newDaringPiracyToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Pirate",
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Pirate},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.MenaceStaticBody,
				game.HasteStaticBody,
			},
		},
	}
}
