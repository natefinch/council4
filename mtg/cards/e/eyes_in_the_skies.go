package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EyesInTheSkies is the card definition for Eyes in the Skies.
//
// Type: Instant
// Cost: {3}{W}
//
// Oracle text:
//
//	Create a 1/1 white Bird creature token with flying, then populate. (Create a token that's a copy of a creature token you control.)
var EyesInTheSkies = newEyesInTheSkies

func newEyesInTheSkies() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Eyes in the Skies",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenDef(eyesInTheSkiesToken),
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
			Create a 1/1 white Bird creature token with flying, then populate. (Create a token that's a copy of a creature token you control.)
		`,
		},
	}
}

var eyesInTheSkiesToken = newEyesInTheSkiesToken()

func newEyesInTheSkiesToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Bird",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bird},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
