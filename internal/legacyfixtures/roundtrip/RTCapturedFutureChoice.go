// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RTCapturedFutureChoice is the card definition for RT Captured Future Choice.
//
// Type: Sorcery
//
// Oracle text:
//
//	Tap target creature. At the beginning of the next end step, you may exile it.
var RTCapturedFutureChoice = newRTCapturedFutureChoice

func newRTCapturedFutureChoice() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "RT Captured Future Choice",
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
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Tap{
							Object: game.TargetPermanentReference(0),
						},
					},
					{
						Primitive: game.CreateDelayedTrigger{
							Trigger: game.DelayedTriggerDef{
								Timing:   game.DelayedAtBeginningOfNextEndStep,
								Optional: true,
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
								CapturedObject: opt.Val(game.TargetPermanentReference(0)),
							},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Tap target creature. At the beginning of the next end step, you may exile it.
		`,
		},
	}
}
