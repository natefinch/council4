package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TurnToFrog is the card definition for Turn to Frog.
//
// Type: Instant
// Cost: {1}{U}
//
// Oracle text:
//
//	Until end of turn, target creature loses all abilities and becomes a blue Frog with base power and toughness 1/1.
var TurnToFrog = newTurnToFrog

func newTurnToFrog() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Turn to Frog",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
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
						Primitive: game.ApplyContinuous{
							Object: opt.Val(game.TargetPermanentReference(0)),
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Duration:           game.DurationPermanent,
									Layer:              game.LayerAbility,
									RemoveAllAbilities: true,
								},
								game.ContinuousEffect{
									Duration:  game.DurationPermanent,
									Layer:     game.LayerColor,
									SetColors: []color.Color{color.Blue},
								},
								game.ContinuousEffect{
									Duration:    game.DurationPermanent,
									Layer:       game.LayerType,
									SetTypes:    []types.Card{types.Creature},
									SetSubtypes: []types.Sub{types.Frog},
								},
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
			}.Ability()),
			OracleText: `
			Until end of turn, target creature loses all abilities and becomes a blue Frog with base power and toughness 1/1.
		`,
		},
	}
}
