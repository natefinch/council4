package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Vivify is the card definition for Vivify.
//
// Type: Instant
// Cost: {2}{G}
//
// Oracle text:
//
//	Target land becomes a 3/3 creature until end of turn. It's still a land.
//	Draw a card.
var Vivify = newVivify()

func newVivify() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Vivify",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
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
					{
						Primitive: game.Draw{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target land becomes a 3/3 creature until end of turn. It's still a land.
			Draw a card.
		`,
		},
	}
}
