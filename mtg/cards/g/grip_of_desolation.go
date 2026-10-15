package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GripOfDesolation is the card definition for Grip of Desolation.
//
// Type: Instant
// Cost: {4}{B}{B}
//
// Oracle text:
//
//	Devoid (This card has no color.)
//	Exile target creature and target land.
var GripOfDesolation = newGripOfDesolation

func newGripOfDesolation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Grip of Desolation",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
				cost.B,
			}),
			Types: []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.DevoidStaticBody,
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
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
						Primitive: game.MovePermanent{
							Object:      game.TargetPermanentReference(0),
							Destination: zone.Exile,
						},
					},
					{
						Primitive: game.MovePermanent{
							Object:      game.TargetPermanentReference(1),
							Destination: zone.Exile,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Devoid (This card has no color.)
			Exile target creature and target land.
		`,
		},
	}
}
