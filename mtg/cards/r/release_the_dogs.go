package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ReleaseTheDogs is the card definition for Release the Dogs.
//
// Type: Sorcery
// Cost: {3}{W}
//
// Oracle text:
//
//	Create four 1/1 white Dog creature tokens.
var ReleaseTheDogs = newReleaseTheDogs

func newReleaseTheDogs() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Release the Dogs",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(4),
							Source: game.TokenDef(releaseTheDogsToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create four 1/1 white Dog creature tokens.
		`,
		},
	}
}

var releaseTheDogsToken = newReleaseTheDogsToken()

func newReleaseTheDogsToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Dog",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dog},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
