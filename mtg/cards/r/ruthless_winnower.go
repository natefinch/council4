package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RuthlessWinnower is the card definition for Ruthless Winnower.
//
// Type: Creature — Elf Rogue
// Cost: {3}{B}{B}
//
// Oracle text:
//
//	At the beginning of each player's upkeep, that player sacrifices a non-Elf creature of their choice.
var RuthlessWinnower = newRuthlessWinnower()

func newRuthlessWinnower() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Ruthless Winnower",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Rogue},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event: game.EventBeginningOfStep,
							Step:  game.StepUpkeep,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.SacrificePermanents{
									Amount:    game.Fixed(1),
									Player:    game.EventPlayerReference(),
									Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludedSubtype: types.Sub("Elf")},
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			At the beginning of each player's upkeep, that player sacrifices a non-Elf creature of their choice.
		`,
		},
	}
}
