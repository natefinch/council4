package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// VaevictisAsmadi is the card definition for Vaevictis Asmadi.
//
// Type: Legendary Creature — Elder Dragon
// Cost: {2}{B}{B}{R}{R}{G}{G}
//
// Oracle text:
//
//	Flying
//	At the beginning of your upkeep, sacrifice Vaevictis Asmadi unless you pay {B}{R}{G}.
//	{B}: Vaevictis Asmadi gets +1/+0 until end of turn.
//	{R}: Vaevictis Asmadi gets +1/+0 until end of turn.
//	{G}: Vaevictis Asmadi gets +1/+0 until end of turn.
var VaevictisAsmadi = newVaevictisAsmadi

func newVaevictisAsmadi() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Vaevictis Asmadi",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.B,
				cost.R,
				cost.R,
				cost.G,
				cost.G,
			}),
			Colors:     []color.Color{color.Black, color.Green, color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Elder, types.Dragon},
			Power:      opt.Val(game.PT{Value: 7}),
			Toughness:  opt.Val(game.PT{Value: 7}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{B}: Vaevictis Asmadi gets +1/+0 until end of turn.",
					ManaCost: opt.Val(cost.Mana{
						cost.B,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ModifyPT{
											Object:     game.SourcePermanentReference(),
											PowerDelta: game.Fixed(1),
											Duration:   game.DurationUntilEndOfTurn,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.ActivatedAbility{
					Text: "{R}: Vaevictis Asmadi gets +1/+0 until end of turn.",
					ManaCost: opt.Val(cost.Mana{
						cost.R,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ModifyPT{
											Object:     game.SourcePermanentReference(),
											PowerDelta: game.Fixed(1),
											Duration:   game.DurationUntilEndOfTurn,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.ActivatedAbility{
					Text: "{G}: Vaevictis Asmadi gets +1/+0 until end of turn.",
					ManaCost: opt.Val(cost.Mana{
						cost.G,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ModifyPT{
											Object:     game.SourcePermanentReference(),
											PowerDelta: game.Fixed(1),
											Duration:   game.DurationUntilEndOfTurn,
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
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "At the beginning of your upkeep, sacrifice Vaevictis Asmadi unless you pay {B}{R}{G}.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:           game.EventBeginningOfStep,
							Controller:      game.TriggerControllerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							Step:            game.StepUpkeep,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Pay{
											Payment: game.ResolutionPayment{
												Prompt: "Pay {B}{R}{G}?",
												ManaCost: opt.Val(cost.Mana{
													cost.B,
													cost.R,
													cost.G,
												}),
											},
										},
										PublishResult: game.ResultKey("sacrifice-unless-paid"),
									},
									game.Instruction{
										Primitive: game.Sacrifice{
											Object: game.SourcePermanentReference(),
										},
										ResultGate: opt.Val(game.InstructionResultGate{
											Key:       game.ResultKey("sacrifice-unless-paid"),
											Succeeded: game.TriFalse,
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
			At the beginning of your upkeep, sacrifice Vaevictis Asmadi unless you pay {B}{R}{G}.
			{B}: Vaevictis Asmadi gets +1/+0 until end of turn.
			{R}: Vaevictis Asmadi gets +1/+0 until end of turn.
			{G}: Vaevictis Asmadi gets +1/+0 until end of turn.
		`,
		},
	}
}
