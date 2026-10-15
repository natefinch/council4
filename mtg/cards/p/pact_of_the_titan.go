package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PactOfTheTitan is the card definition for Pact of the Titan.
//
// Type: Instant
// Cost: {0}
//
// Oracle text:
//
//	Create a 4/4 red Giant creature token.
//	At the beginning of your next upkeep, pay {4}{R}. If you don't, you lose the game.
var PactOfTheTitan = newPactOfTheTitan

func newPactOfTheTitan() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Pact of the Titan",
			ManaCost: opt.Val(cost.Mana{
				cost.O(0),
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenDef(pactOfTheTitanToken),
						},
					},
					{
						Primitive: game.CreateDelayedTrigger{
							Trigger: game.DelayedTriggerDef{
								Timing: game.DelayedAtBeginningOfNextUpkeep,
								Content: game.AbilityContent{
									Modes: []game.Mode{
										game.Mode{
											Sequence: []game.Instruction{
												game.Instruction{
													Primitive: game.Pay{
														Payment: game.ResolutionPayment{
															Prompt: "Pay {4}{R}?",
															ManaCost: opt.Val(cost.Mana{
																cost.O(4),
																cost.R,
															}),
														},
													},
													PublishResult: game.ResultKey("pact-paid"),
												},
												game.Instruction{
													Primitive: game.PlayerLosesGame{
														Player: game.ControllerReference(),
													},
													ResultGate: opt.Val(game.InstructionResultGate{
														Key:       game.ResultKey("pact-paid"),
														Succeeded: game.TriFalse,
													}),
												},
											},
										},
									},
									MinModes: 1,
									MaxModes: 1,
								},
							},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create a 4/4 red Giant creature token.
			At the beginning of your next upkeep, pay {4}{R}. If you don't, you lose the game.
		`,
		},
	}
}

var pactOfTheTitanToken = newPactOfTheTitanToken()

func newPactOfTheTitanToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Giant",
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Giant},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
		},
	}
}
