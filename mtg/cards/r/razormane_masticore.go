package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RazormaneMasticore is the card definition for Razormane Masticore.
//
// Type: Artifact Creature — Masticore
// Cost: {5}
//
// Oracle text:
//
//	First strike (This creature deals combat damage before creatures without first strike.)
//	At the beginning of your upkeep, sacrifice this creature unless you discard a card.
//	At the beginning of your draw step, you may have this creature deal 3 damage to target creature.
var RazormaneMasticore = newRazormaneMasticore()

func newRazormaneMasticore() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Razormane Masticore",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Masticore},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepUpkeep,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Pay{
									Payment: game.ResolutionPayment{
										Prompt: "Discard a card?",
										AdditionalCosts: []cost.Additional{
											{
												Kind:   cost.AdditionalDiscard,
												Text:   "discard a card",
												Amount: 1,
												Source: zone.Hand,
											},
										},
									},
								},
								PublishResult: game.ResultKey("sacrifice-unless-paid"),
							},
							{
								Primitive: game.Sacrifice{
									Object: game.SourcePermanentReference(),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "sacrifice-unless-paid",
									Succeeded: game.TriFalse,
								}),
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepDraw,
						},
					},
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
								Primitive: game.Damage{
									Amount:       game.Fixed(3),
									Recipient:    game.AnyTargetDamageRecipient(0),
									DamageSource: opt.Val(game.SourcePermanentReference()),
								},
								Optional: true,
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			First strike (This creature deals combat damage before creatures without first strike.)
			At the beginning of your upkeep, sacrifice this creature unless you discard a card.
			At the beginning of your draw step, you may have this creature deal 3 damage to target creature.
		`,
		},
	}
}
