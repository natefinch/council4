package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AnimateLand is the card definition for Animate Land.
//
// Type: Instant
// Cost: {G}
//
// Oracle text:
//
//	Until end of turn, target land becomes a 3/3 creature that's still a land.
var AnimateLand = newAnimateLand()

func newAnimateLand() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Animate Land",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target land",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Land}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyContinuous{
							Object: opt.Val(game.TargetPermanentReference(0)),
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer:    game.LayerType,
									AddTypes: []types.Card{types.Creature},
								},
								game.ContinuousEffect{
									Layer:        game.LayerPowerToughnessSet,
									SetPower:     opt.Val(game.PT{Value: 3}),
									SetToughness: opt.Val(game.PT{Value: 3}),
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Until end of turn, target land becomes a 3/3 creature that's still a land.
		`,
		},
	}
}
