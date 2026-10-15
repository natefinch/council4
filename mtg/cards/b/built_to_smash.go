package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BuiltToSmash is the card definition for Built to Smash.
//
// Type: Instant
// Cost: {R}
//
// Oracle text:
//
//	Target attacking creature gets +3/+3 until end of turn. If it's an artifact creature, it gains trample until end of turn.
var BuiltToSmash = newBuiltToSmash

func newBuiltToSmash() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Built to Smash",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target attacking creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, CombatState: game.CombatStateAttacking}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ModifyPT{
							Object:         game.TargetPermanentReference(0),
							PowerDelta:     game.Fixed(3),
							ToughnessDelta: game.Fixed(3),
							Duration:       game.DurationUntilEndOfTurn,
							PublishLinked:  game.LinkedKey("gain-keyword-1"),
						},
					},
					{
						Primitive: game.ApplyContinuous{
							Object: opt.Val(game.LinkedObjectReference("gain-keyword-1")),
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Duration:    game.DurationPermanent,
									Layer:       game.LayerAbility,
									AddKeywords: []game.Keyword{game.Trample},
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Object:        opt.Val(game.TargetPermanentReference(0)),
								ObjectMatches: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Artifact, types.Creature}}),
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Target attacking creature gets +3/+3 until end of turn. If it's an artifact creature, it gains trample until end of turn.
		`,
		},
	}
}
