// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RTConditionGroup is the card definition for RT Condition Group.
//
// Type: Instant
//
// Oracle text:
//
//	If you have no cards in hand, draw a card, then draw a card.
var RTConditionGroup = newRTConditionGroup

func newRTConditionGroup() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "RT Condition Group",
			Types: []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								ControllerHandEmpty: true,
							}),
						}),
						PublishCondition: "condition-0",
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
						ConditionGate: "condition-0",
					},
				},
			}.Ability()),
			OracleText: `
			If you have no cards in hand, draw a card, then draw a card.
		`,
		},
	}
}
