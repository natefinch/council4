package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PerfectedTheory is the card definition for Perfected Theory.
//
// Type: Instant
// Cost: {U}
//
// Oracle text:
//
//	Choose one —
//	• Target creature has base power and toughness 1/1 until end of turn.
//	• Target creature has base power and toughness 4/5 until end of turn.
var PerfectedTheory = newPerfectedTheory

func newPerfectedTheory() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Perfected Theory",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Target creature has base power and toughness 1/1 until end of turn.",
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
											Duration: game.DurationPermanent,
											Layer:    game.LayerPowerToughnessSet,
											SetPower: opt.Val(game.PT{
												Value: 1,
											}),
											SetToughness: opt.Val(game.PT{
												Value: 1,
											}),
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					},
					game.Mode{
						Text: "Target creature has base power and toughness 4/5 until end of turn.",
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
											Duration: game.DurationPermanent,
											Layer:    game.LayerPowerToughnessSet,
											SetPower: opt.Val(game.PT{
												Value: 4,
											}),
											SetToughness: opt.Val(game.PT{
												Value: 5,
											}),
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
			• Target creature has base power and toughness 1/1 until end of turn.
			• Target creature has base power and toughness 4/5 until end of turn.
		`,
		},
	}
}
