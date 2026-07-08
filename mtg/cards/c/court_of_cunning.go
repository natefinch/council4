package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CourtOfCunning is the card definition for Court of Cunning.
//
// Type: Enchantment
// Cost: {1}{U}{U}
//
// Oracle text:
//
//	When this enchantment enters, you become the monarch.
//	At the beginning of your upkeep, any number of target players each mill two cards. If you're the monarch, each of those players mills ten cards instead. (To mill a card, a player puts the top card of their library into their graveyard.)
var CourtOfCunning = newCourtOfCunning

func newCourtOfCunning() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Court of Cunning",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Enchantment},
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
								Primitive: game.BecomeMonarch{
									Player: game.ControllerReference(),
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
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 0,
								MaxTargets: 99,
								Constraint: "any number of target players",
								Allow:      game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Mill{
									Amount:      game.Fixed(2),
									PlayerGroup: game.TargetedPlayersReference(),
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										Negate:              true,
										ControllerIsMonarch: true,
									}),
								}),
							},
							{
								Primitive: game.Mill{
									Amount:      game.Fixed(10),
									PlayerGroup: game.TargetedPlayersReference(),
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										ControllerIsMonarch: true,
									}),
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this enchantment enters, you become the monarch.
			At the beginning of your upkeep, any number of target players each mill two cards. If you're the monarch, each of those players mills ten cards instead. (To mill a card, a player puts the top card of their library into their graveyard.)
		`,
		},
	}
}
