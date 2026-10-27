// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RTCapturedCard is the card definition for RT Captured Card.
//
// Type: Sorcery
//
// Oracle text:
//
//	Exile target creature card from your graveyard. You gain 1 life. Return that card to the battlefield at the beginning of the next end step.
var RTCapturedCard = newRTCapturedCard

func newRTCapturedCard() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "RT Captured Card",
			Types: []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature card from your graveyard",
						Allow:      game.TargetAllowCard,
						TargetZone: zone.Graveyard,
						Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.MoveCard{
							Card: game.CardReference{
								Kind: game.CardReferenceTarget,
							},
							FromZone:               zone.Graveyard,
							Destination:            zone.Exile,
							PublishLinked:          game.LinkedKey("sequence-effect-0-product"),
							ReplacePublishedLinked: true,
						},
						LocalProducts: game.LocalProducts{
							Links: []game.LinkedKey{game.LinkedKey("sequence-effect-0-product")},
						},
						ClearLinkedBeforeGate: true,
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
													Primitive: game.PutOnBattlefield{
														Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceCaptured}),
													},
												},
											},
										},
									},
									MinModes: 1,
									MaxModes: 1,
								},
								CapturedCard: opt.Val(game.LinkedObjectReference("sequence-effect-0-product")),
							},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Exile target creature card from your graveyard. You gain 1 life. Return that card to the battlefield at the beginning of the next end step.
		`,
		},
	}
}
