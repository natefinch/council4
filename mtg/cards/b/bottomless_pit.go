package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BottomlessPit is the card definition for Bottomless Pit.
//
// Type: Enchantment
// Cost: {1}{B}{B}
//
// Oracle text:
//
//	At the beginning of each player's upkeep, that player discards a card at random.
var BottomlessPit = newBottomlessPit()

func newBottomlessPit() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Bottomless Pit",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Enchantment},
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
								Primitive: game.Discard{
									Amount:   game.Fixed(1),
									Player:   game.EventPlayerReference(),
									AtRandom: true,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			At the beginning of each player's upkeep, that player discards a card at random.
		`,
		},
	}
}
