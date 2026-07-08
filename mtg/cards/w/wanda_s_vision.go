package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WandaSVision is the card definition for Wanda's Vision.
//
// Type: Enchantment
// Cost: {3}{R}{R}
//
// Oracle text:
//
//	Whenever you cast your second spell each turn, exile cards from the top of your library until you exile a nonland card. You may cast that card without paying its mana cost.
var WandaSVision = newWandaSVision

func newWandaSVision() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Wanda's Vision",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Enchantment},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                      game.EventSpellCast,
							Controller:                 game.TriggerControllerYou,
							PlayerEventOrdinalThisTurn: 2,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ExileLibraryUntilNonlandCast{
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever you cast your second spell each turn, exile cards from the top of your library until you exile a nonland card. You may cast that card without paying its mana cost.
		`,
		},
	}
}
