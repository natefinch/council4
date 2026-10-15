package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BearerOfSilence is the card definition for Bearer of Silence.
//
// Type: Creature — Eldrazi
// Cost: {1}{B}
//
// Oracle text:
//
//	Devoid (This card has no color.)
//	When you cast this spell, you may pay {1}{C}. If you do, target opponent sacrifices a creature. ({C} represents colorless mana.)
//	Flying
//	This creature can't block.
var BearerOfSilence = newBearerOfSilence

func newBearerOfSilence() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Bearer of Silence",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.DevoidStaticBody,
				game.FlyingStaticBody,
				game.CantBlockStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When you cast this spell, you may pay {1}{C}. If you do, target opponent sacrifices a creature. ({C} represents colorless mana.)",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventSpellCast,
							Controller:      game.TriggerControllerYou,
							Source:          game.TriggerSourceSelf,
							SelfWasCast:     true,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target opponent",
										Allow:      game.TargetAllowPlayer,
										Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Pay{
											Payment: game.ResolutionPayment{
												Prompt: "Pay {1}{C}?",
												ManaCost: opt.Val(cost.Mana{
													cost.O(1),
													cost.C,
												}),
											},
										},
										PublishResult: game.ResultKey("controller-paid"),
									},
									game.Instruction{
										Primitive: game.SacrificePermanents{
											Player:    game.TargetPlayerReference(0),
											Amount:    game.Fixed(1),
											Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
										},
										ResultGate: opt.Val(game.InstructionResultGate{
											Key:       game.ResultKey("controller-paid"),
											Succeeded: game.TriTrue,
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
			Devoid (This card has no color.)
			When you cast this spell, you may pay {1}{C}. If you do, target opponent sacrifices a creature. ({C} represents colorless mana.)
			Flying
			This creature can't block.
		`,
		},
	}
}
