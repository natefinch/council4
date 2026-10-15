package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PrecognitivePerception is the card definition for Precognitive Perception.
//
// Type: Instant
// Cost: {3}{U}{U}
//
// Oracle text:
//
//	Draw three cards.
//	Addendum — If you cast this spell during your main phase, instead scry 3, then draw three cards.
var PrecognitivePerception = newPrecognitivePerception

func newPrecognitivePerception() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Precognitive Perception",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(3),
							Player: game.ControllerReference(),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Negate:                        true,
								CastDuringControllerMainPhase: true,
							}),
						}),
						PublishCondition: "condition-0",
					},
					{
						Primitive: game.Scry{
							Amount: game.Fixed(3),
							Player: game.ControllerReference(),
						},
						ConditionGate:       "condition-0",
						ConditionGateNegate: true,
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(3),
							Player: game.ControllerReference(),
						},
						ConditionGate:       "condition-0",
						ConditionGateNegate: true,
					},
				},
			}.Ability()),
			OracleText: `
			Draw three cards.
			Addendum — If you cast this spell during your main phase, instead scry 3, then draw three cards.
		`,
		},
	}
}
