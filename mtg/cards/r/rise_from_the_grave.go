package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RiseFromTheGrave is the card definition for Rise from the Grave.
//
// Type: Sorcery
// Cost: {4}{B}
//
// Oracle text:
//
//	Put target creature card from a graveyard onto the battlefield under your control. That creature is a black Zombie in addition to its other colors and types.
var RiseFromTheGrave = newRiseFromTheGrave()

func newRiseFromTheGrave() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Rise from the Grave",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature card from a graveyard",
						Allow:      game.TargetAllowCard,
						TargetZone: zone.Graveyard,
						Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.PutOnBattlefield{
							Source:        game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
							Recipient:     opt.Val(game.ControllerReference()),
							PublishLinked: game.LinkedKey("leave-bf-exile-1"),
						},
					},
					{
						Primitive: game.ApplyContinuous{
							Object: opt.Val(game.LinkedObjectReference("leave-bf-exile-1")),
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer:     game.LayerColor,
									AddColors: []color.Color{color.Black},
								},
								game.ContinuousEffect{
									Layer:       game.LayerType,
									AddSubtypes: []types.Sub{types.Sub("Zombie")},
								},
							},
							Duration: game.DurationPermanent,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Put target creature card from a graveyard onto the battlefield under your control. That creature is a black Zombie in addition to its other colors and types.
		`,
		},
	}
}
