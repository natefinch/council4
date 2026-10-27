// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RTCounterActivated is the card definition for RT Counter Activated.
//
// Type: Artifact
//
// Oracle text:
//
//	{1}: Counter target spell. If that spell is countered this way, exile it instead of putting it into its owner's graveyard. If that spell is countered this way, you gain 2 life.
var RTCounterActivated = newRTCounterActivated

func newRTCounterActivated() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "RT Counter Activated",
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}: Counter target spell. If that spell is countered this way, exile it instead of putting it into its owner's graveyard. If that spell is countered this way, you gain 2 life.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
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
									game.Instruction{
										Primitive: game.CounterObject{
											Object:       game.TargetStackObjectReference(0),
											ExileInstead: true,
										},
										PublishResult: game.ResultKey("if-you-do"),
										LocalProducts: game.LocalProducts{
											Results: []game.ResultKey{game.ResultKey("if-you-do")},
										},
									},
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Fixed(2),
											Player: game.ControllerReference(),
										},
										ResultGate: opt.Val(game.InstructionResultGate{
											Key:       game.ResultKey("if-you-do"),
											Succeeded: game.TriTrue,
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
			OracleText: `
			{1}: Counter target spell. If that spell is countered this way, exile it instead of putting it into its owner's graveyard. If that spell is countered this way, you gain 2 life.
		`,
		},
	}
}
