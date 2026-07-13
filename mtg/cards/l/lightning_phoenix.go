package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LightningPhoenix is the card definition for Lightning Phoenix.
//
// Type: Creature — Phoenix
// Cost: {2}{R}
//
// Oracle text:
//
//	Flying, haste
//	This creature can't block.
//	At the beginning of your end step, if an opponent was dealt 3 or more damage this turn, you may pay {R}. If you do, return this card from your graveyard to the battlefield.
var LightningPhoenix = newLightningPhoenix

func newLightningPhoenix() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Lightning Phoenix",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Phoenix},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.HasteStaticBody,
				game.CantBlockStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepEnd,
						},
						InterveningIf: "if an opponent was dealt 3 or more damage this turn",
						InterveningCondition: opt.Val(game.Condition{
							Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateAnyOpponentDamageTakenThisTurn, Op: compare.GreaterOrEqual, Value: 3}},
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Pay{
									Payment: game.ResolutionPayment{
										Prompt: "Pay {R}?",
										ManaCost: opt.Val(cost.Mana{
											cost.R,
										}),
									},
								},
								PublishResult: game.ResultKey("controller-paid"),
							},
							{
								Primitive: game.PutOnBattlefield{
									Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceSource}),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "controller-paid",
									Succeeded: game.TriTrue,
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying, haste
			This creature can't block.
			At the beginning of your end step, if an opponent was dealt 3 or more damage this turn, you may pay {R}. If you do, return this card from your graveyard to the battlefield.
		`,
		},
	}
}
