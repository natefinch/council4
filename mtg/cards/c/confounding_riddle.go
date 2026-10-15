package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ConfoundingRiddle is the card definition for Confounding Riddle.
//
// Type: Instant
// Cost: {2}{U}
//
// Oracle text:
//
//	Choose one —
//	• Look at the top four cards of your library. Put one of them into your hand and the rest into your graveyard.
//	• Counter target spell unless its controller pays {4}.
var ConfoundingRiddle = newConfoundingRiddle

func newConfoundingRiddle() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Confounding Riddle",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Look at the top four cards of your library. Put one of them into your hand and the rest into your graveyard.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Dig{
									Player: game.ControllerReference(),
									Look:   game.Fixed(4),
									Take:   game.Fixed(1),
								},
							},
						},
					},
					game.Mode{
						Text: "Counter target spell unless its controller pays {4}.",
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
								Primitive: game.Pay{
									Payment: game.ResolutionPayment{
										Prompt: "Pay {4}?",
										Payer:  opt.Val(game.ObjectControllerReference(game.TargetStackObjectReference(0))),
										ManaCost: opt.Val(cost.Mana{
											cost.O(4),
										}),
									},
								},
								PublishResult: game.ResultKey("unless-paid"),
							},
							{
								Primitive: game.CounterObject{
									Object: game.TargetStackObjectReference(0),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "unless-paid",
									Succeeded: game.TriFalse,
								}),
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 1,
			}),
			OracleText: `
			Choose one —
			• Look at the top four cards of your library. Put one of them into your hand and the rest into your graveyard.
			• Counter target spell unless its controller pays {4}.
		`,
		},
	}
}
