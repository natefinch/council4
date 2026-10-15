package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SlaughterPact is the card definition for Slaughter Pact.
//
// Type: Instant
// Cost: {0}
//
// Oracle text:
//
//	Destroy target nonblack creature.
//	At the beginning of your next upkeep, pay {2}{B}. If you don't, you lose the game.
var SlaughterPact = newSlaughterPact

func newSlaughterPact() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Slaughter Pact",
			ManaCost: opt.Val(cost.Mana{
				cost.O(0),
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target nonblack creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, ExcludedColors: []color.Color{color.Black}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(0),
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
															Prompt: "Pay {2}{B}?",
															ManaCost: opt.Val(cost.Mana{
																cost.O(2),
																cost.B,
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
			Destroy target nonblack creature.
			At the beginning of your next upkeep, pay {2}{B}. If you don't, you lose the game.
		`,
		},
	}
}
