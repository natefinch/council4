package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CombustibleGearhulk is the card definition for Combustible Gearhulk.
//
// Type: Artifact Creature — Construct
// Cost: {4}{R}{R}
//
// Oracle text:
//
//	First strike
//	When this creature enters, target opponent may have you draw three cards. If the player doesn't, you mill three cards, then this creature deals damage to that player equal to the total mana value of those cards.
var CombustibleGearhulk = newCombustibleGearhulk

func newCombustibleGearhulk() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Combustible Gearhulk",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Construct},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature enters, target opponent may have you draw three cards. If the player doesn't, you mill three cards, then this creature deals damage to that player equal to the total mana value of those cards.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
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
										Primitive: game.Draw{
											Amount: game.Fixed(3),
											Player: game.ControllerReference(),
										},
										Optional:      true,
										OptionalActor: opt.Val(game.TargetPlayerReference(0)),
										PublishResult: game.ResultKey("may-have-action"),
									},
									game.Instruction{
										Primitive: game.MoveTopOfLibrary{
											Amount:        game.Fixed(3),
											Player:        game.ControllerReference(),
											Destination:   zone.Graveyard,
											PublishLinked: game.LinkedKey("may-have-milled-cards"),
										},
										ResultGate: opt.Val(game.InstructionResultGate{
											Key:      game.ResultKey("may-have-action"),
											Accepted: game.TriFalse,
										}),
									},
									game.Instruction{
										Primitive: game.Damage{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountReferencedCardsTotalManaValue,
												Multiplier: 1,
												LinkedKey:  game.LinkedKey("may-have-milled-cards"),
											}),
											Recipient:    game.PlayerDamageRecipient(game.TargetPlayerReference(0)),
											DamageSource: opt.Val(game.SourcePermanentReference()),
										},
										ResultGate: opt.Val(game.InstructionResultGate{
											Key:      game.ResultKey("may-have-action"),
											Accepted: game.TriFalse,
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
			First strike
			When this creature enters, target opponent may have you draw three cards. If the player doesn't, you mill three cards, then this creature deals damage to that player equal to the total mana value of those cards.
		`,
		},
	}
}
