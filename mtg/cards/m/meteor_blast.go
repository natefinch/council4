package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MeteorBlast is the card definition for Meteor Blast.
//
// Type: Sorcery
// Cost: {X}{R}{R}{R}
//
// Oracle text:
//
//	Meteor Blast deals 4 damage to each of X targets.
var MeteorBlast = newMeteorBlast

func newMeteorBlast() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Meteor Blast",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.R,
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets:   0,
						MaxTargets:   20,
						Constraint:   "X targets",
						Allow:        game.TargetAllowPermanent | game.TargetAllowPlayer,
						Selection:    opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker, types.Battle}}),
						CountEqualsX: true,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:     game.Fixed(4),
							Recipient:  game.AnyTargetDamageRecipient(0),
							EachTarget: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Meteor Blast deals 4 damage to each of X targets.
		`,
		},
	}
}
