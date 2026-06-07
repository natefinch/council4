package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CosmicHunger is the card definition for Cosmic Hunger.
//
// Type: Instant
// Cost: {1}{G}
//
// Oracle text:
//
//	Target creature you control deals damage equal to its power to another target creature, planeswalker, or battle.
var CosmicHunger = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name: "Cosmic Hunger",
		ManaCost: opt.Val(cost.Mana{
			cost.O(1),
			cost.G,
		}),
		Colors: []color.Color{color.Green},
		Types:  []types.Card{types.Instant},
		OracleText: `
			Target creature you control deals damage equal to its power to another target creature, planeswalker, or battle.
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
					{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "another creature, planeswalker, or battle",
						Allow:      game.TargetAllowPermanent,
						Predicate: game.TargetPredicate{
							PermanentTypes: []types.Card{
								types.Creature,
								types.Planeswalker,
								types.Battle,
							},
							Another: true,
						},
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:        game.DynamicAmountTargetPower,
								TargetIndex: 0,
							}),
							Recipient: game.TargetRecipient(1),
							DamageSource: opt.Val(game.ObjectReference{
								Kind:        game.ObjectReferenceTargetPermanent,
								TargetIndex: 0,
							}),
						},
					},
				},
			}.Ability(),
		),
	},
}
