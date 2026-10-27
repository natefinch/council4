// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RTCounterOutcome is the card definition for RT Counter Outcome.
//
// Type: Instant
//
// Oracle text:
//
//	Counter target spell unless its controller pays {3}. If that spell is countered this way, exile it instead of putting it into its owner's graveyard. If that spell is countered this way, you gain 2 life. You gain 1 life.
var RTCounterOutcome = newRTCounterOutcome

func newRTCounterOutcome() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "RT Counter Outcome",
			Types: []types.Card{types.Instant},
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
						Primitive: game.Pay{
							Payment: game.ResolutionPayment{
								Prompt: "Pay {3}?",
								Payer:  opt.Val(game.ObjectControllerReference(game.TargetStackObjectReference(0))),
								ManaCost: opt.Val(cost.Mana{
									cost.O(3),
								}),
							},
						},
						PublishResult: game.ResultKey("unless-paid"),
					},
					{
						Primitive: game.CounterObject{
							Object:       game.TargetStackObjectReference(0),
							ExileInstead: true,
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:       "unless-paid",
							Succeeded: game.TriFalse,
						}),
						PublishResult: game.ResultKey("if-you-do"),
						LocalProducts: game.LocalProducts{
							Results: []game.ResultKey{game.ResultKey("if-you-do")},
						},
					},
					{
						Primitive: game.GainLife{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:       "if-you-do",
							Succeeded: game.TriTrue,
						}),
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
			Counter target spell unless its controller pays {3}. If that spell is countered this way, exile it instead of putting it into its owner's graveyard. If that spell is countered this way, you gain 2 life. You gain 1 life.
		`,
		},
	}
}
