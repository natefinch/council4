package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PactOfNegation is the card definition for Pact of Negation.
//
// Type: Instant
// Cost: {0}
//
// Oracle text:
//
//	Counter target spell.
//	At the beginning of your next upkeep, pay {3}{U}{U}. If you don't, you lose the game.
var PactOfNegation = newPactOfNegation

func newPactOfNegation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Pact of Negation",
			ManaCost: opt.Val(cost.Mana{
				cost.O(0),
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target spell",
						Allow:      game.TargetAllowStackObject,
						Predicate: game.TargetPredicate{
							StackObjectKinds: []game.StackObjectKind{game.StackSpell},
						},
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.CounterObject{
							Object: game.TargetStackObjectReference(0),
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
															Prompt: "Pay {3}{U}{U}?",
															ManaCost: opt.Val(cost.Mana{
																cost.O(3),
																cost.U,
																cost.U,
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
			Counter target spell.
			At the beginning of your next upkeep, pay {3}{U}{U}. If you don't, you lose the game.
		`,
		},
	}
}
