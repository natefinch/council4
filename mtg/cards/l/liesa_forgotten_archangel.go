package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// LiesaForgottenArchangel is the card definition for Liesa, Forgotten Archangel.
//
// Type: Legendary Creature — Angel
// Cost: {2}{W}{W}{B}
//
// Oracle text:
//
//	Flying, lifelink
//	Whenever another nontoken creature you control dies, return that card to its owner's hand at the beginning of the next end step.
//	If a creature an opponent controls would die, exile it instead.
var LiesaForgottenArchangel = newLiesaForgottenArchangel

func newLiesaForgottenArchangel() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name: "Liesa, Forgotten Archangel",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.W,
				cost.B,
			}),
			Colors:     []color.Color{color.Black, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Angel},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.LifelinkStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever another nontoken creature you control dies, return that card to its owner's hand at the beginning of the next end step.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							Controller:       game.TriggerControllerYou,
							ExcludeSelf:      true,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}, NonToken: true},
							StackObjectKind:  game.StackSpell,
							CounterKind:      counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateDelayedTrigger{
											Trigger: game.DelayedTriggerDef{
												Timing: game.DelayedAtBeginningOfNextEndStep,
												Content: game.AbilityContent{
													Modes: []game.Mode{
														game.Mode{
															Sequence: []game.Instruction{
																game.Instruction{
																	Primitive: game.MoveCard{
																		Card: game.CardReference{
																			Kind: game.CardReferenceCaptured,
																		},
																		FromZone:    zone.Graveyard,
																		Destination: zone.Hand,
																	},
																},
															},
														},
													},
													MinModes: 1,
													MaxModes: 1,
												},
												CapturedCard: opt.Val(game.EventPermanentReference()),
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
			ReplacementAbilities: []game.ReplacementAbility{
				game.GraveyardRedirectReplacement("If a creature an opponent controls would die, exile it instead.", game.TriggerControllerAny, game.TriggerControllerOpponent, true, types.Creature),
			},
			OracleText: `
			Flying, lifelink
			Whenever another nontoken creature you control dies, return that card to its owner's hand at the beginning of the next end step.
			If a creature an opponent controls would die, exile it instead.
		`,
		},
	}
}
