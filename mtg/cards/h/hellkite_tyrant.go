package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HellkiteTyrant is the card definition for Hellkite Tyrant.
//
// Type: Creature — Dragon
// Cost: {4}{R}{R}
//
// Oracle text:
//
//	Flying, trample
//	Whenever this creature deals combat damage to a player, gain control of all artifacts that player controls.
//	At the beginning of your upkeep, if you control twenty or more artifacts, you win the game.
var HellkiteTyrant = newHellkiteTyrant

func newHellkiteTyrant() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Hellkite Tyrant",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dragon},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
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
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:         game.LayerControl,
											NewController: opt.Val(game.Player1),
											Group:         game.PlayerControlledGroup(game.EventPlayerReference(), game.Selection{RequiredTypes: []types.Card{types.Artifact}}),
										},
									},
									Duration: game.DurationPermanent,
								},
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
							Step:       game.StepUpkeep,
						},
						InterveningIf: "if you control twenty or more artifacts",
						InterveningCondition: opt.Val(game.Condition{
							ControlsMatching: opt.Val(game.SelectionCount{
								Selection: game.Selection{RequiredTypes: []types.Card{types.Artifact}},
								MinCount:  20,
							}),
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.PlayerWinsGame{
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying, trample
			Whenever this creature deals combat damage to a player, gain control of all artifacts that player controls.
			At the beginning of your upkeep, if you control twenty or more artifacts, you win the game.
		`,
		},
	}
}
