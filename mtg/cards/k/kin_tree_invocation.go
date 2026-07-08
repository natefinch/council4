package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KinTreeInvocation is the card definition for Kin-Tree Invocation.
//
// Type: Sorcery
// Cost: {B}{G}
//
// Oracle text:
//
//	Create an X/X black and green Spirit Warrior creature token, where X is the greatest toughness among creatures you control.
var KinTreeInvocation = newKinTreeInvocation

func newKinTreeInvocation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Kin-Tree Invocation",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
				cost.G,
			}),
			Colors: []color.Color{color.Black, color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenDef(kinTreeInvocationToken),
							Power: opt.Val(game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountGreatestToughnessInGroup,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							})),
							Toughness: opt.Val(game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountGreatestToughnessInGroup,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							})),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create an X/X black and green Spirit Warrior creature token, where X is the greatest toughness among creatures you control.
		`,
		},
	}
}

var kinTreeInvocationToken = newKinTreeInvocationToken()

func newKinTreeInvocationToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Spirit Warrior",
			Colors:   []color.Color{color.Black, color.Green},
			Types:    []types.Card{types.Creature},
			Subtypes: []types.Sub{types.Spirit, types.Warrior},
		},
	}
}
