package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SeismicShift is the card definition for Seismic Shift.
//
// Type: Sorcery
// Cost: {3}{R}
//
// Oracle text:
//
//	Destroy target land. Up to two target creatures can't block this turn.
var SeismicShift = newSeismicShift

func newSeismicShift() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Seismic Shift",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target land",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Land}}),
					},
					game.TargetSpec{
						MinTargets: 0,
						MaxTargets: 2,
						Constraint: "up to two target creatures",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(0),
						},
					},
					{
						Primitive: game.ApplyRule{
							Object: opt.Val(game.TargetPermanentReference(1)),
							RuleEffects: []game.RuleEffect{
								game.RuleEffect{
									Kind: game.RuleEffectCantBlock,
								},
							},
							Duration: game.DurationThisTurn,
						},
					},
					{
						Primitive: game.ApplyRule{
							Object: opt.Val(game.TargetPermanentReference(2)),
							RuleEffects: []game.RuleEffect{
								game.RuleEffect{
									Kind: game.RuleEffectCantBlock,
								},
							},
							Duration: game.DurationThisTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy target land. Up to two target creatures can't block this turn.
		`,
		},
	}
}
