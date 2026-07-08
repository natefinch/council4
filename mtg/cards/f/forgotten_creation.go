package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ForgottenCreation is the card definition for Forgotten Creation.
//
// Type: Creature — Zombie Horror
// Cost: {3}{U}
//
// Oracle text:
//
//	Skulk (This creature can't be blocked by creatures with greater power.)
//	At the beginning of your upkeep, you may discard all the cards in your hand. If you do, draw that many cards.
var ForgottenCreation = newForgottenCreation

func newForgottenCreation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Forgotten Creation",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Horror},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.SkulkStaticBody,
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
			Skulk (This creature can't be blocked by creatures with greater power.)
			At the beginning of your upkeep, you may discard all the cards in your hand. If you do, draw that many cards.
		`,
		},
	}
}
