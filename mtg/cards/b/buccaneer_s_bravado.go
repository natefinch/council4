package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BuccaneerSBravado is the card definition for Buccaneer's Bravado.
//
// Type: Instant
// Cost: {1}{R}
//
// Oracle text:
//
//	Choose one —
//	• Target creature gets +1/+1 and gains first strike until end of turn.
//	• Target Pirate gets +1/+1 and gains double strike until end of turn.
var BuccaneerSBravado = newBuccaneerSBravado

func newBuccaneerSBravado() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Buccaneer's Bravado",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Target creature gets +1/+1 and gains first strike until end of turn.",
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
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.TargetPermanentReference(0)),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Duration:       game.DurationPermanent,
											Layer:          game.LayerPowerToughnessModify,
											PowerDelta:     1,
											ToughnessDelta: 1,
										},
										game.ContinuousEffect{
											Duration:    game.DurationPermanent,
											Layer:       game.LayerAbility,
											AddKeywords: []game.Keyword{game.FirstStrike},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					},
					game.Mode{
						Text: "Target Pirate gets +1/+1 and gains double strike until end of turn.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target Pirate",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Sub("Pirate")}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.TargetPermanentReference(0)),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Duration:       game.DurationPermanent,
											Layer:          game.LayerPowerToughnessModify,
											PowerDelta:     1,
											ToughnessDelta: 1,
										},
										game.ContinuousEffect{
											Duration:    game.DurationPermanent,
											Layer:       game.LayerAbility,
											AddKeywords: []game.Keyword{game.DoubleStrike},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
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
			• Target creature gets +1/+1 and gains first strike until end of turn.
			• Target Pirate gets +1/+1 and gains double strike until end of turn.
		`,
		},
	}
}
