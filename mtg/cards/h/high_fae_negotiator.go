package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HighFaeNegotiator is the card definition for High Fae Negotiator.
//
// Type: Creature — Faerie Warlock
// Cost: {3}{B}{B}
//
// Oracle text:
//
//	Bargain (You may sacrifice an artifact, enchantment, or token as you cast this spell.)
//	Flying
//	When this creature enters, if it was bargained, each opponent loses 3 life and you gain 3 life.
var HighFaeNegotiator = newHighFaeNegotiator

func newHighFaeNegotiator() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "High Fae Negotiator",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Faerie, types.Warlock},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.BargainStaticBody,
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
						InterveningIf:                           "if it was bargained",
						InterveningIfEventPermanentWasBargained: true,
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.LoseLife{
									Amount:      game.Fixed(3),
									PlayerGroup: game.OpponentsReference(),
								},
								PublishResult: game.ResultKey("life-change"),
							},
							{
								Primitive: game.GainLife{
									Amount: game.Fixed(3),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Bargain (You may sacrifice an artifact, enchantment, or token as you cast this spell.)
			Flying
			When this creature enters, if it was bargained, each opponent loses 3 life and you gain 3 life.
		`,
		},
	}
}
