package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ForbiddenFriendship is the card definition for Forbidden Friendship.
//
// Type: Sorcery
// Cost: {1}{R}
//
// Oracle text:
//
//	Create a 1/1 red Dinosaur creature token with haste and a 1/1 white Human Soldier creature token.
var ForbiddenFriendship = newForbiddenFriendship

func newForbiddenFriendship() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Forbidden Friendship",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenDef(forbiddenFriendshipToken),
						},
					},
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenDef(forbiddenFriendshipToken2),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create a 1/1 red Dinosaur creature token with haste and a 1/1 white Human Soldier creature token.
		`,
		},
	}
}

var forbiddenFriendshipToken = newForbiddenFriendshipToken()

func newForbiddenFriendshipToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Dinosaur",
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dinosaur},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
			},
		},
	}
}

var forbiddenFriendshipToken2 = newForbiddenFriendshipToken2()

func newForbiddenFriendshipToken2() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Human Soldier",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Soldier},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
