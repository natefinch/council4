package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CryptLurker is the card definition for Crypt Lurker.
//
// Type: Creature — Horror
// Cost: {3}{B}
//
// Oracle text:
//
//	When this creature enters, you may sacrifice a creature or discard a creature card. If you do, draw a card.
var CryptLurker = newCryptLurker()

func newCryptLurker() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Crypt Lurker",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Horror},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 4}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.SacrificePermanents{
									Amount:    game.Fixed(1),
									Player:    game.ControllerReference(),
									Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
								},
								Optional:      true,
								PublishResult: game.ResultKey("disjunctive-cost-a"),
							},
							{
								Primitive: game.ChooseDiscardFromHand{
									Player:    game.ControllerReference(),
									Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:      "disjunctive-cost-a",
									Accepted: game.TriFalse,
								}),
								Optional:      true,
								PublishResult: game.ResultKey("disjunctive-cost-b"),
							},
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "disjunctive-cost-a",
									Succeeded: game.TriTrue,
								}),
							},
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "disjunctive-cost-b",
									Succeeded: game.TriTrue,
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this creature enters, you may sacrifice a creature or discard a creature card. If you do, draw a card.
		`,
		},
	}
}
