package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CoursersAccord is the card definition for Coursers' Accord.
//
// Type: Sorcery
// Cost: {4}{G}{W}
//
// Oracle text:
//
//	Create a 3/3 green Centaur creature token, then populate. (Create a token that's a copy of a creature token you control.)
var CoursersAccord = newCoursersAccord

func newCoursersAccord() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Coursers' Accord",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.W,
			}),
			Colors: []color.Color{color.Green, color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenDef(coursersAccordToken),
						},
					},
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenCopyOf(game.TokenCopySpec{
								Source: game.TokenCopySourceChosenControlledCreatureToken,
							}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create a 3/3 green Centaur creature token, then populate. (Create a token that's a copy of a creature token you control.)
		`,
		},
	}
}

var coursersAccordToken = newCoursersAccordToken()

func newCoursersAccordToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Centaur",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Centaur},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
		},
	}
}
