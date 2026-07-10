package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MoltenDuplication is the card definition for Molten Duplication.
//
// Type: Sorcery
// Cost: {1}{R}
//
// Oracle text:
//
//	Create a token that's a copy of target artifact or creature you control, except it's an artifact in addition to its other types. It gains haste until end of turn. Sacrifice it at the beginning of the next end step.
var MoltenDuplication = newMoltenDuplication

func newMoltenDuplication() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Molten Duplication",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target artifact or creature you control",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature}, Controller: game.ControllerYou}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenCopyOf(game.TokenCopySpec{
								Source:   game.TokenCopySourceObject,
								Object:   game.TargetPermanentReference(0),
								AddTypes: []types.Card{types.Artifact},
							}),
						},
					},
					{
						Primitive: game.ApplyContinuous{
							Object: opt.Val(game.TargetPermanentReference(0)),
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer: game.LayerAbility,
									AddKeywords: []game.Keyword{
										game.Haste,
									},
								},
							},
							Duration:      game.DurationUntilEndOfTurn,
							PublishLinked: game.LinkedKey("delayed-sacrifice-2"),
						},
					},
					{
						Primitive: game.CreateDelayedTrigger{
							Trigger: game.DelayedTriggerDef{
								Timing:         game.DelayedAtBeginningOfNextEndStep,
								CapturedObject: opt.Val(game.LinkedObjectReference("delayed-sacrifice-2")),
								Content: game.Mode{
									Sequence: []game.Instruction{
										{
											Primitive: game.Sacrifice{
												Object: game.CapturedObjectReference(),
											},
										},
									},
								}.Ability(),
							},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create a token that's a copy of target artifact or creature you control, except it's an artifact in addition to its other types. It gains haste until end of turn. Sacrifice it at the beginning of the next end step.
		`,
		},
	}
}
