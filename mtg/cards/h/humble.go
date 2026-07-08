package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Humble is the card definition for Humble.
//
// Type: Instant
// Cost: {1}{W}
//
// Oracle text:
//
//	Until end of turn, target creature loses all abilities and has base power and toughness 0/1.
var Humble = newHumble

func newHumble() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Humble",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors: []color.Color{color.White},
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
									Layer:        game.LayerPowerToughnessSet,
									SetPower:     opt.Val(game.PT{Value: 0}),
									SetToughness: opt.Val(game.PT{Value: 1}),
								},
								game.ContinuousEffect{
									Layer:              game.LayerAbility,
									RemoveAllAbilities: true,
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Until end of turn, target creature loses all abilities and has base power and toughness 0/1.
		`,
		},
	}
}
