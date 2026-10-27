// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RTOptionalGroup is the card definition for RT Optional Group.
//
// Type: Instant
//
// Oracle text:
//
//	If you have no cards in hand, you may draw a card and gain 2 life. You gain 1 life.
var RTOptionalGroup = newRTOptionalGroup

func newRTOptionalGroup() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "RT Optional Group",
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
						PublishCondition:        "condition-0",
						Optional:                true,
						PublishOptionalDecision: "optional-clause-1",
					},
					{
						Primitive: game.GainLife{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
						ConditionGate:        "condition-0",
						OptionalDecisionGate: "optional-clause-1",
					},
					{
						Primitive: game.GainLife{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			If you have no cards in hand, you may draw a card and gain 2 life. You gain 1 life.
		`,
		},
	}
}
