package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EchoingCourage is the card definition for Echoing Courage.
//
// Type: Instant
// Cost: {1}{G}
//
// Oracle text:
//
//	Target creature and all other creatures with the same name as that creature get +2/+2 until end of turn.
var EchoingCourage = newEchoingCourage

func newEchoingCourage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Echoing Courage",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature and all other creatures with the same name as that creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer:          game.LayerPowerToughnessModify,
									Group:          game.SameNamePermanentGroup(game.TargetPermanentReference(0), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
									PowerDelta:     2,
									ToughnessDelta: 2,
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target creature and all other creatures with the same name as that creature get +2/+2 until end of turn.
		`,
		},
	}
}
