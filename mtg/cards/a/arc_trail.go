package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ArcTrail is the card definition for Arc Trail.
//
// Type: Sorcery
// Cost: {1}{R}
//
// Oracle text:
//
//	Arc Trail deals 2 damage to any target and 1 damage to any other target.
var ArcTrail = newArcTrail

func newArcTrail() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Arc Trail",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "any target",
						Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker, types.Battle}}),
					},
					game.TargetSpec{
						MinTargets:               1,
						MaxTargets:               1,
						Constraint:               "any other target",
						Allow:                    game.TargetAllowPermanent | game.TargetAllowPlayer,
						Selection:                opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker, types.Battle}}),
						DistinctFromPriorTargets: true,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(2),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
					},
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(1),
							Recipient: game.AnyTargetDamageRecipient(1),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Arc Trail deals 2 damage to any target and 1 damage to any other target.
		`,
		},
	}
}
