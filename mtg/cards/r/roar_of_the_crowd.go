package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RoarOfTheCrowd is the card definition for Roar of the Crowd.
//
// Type: Sorcery
// Cost: {3}{R}
//
// Oracle text:
//
//	Choose a creature type. Roar of the Crowd deals damage to any target equal to the number of permanents you control of the chosen type.
var RoarOfTheCrowd = newRoarOfTheCrowd

func newRoarOfTheCrowd() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Roar of the Crowd",
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
						Constraint: "any target",
						Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker, types.Battle}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Choose{
							Choice: game.ResolutionChoice{
								Kind:          game.ResolutionChoiceSubtype,
								Prompt:        "Choose a creature type",
								SubtypeOfType: types.Creature,
							},
							PublishChoice: game.SpellChosenTypeChoiceKey,
						},
					},
					{
						Primitive: game.Damage{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountCountSelector,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{SubtypeChoice: game.SubtypeChoiceSourceEntry, Controller: game.ControllerYou}),
							}),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Choose a creature type. Roar of the Crowd deals damage to any target equal to the number of permanents you control of the chosen type.
		`,
		},
	}
}
