package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MasterSRebuke is the card definition for Master's Rebuke.
//
// Type: Instant
// Cost: {1}{G}
//
// Oracle text:
//
//	Target creature you control deals damage equal to its power to target creature or planeswalker you don't control.
var MasterSRebuke = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name: "Master's Rebuke",
		ManaCost: opt.Val(cost.Mana{
			cost.O(1),
			cost.G,
		}),
		Colors: []color.Color{color.Green},
		Types:  []types.Card{types.Instant},
		OracleText: `
			Target creature you control deals damage equal to its power to target creature or planeswalker you don't control.
		`,
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{
				{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature you control",
					Allow:      game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{
						RequiredTypesAny: []types.Card{
							types.Creature,
						},
						Controller: game.ControllerYou,
					}),
				},
				{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature or planeswalker you don't control",
					Allow:      game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{
						RequiredTypesAny: []types.Card{
							types.Creature,
							types.Planeswalker,
						},
						Controller: game.ControllerOpponent,
					}),
				},
			},
			Sequence: []game.Instruction{
				{
					Primitive: game.Damage{
						Amount: game.Dynamic(game.DynamicAmount{
							Kind:   game.DynamicAmountTargetPower,
							Object: game.TargetPermanentReference(0),
						}),
						Recipient: game.ObjectDamageRecipient(game.TargetPermanentReference(1)),
					},
				},
			},
		}.Ability()),
	},
}
