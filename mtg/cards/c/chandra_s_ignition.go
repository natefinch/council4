package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"

	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ChandraSIgnition is the card definition for Chandra's Ignition.
//
// Type: Sorcery
// Cost: {3}{R}{R}
//
// Oracle text:
//
//	Target creature you control deals damage equal to its power to each other creature and each opponent.
var ChandraSIgnition = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Red),
	CardFace: game.CardFace{
		Name: "Chandra's Ignition",
		ManaCost: opt.Val(cost.Mana{
			cost.O(3),
			cost.R,
			cost.R,
		}),
		Colors: []color.Color{color.Red},
		Types:  []types.Card{types.Sorcery},
		OracleText: `
			Target creature you control deals damage equal to its power to each other creature and each opponent.
		`,
		SpellAbility: opt.Val(
			game.Mode{
				Targets: []game.TargetSpec{
					{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "creature you control",
						Allow:      game.TargetAllowPermanent,
						Predicate: game.TargetPredicate{
							PermanentTypes: []types.Card{
								types.Creature,
							},
							Controller: game.ControllerYou,
						},
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:   game.DynamicAmountTargetPower,
								Object: game.TargetPermanentReference(0),
							}),
							Recipient: game.GroupDamageRecipient(game.BattlefieldGroupExcluding(
								game.Selection{RequiredTypes: []types.Card{types.Creature}},
								game.TargetPermanentReference(0),
							)),
							DamageSource: opt.Val(game.TargetPermanentReference(0)),
						},
						Description: "deals damage equal to its power to each other creature",
					},
					{
						Primitive: game.Damage{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:   game.DynamicAmountTargetPower,
								Object: game.TargetPermanentReference(0),
							}),
							Recipient:    game.PlayerGroupDamageRecipient(game.OpponentsReference()),
							DamageSource: opt.Val(game.TargetPermanentReference(0)),
						},
						Description: "deals damage equal to its power to each opponent",
					},
				},
			}.Ability(),
		),
	},
}
