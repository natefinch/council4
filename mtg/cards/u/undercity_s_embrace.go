package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// UndercitySEmbrace is the card definition for Undercity's Embrace.
//
// Type: Instant
// Cost: {2}{B}
//
// Oracle text:
//
//	Target opponent sacrifices a creature of their choice. If you control a creature with power 4 or greater, you gain 4 life.
var UndercitySEmbrace = newUndercitySEmbrace()

func newUndercitySEmbrace() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Undercity's Embrace",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "Target opponent",
						Allow:      game.TargetAllowPlayer,
						Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.SacrificePermanents{
							Amount:    game.Fixed(1),
							Player:    game.TargetPlayerReference(0),
							Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					{
						Primitive: game.GainLife{
							Amount: game.Fixed(4),
							Player: game.ControllerReference(),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								ControlsMatching: opt.Val(game.SelectionCount{
									Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Power: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4})},
								}),
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Target opponent sacrifices a creature of their choice. If you control a creature with power 4 or greater, you gain 4 life.
		`,
		},
	}
}
