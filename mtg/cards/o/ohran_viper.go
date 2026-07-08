package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// OhranViper is the card definition for Ohran Viper.
//
// Type: Snow Creature — Snake
// Cost: {1}{G}{G}
//
// Oracle text:
//
//	Whenever this creature deals combat damage to a creature, destroy that creature at end of combat.
//	Whenever this creature deals combat damage to a player, you may draw a card.
var OhranViper = newOhranViper

func newOhranViper() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Ohran Viper",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Snow},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Snake},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                game.EventDamageDealt,
							Source:               game.TriggerSourceSelf,
							Subject:              game.TriggerSubjectDamageSource,
							RequireCombatDamage:  true,
							DamageRecipient:      game.DamageRecipientPermanent,
							DamageRecipientTypes: []types.Card{types.Creature},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateDelayedTrigger{
									Trigger: game.DelayedTriggerDef{
										Timing:         game.DelayedAtEndOfCombat,
										CapturedObject: opt.Val(game.EventPermanentReference()),
										Content: game.Mode{
											Sequence: []game.Instruction{
												{
													Primitive: game.Destroy{
														Object: game.CapturedObjectReference(),
													},
												},
											},
										}.Ability(),
									},
								},
							},
						},
					}.Ability(),
				},
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
					Optional: true,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever this creature deals combat damage to a creature, destroy that creature at end of combat.
			Whenever this creature deals combat damage to a player, you may draw a card.
		`,
		},
	}
}
