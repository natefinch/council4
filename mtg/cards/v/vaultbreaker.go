package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Vaultbreaker is the card definition for Vaultbreaker.
//
// Type: Creature — Orc Rogue
// Cost: {3}{R}
//
// Oracle text:
//
//	Whenever this creature attacks, you may discard a card. If you do, draw a card.
//	Dash {2}{R} (You may cast this spell for its dash cost. If you do, it gains haste, and it's returned from the battlefield to its owner's hand at the beginning of the next end step.)
var Vaultbreaker = newVaultbreaker

func newVaultbreaker() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Vaultbreaker",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Orc, types.Rogue},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
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
								Primitive: game.Discard{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
								Optional:      true,
								PublishResult: game.ResultKey("if-you-do"),
							},
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "if-you-do",
									Succeeded: game.TriTrue,
								}),
							},
						},
					}.Ability(),
				},
				game.DashTriggeredAbility(),
			},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:    "Dash",
					ManaCost: opt.Val(cost.Mana{cost.O(2), cost.R}),
					Mechanic: cost.AlternativeMechanicDash,
				},
			},
			OracleText: `
			Whenever this creature attacks, you may discard a card. If you do, draw a card.
			Dash {2}{R} (You may cast this spell for its dash cost. If you do, it gains haste, and it's returned from the battlefield to its owner's hand at the beginning of the next end step.)
		`,
		},
	}
}
