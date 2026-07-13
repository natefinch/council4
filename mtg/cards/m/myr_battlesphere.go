package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MyrBattlesphere is the card definition for Myr Battlesphere.
//
// Type: Artifact Creature — Myr Construct
// Cost: {7}
//
// Oracle text:
//
//	When this creature enters, create four 1/1 colorless Myr artifact creature tokens.
//	Whenever this creature attacks, you may tap X untapped Myr you control. If you do, this creature gets +X/+0 until end of turn and deals X damage to the player or planeswalker it's attacking.
var MyrBattlesphere = newMyrBattlesphere

func newMyrBattlesphere() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Myr Battlesphere",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Myr, types.Construct},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 7}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(4),
									Source: game.TokenDef(myrBattlesphereToken),
								},
							},
						},
					}.Ability(),
				},
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
								Primitive: game.TapChosenGroup{
									ChooseFrom:   game.PlayerControlledGroup(game.ControllerReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Myr")}, Controller: game.ControllerYou, Tapped: game.TriFalse}),
									PublishCount: game.ResultKey("optional-tap-group-count"),
									Prompt:       "Tap any number of the matching untapped permanents you control.",
								},
								PublishResult: game.ResultKey("optional-tap-group-count"),
							},
							{
								Primitive: game.ModifyPT{
									Object: game.SourcePermanentReference(),
									PowerDelta: game.Dynamic(game.DynamicAmount{
										Kind:      game.DynamicAmountChosenNumber,
										ResultKey: game.ResultKey("optional-tap-group-count"),
									}),
									ToughnessDelta: game.Fixed(0),
									Duration:       game.DurationUntilEndOfTurn,
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "optional-tap-group-count",
									Succeeded: game.TriTrue,
								}),
							},
							{
								Primitive: game.Damage{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:      game.DynamicAmountChosenNumber,
										ResultKey: game.ResultKey("optional-tap-group-count"),
									}),
									Recipient:    game.AttackedDefenderDamageRecipient(),
									DamageSource: opt.Val(game.SourcePermanentReference()),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "optional-tap-group-count",
									Succeeded: game.TriTrue,
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this creature enters, create four 1/1 colorless Myr artifact creature tokens.
			Whenever this creature attacks, you may tap X untapped Myr you control. If you do, this creature gets +X/+0 until end of turn and deals X damage to the player or planeswalker it's attacking.
		`,
		},
	}
}

var myrBattlesphereToken = newMyrBattlesphereToken()

func newMyrBattlesphereToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Myr",
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Myr},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
