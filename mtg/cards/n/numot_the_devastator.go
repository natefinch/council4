package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NumotTheDevastator is the card definition for Numot, the Devastator.
//
// Type: Legendary Creature — Dragon
// Cost: {3}{U}{R}{W}
//
// Oracle text:
//
//	Flying
//	Whenever Numot deals combat damage to a player, you may pay {2}{R}. If you do, destroy up to two target lands.
var NumotTheDevastator = newNumotTheDevastator

func newNumotTheDevastator() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Numot, the Devastator",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
				cost.R,
				cost.W,
			}),
			Colors:     []color.Color{color.Red, color.Blue, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Dragon},
			Power:      opt.Val(game.PT{Value: 6}),
			Toughness:  opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:               game.EventDamageDealt,
							Source:              game.TriggerSourceSelf,
							Subject:             game.TriggerSubjectDamageSource,
							RequireCombatDamage: true,
							DamageRecipient:     game.DamageRecipientPlayer,
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 0,
								MaxTargets: 2,
								Constraint: "up to two target lands",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Land}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Pay{
									Payment: game.ResolutionPayment{
										Prompt: "Pay {2}{R}?",
										ManaCost: opt.Val(cost.Mana{
											cost.O(2),
											cost.R,
										}),
									},
								},
								PublishResult: game.ResultKey("controller-paid"),
							},
							{
								Primitive: game.Destroy{
									Object: game.TargetPermanentReference(0),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "controller-paid",
									Succeeded: game.TriTrue,
								}),
							},
							{
								Primitive: game.Destroy{
									Object: game.TargetPermanentReference(1),
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
			Flying
			Whenever Numot deals combat damage to a player, you may pay {2}{R}. If you do, destroy up to two target lands.
		`,
		},
	}
}
