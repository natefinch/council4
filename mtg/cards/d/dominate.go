package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Dominate is the card definition for Dominate.
//
// Type: Instant
// Cost: {X}{1}{U}{U}
//
// Oracle text:
//
//	Gain control of target creature with mana value X or less.
var Dominate = newDominate

func newDominate() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Dominate",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.O(1),
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets:       1,
						MaxTargets:       1,
						Constraint:       "target creature with mana value X or less",
						Allow:            game.TargetAllowPermanent,
						Selection:        opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
						ManaValueAtMostX: true,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyContinuous{
							Object: opt.Val(game.TargetPermanentReference(0)),
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer:         game.LayerControl,
									NewController: opt.Val(game.Player1),
								},
							},
							Duration: game.DurationPermanent,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Gain control of target creature with mana value X or less.
		`,
		},
	}
}
