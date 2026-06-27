package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Hydroform is the card definition for Hydroform.
//
// Type: Instant
// Cost: {G}{U}
//
// Oracle text:
//
//	Target land becomes a 3/3 Elemental creature with flying until end of turn. It's still a land.
var Hydroform = newHydroform()

func newHydroform() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Hydroform",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
				cost.U,
			}),
			Colors: []color.Color{color.Green, color.Blue},
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
									Layer:       game.LayerType,
									AddTypes:    []types.Card{types.Creature},
									AddSubtypes: []types.Sub{types.Elemental},
								},
								game.ContinuousEffect{
									Layer: game.LayerAbility,
									AddKeywords: []game.Keyword{
										game.Flying,
									},
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
			Target land becomes a 3/3 Elemental creature with flying until end of turn. It's still a land.
		`,
		},
	}
}
