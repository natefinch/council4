package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RishkarSExpertise is the card definition for Rishkar's Expertise.
//
// Type: Sorcery
// Cost: {4}{G}{G}
//
// Oracle text:
//
//	Draw cards equal to the greatest power among creatures you control.
//	You may cast a spell with mana value 5 or less from your hand without paying its mana cost.
var RishkarSExpertise = newRishkarSExpertise

func newRishkarSExpertise() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Rishkar's Expertise",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountGreatestPowerInGroup,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							}),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.CastForFree{
							Player: game.ControllerReference(),
							Selection: game.Selection{ExcludedTypes: []types.Card{types.Land}, ManaValue: opt.Val(compare.Int{
								Op:    compare.LessOrEqual,
								Value: 5,
							})},
							Zone: zone.Hand,
						},
						Optional: true,
					},
				},
			}.Ability()),
			OracleText: `
			Draw cards equal to the greatest power among creatures you control.
			You may cast a spell with mana value 5 or less from your hand without paying its mana cost.
		`,
		},
	}
}
