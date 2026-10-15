package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TemperedSliver is the card definition for Tempered Sliver.
//
// Type: Creature — Sliver
// Cost: {2}{G}
//
// Oracle text:
//
//	Sliver creatures you control have "Whenever this creature deals combat damage to a player, put a +1/+1 counter on it."
var TemperedSliver = newTemperedSliver

func newTemperedSliver() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Tempered Sliver",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Sliver},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Sliver creatures you control have \"Whenever this creature deals combat damage to a player, put a +1/+1 counter on it.\"",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Sliver")}}),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.TriggeredAbility{
									Text: "Whenever this creature deals combat damage to a player, put a +1/+1 counter on it.",
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
														Primitive: game.AddCounter{
															Amount:      game.Fixed(1),
															Object:      game.EventPermanentReference(),
															CounterKind: counter.PlusOnePlusOne,
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
			Sliver creatures you control have "Whenever this creature deals combat damage to a player, put a +1/+1 counter on it."
		`,
		},
	}
}
