// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RTCounterModal is the card definition for RT Counter Modal.
//
// Type: Instant
//
// Oracle text:
//
//	Choose one —
//	• Counter target spell. If that spell is countered this way, exile it instead of putting it into its owner's graveyard. If that spell is countered this way, you gain 2 life.
//	• Draw a card.
var RTCounterModal = newRTCounterModal

func newRTCounterModal() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "RT Counter Modal",
			Types: []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Counter target spell. If that spell is countered this way, exile it instead of putting it into its owner's graveyard. If that spell is countered this way, you gain 2 life.",
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
									Object:       game.TargetStackObjectReference(0),
									ExileInstead: true,
								},
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
						},
					},
					game.Mode{
						Text: "Draw a card.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 1,
			}),
			OracleText: `
			Choose one —
			• Counter target spell. If that spell is countered this way, exile it instead of putting it into its owner's graveyard. If that spell is countered this way, you gain 2 life.
			• Draw a card.
		`,
		},
	}
}
