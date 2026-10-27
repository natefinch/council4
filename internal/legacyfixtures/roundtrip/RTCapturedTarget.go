// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RTCapturedTarget is the card definition for RT Captured Target.
//
// Type: Sorcery
//
// Oracle text:
//
//	Tap target creature. Untap target creature. You gain 1 life. Exile it at the beginning of the next end step.
var RTCapturedTarget = newRTCapturedTarget

func newRTCapturedTarget() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "RT Captured Target",
			Types: []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Tap{
							Object: game.TargetPermanentReference(0),
						},
					},
					{
						Primitive: game.Untap{
							Object: game.TargetPermanentReference(1),
						},
					},
					{
						Primitive: game.GainLife{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.CreateDelayedTrigger{
							Trigger: game.DelayedTriggerDef{
								Timing: game.DelayedAtBeginningOfNextEndStep,
								Content: game.AbilityContent{
									Modes: []game.Mode{
										game.Mode{
											Sequence: []game.Instruction{
												game.Instruction{
													Primitive: game.MovePermanent{
														Object:      game.CapturedObjectReference(),
														Destination: zone.Exile,
													},
												},
											},
										},
									},
									MinModes: 1,
									MaxModes: 1,
								},
								CapturedObject: opt.Val(game.TargetPermanentReference(1)),
							},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Tap target creature. Untap target creature. You gain 1 life. Exile it at the beginning of the next end step.
		`,
		},
	}
}
