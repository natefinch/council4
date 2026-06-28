package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BookDevourer is the card definition for Book Devourer.
//
// Type: Creature — Beast
// Cost: {5}{R}
//
// Oracle text:
//
//	Trample
//	Whenever this creature deals combat damage to a player, you may discard all the cards in your hand. If you do, draw that many cards.
var BookDevourer = newBookDevourer()

func newBookDevourer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Book Devourer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
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
						Text: "You may discard all the cards in your hand. If you do, draw that many cards.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Discard{
									EntireHand: true,
									Player:     game.ControllerReference(),
								},
								Optional:      true,
								PublishResult: game.ResultKey("wheel-discarded-this-way"),
							},
							{
								Primitive: game.Draw{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:      game.DynamicAmountPreviousEffectResult,
										ResultKey: game.ResultKey("wheel-discarded-this-way"),
									}),
									Player: game.ControllerReference(),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:      "wheel-discarded-this-way",
									Accepted: game.TriTrue,
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Trample
			Whenever this creature deals combat damage to a player, you may discard all the cards in your hand. If you do, draw that many cards.
		`,
		},
	}
}
