package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CapriciousSliver is the card definition for Capricious Sliver.
//
// Type: Creature — Sliver
// Cost: {3}{R}
//
// Oracle text:
//
//	Sliver creatures you control have "Whenever this creature deals combat damage to a player, exile the top card of your library. You may play that card this turn."
var CapriciousSliver = newCapriciousSliver

func newCapriciousSliver() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Capricious Sliver",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Sliver},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Sliver creatures you control have \"Whenever this creature deals combat damage to a player, exile the top card of your library. You may play that card this turn.\"",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Sliver")}}),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.TriggeredAbility{
									Text: "Whenever this creature deals combat damage to a player, exile the top card of your library. You may play that card this turn.",
									Trigger: game.TriggerCondition{
										Type: game.TriggerWhenever,
										Pattern: game.TriggerPattern{
											Event:               game.EventDamageDealt,
											Source:              game.TriggerSourceSelf,
											Subject:             game.TriggerSubjectDamageSource,
											StackObjectKind:     game.StackSpell,
											DamageRecipient:     game.DamageRecipientPlayer,
											RequireCombatDamage: true,
											CounterKind:         counter.PlusOnePlusOne,
										},
									},
									Content: game.AbilityContent{
										Modes: []game.Mode{
											game.Mode{
												Sequence: []game.Instruction{
													game.Instruction{
														Primitive: game.ImpulseExile{
															Player:   game.ControllerReference(),
															Amount:   game.Fixed(1),
															Duration: game.DurationThisTurn,
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
						},
					},
				},
			},
			OracleText: `
			Sliver creatures you control have "Whenever this creature deals combat damage to a player, exile the top card of your library. You may play that card this turn."
		`,
		},
	}
}
