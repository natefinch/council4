// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RTGatedMana is the card definition for RT Gated Mana.
//
// Type: Instant
//
// Oracle text:
//
//	If you have no cards in hand, draw a card, then add {R}{G}.
var RTGatedMana = newRTGatedMana

func newRTGatedMana() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "RT Gated Mana",
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
						Primitive: game.AddMana{
							Amount:    game.Fixed(1),
							ManaColor: mana.R,
						},
						ConditionGate: "condition-0",
					},
					{
						Primitive: game.AddMana{
							Amount:    game.Fixed(1),
							ManaColor: mana.G,
						},
						ConditionGate: "condition-0",
					},
				},
			}.Ability()),
			OracleText: `
			If you have no cards in hand, draw a card, then add {R}{G}.
		`,
		},
	}
}
