package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PitilessHorde is the card definition for Pitiless Horde.
//
// Type: Creature — Orc Berserker
// Cost: {2}{B}
//
// Oracle text:
//
//	At the beginning of your upkeep, you lose 2 life.
//	Dash {2}{B}{B} (You may cast this spell for its dash cost. If you do, it gains haste, and it's returned from the battlefield to its owner's hand at the beginning of the next end step.)
var PitilessHorde = newPitilessHorde

func newPitilessHorde() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Pitiless Horde",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Orc, types.Berserker},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 3}),
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
						Sequence: []game.Instruction{
							{
								Primitive: game.LoseLife{
									Amount: game.Fixed(2),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
				game.DashTriggeredAbility(),
			},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:    "Dash",
					ManaCost: opt.Val(cost.Mana{cost.O(2), cost.B, cost.B}),
					Mechanic: cost.AlternativeMechanicDash,
				},
			},
			OracleText: `
			At the beginning of your upkeep, you lose 2 life.
			Dash {2}{B}{B} (You may cast this spell for its dash cost. If you do, it gains haste, and it's returned from the battlefield to its owner's hand at the beginning of the next end step.)
		`,
		},
	}
}
