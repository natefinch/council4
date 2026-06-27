package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MimingSlime is the card definition for Miming Slime.
//
// Type: Sorcery
// Cost: {2}{G}
//
// Oracle text:
//
//	Create an X/X green Ooze creature token, where X is the greatest power among creatures you control.
var MimingSlime = newMimingSlime()

func newMimingSlime() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Miming Slime",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenDef(mimingSlimeToken),
							Power: opt.Val(game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountGreatestPowerInGroup,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							})),
							Toughness: opt.Val(game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountGreatestPowerInGroup,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							})),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create an X/X green Ooze creature token, where X is the greatest power among creatures you control.
		`,
		},
	}
}

var mimingSlimeToken = newMimingSlimeToken()

func newMimingSlimeToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Ooze",
			Colors:   []color.Color{color.Green},
			Types:    []types.Card{types.Creature},
			Subtypes: []types.Sub{types.Ooze},
		},
	}
}
