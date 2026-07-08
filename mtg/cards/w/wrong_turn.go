package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WrongTurn is the card definition for Wrong Turn.
//
// Type: Instant
// Cost: {2}{U}
//
// Oracle text:
//
//	Target opponent gains control of target creature. (If an attacking or blocking creature changes controllers, it's removed from combat.)
var WrongTurn = newWrongTurn

func newWrongTurn() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Wrong Turn",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
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
							Object: opt.Val(game.TargetPermanentReference(1)),
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer:            game.LayerControl,
									NewControllerRef: opt.Val(game.TargetPlayerReference(0)),
								},
							},
							Duration: game.DurationPermanent,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target opponent gains control of target creature. (If an attacking or blocking creature changes controllers, it's removed from combat.)
		`,
		},
	}
}
