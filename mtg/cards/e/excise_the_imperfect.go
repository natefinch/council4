package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ExciseTheImperfect is the card definition for Excise the Imperfect.
//
// Type: Instant
// Cost: {1}{W}{W}
//
// Oracle text:
//
//	Exile target nonland permanent. Its controller incubates X, where X is its mana value. (They create an Incubator token with X +1/+1 counters on it and "{2}: Transform this token." It transforms into a 0/0 Phyrexian artifact creature.)
var ExciseTheImperfect = newExciseTheImperfect

func newExciseTheImperfect() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Excise the Imperfect",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target nonland permanent",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{ExcludedTypes: []types.Card{types.Land}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.MovePermanent{
							Object:        game.TargetPermanentReference(0),
							Destination:   zone.Exile,
							PublishLinked: game.LinkedKey("life-rider-1"),
						},
					},
					{
						Primitive: game.Incubate{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountObjectManaValue,
								Multiplier: 1,
								Object:     game.LinkedObjectReference("life-rider-1"),
							}),
							Recipient: opt.Val(game.ObjectControllerReference(game.TargetPermanentReference(0))),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Exile target nonland permanent. Its controller incubates X, where X is its mana value. (They create an Incubator token with X +1/+1 counters on it and "{2}: Transform this token." It transforms into a 0/0 Phyrexian artifact creature.)
		`,
		},
	}
}
